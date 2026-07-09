<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import AdminLayout from "../../components/AdminLayout.vue";
import RichMarkdownEditor from "../../components/RichMarkdownEditor.vue";
import {
  createAdminPost,
  createAdminPostPreview,
  getCategories,
  getAdminPost,
  getAdminPostRevisions,
  getTags,
  publishAdminPost,
  restoreAdminPostRevision,
  uploadAdminMedia,
  updateAdminPost,
  type AdminPost,
  type AdminPostPayload,
  type AdminPostRevision,
  type AdminPostVisibility,
  type AdminPostStatus,
  type Category,
  type Tag
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";

const route = useRoute();
const confirmDialog = useConfirmStore();
const toast = useToastStore();

const current = ref<AdminPost | null>(null);
const loading = ref(false);
const saving = ref(false);
const previewing = ref(false);
const revisionLoading = ref(false);
const restoringId = ref("");
const error = ref("");
const message = ref("");
const categoryOptions = ref<Category[]>([]);
const tagOptions = ref<Tag[]>([]);
const revisions = ref<AdminPostRevision[]>([]);
const coverFileInput = ref<HTMLInputElement | null>(null);

const title = ref("");
const summary = ref("");
const content = ref("");
const slug = ref("");
const category = ref("工程实践");
const tagsText = ref("");
const coverImage = ref("https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80");
const seoTitle = ref("");
const seoDescription = ref("");
const status = ref<AdminPostStatus>("draft");
const visibility = ref<AdminPostVisibility>("public");
const scheduledAt = ref(nextScheduleValue());
const coverUploading = ref(false);

const description = computed(() => current.value ? `自动保存于 ${formatDateTime(current.value.updatedAt)}，当前版本 ${current.value.version}。` : "新文章草稿。");

onMounted(() => {
  void load();
  void loadTaxonomies();
});

async function loadTaxonomies() {
  try {
    const [categoryResult, tagResult] = await Promise.all([getCategories(), getTags()]);
    categoryOptions.value = categoryResult.items;
    tagOptions.value = tagResult.items;
  } catch {
    categoryOptions.value = [];
    tagOptions.value = [];
  }
}

async function load() {
  const id = String(route.query.id || "");
  if (!id) {
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    const post = await getAdminPost(id);
    applyPost(post);
    await loadRevisions(post.id);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadRevisions(id = current.value?.id || "") {
  if (!id) {
    revisions.value = [];
    return;
  }

  revisionLoading.value = true;

  try {
    revisions.value = (await getAdminPostRevisions(id)).items;
  } catch {
    revisions.value = [];
  } finally {
    revisionLoading.value = false;
  }
}

function applyPost(post: AdminPost) {
  current.value = post;
  title.value = post.title;
  summary.value = post.summary;
  content.value = post.content;
  slug.value = post.slug;
  category.value = post.category;
  tagsText.value = post.tags.join(", ");
  coverImage.value = post.coverImage;
  seoTitle.value = post.seoTitle;
  seoDescription.value = post.seoDescription;
  status.value = post.status;
  visibility.value = post.visibility || "public";
  scheduledAt.value = post.scheduledAt ? toDateTimeLocal(post.scheduledAt) : nextScheduleValue();
}

function openCoverPicker() {
  if (coverUploading.value) {
    return;
  }

  coverFileInput.value?.click();
}

async function handleCoverPicked(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0] || null;
  input.value = "";
  if (!file) {
    return;
  }

  await uploadCover(file);
}

async function uploadCover(file: File) {
  if (!file.type.startsWith("image/")) {
    toast.error("封面上传失败", "只能上传图片文件。");
    return;
  }

  coverUploading.value = true;
  error.value = "";
  message.value = "";

  try {
    const asset = await uploadAdminMedia(file, {
      alt: title.value.trim() || defaultAlt(file.name),
      category: "封面图"
    });
    coverImage.value = asset.url;
    toast.success("封面已上传", "已更新封面预览。");
  } catch (err) {
    const uploadError = err instanceof Error ? err.message : "封面上传失败";
    toast.error("封面上传失败", uploadError);
  } finally {
    coverUploading.value = false;
  }
}

function payload(nextStatus: AdminPostStatus): AdminPostPayload {
  return {
    title: title.value,
    summary: summary.value,
    content: content.value,
    slug: slug.value,
    category: category.value,
    tags: tagsText.value.split(/[,，]/).map((item) => item.trim()).filter(Boolean),
    coverImage: coverImage.value,
    seoTitle: seoTitle.value,
    seoDescription: seoDescription.value,
    visibility: visibility.value,
    scheduledAt: scheduledAt.value ? new Date(scheduledAt.value).toISOString() : undefined,
    status: nextStatus
  };
}

async function saveDraft() {
  await save("draft");
}

async function save(nextStatus: AdminPostStatus) {
  if (!title.value.trim()) {
    error.value = "请输入标题后再保存";
    message.value = "";
    return;
  }
  if (nextStatus === "review" && !content.value.trim()) {
    error.value = "请填写正文后再提交审核";
    message.value = "";
    return;
  }

  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    const saved = current.value
      ? await updateAdminPost(current.value.id, payload(nextStatus))
      : await createAdminPost(payload(nextStatus));
    applyPost(saved);
    await loadRevisions(saved.id);
    message.value = "草稿已保存。";
    toast.success("草稿已保存", `当前版本 ${saved.version}。`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

async function publish() {
  if (visibility.value === "members") {
    notifyPublishBlocked("发布");
    return;
  }
  if (!title.value.trim() || !content.value.trim()) {
    error.value = "请填写标题和正文后再发布。";
    message.value = "";
    return;
  }

  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    const saved = current.value
      ? await updateAdminPost(current.value.id, payload("draft"))
      : await createAdminPost(payload("draft"));
    const published = await publishAdminPost(saved.id);
    applyPost(published);
    await loadRevisions(published.id);
    message.value = `已发布到 /posts/${published.publishedPostSlug || published.slug}`;
    toast.success("文章已发布", `/posts/${published.publishedPostSlug || published.slug}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "发布失败";
  } finally {
    saving.value = false;
  }
}

async function schedulePost() {
  if (visibility.value === "members") {
    notifyPublishBlocked("定时发布");
    return;
  }
  if (!scheduledAt.value) {
    error.value = "请选择发布时间。";
    message.value = "";
    return;
  }
  if (!title.value.trim() || !content.value.trim()) {
    error.value = "请填写标题和正文后再定时发布。";
    message.value = "";
    return;
  }

  await save("scheduled");
  if (!error.value) {
    message.value = `已保存为待发布，预约时间 ${formatDate(new Date(scheduledAt.value).toISOString())}`;
    toast.success("已保存为待发布", `预约时间 ${formatDate(new Date(scheduledAt.value).toISOString())}`);
  }
}

function notifyPublishBlocked(action: string) {
  const text = `${visibilityText(visibility.value)}文章暂不支持${action}，请先把可见性改为公开或私密。`;
  error.value = text;
  message.value = "";
  toast.warning("无法发布", text);
}

async function openPreview() {
  if (!title.value.trim()) {
    error.value = "请输入标题后再生成预览";
    message.value = "";
    return;
  }

  previewing.value = true;
  error.value = "";
  message.value = "";

  try {
    const saved = current.value
      ? await updateAdminPost(current.value.id, payload(status.value || "draft"))
      : await createAdminPost(payload("draft"));
    applyPost(saved);
    await loadRevisions(saved.id);

    const preview = await createAdminPostPreview(saved.id);
    window.open(preview.previewUrl, "_blank", "noopener");
    message.value = `预览链接已生成，${formatDate(preview.expiresAt)} 前有效。`;
    toast.success("预览链接已生成", `${formatDate(preview.expiresAt)} 前有效。`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "预览生成失败";
  } finally {
    previewing.value = false;
  }
}

async function restoreRevision(revision: AdminPostRevision) {
  if (!current.value) {
    return;
  }

  const confirmed = await confirmDialog.open({
    title: `恢复到版本 ${revision.version}`,
    message: "当前编辑内容会被该历史版本覆盖，恢复后会生成新的版本记录。",
    confirmText: "恢复版本",
    tone: "success"
  });
  if (!confirmed) {
    return;
  }

  restoringId.value = revision.id;
  error.value = "";
  message.value = "";

  try {
    const restored = await restoreAdminPostRevision(current.value.id, revision.id);
    applyPost(restored);
    await loadRevisions(restored.id);
    message.value = `已恢复到版本 ${revision.version}。`;
    toast.success("版本已恢复", `已恢复到版本 ${revision.version}。`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "版本恢复失败";
  } finally {
    restoringId.value = "";
  }
}

function statusText(value: AdminPostStatus) {
  if (value === "published") return "已发布";
  if (value === "scheduled") return "待发布";
  if (value === "review") return "待审核";
  if (value === "archived") return "已归档";
  return "草稿";
}

function visibilityText(value: AdminPostVisibility) {
  if (value === "private") return "私密";
  if (value === "members") return "会员可见";
  return "公开";
}

function formatDate(value: string) {
  return formatDateTime(value);
}

function nextScheduleValue() {
  const date = new Date();
  date.setHours(date.getHours() + 1, 0, 0, 0);
  return toDateTimeLocal(date.toISOString());
}

function toDateTimeLocal(value: string) {
  const date = new Date(value);
  const pad = (item: number) => String(item).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function defaultAlt(fileName: string) {
  return fileName.replace(/\.[^.]+$/, "");
}
</script>

<template>
  <AdminLayout title="编辑文章" :description="description" mobile-title="写作" primary-action="发布">
    <template #mobile-action>
      <button class="button" type="button" :disabled="saving" title="私密文章可发布为受限访问，会员可见暂不支持发布" @click="publish">
        {{ saving ? "发布中..." : "发布" }}
      </button>
    </template>

    <template #actions>
      <div class="header-actions">
        <RouterLink v-if="current?.publishedPostSlug" class="button-secondary" :to="`/posts/${current.publishedPostSlug}`">查看已发布</RouterLink>
        <button class="button-secondary" type="button" :disabled="previewing || saving" @click="openPreview">{{ previewing ? "生成中..." : "预览" }}</button>
        <button class="button-secondary" type="button" :disabled="saving" @click="saveDraft">{{ saving ? "保存中..." : "保存草稿" }}</button>
        <button class="button" type="button" :disabled="saving" title="私密文章可发布为受限访问，会员可见暂不支持发布" @click="publish">{{ saving ? "发布中..." : "发布" }}</button>
      </div>
    </template>

    <p v-if="loading" class="muted">正在加载文章...</p>
    <p v-if="error" class="error">{{ error }}</p>

    <section class="editor-layout">
      <div class="editor-panel">
        <RichMarkdownEditor
          v-model="content"
          editor-id="admin-post-editor"
          upload-category="文章插图"
          placeholder="撰写文章正文，支持 Markdown、实时预览、表情、粘贴或拖拽上传图片。"
          @save="saveDraft"
          @upload-error="(value) => { error = value; message = ''; }"
        />
      </div>

      <aside class="settings-stack" aria-label="文章设置">
        <section class="panel">
          <div class="panel-title">
            <h2>发布</h2>
            <span class="status draft">{{ statusText(status) }}</span>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="publish-time">发布时间</label>
              <input v-model="scheduledAt" class="input" id="publish-time" type="datetime-local">
            </div>
            <div class="field">
              <label for="visibility">可见性</label>
              <select v-model="visibility" class="input" id="visibility">
                <option value="public">公开</option>
                <option value="private">私密</option>
                <option value="members">会员可见</option>
              </select>
            </div>
            <button class="button" type="button" :disabled="saving" @click="save('review')">提交审核</button>
            <button class="button-secondary" type="button" :disabled="saving" @click="schedulePost">保存定时</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>内容信息</h2>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="title">标题</label><input v-model="title" class="input" id="title"></div>
            <div class="field"><label for="summary">摘要</label><textarea v-model="summary" class="input" id="summary"></textarea></div>
            <div class="field"><label for="slug">Slug</label><input v-model="slug" class="input" id="slug"></div>
            <div class="field">
              <label for="category">分类</label>
              <select v-model="category" class="input" id="category">
                <option v-for="item in categoryOptions" :key="item.id" :value="item.name">{{ item.name }}</option>
                <option v-if="!categoryOptions.length">工程实践</option>
              </select>
            </div>
            <div class="field">
              <label for="tags">标签</label>
              <input v-model="tagsText" class="input" id="tags" list="admin-tag-options">
              <datalist id="admin-tag-options">
                <option v-for="item in tagOptions" :key="item.id" :value="item.name"></option>
              </datalist>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>封面和主题</h2>
          </div>
          <div class="settings-stack">
            <img
              :src="coverImage"
              alt="当前文章封面图"
              style="border-radius: 8px; aspect-ratio: 16 / 9; object-fit: cover;"
            >
            <div class="cover-actions">
              <button class="button-secondary" type="button" :disabled="coverUploading" @click="openCoverPicker">
                {{ coverUploading ? "上传中..." : "上传封面" }}
              </button>
              <input
                ref="coverFileInput"
                class="sr-only"
                type="file"
                accept="image/*"
                :disabled="coverUploading"
                @change="handleCoverPicked"
              >
            </div>
            <input v-model="coverImage" class="input" aria-label="封面图片 URL">
            <div class="field">
              <label>强调色</label>
              <div class="swatches" aria-label="强调色选择">
                <span class="swatch" style="background:#295b4b"></span>
                <span class="swatch" style="background:#b95f2d"></span>
                <span class="swatch" style="background:#e3b45d"></span>
              </div>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>SEO</h2>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="seo-title">SEO 标题</label><input v-model="seoTitle" class="input" id="seo-title"></div>
            <div class="field"><label for="seo-description">SEO 描述</label><textarea v-model="seoDescription" class="input" id="seo-description"></textarea></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>版本历史</h2>
            <span class="tag">{{ revisions.length }} 个版本</span>
          </div>
          <div class="timeline">
            <p v-if="revisionLoading" class="muted">正在加载版本...</p>
            <p v-else-if="!current" class="muted">保存草稿后生成版本记录。</p>
            <p v-else-if="!revisions.length" class="muted">暂无版本记录。</p>
            <article v-for="revision in revisions" :key="revision.id" class="timeline-item">
              <strong>版本 {{ revision.version }} · {{ revision.title }}</strong>
              <p>{{ revision.summary || "无摘要" }}</p>
              <div class="meta-row">
                <span>{{ formatDate(revision.createdAt) }}</span>
                <span>{{ visibilityText(revision.visibility) }}</span>
                <span>{{ revision.authorName }}</span>
                <button class="button-secondary" type="button" :disabled="restoringId === revision.id || revision.version === current?.version" @click="restoreRevision(revision)">
                  {{ restoringId === revision.id ? "恢复中..." : "恢复" }}
                </button>
              </div>
            </article>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
