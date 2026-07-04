<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from "vue";
import { useRoute } from "vue-router";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  createAdminPost,
  getCategories,
  getAdminPost,
  getAdminPostRevisions,
  getTags,
  publishAdminPost,
  restoreAdminPostRevision,
  updateAdminPost,
  type AdminPost,
  type AdminPostPayload,
  type AdminPostRevision,
  type AdminPostVisibility,
  type AdminPostStatus,
  type Category,
  type Tag
} from "../../shared/api";

const route = useRoute();

const current = ref<AdminPost | null>(null);
const loading = ref(false);
const saving = ref(false);
const revisionLoading = ref(false);
const restoringId = ref("");
const error = ref("");
const message = ref("");
const categoryOptions = ref<Category[]>([]);
const tagOptions = ref<Tag[]>([]);
const revisions = ref<AdminPostRevision[]>([]);
const editorArea = ref<HTMLTextAreaElement | null>(null);
const previewArea = ref<HTMLElement | null>(null);

const title = ref("如何设计一个内容长期增长的博客系统");
const summary = ref("博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。");
const content = ref(`# 如何设计一个内容长期增长的博客系统

博客不是文章列表加详情页。真正可持续的系统需要同时照顾写作、发布、搜索、运营、迁移和长期维护。

## 内容模型先于页面

文章需要拥有稳定的 slug、分类、标签、SEO 元数据、封面图、摘要、阅读时长和发布时间。

> 内容系统的核心不是页面，而是可被长期复用、迁移和分发的数据。

## 发布流程需要留出空间

成熟博客通常支持草稿、预览、审核、定时发布和版本历史。`);
const slug = ref("blog-system-design");
const category = ref("工程实践");
const tagsText = ref("Vue3, SEO, 内容系统");
const coverImage = ref("https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=700&q=80");
const seoTitle = ref("如何设计一个现代化博客系统");
const seoDescription = ref("从内容模型、发布流程、SEO、缓存和运营能力设计一个可长期维护的博客系统。");
const status = ref<AdminPostStatus>("draft");
const visibility = ref<AdminPostVisibility>("public");

const previewLines = computed(() => content.value.split(/\n+/).map((item) => item.trim()).filter(Boolean));
const description = computed(() => current.value ? `自动保存于 ${new Date(current.value.updatedAt).toLocaleTimeString("zh-CN")}，当前版本 ${current.value.version}。` : "新文章草稿。");

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
    status: nextStatus
  };
}

async function saveDraft() {
  await save("draft");
}

async function save(nextStatus: AdminPostStatus) {
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
  } catch (err) {
    error.value = err instanceof Error ? err.message : "保存失败";
  } finally {
    saving.value = false;
  }
}

async function publish() {
  if (visibility.value !== "public") {
    error.value = "非公开文章暂不支持发布到公开站点。";
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
  } catch (err) {
    error.value = err instanceof Error ? err.message : "发布失败";
  } finally {
    saving.value = false;
  }
}

async function restoreRevision(revision: AdminPostRevision) {
  if (!current.value || !window.confirm(`恢复到版本 ${revision.version}？`)) {
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
  } catch (err) {
    error.value = err instanceof Error ? err.message : "版本恢复失败";
  } finally {
    restoringId.value = "";
  }
}

function applyMarkdown(type: "bold" | "italic" | "heading" | "quote" | "code" | "link") {
  const textarea = editorArea.value;
  const start = textarea?.selectionStart ?? content.value.length;
  const end = textarea?.selectionEnd ?? content.value.length;
  const selected = content.value.slice(start, end);
  let inner = selected;
  let replacement = "";
  let selectionStart = start;
  let selectionEnd = start;

  if (type === "bold") {
    inner = selected || "加粗文字";
    replacement = `**${inner}**`;
    selectionStart = start + 2;
    selectionEnd = selectionStart + inner.length;
  } else if (type === "italic") {
    inner = selected || "斜体文字";
    replacement = `*${inner}*`;
    selectionStart = start + 1;
    selectionEnd = selectionStart + inner.length;
  } else if (type === "heading") {
    inner = selected || "小标题";
    replacement = `## ${inner}`;
    selectionStart = start + 3;
    selectionEnd = selectionStart + inner.length;
  } else if (type === "quote") {
    inner = selected || "引用内容";
    replacement = inner.split("\n").map((line) => `> ${line}`).join("\n");
    selectionStart = start;
    selectionEnd = start + replacement.length;
  } else if (type === "code") {
    inner = selected || "code";
    if (inner.includes("\n")) {
      replacement = `\`\`\`\n${inner}\n\`\`\``;
      selectionStart = start + 4;
    } else {
      replacement = `\`${inner}\``;
      selectionStart = start + 1;
    }
    selectionEnd = selectionStart + inner.length;
  } else {
    const url = window.prompt("链接地址", "https://");
    if (!url) {
      return;
    }
    inner = selected || "链接文字";
    replacement = `[${inner}](${url})`;
    selectionStart = start + 1;
    selectionEnd = selectionStart + inner.length;
  }

  content.value = `${content.value.slice(0, start)}${replacement}${content.value.slice(end)}`;
  void nextTick(() => {
    editorArea.value?.focus();
    editorArea.value?.setSelectionRange(selectionStart, selectionEnd);
  });
}

function scrollToPreview() {
  previewArea.value?.scrollIntoView({ behavior: "smooth", block: "center" });
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
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}
</script>

<template>
  <AdminLayout title="编辑文章" :description="description" mobile-title="写作" primary-action="发布">
    <template #mobile-action>
      <button class="button" type="button" :disabled="saving || visibility !== 'public'" title="私密和会员可见文章暂不支持发布到公开站点" @click="publish">
        {{ saving ? "发布中..." : "发布" }}
      </button>
    </template>

    <template #actions>
      <div class="header-actions">
        <RouterLink v-if="current?.publishedPostSlug" class="button-secondary" :to="`/posts/${current.publishedPostSlug}`">预览</RouterLink>
        <button v-else class="button-secondary" type="button" @click="scrollToPreview">预览</button>
        <button class="button-secondary" type="button" :disabled="saving" @click="saveDraft">{{ saving ? "保存中..." : "保存草稿" }}</button>
        <button class="button" type="button" :disabled="saving || visibility !== 'public'" title="私密和会员可见文章暂不支持发布到公开站点" @click="publish">{{ saving ? "发布中..." : "发布" }}</button>
      </div>
    </template>

    <p v-if="loading" class="muted">正在加载文章...</p>
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="editor-layout">
      <div class="editor-panel">
          <div class="editor-toolbar">
            <div class="tool-group" aria-label="编辑工具栏">
            <button class="tool" type="button" aria-label="加粗" @click="applyMarkdown('bold')">B</button>
            <button class="tool" type="button" aria-label="斜体" @click="applyMarkdown('italic')">I</button>
            <button class="tool" type="button" aria-label="标题" @click="applyMarkdown('heading')">H</button>
            <button class="tool" type="button" aria-label="引用" @click="applyMarkdown('quote')">"</button>
            <button class="tool" type="button" aria-label="代码" @click="applyMarkdown('code')">{ }</button>
            <button class="tool" type="button" aria-label="链接" @click="applyMarkdown('link')">↗</button>
          </div>
          <div class="meta-row">
            <span>Markdown</span>
            <span>预览</span>
          </div>
        </div>

        <div class="editor-grid">
          <textarea ref="editorArea" v-model="content" class="markdown-area" aria-label="Markdown 编辑区"></textarea>

          <article ref="previewArea" class="preview-area">
            <h1>{{ title }}</h1>
            <p>{{ summary }}</p>
            <template v-for="line in previewLines" :key="line">
              <h2 v-if="line.startsWith('## ')">{{ line.slice(3) }}</h2>
              <blockquote v-else-if="line.startsWith('>')">{{ line.replace(/^>\s?/, "") }}</blockquote>
              <p v-else-if="!line.startsWith('# ')">{{ line }}</p>
            </template>
          </article>
        </div>
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
              <input class="input" id="publish-time" type="datetime-local" value="2026-07-04T20:00">
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
