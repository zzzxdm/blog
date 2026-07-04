<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";

import {
  getAccountSettings,
  getMessages,
  getMyBookmarks,
  getMyComments,
  getMySubmissions,
  type AccountSettings,
  type BookmarkItem,
  type Comment,
  type MessageStats,
  type StationMessage,
  type Submission,
  type SubmissionStats
} from "../shared/api";

const route = useRoute();

const account = ref<AccountSettings | null>(null);
const submissions = ref<Submission[]>([]);
const submissionStats = ref<SubmissionStats>({ draft: 0, submitted: 0, returned: 0, rejected: 0, published: 0, total: 0 });
const comments = ref<Comment[]>([]);
const bookmarks = ref<BookmarkItem[]>([]);
const messages = ref<StationMessage[]>([]);
const messageStats = ref<MessageStats>({ unread: 0, review: 0, admin: 0, archived: 0, total: 0 });
const loading = ref(false);
const error = ref("");

const commentTotal = computed(() => comments.value.length);
const bookmarkTotal = computed(() => bookmarks.value.length);

onMounted(load);

function active(path: string) {
  return route.path === path;
}

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const [accountResponse, submissionsResponse, commentsResponse, bookmarksResponse, messagesResponse] = await Promise.all([
      getAccountSettings(),
      getMySubmissions(),
      getMyComments(),
      getMyBookmarks(),
      getMessages()
    ]);

    account.value = accountResponse;
    submissions.value = submissionsResponse.items.slice(0, 2);
    submissionStats.value = submissionsResponse.stats;
    comments.value = commentsResponse.items.slice(0, 2);
    bookmarks.value = bookmarksResponse.items.slice(0, 2);
    messages.value = messagesResponse.items.slice(0, 2);
    messageStats.value = messagesResponse.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "个人中心加载失败";
  } finally {
    loading.value = false;
  }
}

function statusText(status: string) {
  if (status === "submitted" || status === "pending") return "待审核";
  if (status === "returned") return "退回修改";
  if (status === "rejected") return "已拒绝";
  if (status === "published" || status === "approved") return "已通过";
  if (status === "draft") return "草稿";
  return "已删除";
}

function statusClass(status: string) {
  if (status === "published" || status === "approved") return "published";
  if (status === "draft") return "draft";
  if (status === "rejected" || status === "deleted" || status === "spam") return "rejected";
  return "review";
}

function formatTime(value?: string) {
  if (!value) {
    return "未提交";
  }

  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>个人中心</h1>
        <p>管理资料、评论、收藏、投稿和站内信。</p>
      </div>
    </section>

    <p v-if="loading" class="muted">正在加载个人中心...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <section v-else class="account-layout">
      <aside class="panel">
        <div class="profile-card">
          <div class="profile-hero">
            <span class="avatar">{{ account?.avatarText || "用" }}</span>
            <div>
              <strong>{{ account?.displayName || "用户" }}</strong>
              <div class="meta-row">
                <span>{{ account?.email }}</span>
                <span>资料完整度 {{ account?.profileCompleteness || 0 }}%</span>
              </div>
            </div>
          </div>
          <nav class="account-nav" aria-label="个人中心导航">
            <RouterLink :class="{ active: active('/account') }" to="/account">概览</RouterLink>
            <RouterLink :class="{ active: active('/account/comments') }" to="/account/comments">我的评论</RouterLink>
            <RouterLink :class="{ active: active('/account/bookmarks') }" to="/account/bookmarks">我的收藏</RouterLink>
            <RouterLink :class="{ active: active('/account/submissions') }" to="/account/submissions">我的投稿</RouterLink>
            <RouterLink :class="{ active: active('/account/messages') }" to="/account/messages">站内信</RouterLink>
            <RouterLink :class="{ active: active('/account/settings') }" to="/account/settings">账号设置</RouterLink>
          </nav>
        </div>
      </aside>

      <div class="settings-stack">
        <section class="stats-grid" aria-label="用户统计">
          <div class="stat-card"><span>评论</span><strong>{{ commentTotal }}</strong></div>
          <div class="stat-card"><span>收藏</span><strong>{{ bookmarkTotal }}</strong></div>
          <div class="stat-card"><span>投稿</span><strong>{{ submissionStats.total }}</strong></div>
          <div class="stat-card"><span>未读站内信</span><strong>{{ messageStats.unread }}</strong></div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>我的投稿</h2>
            <RouterLink class="button-secondary" to="/submit">继续投稿</RouterLink>
          </div>
          <div class="timeline">
            <article v-for="item in submissions" :key="item.id" class="timeline-item">
              <strong>{{ item.title }}</strong>
              <p>{{ item.reviewNote || item.summary || "提交后进入审核队列，编辑会检查结构和引用来源。" }}</p>
              <div class="meta-row">
                <span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span>
                <span>{{ formatTime(item.submittedAt || item.updatedAt) }}</span>
                <RouterLink v-if="item.publishedPostSlug" :to="`/posts/${item.publishedPostSlug}`">查看文章</RouterLink>
              </div>
            </article>
            <p v-if="submissions.length === 0" class="muted">还没有投稿。</p>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>我的评论</h2>
            <RouterLink class="button-secondary" to="/posts/blog-system-design">去评论</RouterLink>
          </div>
          <div class="timeline">
            <article v-for="item in comments" :key="item.id" class="timeline-item">
              <strong>{{ item.body }}</strong>
              <p>评论于《{{ item.postTitle || item.postSlug }}》</p>
              <div class="meta-row">
                <span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span>
                <span>{{ item.likeCount }} 次点赞</span>
                <RouterLink :to="`/posts/${item.postSlug}`">查看上下文</RouterLink>
              </div>
            </article>
            <p v-if="comments.length === 0" class="muted">还没有评论。</p>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>我的收藏</h2>
            <RouterLink class="button-secondary" to="/account/bookmarks">查看全部</RouterLink>
          </div>
          <ul class="link-list">
            <li v-for="item in bookmarks" :key="item.slug">
              <RouterLink :to="`/posts/${item.slug}`">
                <strong>{{ item.title }}</strong>
                <span>{{ item.category }} · {{ item.readingTime }} 分钟阅读</span>
              </RouterLink>
            </li>
          </ul>
          <p v-if="bookmarks.length === 0" class="muted">还没有收藏文章。</p>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>站内信</h2>
            <RouterLink class="button-secondary" to="/account/messages">查看全部</RouterLink>
          </div>
          <div class="timeline">
            <article v-for="item in messages" :key="item.id" class="timeline-item">
              <strong>{{ item.title }}</strong>
              <p>{{ item.body }}</p>
              <div class="meta-row">
                <span>{{ formatTime(item.createdAt) }}</span>
                <RouterLink v-if="item.targetType === 'submission'" to="/account/submissions">查看投稿</RouterLink>
              </div>
            </article>
            <p v-if="messages.length === 0" class="muted">暂无站内信。</p>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>账号设置</h2>
            <RouterLink class="button-secondary" to="/account/settings">编辑</RouterLink>
          </div>
          <div class="settings-grid">
            <div class="field">
              <label for="display-name">昵称</label>
              <input class="input" id="display-name" :value="account?.displayName" readonly>
            </div>
            <div class="field">
              <label for="email">邮箱</label>
              <input class="input" id="email" :value="account?.email" readonly>
            </div>
            <div class="field">
              <label for="security">安全等级</label>
              <input class="input" id="security" :value="account?.securityLevel" readonly>
            </div>
            <div class="field">
              <label for="notification">通知偏好</label>
              <input class="input" id="notification" :value="account?.emailNotification ? '站内信 + 邮件提醒' : '仅站内信'" readonly>
            </div>
          </div>
        </section>
      </div>
    </section>
  </main>
</template>
