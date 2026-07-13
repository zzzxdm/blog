<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  exportAdminComments,
  getAdminComments,
  updateCommentStatus,
  type Comment,
  type CommentStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { downloadJson, exportFileName } from "../../shared/download";
import { useToastStore } from "../../stores/toast";

const toast = useToastStore();

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
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const approvableComments = computed(() => comments.value.filter((item) => item.status === "pending"));

onMounted(load);

async function load(notify = false) {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminComments({
      status: status.value,
      q: searchQuery.value,
      sort: sortMode.value,
      page: page.value,
      pageSize: pageSize.value
    });
    comments.value = response.items;
    stats.value = response.stats;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
    if (notify) {
      toast.success("评论列表已刷新", `当前筛选共 ${total.value} 条评论。`);
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论列表加载失败";
    toast.error("评论列表加载失败", error.value);
  } finally {
    loading.value = false;
  }
}

async function applyFilters() {
  page.value = 1;
  await load();
}

async function setPage(value: number) {
  page.value = value;
  await load();
}

async function setPageSize(value: number) {
  pageSize.value = value;
  page.value = 1;
  await load();
}

async function setStatus(item: Comment, nextStatus: Comment["status"]) {
  actingId.value = item.id;
  error.value = "";
  message.value = "";

  try {
    await updateCommentStatus(item.id, nextStatus);
    message.value = "评论状态已更新。";
    toast.success("评论状态已更新", statusText(nextStatus));
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论状态更新失败";
    toast.error("评论状态更新失败", error.value);
  } finally {
    actingId.value = "";
  }
}

async function approveVisiblePending() {
  if (!approvableComments.value.length) {
    toast.info("没有待审核评论", "当前列表没有可批量通过的评论。");
    return;
  }

  bulkActing.value = true;
  error.value = "";
  message.value = "";

  try {
    const approvedCount = approvableComments.value.length;
    for (const item of approvableComments.value) {
      await updateCommentStatus(item.id, "approved");
    }
    message.value = `已通过 ${approvedCount} 条评论。`;
    toast.success("批量审核完成", `已通过 ${approvedCount} 条评论。`);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "批量审核失败";
    toast.error("批量审核失败", error.value);
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
    toast.success("评论导出已生成", "下载文件已创建。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论导出失败";
    toast.error("评论导出失败", error.value);
  } finally {
    exporting.value = false;
  }
}

function formatDate(value: string) {
  return formatDateTime(value);
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
        <button class="button" type="button" @click="load(true)">刷新</button>
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
      <form class="table-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索评论内容、用户、文章" aria-label="搜索评论">
        <select v-model="status" class="input" aria-label="评论状态" @change="applyFilters">
          <option value="">全部状态</option>
          <option value="pending">待审核</option>
          <option value="approved">已通过</option>
          <option value="rejected">已拒绝</option>
          <option value="spam">垃圾评论</option>
          <option value="deleted">已删除</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
          <option value="latest">最新提交</option>
          <option value="risk">风险优先</option>
          <option value="likes">点赞最多</option>
        </select>
      </form>

      <LoadingState v-if="loading" variant="table" text="正在加载评论..." :rows="4" />
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
          <tr v-if="comments.length === 0">
            <td colspan="7" class="muted">没有匹配的评论。</td>
          </tr>
        </tbody>
      </table>
      <PaginationControls
        v-if="!loading"
        :page="page"
        :page-size="pageSize"
        :total="total"
        :loading="loading"
        item-label="条评论"
        show-page-size
        :page-size-options="[5, 10, 20, 50, 100]"
        @update:page="setPage"
        @update:page-size="setPageSize"
      />
    </section>
  </AdminLayout>
</template>
