<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  getMySubmissions,
  getMyPrivatePosts,
  type Post,
  type Submission
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useToastStore } from "../../stores/toast";

const toast = useToastStore();
const posts = ref<Post[]>([]);
const loading = ref(false);
const error = ref("");
const searchQuery = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const submissionsBySlug = ref<Record<string, Submission>>({});

const currentCount = computed(() => posts.value.length);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const [response, submissionResponse] = await Promise.all([getMyPrivatePosts({
      q: searchQuery.value,
      page: page.value,
      pageSize: pageSize.value
    }), getMySubmissions({ status: "published", all: true })]);
    posts.value = response.items;
    submissionsBySlug.value = Object.fromEntries(
      submissionResponse.items
        .filter((item) => item.visibility === "private" && item.publishedPostSlug)
        .map((item) => [item.publishedPostSlug || "", item])
    );
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "私密文章加载失败";
    toast.error("私密文章加载失败", error.value);
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

function formatDate(value: string) {
  return formatDateTime(value);
}

function submissionForPost(post: Post) {
  return submissionsBySlug.value[post.slug];
}
</script>

<template>
  <AccountLayout title="私密文章" description="查看你有权限访问的私密发布内容。">
    <template #actions>
      <RouterLink class="button-secondary" to="/account">返回概览</RouterLink>
    </template>

    <section class="stats-grid" aria-label="私密文章统计">
      <div class="stat-card"><span>私密文章</span><strong>{{ total }}</strong></div>
      <div class="stat-card"><span>当前显示</span><strong>{{ currentCount }}</strong></div>
      <div class="stat-card"><span>访问范围</span><strong>作者</strong></div>
      <div class="stat-card"><span>公开检索</span><strong>隐藏</strong></div>
    </section>

    <section class="panel">
      <form class="archive-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索私密文章" aria-label="搜索私密文章">
        <button class="button-secondary" type="submit">搜索</button>
      </form>

      <LoadingState v-if="loading" variant="table" text="正在加载私密文章..." :rows="4" />
      <p v-else-if="error" class="error">{{ error }}</p>

      <div v-else class="article-list">
        <article v-for="item in posts" :key="item.slug" class="article-card">
          <img :src="item.coverImage" :alt="item.title">
          <div class="article-card-body">
            <div class="meta-row">
              <span class="tag">私密</span>
              <span>{{ item.category }}</span>
              <span>{{ formatDate(item.publishedAt) }}</span>
            </div>
            <h3><RouterLink :to="`/posts/${item.slug}`">{{ item.title }}</RouterLink></h3>
            <p>{{ item.summary }}</p>
            <div class="meta-row">
              <span>{{ item.readingTime }} 分钟阅读</span>
              <span>{{ item.viewCount }} 次阅读</span>
              <RouterLink class="button-secondary" :to="`/posts/${item.slug}`">查看文章</RouterLink>
              <RouterLink v-if="submissionForPost(item)" class="button-secondary" :to="`/submit?id=${encodeURIComponent(submissionForPost(item).id)}`">编辑私密文章</RouterLink>
              <RouterLink v-if="submissionForPost(item)" class="button-secondary" :to="`/submit?id=${encodeURIComponent(submissionForPost(item).id)}&visibility=public`">转公开投稿</RouterLink>
            </div>
          </div>
        </article>
        <p v-if="posts.length === 0" class="muted">没有可访问的私密文章。</p>
      </div>

      <PaginationControls
        v-if="!loading && !error"
        :page="page"
        :page-size="pageSize"
        :total="total"
        :loading="loading"
        item-label="篇私密文章"
        show-page-size
        :page-size-options="[5, 10, 20, 50, 100]"
        @update:page="setPage"
        @update:page-size="setPageSize"
      />
    </section>
  </AccountLayout>
</template>
