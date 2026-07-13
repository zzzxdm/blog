<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter, type LocationQueryRaw } from "vue-router";

import PaginationControls from "../components/PaginationControls.vue";
import { usePostsStore } from "../stores/posts";
import { getCategories, type Category, type Post } from "../shared/api";

type ArchiveView = "grid" | "list";

const pageSize = 12;
const route = useRoute();
const router = useRouter();
const posts = usePostsStore();

const keyword = ref("");
const category = ref("");
const sort = ref("latest");
const view = ref<ArchiveView>("grid");
const categories = ref<Category[]>([]);

const currentPage = computed(() => normalizePage(route.query.page));
const selectedTag = computed(() => stringQuery(route.query.tag));
const total = computed(() => posts.list?.total ?? 0);
const categoryCount = computed(() => categories.value.length || 4);
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));

watch(
  () => route.query,
  () => {
    keyword.value = stringQuery(route.query.q);
    category.value = stringQuery(route.query.category);
    sort.value = stringQuery(route.query.sort) || "latest";
    view.value = initialView();

    void posts.loadList({
      page: currentPage.value,
      pageSize,
      q: keyword.value,
      category: category.value,
      tag: selectedTag.value,
      sort: sort.value === "views" || sort.value === "comments" ? sort.value : undefined
    });
  },
  { immediate: true }
);

onMounted(() => {
  void loadCategories();
});

async function loadCategories() {
  try {
    categories.value = (await getCategories()).items;
  } catch {
    categories.value = [];
  }
}

function initialView(): ArchiveView {
  const routeView = stringQuery(route.query.view);
  if (routeView === "list" || routeView === "grid") {
    window.localStorage.setItem("archive:view", routeView);
    return routeView;
  }

  return window.localStorage.getItem("archive:view") === "list" ? "list" : "grid";
}

function setView(nextView: ArchiveView) {
  view.value = nextView;
  window.localStorage.setItem("archive:view", nextView);
  void router.replace({
    path: "/archive",
    query: cleanQuery({ ...route.query, view: nextView })
  });
}

function applyFilters() {
  void router.push({
    path: "/archive",
    query: cleanQuery({
      page: "1",
      view: view.value,
      q: keyword.value,
      category: category.value,
      tag: selectedTag.value,
      sort: sort.value === "latest" ? "" : sort.value
    })
  });
}

function handleKeywordInput() {
  if (keyword.value.trim() || !route.query.q) {
    return;
  }

  applyFilters();
}

function goPage(page: number) {
  const nextPage = Math.min(Math.max(page, 1), totalPages.value);
  if (nextPage === currentPage.value) {
    return;
  }

  void router.push({
    path: "/archive",
    query: cleanQuery({ ...route.query, page: String(nextPage), view: view.value })
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

    if (Array.isArray(value)) {
      const values = value
        .filter((item) => item !== undefined && item !== null && item !== "")
        .map((item) => String(item));
      if (values.length) {
        result[key] = values;
      }
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

function tagTone(post: Post, index = 0) {
  if (post.category === "架构" || index % 3 === 1) {
    return "rust";
  }

  if (post.category === "运营" || index % 3 === 2) {
    return "amber";
  }

  return "";
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>文章归档</h1>
        <p>按专题、标签和关键词查找内容。</p>
      </div>
      <div class="meta-row">
        <span>{{ total }} 篇文章</span>
        <span>{{ selectedTag || `${categoryCount} 个分类` }}</span>
        <span>{{ formatNumber(posts.list?.items.reduce((sum, post) => sum + post.viewCount, 0) ?? 0) }} 次阅读</span>
      </div>
    </section>

    <form class="archive-toolbar" @submit.prevent="applyFilters">
      <input v-model="keyword" class="input" type="search" placeholder="搜索标题、摘要或标签" aria-label="搜索文章" @input="handleKeywordInput">
      <select v-model="category" class="input" aria-label="选择分类" @change="applyFilters">
        <option value="">全部分类</option>
        <option v-for="item in categories" :key="item.id" :value="item.name">
          {{ item.name }}{{ item.postCount ? ` (${item.postCount})` : "" }}
        </option>
      </select>
      <select v-model="sort" class="input" aria-label="排序方式" @change="applyFilters">
        <option value="latest">最新发布</option>
        <option value="views">阅读最多</option>
        <option value="comments">评论最多</option>
      </select>
    </form>

    <div class="archive-viewbar">
      <div class="meta-row">
        <span>第 {{ currentPage }} 页</span>
        <span>共 {{ total }} 篇文章</span>
        <span>每页 {{ pageSize }} 篇</span>
      </div>
      <div class="view-toggle" aria-label="归档展示方式">
        <button :class="{ active: view === 'grid' }" type="button" @click="setView('grid')">卡片</button>
        <button :class="{ active: view === 'list' }" type="button" @click="setView('list')">列表</button>
      </div>
    </div>

    <LoadingState v-if="posts.loading" variant="page" text="正在加载文章..." :rows="5" />
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>
    <p v-else-if="!posts.list?.items.length" class="muted">没有找到匹配的文章。</p>

    <template v-else>
      <section v-show="view === 'grid'" class="compact-grid archive-view" aria-label="文章卡片列表">
        <article v-for="(post, index) in posts.list.items" :key="post.id" class="compact-card">
          <img :src="post.coverImage" :alt="post.title">
          <div class="compact-card-body">
            <div class="meta-row">
              <span class="tag" :class="tagTone(post, index)">{{ post.category }}</span>
              <span>{{ post.readingTime }} 分钟</span>
            </div>
            <h3>
              <RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink>
            </h3>
            <p>{{ post.summary }}</p>
            <div class="meta-row">
              <span>{{ formatDate(post.publishedAt) }}</span>
              <span>{{ formatNumber(post.viewCount) }} 阅读</span>
            </div>
          </div>
        </article>
      </section>

      <section v-show="view === 'list'" class="archive-list archive-view" aria-label="文章列表">
        <article v-for="(post, index) in posts.list.items" :key="post.id" class="archive-list-item">
          <div class="archive-list-main">
            <div class="meta-row">
              <span class="tag" :class="tagTone(post, index)">{{ post.category }}</span>
              <span>{{ formatDate(post.publishedAt) }}</span>
              <span>{{ post.readingTime }} 分钟阅读</span>
            </div>
            <h3>
              <RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink>
            </h3>
            <p>{{ post.summary }}</p>
          </div>
          <div class="archive-list-side">
            <span>{{ formatNumber(post.viewCount) }} 阅读</span>
            <span>{{ post.commentCount }} 评论</span>
          </div>
        </article>
      </section>

      <PaginationControls
        :page="currentPage"
        :page-size="pageSize"
        :total="total"
        :loading="posts.loading"
        item-label="篇文章"
        @update:page="goPage"
      />
    </template>
  </main>
</template>
