<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminComments,
  getAdminMessages,
  getAdminPosts,
  getAdminStats,
  getAdminSubmissions,
  getHealth,
  type AdminPostStats,
  type AdminStats,
  type Comment,
  type CommentStats,
  type HealthResponse,
  type MessageStats,
  type StationMessage,
  type Submission,
  type SubmissionStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useToastStore } from "../../stores/toast";

const toast = useToastStore();
interface TodoItem {
  title: string;
  source: string;
  status: string;
  tone: string;
  time: string;
  actionLabel: string;
  actionTo: string;
}

const postStats = ref<AdminPostStats>({ published: 0, draft: 0, review: 0, scheduled: 0, monthlyViews: "0", total: 0 });
const submissionStats = ref<SubmissionStats>({ draft: 0, submitted: 0, returned: 0, rejected: 0, published: 0, archived: 0, total: 0 });
const commentStats = ref<CommentStats>({ total: 0, pending: 0, approved: 0, rejected: 0, spam: 0, deleted: 0, likes: 0, replies: 0 });
const messageStats = ref<MessageStats>({ unread: 0, review: 0, admin: 0, archived: 0, scheduled: 0, total: 0 });
const health = ref<HealthResponse | null>(null);
const stats = ref<AdminStats | null>(null);
const submissions = ref<Submission[]>([]);
const comments = ref<Comment[]>([]);
const messages = ref<StationMessage[]>([]);
const loading = ref(false);
const error = ref("");

const todos = computed<TodoItem[]>(() => {
  const items: TodoItem[] = [];
  const submission = submissions.value[0];
  const comment = comments.value[0];
  const message = messages.value[0];

  if (submission) {
    items.push({
      title: submission.title,
      source: "投稿审核",
      status: "待审核",
      tone: "review",
      time: formatTime(submission.submittedAt || submission.updatedAt),
      actionLabel: "查看",
      actionTo: "/admin/submissions"
    });
  }
  if (comment) {
    items.push({
      title: comment.body,
      source: "评论审核",
      status: comment.riskLevel === "高" ? "高风险" : "待审核",
      tone: comment.riskLevel === "高" ? "banned" : "review",
      time: formatTime(comment.createdAt),
      actionLabel: "处理",
      actionTo: "/admin/comments"
    });
  }
  if (message) {
    items.push({
      title: message.title,
      source: "站内信",
      status: message.status === "unread" ? "未读" : "已发送",
      tone: message.status === "unread" ? "draft" : "published",
      time: formatTime(message.createdAt),
      actionLabel: "查看",
      actionTo: "/admin/messages"
    });
  }

  return items;
});

const healthOk = computed(() => health.value?.status === "ok");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const [healthResponse, postsResponse, submissionsResponse, commentsResponse, messagesResponse, statsResponse] = await Promise.all([
      getHealth(),
      getAdminPosts(),
      getAdminSubmissions("submitted"),
      getAdminComments("pending"),
      getAdminMessages({ status: "unread" }),
      getAdminStats()
    ]);

    health.value = healthResponse;
    postStats.value = postsResponse.stats;
    submissionStats.value = submissionsResponse.stats;
    commentStats.value = commentsResponse.stats;
    submissions.value = submissionsResponse.items.slice(0, 1);
    comments.value = commentsResponse.items.slice(0, 1);
    messages.value = messagesResponse.items.slice(0, 1);
    messageStats.value = messagesResponse.stats;
    stats.value = statsResponse;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "后台概览加载失败";
    toast.error("后台概览加载失败", error.value);
  } finally {
    loading.value = false;
  }
}

function formatTime(value?: string) {
  return formatDateTime(value);
}

function healthSummary() {
  if (!health.value) {
    return "等待健康检查返回";
  }

  return `运行中 · ${health.value.env} · ${formatTime(health.value.time)}`;
}
</script>

<template>
  <AdminLayout title="后台概览" description="查看待办、最近操作、内容状态和系统运行情况。" mobile-title="后台概览" primary-action="新建" primary-action-to="/admin/editor">
    <template #actions>
      <div class="header-actions">
        <RouterLink class="button-secondary" to="/">查看站点</RouterLink>
        <RouterLink class="button" to="/admin/editor">新建文章</RouterLink>
      </div>
    </template>

    <LoadingState v-if="loading" variant="page" text="正在加载后台概览..." :rows="5" />
    <p v-else-if="error" class="error">{{ error }}</p>

    <template v-else>
      <section class="stats-grid" aria-label="后台核心指标">
        <div class="stat-card">
          <span>已发布文章</span>
          <strong>{{ postStats.published }}</strong>
        </div>
        <div class="stat-card">
          <span>待审核投稿</span>
          <strong>{{ submissionStats.submitted }}</strong>
        </div>
        <div class="stat-card">
          <span>未读站内信</span>
          <strong>{{ messageStats.unread }}</strong>
        </div>
        <div class="stat-card">
          <span>本月阅读</span>
          <strong>{{ postStats.monthlyViews }}</strong>
        </div>
      </section>

      <section class="admin-grid-2">
        <section class="table-panel">
          <div class="panel-title" style="padding: 16px 16px 0;">
            <h2>待办事项</h2>
            <RouterLink class="button-secondary" to="/admin/submissions">处理投稿</RouterLink>
          </div>
          <table>
            <thead>
              <tr>
                <th>事项</th>
                <th>来源</th>
                <th>状态</th>
                <th>时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in todos" :key="`${item.source}-${item.title}`">
                <td>{{ item.title }}</td>
                <td>{{ item.source }}</td>
                <td><span class="status" :class="item.tone">{{ item.status }}</span></td>
                <td>{{ item.time }}</td>
                <td><RouterLink class="button-secondary" :to="item.actionTo">{{ item.actionLabel }}</RouterLink></td>
              </tr>
            </tbody>
          </table>
          <p v-if="todos.length === 0" class="muted" style="padding: 0 16px 16px;">暂无待办事项。</p>
        </section>

        <aside class="settings-stack">
          <section class="panel">
            <div class="panel-title">
              <h2>系统状态</h2>
              <span class="status" :class="healthOk ? 'published' : 'banned'">{{ healthOk ? "正常" : "异常" }}</span>
            </div>
            <ul class="link-list">
              <li>
                <strong>API 服务</strong>
                <span>{{ healthSummary() }}</span>
              </li>
              <li>
                <strong>访问保护</strong>
                <span>Cookie 会话、CSRF 校验和每分钟 120 次限流已启用</span>
              </li>
              <li>
                <strong>审核队列</strong>
                <span>{{ submissionStats.submitted }} 篇投稿待审，{{ commentStats.pending }} 条评论待审，{{ messageStats.unread }} 封未读站内信</span>
              </li>
              <li>
                <strong>媒体存储</strong>
                <span>本地 /uploads 提供文件访问，媒体库接口负责元数据和删除保护</span>
              </li>
            </ul>
          </section>

          <section class="panel">
            <div class="panel-title">
              <h2>最近操作</h2>
            </div>
            <div class="timeline">
              <article v-for="post in stats?.topPosts.slice(0, 2)" :key="post.title" class="timeline-item">
                <strong>热门内容更新</strong>
                <p>{{ post.title }}</p>
                <div class="meta-row"><span>{{ post.views }} 次阅读</span><span>{{ post.comments }} 条评论</span></div>
              </article>
            </div>
          </section>
        </aside>
      </section>
    </template>
  </AdminLayout>
</template>
