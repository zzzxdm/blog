<script setup lang="ts">
import { onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminPosts,
  type AdminPost,
  type AdminPostStats
} from "../../shared/api";

const posts = ref<AdminPost[]>([]);
const stats = ref<AdminPostStats>({ published: 0, draft: 0, review: 0, monthlyViews: "0", total: 0 });
const loading = ref(false);
const error = ref("");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminPosts();
    posts.value = response.items;
    stats.value = response.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章列表加载失败";
  } finally {
    loading.value = false;
  }
}

function statusText(status: AdminPost["status"]) {
  if (status === "published") return "已发布";
  if (status === "scheduled") return "待发布";
  if (status === "review") return "待审核";
  if (status === "archived") return "已归档";
  return "草稿";
}

function statusClass(status: AdminPost["status"]) {
  if (status === "published") return "published";
  if (status === "draft") return "draft";
  if (status === "archived") return "muted";
  return "review";
}

function formatDate(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}
</script>

<template>
  <AdminLayout title="文章管理" description="管理草稿、审核、定时发布和已发布内容。" mobile-title="文章管理" primary-action="新建">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">导入</button>
        <RouterLink class="button" to="/admin/editor">新建文章</RouterLink>
      </div>
    </template>

    <section class="stats-grid" aria-label="文章统计">
      <div class="stat-card"><span>已发布</span><strong>{{ stats.published }}</strong></div>
      <div class="stat-card"><span>草稿</span><strong>{{ stats.draft }}</strong></div>
      <div class="stat-card"><span>待审核</span><strong>{{ stats.review }}</strong></div>
      <div class="stat-card"><span>本月阅读</span><strong>{{ stats.monthlyViews }}</strong></div>
    </section>

    <section class="table-panel" aria-label="文章列表">
      <form class="table-toolbar" @submit.prevent="load">
        <input class="input" type="search" placeholder="搜索标题、作者、标签" aria-label="搜索文章">
        <select class="input" aria-label="文章状态">
          <option>全部状态</option>
          <option>已发布</option>
          <option>草稿</option>
          <option>待审核</option>
        </select>
        <select class="input" aria-label="排序">
          <option>最近更新</option>
          <option>最多阅读</option>
          <option>定时发布</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载文章...</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <table v-else>
        <thead>
          <tr>
            <th>标题</th>
            <th>状态</th>
            <th>分类</th>
            <th>阅读</th>
            <th>评论</th>
            <th>更新时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="post in posts" :key="post.id">
            <td>
              <strong>{{ post.title }}</strong>
              <div class="meta-row">
                <span>{{ post.authorName }}</span>
                <span v-if="post.status === 'scheduled' && post.scheduledAt">定时发布：{{ formatDate(post.scheduledAt) }}</span>
                <span v-else>/posts/{{ post.publishedPostSlug || post.slug }}</span>
              </div>
            </td>
            <td><span class="status" :class="statusClass(post.status)">{{ statusText(post.status) }}</span></td>
            <td>{{ post.category }}</td>
            <td>{{ post.viewCount }}</td>
            <td>{{ post.commentCount }}</td>
            <td>{{ formatDate(post.updatedAt) }}</td>
            <td><RouterLink class="button-secondary" :to="`/admin/editor?id=${post.id}`">编辑</RouterLink></td>
          </tr>
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
