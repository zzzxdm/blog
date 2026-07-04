<script setup lang="ts">
import { onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminComments,
  updateCommentStatus,
  type Comment,
  type CommentStats
} from "../../shared/api";

const comments = ref<Comment[]>([]);
const stats = ref<CommentStats>({ total: 0, pending: 0, approved: 0, rejected: 0, spam: 0, deleted: 0, likes: 0, replies: 0 });
const status = ref("pending");
const loading = ref(false);
const actingId = ref("");
const error = ref("");
const message = ref("");

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
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">导出</button>
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
        <input class="input" type="search" placeholder="搜索评论内容、用户、文章" aria-label="搜索评论">
        <select v-model="status" class="input" aria-label="评论状态" @change="load">
          <option value="">全部状态</option>
          <option value="pending">待审核</option>
          <option value="approved">已通过</option>
          <option value="rejected">已拒绝</option>
          <option value="spam">垃圾评论</option>
          <option value="deleted">已删除</option>
        </select>
        <select class="input" aria-label="排序">
          <option>最新提交</option>
          <option>举报优先</option>
          <option>点赞最多</option>
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
          <tr v-for="item in comments" :key="item.id">
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
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
