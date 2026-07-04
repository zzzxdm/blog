<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";

import {
  createSubmission,
  getCategories,
  getMySubmissions,
  getSiteSettings,
  getTags,
  updateSubmission,
  type Category,
  type SiteSettings,
  type Submission,
  type SubmissionPayload,
  type Tag
} from "../shared/api";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const route = useRoute();

const current = ref<Submission | null>(null);
const loadingSubmission = ref(false);
const saving = ref(false);
const message = ref("");
const error = ref("");
const categoryOptions = ref<Category[]>([]);
const tagOptions = ref<Tag[]>([]);
const editorArea = ref<HTMLTextAreaElement | null>(null);
const siteSettings = ref<SiteSettings | null>(null);
const linkUrl = ref("https://");

const title = ref("");
const summary = ref("");
const category = ref("工程实践");
const tagsText = ref("");
const slug = ref("");
const coverImage = ref("https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=700&q=80");
const content = ref("");

const tags = computed(() => tagsText.value.split(/[,，]/).map((item) => item.trim()).filter(Boolean));
const previewLines = computed(() => content.value.split(/\n+/).map((item) => item.trim()).filter(Boolean));
const status = computed(() => current.value?.status || "draft");
const submissionsEnabled = computed(() => siteSettings.value?.submissionsEnabled ?? true);
const submissionGuide = computed(() => siteSettings.value?.submissionGuide || "登录用户可以提交文章草稿，审核通过后会发布到站点。");
const canEdit = computed(() => submissionsEnabled.value && (!current.value || current.value.status === "draft" || current.value.status === "returned"));

onMounted(() => {
  void loadSiteSettings();
  void loadTaxonomies();
  void loadSubmissionFromQuery();
});

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

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

async function loadSubmissionFromQuery() {
  const id = String(route.query.id || "");
  if (!id || !auth.user) {
    return;
  }

  loadingSubmission.value = true;
  error.value = "";

  try {
    const response = await getMySubmissions();
    const item = response.items.find((submission) => submission.id === id);
    if (!item) {
      error.value = "投稿不存在或无权访问。";
      return;
    }

    applySubmission(item);
    if (!canEdit.value) {
      message.value = "当前投稿已进入审核流程，暂不能修改。";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿加载失败";
  } finally {
    loadingSubmission.value = false;
  }
}

function applySubmission(submission: Submission) {
  current.value = submission;
  title.value = submission.title;
  summary.value = submission.summary;
  category.value = submission.category;
  tagsText.value = submission.tags.join(", ");
  slug.value = submission.slug;
  coverImage.value = submission.coverImage;
  content.value = submission.content;
}

function payload(submit = false): SubmissionPayload {
  return {
    title: title.value,
    summary: summary.value,
    content: content.value,
    category: category.value,
    tags: tags.value,
    coverImage: coverImage.value,
    slug: slug.value,
    submit
  };
}

async function saveDraft() {
  await persist(false);
}

async function submitForReview() {
  await persist(true);
}

async function persist(submit: boolean) {
  if (!auth.user) {
    error.value = "请先登录后再投稿";
    return;
  }
  if (!canEdit.value) {
    error.value = "当前投稿状态不能修改";
    return;
  }
  if (!title.value.trim()) {
    error.value = "请输入标题后再保存";
    message.value = "";
    return;
  }
  if (submit && !content.value.trim()) {
    error.value = "请填写正文后再提交审核";
    message.value = "";
    return;
  }

  saving.value = true;
  message.value = "";
  error.value = "";

  try {
    current.value = current.value
      ? await updateSubmission(current.value.id, payload(submit))
      : await createSubmission(payload(submit));
    message.value = submit ? "已提交审核，审核结果会通过站内信通知你。" : "草稿已保存。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿保存失败";
  } finally {
    saving.value = false;
  }
}

function applyMarkdown(type: "bold" | "italic" | "heading" | "quote" | "code" | "link") {
  if (!canEdit.value) {
    return;
  }

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
    const url = linkUrl.value.trim();
    if (!url || url === "https://") {
      error.value = "请输入链接地址。";
      message.value = "";
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

function statusText(value: string) {
  if (value === "submitted") {
    return "待审核";
  }
  if (value === "returned") {
    return "退回修改";
  }
  if (value === "rejected") {
    return "已拒绝";
  }
  if (value === "published") {
    return "已发布";
  }
  return "草稿";
}

function statusClass(value: string) {
  if (value === "submitted") {
    return "review";
  }
  if (value === "returned" || value === "rejected") {
    return "rejected";
  }
  if (value === "published") {
    return "published";
  }
  return "draft";
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>投稿</h1>
        <p>{{ submissionGuide }}</p>
      </div>
      <div class="meta-row">
        <span>{{ loadingSubmission ? "正在加载投稿" : (current ? `版本 ${current.version}` : "新草稿") }}</span>
        <span class="status" :class="statusClass(status)">{{ statusText(status) }}</span>
      </div>
    </section>

    <section v-if="!auth.user && !auth.loading" class="panel">
      <div class="panel-title"><h2>需要登录</h2></div>
      <p class="muted">登录后可以保存草稿、提交审核，并在站内信里接收审核结果。</p>
      <RouterLink class="button" to="/login">去登录</RouterLink>
    </section>

    <section v-if="!submissionsEnabled" class="panel">
      <div class="panel-title"><h2>投稿暂未开放</h2></div>
      <p class="muted">管理员已关闭用户投稿入口，已有投稿可在个人中心查看审核结果。</p>
      <RouterLink class="button-secondary" to="/account/submissions">查看我的投稿</RouterLink>
    </section>

    <section v-else class="editor-layout">
      <div class="editor-panel">
        <div class="editor-toolbar">
          <div class="tool-group" aria-label="投稿编辑工具栏">
            <button class="tool" type="button" aria-label="加粗" :disabled="!canEdit" @click="applyMarkdown('bold')">B</button>
            <button class="tool" type="button" aria-label="斜体" :disabled="!canEdit" @click="applyMarkdown('italic')">I</button>
            <button class="tool" type="button" aria-label="标题" :disabled="!canEdit" @click="applyMarkdown('heading')">H</button>
            <button class="tool" type="button" aria-label="引用" :disabled="!canEdit" @click="applyMarkdown('quote')">"</button>
            <button class="tool" type="button" aria-label="代码" :disabled="!canEdit" @click="applyMarkdown('code')">{ }</button>
            <button class="tool" type="button" aria-label="链接" :disabled="!canEdit" @click="applyMarkdown('link')">↗</button>
          </div>
          <label class="editor-link-input"><span class="sr-only">链接地址</span><input v-model="linkUrl" class="input" type="url" placeholder="https://example.com" aria-label="链接地址" :disabled="!canEdit"></label>
        </div>

        <div class="editor-grid">
          <textarea ref="editorArea" v-model="content" class="markdown-area" aria-label="投稿 Markdown 编辑区" :disabled="!canEdit"></textarea>

          <article class="preview-area">
            <h1>{{ title || "未命名投稿" }}</h1>
            <p>{{ summary }}</p>
            <template v-for="line in previewLines" :key="line">
              <h2 v-if="line.startsWith('## ')" :id="line.slice(3)">{{ line.slice(3) }}</h2>
              <blockquote v-else-if="line.startsWith('>')">{{ line.replace(/^>\s?/, "") }}</blockquote>
              <ul v-else-if="line.startsWith('- ')">
                <li>{{ line.slice(2) }}</li>
              </ul>
              <p v-else-if="!line.startsWith('# ')">{{ line }}</p>
            </template>
          </article>
        </div>
      </div>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>投稿流程</h2>
          </div>
          <div class="stepper">
            <div class="step done">
              <span class="step-index">1</span>
              <div>
                <strong>填写内容</strong>
                <div class="meta-row"><span>{{ current ? "草稿已保存" : "正在编辑" }}</span></div>
              </div>
            </div>
            <div class="step" :class="{ current: status === 'draft' || status === 'returned', done: status === 'submitted' || status === 'published' }">
              <span class="step-index">2</span>
              <div>
                <strong>提交审核</strong>
                <div class="meta-row"><span>编辑会检查质量、格式和安全</span></div>
              </div>
            </div>
            <div class="step" :class="{ current: status === 'submitted', done: status === 'published' }">
              <span class="step-index">3</span>
              <div>
                <strong>通过后发布</strong>
                <div class="meta-row"><span>发布后进入公开文章列表</span></div>
              </div>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>文章信息</h2>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="title">标题</label>
              <input v-model="title" class="input" id="title" :disabled="!canEdit">
            </div>
            <div class="field">
              <label for="summary">摘要</label>
              <textarea v-model="summary" class="input" id="summary" :disabled="!canEdit"></textarea>
            </div>
            <div class="field">
              <label for="category">建议分类</label>
              <select v-model="category" class="input" id="category" :disabled="!canEdit">
                <option v-for="item in categoryOptions" :key="item.id" :value="item.name">{{ item.name }}</option>
                <option v-if="!categoryOptions.length">工程实践</option>
              </select>
            </div>
            <div class="field">
              <label for="slug">Slug</label>
              <input v-model="slug" class="input" id="slug" :disabled="!canEdit">
            </div>
            <div class="field">
              <label for="tags">标签</label>
              <input v-model="tagsText" class="input" id="tags" list="submit-tag-options" :disabled="!canEdit">
              <datalist id="submit-tag-options">
                <option v-for="item in tagOptions" :key="item.id" :value="item.name"></option>
              </datalist>
            </div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>封面</h2>
          </div>
          <div class="settings-stack">
            <img
              :src="coverImage"
              alt="投稿封面预览"
              style="border-radius: 8px; aspect-ratio: 16 / 9; object-fit: cover;"
            >
            <input v-model="coverImage" class="input" aria-label="封面图片 URL" :disabled="!canEdit">
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>提交</h2>
          </div>
          <div class="settings-stack">
            <div class="review-note">
              <strong>投稿不会直接公开</strong>
              <p>提交后进入待审核状态。编辑可能会通过、退回修改或拒绝投稿。</p>
            </div>
            <p v-if="message" class="muted">{{ message }}</p>
            <p v-if="error" class="error">{{ error }}</p>
            <button class="button-secondary" type="button" :disabled="saving || loadingSubmission || !auth.user || !canEdit" @click="saveDraft">
              {{ saving ? "保存中..." : "保存草稿" }}
            </button>
            <button class="button" type="button" :disabled="saving || loadingSubmission || !auth.user || !canEdit" @click="submitForReview">
              {{ saving ? "提交中..." : "提交审核" }}
            </button>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
