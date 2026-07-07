<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import {
  getMySubmissions,
  type Submission,
  type SubmissionStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";

const submissions = ref<Submission[]>([]);
const stats = ref<SubmissionStats>({ draft: 0, submitted: 0, returned: 0, rejected: 0, published: 0, total: 0 });
const loading = ref(false);
const error = ref("");
const status = ref("");
const searchQuery = ref("");
const sortMode = ref("updated");

const returnedSubmission = computed(() => submissions.value.find((item) => item.status === "returned" && item.reviewNote));
const visibleSubmissions = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  const filtered = submissions.value.filter((item) => {
    if (!keyword) {
      return true;
    }

    return [
      item.title,
      item.summary,
      item.category,
      item.slug,
      item.reviewNote,
      item.tags.join(" ")
    ].join(" ").toLowerCase().includes(keyword);
  });

  return [...filtered].sort((left, right) => {
    if (sortMode.value === "submitted") {
      return dateValue(right.submittedAt || right.updatedAt) - dateValue(left.submittedAt || left.updatedAt);
    }
    if (sortMode.value === "published") {
      const leftPublished = left.status === "published" ? 1 : 0;
      const rightPublished = right.status === "published" ? 1 : 0;
      if (leftPublished !== rightPublished) {
        return rightPublished - leftPublished;
      }
      return dateValue(right.publishedAt || right.updatedAt) - dateValue(left.publishedAt || left.updatedAt);
    }
    return dateValue(right.updatedAt) - dateValue(left.updatedAt);
  });
});

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getMySubmissions(status.value);
    submissions.value = response.items;
    stats.value = response.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿列表加载失败";
  } finally {
    loading.value = false;
  }
}

function formatDate(value?: string) {
  return formatDateTime(value, "未提交");
}

function dateValue(value?: string) {
  return value ? new Date(value).getTime() : 0;
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
      <form class="table-toolbar" @submit.prevent="load">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索投稿标题" aria-label="搜索投稿">
        <select v-model="status" class="input" aria-label="投稿状态" @change="load">
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="submitted">待审核</option>
          <option value="returned">退回修改</option>
          <option value="published">已发布</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序">
          <option value="updated">最近更新</option>
          <option value="submitted">最近提交</option>
          <option value="published">已发布优先</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载投稿...</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <table v-else>
        <thead>
          <tr><th>投稿</th><th>状态</th><th>分类</th><th>提交时间</th><th>审核意见</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="item in visibleSubmissions" :key="item.id">
            <td><strong>{{ item.title }}</strong><div class="meta-row"><span>版本 {{ item.version }}</span><span>{{ item.wordCount }} 字</span></div></td>
            <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
            <td>{{ item.category }}</td>
            <td>{{ formatDate(item.submittedAt) }}</td>
            <td>{{ item.reviewNote || (item.status === "submitted" ? "等待编辑审核" : "未提交审核") }}</td>
            <td>
              <RouterLink v-if="item.status === 'published' && item.publishedPostSlug" class="button-secondary" :to="`/posts/${item.publishedPostSlug}`">查看文章</RouterLink>
              <RouterLink v-else class="button-secondary" :to="`/submit?id=${encodeURIComponent(item.id)}`">{{ item.status === "draft" || item.status === "returned" ? "继续编辑" : "查看" }}</RouterLink>
            </td>
          </tr>
          <tr v-if="visibleSubmissions.length === 0">
            <td colspan="6" class="muted">没有匹配的投稿。</td>
          </tr>
        </tbody>
      </table>
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
