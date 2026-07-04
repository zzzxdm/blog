<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter, type LocationQueryRaw } from "vue-router";

import { usePostsStore } from "../stores/posts";
import type { Post } from "../shared/api";

type ArchiveView = "grid" | "list";

const pageSize = 12;
const route = useRoute();
const router = useRouter();
const posts = usePostsStore();

const keyword = ref("");
const category = ref("");
const sort = ref("latest");
const view = ref<ArchiveView>("grid");

const currentPage = computed(() => normalizePage(route.query.page));
const selectedTag = computed(() => stringQuery(route.query.tag));
const total = computed(() => posts.list?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));
const pageNumbers = computed(() => {
  const pages = [1, 2, 3].filter((page) => page <= totalPages.value);
  if (totalPages.value > 3) {
    pages.push(totalPages.value);
  }

  return [...new Set(pages)];
});

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
        <span>{{ selectedTag || "24 个专题" }}</span>
        <span>{{ formatNumber(posts.list?.items.reduce((sum, post) => sum + post.viewCount, 0) ?? 0) }} 次阅读</span>
      </div>
    </section>

    <form class="archive-toolbar" @submit.prevent="applyFilters">
      <input v-model="keyword" class="input" type="search" placeholder="搜索标题、摘要或标签" aria-label="搜索文章">
      <select v-model="category" class="input" aria-label="选择分类">
        <option value="">全部分类</option>
        <option value="工程实践">工程实践</option>
        <option value="产品设计">产品设计</option>
        <option value="运营">内容运营</option>
        <option value="架构">架构</option>
        <option value="Vue3">Vue3</option>
      </select>
      <select v-model="sort" class="input" aria-label="排序方式">
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

    <p v-if="posts.loading" class="muted">正在加载文章...</p>
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

      <nav class="pagination" aria-label="归档分页">
        <button
          class="page-button"
          :class="{ disabled: currentPage <= 1 }"
          type="button"
          :disabled="currentPage <= 1"
          @click="goPage(currentPage - 1)"
        >
          ←
        </button>
        <template v-for="(page, index) in pageNumbers" :key="page">
          <span v-if="index > 0 && page - pageNumbers[index - 1] > 1" class="page-button">...</span>
          <button
            class="page-button"
            :class="{ current: currentPage === page }"
            type="button"
            @click="goPage(page)"
          >
            {{ page }}
          </button>
        </template>
        <button
          class="page-button"
          :class="{ disabled: currentPage >= totalPages }"
          type="button"
          :disabled="currentPage >= totalPages"
          @click="goPage(currentPage + 1)"
        >
          →
        </button>
      </nav>
    </template>
  </main>
</template>
