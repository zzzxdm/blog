<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  getMySubmissions,
  type Submission,
  type SubmissionStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";

const submissions = ref<Submission[]>([]);
const stats = ref<SubmissionStats>({ draft: 0, submitted: 0, returned: 0, rejected: 0, published: 0, archived: 0, total: 0 });
const loading = ref(false);
const error = ref("");
const status = ref("");
const searchQuery = ref("");
const sortMode = ref("updated");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const returnedSubmission = computed(() => submissions.value.find((item) => item.status === "returned" && item.reviewNote));

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getMySubmissions({
      status: status.value,
      q: searchQuery.value,
      sort: sortMode.value,
      page: page.value,
      pageSize: pageSize.value
    });
    submissions.value = response.items;
    stats.value = response.stats;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿列表加载失败";
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

function formatDate(value?: string) {
  return formatDateTime(value, "未提交");
}

function visibilityText(value: Submission["visibility"]) {
  return value === "private" ? "私密" : "公开";
}

function statusText(value: Submission["status"]) {
  if (value === "submitted") {
    return "待审核";
  }
  if (value === "returned") {
    return "退回";
  }
  if (value === "rejected") {
    return "已拒绝";
  }
  if (value === "published") {
    return "已发布";
  }
  if (value === "archived") {
    return "已下架";
  }
  return "草稿";
}

function statusClass(value: Submission["status"]) {
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
  <AccountLayout title="我的投稿" description="查看投稿草稿、审核进度、退回原因和已发布投稿。">
    <template #actions>
      <RouterLink class="button" to="/submit">新建投稿</RouterLink>
    </template>

    <section class="stats-grid" aria-label="投稿统计">
      <div class="stat-card"><span>草稿</span><strong>{{ stats.draft }}</strong></div>
      <div class="stat-card"><span>待审核</span><strong>{{ stats.submitted }}</strong></div>
      <div class="stat-card"><span>退回修改</span><strong>{{ stats.returned }}</strong></div>
      <div class="stat-card"><span>已发布</span><strong>{{ stats.published }}</strong></div>
    </section>

    <section class="table-panel">
      <form class="table-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索投稿标题" aria-label="搜索投稿">
        <select v-model="status" class="input" aria-label="投稿状态" @change="applyFilters">
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="submitted">待审核</option>
          <option value="returned">退回修改</option>
              <option value="published">已发布</option>
              <option value="archived">已下架</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
          <option value="updated">最近更新</option>
          <option value="submitted">最近提交</option>
          <option value="published">已发布优先</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载投稿...</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <table v-else>
        <thead>
          <tr><th>投稿</th><th>状态</th><th>可见性</th><th>分类</th><th>提交时间</th><th>审核意见</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="item in submissions" :key="item.id">
            <td><strong>{{ item.title }}</strong><div class="meta-row"><span>版本 {{ item.version }}</span><span>{{ item.wordCount }} 字</span></div></td>
            <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
            <td>{{ visibilityText(item.visibility) }}</td>
            <td>{{ item.category }}</td>
            <td>{{ formatDate(item.submittedAt) }}</td>
            <td>{{ item.reviewNote || (item.visibility === "private" && item.status === "published" ? "私密发布，无需审核" : (item.status === "submitted" ? "等待编辑审核" : "未提交审核")) }}</td>
            <td>
              <div class="header-actions">
                <template v-if="item.status === 'published' && item.publishedPostSlug">
                  <RouterLink class="button-secondary" :to="`/posts/${item.publishedPostSlug}`">查看文章</RouterLink>
                  <RouterLink v-if="item.visibility === 'private'" class="button-secondary" :to="`/submit?id=${encodeURIComponent(item.id)}`">编辑私密文章</RouterLink>
                  <RouterLink v-if="item.visibility === 'private'" class="button-secondary" :to="`/submit?id=${encodeURIComponent(item.id)}&visibility=public`">转公开投稿</RouterLink>
                </template>
                <span v-else-if="item.status === 'archived'" class="muted">已下架</span>
                <RouterLink v-else class="button-secondary" :to="`/submit?id=${encodeURIComponent(item.id)}`">{{ item.status === "draft" || item.status === "returned" ? "继续编辑" : "查看" }}</RouterLink>
              </div>
            </td>
          </tr>
          <tr v-if="submissions.length === 0">
            <td colspan="7" class="muted">没有匹配的投稿。</td>
          </tr>
        </tbody>
      </table>
      <PaginationControls
        v-if="!loading && !error"
        :page="page"
        :page-size="pageSize"
        :total="total"
        :loading="loading"
        item-label="篇投稿"
        show-page-size
        :page-size-options="[5, 10, 20, 50, 100]"
        @update:page="setPage"
        @update:page-size="setPageSize"
      />
    </section>

    <section v-if="returnedSubmission" class="panel">
      <div class="panel-title"><h2>退回修改说明</h2></div>
      <div class="review-note">
        <strong>《{{ returnedSubmission.title }}》</strong>
        <p>{{ returnedSubmission.reviewNote }}</p>
      </div>
    </section>
  </AccountLayout>
</template>
