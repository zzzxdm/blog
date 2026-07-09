<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  getMyComments,
  type Comment,
  type CommentStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";

const comments = ref<Comment[]>([]);
const stats = ref<CommentStats>({ total: 0, pending: 0, approved: 0, rejected: 0, spam: 0, deleted: 0, likes: 0, replies: 0 });
const status = ref("");
const loading = ref(false);
const error = ref("");
const searchQuery = ref("");
const sortMode = ref("created");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const commentEntry = computed(() => comments.value[0] ? `/posts/${comments.value[0].postSlug}` : "/archive");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getMyComments({
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
  } catch (err) {
    error.value = err instanceof Error ? err.message : "评论列表加载失败";
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
  if (value === "deleted" || value === "spam") {
    return "banned";
  }
  if (value === "rejected") {
    return "rejected";
  }
  return "review";
}

function formatDate(value: string) {
  return formatDateTime(value);
}
</script>

<template>
  <AccountLayout title="我的评论" description="查看评论审核状态、回复记录和已通过评论。">
    <template #actions>
      <RouterLink class="button-secondary" :to="commentEntry">{{ comments.length ? "继续讨论" : "去阅读" }}</RouterLink>
    </template>

    <section class="stats-grid" aria-label="评论统计">
      <div class="stat-card"><span>全部评论</span><strong>{{ stats.total }}</strong></div>
      <div class="stat-card"><span>待审核</span><strong>{{ stats.pending }}</strong></div>
      <div class="stat-card"><span>获赞</span><strong>{{ stats.likes }}</strong></div>
      <div class="stat-card"><span>被回复</span><strong>{{ stats.replies }}</strong></div>
    </section>

    <section class="table-panel">
      <form class="table-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索评论内容或文章标题" aria-label="搜索我的评论">
        <select v-model="status" class="input" aria-label="评论状态" @change="applyFilters">
          <option value="">全部状态</option>
          <option value="pending">待审核</option>
          <option value="approved">已通过</option>
          <option value="rejected">已拒绝</option>
          <option value="deleted">已删除</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
          <option value="created">最近评论</option>
          <option value="likes">获赞最多</option>
          <option value="replies">被回复优先</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载评论...</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <table v-else>
        <thead>
          <tr><th>评论</th><th>文章</th><th>状态</th><th>互动</th><th>时间</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="item in comments" :key="item.id">
            <td><strong>{{ item.body }}</strong><div class="meta-row"><span v-if="item.parentId">回复主评论 {{ item.parentId }}</span></div></td>
            <td>{{ item.postTitle || item.postSlug }}</td>
            <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
            <td>{{ item.likeCount }} 赞<span v-if="item.replyCount"> · {{ item.replyCount }} 回复</span></td>
            <td>{{ formatDate(item.createdAt) }}</td>
            <td><RouterLink class="button-secondary" :to="`/posts/${item.postSlug}`">查看</RouterLink></td>
          </tr>
          <tr v-if="comments.length === 0">
            <td colspan="6" class="muted">没有匹配的评论。</td>
          </tr>
        </tbody>
      </table>
      <PaginationControls
        v-if="!loading && !error"
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
  </AccountLayout>
</template>
