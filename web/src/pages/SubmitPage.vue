<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";

import {
  createSubmission,
  updateSubmission,
  type Submission,
  type SubmissionPayload
} from "../shared/api";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();

const current = ref<Submission | null>(null);
const saving = ref(false);
const message = ref("");
const error = ref("");

const title = ref("用户评论系统应该怎么设计");
const summary = ref("从登录用户评论、审核、举报、通知和禁言机制出发，设计一个可维护的评论系统。");
const category = ref("用户系统");
const tagsText = ref("评论, 用户系统, 审核");
const slug = ref("user-comment-system-design");
const coverImage = ref("https://images.unsplash.com/photo-1519389950473-47ba0277781c?auto=format&fit=crop&w=700&q=80");
const content = ref(`# 用户评论系统应该怎么设计

登录用户评论、审核、举报、通知和禁言机制，是开放内容站点的基础能力。

## 评论不是简单留言

评论需要和用户系统、通知系统、反垃圾策略一起设计。

## 审核状态

- 待审核
- 已通过
- 已拒绝
- 垃圾评论

> 好的评论区应该帮助内容继续生长，而不是成为后台负担。`);

const tags = computed(() => tagsText.value.split(/[,，]/).map((item) => item.trim()).filter(Boolean));
const previewLines = computed(() => content.value.split(/\n+/).map((item) => item.trim()).filter(Boolean));
const status = computed(() => current.value?.status || "draft");

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
        <p>登录用户可以提交文章草稿，审核通过后会发布到站点。</p>
      </div>
      <div class="meta-row">
        <span>{{ current ? `版本 ${current.version}` : "新草稿" }}</span>
        <span class="status" :class="statusClass(status)">{{ statusText(status) }}</span>
      </div>
    </section>

    <section v-if="!auth.user && !auth.loading" class="panel">
      <div class="panel-title"><h2>需要登录</h2></div>
      <p class="muted">登录后可以保存草稿、提交审核，并在站内信里接收审核结果。</p>
      <RouterLink class="button" to="/login">去登录</RouterLink>
    </section>

    <section class="editor-layout">
      <div class="editor-panel">
        <div class="editor-toolbar">
          <div class="tool-group" aria-label="投稿编辑工具栏">
            <button class="tool" type="button" aria-label="加粗">B</button>
            <button class="tool" type="button" aria-label="斜体">I</button>
            <button class="tool" type="button" aria-label="标题">H</button>
            <button class="tool" type="button" aria-label="引用">"</button>
            <button class="tool" type="button" aria-label="代码">{ }</button>
            <button class="tool" type="button" aria-label="链接">↗</button>
          </div>
          <div class="meta-row">
            <span>Markdown</span>
            <span>实时预览</span>
          </div>
        </div>

        <div class="editor-grid">
          <textarea v-model="content" class="markdown-area" aria-label="投稿 Markdown 编辑区"></textarea>

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
              <input v-model="title" class="input" id="title">
            </div>
            <div class="field">
              <label for="summary">摘要</label>
              <textarea v-model="summary" class="input" id="summary"></textarea>
            </div>
            <div class="field">
              <label for="category">建议分类</label>
              <select v-model="category" class="input" id="category">
                <option>用户系统</option>
                <option>工程实践</option>
                <option>内容治理</option>
                <option>写作工作流</option>
              </select>
            </div>
            <div class="field">
              <label for="slug">Slug</label>
              <input v-model="slug" class="input" id="slug">
            </div>
            <div class="field">
              <label for="tags">标签</label>
              <input v-model="tagsText" class="input" id="tags">
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
            <input v-model="coverImage" class="input" aria-label="封面图片 URL">
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
            <button class="button-secondary" type="button" :disabled="saving || !auth.user" @click="saveDraft">
              {{ saving ? "保存中..." : "保存草稿" }}
            </button>
            <button class="button" type="button" :disabled="saving || !auth.user" @click="submitForReview">
              {{ saving ? "提交中..." : "提交审核" }}
            </button>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
