<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { ElOption, ElSelect } from "element-plus";
import "element-plus/es/components/select/style/css";

import RichMarkdownEditor from "../components/RichMarkdownEditor.vue";
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
  type SubmissionVisibility,
  type Tag
} from "../shared/api";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string;
      reset: (widgetId: string) => void;
      remove: (widgetId: string) => void;
    };
  }
}

const auth = useAuthStore();
const toast = useToastStore();
const route = useRoute();

const current = ref<Submission | null>(null);
const loadingSubmission = ref(false);
const saving = ref(false);
const message = ref("");
const error = ref("");
const categoryOptions = ref<Category[]>([]);
const tagOptions = ref<Tag[]>([]);
const turnstileEl = ref<HTMLElement | null>(null);
const turnstileWidgetId = ref("");
const turnstileToken = ref("");
const turnstileError = ref("");
const siteSettings = ref<SiteSettings | null>(null);

const title = ref("");
const summary = ref("");
const category = ref("工程实践");
const selectedTags = ref<string[]>([]);
const slug = ref("");
const visibility = ref<SubmissionVisibility>("public");
const coverImage = ref("https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=700&q=80");
const content = ref("");

const normalizedTags = computed(() => normalizeTags(selectedTags.value));
const tagSelectOptions = computed(() => {
  const options = new Map<string, { value: string; label: string; meta: string }>();
  tagOptions.value.forEach((item) => {
    options.set(item.name, {
      value: item.name,
      label: item.name,
      meta: `${item.slug} · ${item.postCount} 篇`
    });
  });
  normalizedTags.value.forEach((value) => {
    if (!options.has(value)) {
      options.set(value, { value, label: value, meta: "已选" });
    }
  });
  return [...options.values()];
});
const status = computed(() => current.value?.status || "draft");
const isPrivate = computed(() => visibility.value === "private");
const submissionsEnabled = computed(() => siteSettings.value?.submissionsEnabled ?? true);
const submissionGuide = computed(() => siteSettings.value?.submissionGuide || "登录用户可以提交公开投稿或私密文章，公开投稿审核通过后会发布到站点。");
const submissionLimit = computed(() => siteSettings.value?.submissionLimit || "每天最多 3 篇");
const canEdit = computed(() => submissionsEnabled.value && (!current.value || current.value.status === "draft" || current.value.status === "returned" || (current.value.status === "published" && current.value.visibility === "private")));
const turnstileRequired = computed(() => Boolean(
  auth.user &&
  canEdit.value &&
  siteSettings.value?.turnstileEnabled &&
  siteSettings.value.turnstileSubmission &&
  siteSettings.value.turnstileSiteKey
));

onMounted(() => {
  void loadSiteSettings();
  void loadTaxonomies();
  void loadSubmissionFromQuery();
});

onBeforeUnmount(() => {
  removeTurnstile();
});

watch(turnstileRequired, (required) => {
  if (required) {
    void renderTurnstile();
    return;
  }

  removeTurnstile();
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
    const response = await getMySubmissions({ all: true });
    const item = response.items.find((submission) => submission.id === id);
    if (!item) {
      error.value = "投稿不存在或无权访问。";
      return;
    }

    applySubmission(item);
    if (route.query.visibility === "public") {
      visibility.value = "public";
    }
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
  selectedTags.value = normalizeTags(submission.tags);
  slug.value = submission.slug;
  visibility.value = submission.visibility;
  coverImage.value = submission.coverImage;
  content.value = submission.content;
}

function payload(submit = false): SubmissionPayload {
  return {
    title: title.value,
    summary: summary.value,
    content: content.value,
    category: category.value,
    tags: normalizedTags.value,
    coverImage: coverImage.value,
    slug: slug.value,
    visibility: visibility.value,
    submit,
    turnstileToken: submit ? turnstileToken.value : ""
  };
}

function normalizeTags(values: string[]) {
  return [...new Set(values.map((value) => value.trim()).filter(Boolean))];
}

function normalizeSelectedTags() {
  selectedTags.value = normalizeTags(selectedTags.value);
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
    error.value = isPrivate.value ? "请填写正文后再发布私密文章" : "请填写正文后再提交审核";
    message.value = "";
    return;
  }
  if (submit && turnstileRequired.value && !turnstileToken.value) {
    error.value = turnstileError.value || "请先完成人机验证";
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
    if (submit) {
      resetTurnstile();
    }
    const submitTitle = isPrivate.value ? "私密文章已发布" : "已提交审核";
    const submitMessage = isPrivate.value ? "只有你和管理员可以访问这篇文章。" : "审核结果会通过站内信通知你。";
    message.value = submit ? submitMessage : "草稿已保存。";
    toast.success(submit ? submitTitle : "草稿已保存", submit ? submitMessage : "内容已保存在你的投稿列表。");
  } catch (err) {
    error.value = submissionErrorMessage(err);
    if (submit) {
      resetTurnstile();
    }
  } finally {
    saving.value = false;
  }
}

function submissionErrorMessage(err: unknown) {
  if (err instanceof Error && err.message === "submission limit exceeded") {
    return `已达到投稿频率限制（${submissionLimit.value}），请稍后再提交。`;
  }
  if (err instanceof Error && err.message.includes("turnstile")) {
    return "人机验证未通过，请重新验证后再提交。";
  }

  return err instanceof Error ? err.message : "投稿保存失败";
}

async function renderTurnstile() {
  await nextTick();
  if (!turnstileRequired.value || !turnstileEl.value || turnstileWidgetId.value) {
    return;
  }

  try {
    await loadTurnstileScript();
  } catch {
    turnstileError.value = "人机验证脚本加载失败，请检查浏览器是否能访问 challenges.cloudflare.com。";
    return;
  }

  if (!window.turnstile || !turnstileEl.value || !siteSettings.value?.turnstileSiteKey) {
    turnstileError.value = "人机验证加载失败，请刷新页面后重试。";
    return;
  }

  turnstileError.value = "";
  turnstileWidgetId.value = window.turnstile.render(turnstileEl.value, {
    sitekey: siteSettings.value.turnstileSiteKey,
    callback: (token: string) => {
      turnstileToken.value = token;
      turnstileError.value = "";
    },
    "expired-callback": () => {
      turnstileToken.value = "";
    },
    "error-callback": () => {
      turnstileToken.value = "";
      turnstileError.value = "人机验证无法连接，请确认 Site Key 允许当前域名（localhost/127.0.0.1），或改用 Cloudflare Turnstile 本地测试 Key。";
    }
  });
}

function loadTurnstileScript() {
  if (window.turnstile) {
    return Promise.resolve();
  }

  return new Promise<void>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>("script[data-turnstile]");
    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener("error", () => reject(new Error("turnstile script failed")), { once: true });
      return;
    }

    const script = document.createElement("script");
    script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";
    script.async = true;
    script.defer = true;
    script.dataset.turnstile = "true";
    script.addEventListener("load", () => resolve(), { once: true });
    script.addEventListener("error", () => reject(new Error("turnstile script failed")), { once: true });
    document.head.appendChild(script);
  });
}

function resetTurnstile() {
  turnstileToken.value = "";
  turnstileError.value = "";
  if (turnstileWidgetId.value && window.turnstile) {
    window.turnstile.reset(turnstileWidgetId.value);
  }
}

function removeTurnstile() {
  turnstileToken.value = "";
  turnstileError.value = "";
  if (turnstileWidgetId.value && window.turnstile) {
    window.turnstile.remove(turnstileWidgetId.value);
  }
  turnstileWidgetId.value = "";
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
    return visibility.value === "private" ? "私密已发布" : "已发布";
  }
  if (value === "archived") {
    return "已下架";
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
  if (value === "archived") {
    return "muted";
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
        <RichMarkdownEditor
          v-model="content"
          editor-id="submission-editor"
          upload-category="投稿插图"
          placeholder="撰写投稿正文，支持 Markdown、实时预览、表情、粘贴或拖拽上传图片。"
          :disabled="!canEdit"
          @save="saveDraft"
          @upload-error="(value) => { error = value; message = ''; }"
        />
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
                <strong>{{ isPrivate ? "私密发布" : "提交审核" }}</strong>
                <div class="meta-row"><span>{{ isPrivate ? "私密文章不进入公开审核队列" : "编辑会检查质量、格式和安全" }}</span></div>
              </div>
            </div>
            <div class="step" :class="{ current: status === 'submitted', done: status === 'published' }">
              <span class="step-index">3</span>
              <div>
                <strong>{{ isPrivate ? "仅作者可见" : "通过后发布" }}</strong>
                <div class="meta-row"><span>{{ isPrivate ? "之后可改为公开并提交审核" : "发布后进入公开文章列表" }}</span></div>
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
              <label for="visibility">可见性</label>
              <select v-model="visibility" class="input" id="visibility" :disabled="!canEdit">
                <option value="public">公开：提交后进入审核，通过后公开展示</option>
                <option value="private">私密：直接发布，仅作者和管理员可见</option>
              </select>
            </div>
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
              <ElSelect
                v-model="selectedTags"
                class="element-select"
                input-id="tags"
                multiple
                filterable
                allow-create
                default-first-option
                clearable
                placeholder="选择或输入标签"
                :disabled="!canEdit"
                @change="normalizeSelectedTags"
              >
                <ElOption v-for="item in tagSelectOptions" :key="item.value" :label="item.label" :value="item.value">
                  <span>{{ item.label }}</span>
                  <span class="select-option-meta">{{ item.meta }}</span>
                </ElOption>
              </ElSelect>
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
              alt="文章封面预览"
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
              <strong>{{ isPrivate ? "私密文章会直接发布" : "公开文章需要审核" }}</strong>
              <p>{{ isPrivate ? "发布后仅你和管理员可以访问，之后可改为公开并提交审核。" : "提交后进入待审核状态。编辑可能会通过、退回修改或拒绝投稿。" }}</p>
              <p v-if="!isPrivate">当前频率限制：{{ submissionLimit }}。</p>
	            </div>
		            <p v-if="error" class="error">{{ error }}</p>
		            <div v-if="turnstileRequired" class="field">
		              <label>人机验证</label>
		              <div ref="turnstileEl"></div>
		              <p v-if="turnstileError" class="error" role="alert">{{ turnstileError }}</p>
		            </div>
            <button class="button-secondary" type="button" :disabled="saving || loadingSubmission || !auth.user || !canEdit || status === 'published'" @click="saveDraft">
              {{ saving ? "保存中..." : "保存草稿" }}
            </button>
            <button class="button" type="button" :disabled="saving || loadingSubmission || !auth.user || !canEdit" @click="submitForReview">
              {{ saving ? "提交中..." : (isPrivate ? "发布私密文章" : "提交审核") }}
            </button>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
