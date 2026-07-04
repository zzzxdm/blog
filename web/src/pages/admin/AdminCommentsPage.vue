<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  exportAdminComments,
  getAdminComments,
  updateCommentStatus,
  type Comment,
  type CommentStats
} from "../../shared/api";
import { downloadJson, exportFileName } from "../../shared/download";

const comments = ref<Comment[]>([]);
const stats = ref<CommentStats>({ total: 0, pending: 0, approved: 0, rejected: 0, spam: 0, deleted: 0, likes: 0, replies: 0 });
const status = ref("pending");
const loading = ref(false);
const exporting = ref(false);
const bulkActing = ref(false);
const actingId = ref("");
const error = ref("");
const message = ref("");
const searchQuery = ref("");
const sortMode = ref("latest");

const visibleComments = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  const riskRank: Record<string, number> = { 高: 3, 中: 2, 低: 1 };
  const filtered = comments.value.filter((item) => {
    if (!keyword) {
      return true;
    }

    return [
      item.body,
      item.authorName,
      item.authorId,
      item.postTitle || "",
      item.postSlug,
      item.riskLevel || ""
    ].join(" ").toLowerCase().includes(keyword);
  });

  return [...filtered].sort((left, right) => {
    if (sortMode.value === "likes") {
      return right.likeCount - left.likeCount;
    }
    if (sortMode.value === "risk") {
      return (riskRank[right.riskLevel || "低"] || 0) - (riskRank[left.riskLevel || "低"] || 0);
    }
    return new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime();
  });
});
const approvableComments = computed(() => visibleComments.value.filter((item) => item.status === "pending"));

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminComments(status.value);
    comments.value = response.items;
    stats.value = response.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论列表加载失败";
  } finally {
    loading.value = false;
  }
}

async function setStatus(item: Comment, nextStatus: Comment["status"]) {
  actingId.value = item.id;
  error.value = "";
  message.value = "";

  try {
    await updateCommentStatus(item.id, nextStatus);
    message.value = "评论状态已更新。";
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论状态更新失败";
  } finally {
    actingId.value = "";
  }
}

async function approveVisiblePending() {
  if (!approvableComments.value.length) {
    return;
  }

  bulkActing.value = true;
  error.value = "";
  message.value = "";

  try {
    for (const item of approvableComments.value) {
      await updateCommentStatus(item.id, "approved");
    }
    message.value = `已通过 ${approvableComments.value.length} 条评论。`;
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "批量审核失败";
  } finally {
    bulkActing.value = false;
  }
}

async function exportComments() {
  exporting.value = true;
  error.value = "";
  message.value = "";

  try {
    downloadJson(exportFileName("comments"), await exportAdminComments(status.value));
    message.value = "评论导出已生成。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论导出失败";
  } finally {
    exporting.value = false;
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function statusText(value: Comment["status"]) {
  if (value === "approved") {
    return "已通过";
  }
  if (value === "pending") {
    return "待审核";
  }
  if (value === "rejected") {
    return "已拒绝";
  }
  if (value === "spam") {
    return "垃圾评论";
  }
  return "已删除";
}

function statusClass(value: Comment["status"]) {
  if (value === "approved") {
    return "published";
  }
  if (value === "spam" || value === "deleted") {
    return "banned";
  }
  if (value === "rejected") {
    return "rejected";
  }
  return "review";
}
</script>

<template>
  <AdminLayout title="评论管理" description="审核用户评论、处理举报，并对异常用户进行禁言或封禁。" mobile-title="评论管理" primary-action="批量通过">
    <template #mobile-action>
      <button class="button" type="button" :disabled="bulkActing || !approvableComments.length" @click="approveVisiblePending">
        {{ bulkActing ? "处理中..." : `通过 ${approvableComments.length}` }}
      </button>
    </template>

    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" :disabled="exporting" @click="exportComments">{{ exporting ? "导出中..." : "导出" }}</button>
        <button class="button-secondary" type="button" :disabled="bulkActing || !approvableComments.length" @click="approveVisiblePending">
          {{ bulkActing ? "处理中..." : `通过当前 ${approvableComments.length}` }}
        </button>
        <button class="button" type="button" @click="load">刷新</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="评论统计">
      <div class="stat-card"><span>待审核</span><strong>{{ stats.pending }}</strong></div>
      <div class="stat-card"><span>全部评论</span><strong>{{ stats.total }}</strong></div>
      <div class="stat-card"><span>已拒绝</span><strong>{{ stats.rejected }}</strong></div>
      <div class="stat-card"><span>垃圾评论</span><strong>{{ stats.spam }}</strong></div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="table-panel" aria-label="评论列表">
      <form class="table-toolbar" @submit.prevent="load">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索评论内容、用户、文章" aria-label="搜索评论">
        <select v-model="status" class="input" aria-label="评论状态" @change="load">
          <option value="">全部状态</option>
          <option value="pending">待审核</option>
          <option value="approved">已通过</option>
          <option value="rejected">已拒绝</option>
          <option value="spam">垃圾评论</option>
          <option value="deleted">已删除</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序">
          <option value="latest">最新提交</option>
          <option value="risk">风险优先</option>
          <option value="likes">点赞最多</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载评论...</p>
      <table v-else>
        <thead>
          <tr>
            <th>评论</th>
            <th>用户</th>
            <th>文章</th>
            <th>状态</th>
            <th>风险</th>
            <th>时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in visibleComments" :key="item.id">
            <td>
              <strong>{{ item.body }}</strong>
              <div class="meta-row"><span>{{ item.likeCount }} 次点赞</span><span v-if="item.parentId">回复 {{ item.parentId }}</span></div>
            </td>
            <td>{{ item.authorName }}<div class="meta-row"><span>{{ item.authorId }}</span></div></td>
            <td>{{ item.postTitle || item.postSlug }}</td>
            <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
            <td>{{ item.riskLevel || "低" }}</td>
            <td>{{ formatDate(item.createdAt) }}</td>
            <td>
              <div class="header-actions">
                <button class="button-secondary" type="button" :disabled="actingId === item.id" @click="setStatus(item, 'approved')">通过</button>
                <button class="button-secondary" type="button" :disabled="actingId === item.id" @click="setStatus(item, 'rejected')">拒绝</button>
                <button class="button-secondary" type="button" :disabled="actingId === item.id" @click="setStatus(item, 'deleted')">删除</button>
              </div>
            </td>
          </tr>
          <tr v-if="visibleComments.length === 0">
            <td colspan="7" class="muted">没有匹配的评论。</td>
          </tr>
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
