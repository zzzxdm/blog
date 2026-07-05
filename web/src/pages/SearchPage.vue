<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter, type LocationQueryRaw } from "vue-router";

import { getCategories, searchPosts, type Category, type ListResponse, type Post } from "../shared/api";

const pageSize = 10;
const route = useRoute();
const router = useRouter();

const keyword = ref("");
const category = ref("");
const sort = ref("latest");
const categories = ref<Category[]>([]);
const results = ref<ListResponse<Post> | null>(null);
const loading = ref(false);
const error = ref("");

const currentPage = computed(() => normalizePage(route.query.page));
const selectedTag = computed(() => stringQuery(route.query.tag));
const totalPages = computed(() => Math.max(1, Math.ceil((results.value?.total ?? 0) / pageSize)));

void loadCategories();

watch(
  () => route.query,
  () => {
    keyword.value = stringQuery(route.query.q);
    category.value = stringQuery(route.query.category);
    sort.value = stringQuery(route.query.sort) || "latest";
    void load();
  },
  { immediate: true }
);

async function loadCategories() {
  try {
    categories.value = (await getCategories()).items;
  } catch {
    categories.value = [];
  }
}

async function load() {
  const q = keyword.value.trim();
  loading.value = true;
  error.value = "";

  try {
    results.value = await searchPosts({
      q,
      category: category.value,
      tag: selectedTag.value,
      sort: sort.value === "views" || sort.value === "comments" || sort.value === "likes" ? sort.value : undefined,
      page: currentPage.value,
      pageSize
    });
  } catch (err) {
    results.value = null;
    error.value = err instanceof Error ? err.message : "搜索失败";
  } finally {
    loading.value = false;
  }
}

function applyFilters() {
  void router.push({
    path: "/search",
    query: cleanQuery({
      q: keyword.value,
      category: category.value,
      tag: selectedTag.value,
      sort: sort.value === "latest" ? "" : sort.value,
      page: "1"
    })
  });
}

function goPage(page: number) {
  const nextPage = Math.min(Math.max(page, 1), totalPages.value);
  if (nextPage === currentPage.value) {
    return;
  }

  void router.push({
    path: "/search",
    query: cleanQuery({ ...route.query, page: String(nextPage) })
  });
}

function normalizePage(value: unknown) {
  const page = Number.parseInt(stringQuery(value) || "1", 10);
  return Number.isFinite(page) && page > 0 ? page : 1;
}

function stringQuery(value: unknown) {
  return Array.isArray(value) ? String(value[0] ?? "") : String(value ?? "");
}

function cleanQuery(query: Record<string, unknown>): LocationQueryRaw {
  const result: LocationQueryRaw = {};
  Object.entries(query).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    result[key] = String(value);
  });
  return result;
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>站内搜索</h1>
        <p>搜索文章标题、摘要、正文、分类和标签。</p>
      </div>
      <div class="meta-row">
        <span>{{ results?.total ?? 0 }} 条结果</span>
        <span>第 {{ currentPage }} 页</span>
      </div>
    </section>

    <form class="archive-toolbar" @submit.prevent="applyFilters">
      <input v-model="keyword" class="input" type="search" placeholder="输入关键词" aria-label="搜索关键词">
      <select v-model="category" class="input" aria-label="选择分类">
        <option value="">全部分类</option>
        <option v-for="item in categories" :key="item.id" :value="item.name">
          {{ item.name }}{{ item.postCount ? ` (${item.postCount})` : "" }}
        </option>
      </select>
      <select v-model="sort" class="input" aria-label="排序方式">
        <option value="latest">相关和最新</option>
        <option value="views">阅读最多</option>
        <option value="comments">评论最多</option>
        <option value="likes">点赞最多</option>
      </select>
      <button class="button" type="submit">搜索</button>
    </form>

    <p v-if="loading" class="muted">正在搜索...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="!keyword.trim()" class="muted">输入关键词开始搜索。</p>
    <p v-else-if="!results?.items.length" class="muted">没有找到匹配内容。</p>

    <section v-else class="archive-list archive-view" aria-label="搜索结果">
      <article v-for="post in results.items" :key="post.id" class="archive-list-item">
        <div class="archive-list-main">
          <div class="meta-row">
            <span class="tag">{{ post.category }}</span>
            <span>{{ formatDate(post.publishedAt) }}</span>
            <span>{{ post.readingTime }} 分钟阅读</span>
          </div>
          <h3><RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink></h3>
          <p>{{ post.summary }}</p>
          <div class="meta-row">
            <RouterLink v-for="tag in post.tags" :key="tag" :to="`/search?q=${encodeURIComponent(keyword)}&tag=${encodeURIComponent(tag)}`">#{{ tag }}</RouterLink>
          </div>
        </div>
        <div class="archive-list-side">
          <span>{{ formatNumber(post.viewCount) }} 阅读</span>
          <span>{{ post.commentCount }} 评论</span>
        </div>
      </article>
    </section>

    <nav v-if="results && results.total > pageSize" class="pagination" aria-label="搜索分页">
      <button class="page-button" :class="{ disabled: currentPage <= 1 }" type="button" :disabled="currentPage <= 1" @click="goPage(currentPage - 1)">←</button>
      <span class="page-button current">{{ currentPage }} / {{ totalPages }}</span>
      <button class="page-button" :class="{ disabled: currentPage >= totalPages }" type="button" :disabled="currentPage >= totalPages" @click="goPage(currentPage + 1)">→</button>
    </nav>
  </main>
</template>
