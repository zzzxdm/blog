<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import {
  getMyBookmarks,
  setBookmark,
  type BookmarkItem
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";

const bookmarks = ref<BookmarkItem[]>([]);
const loading = ref(false);
const error = ref("");
const searchQuery = ref("");
const categoryFilter = ref("");
const sortMode = ref("bookmarked");

const engineeringCount = computed(() => bookmarks.value.filter((item) => item.category === "工程实践").length);
const thisMonthCount = computed(() => {
  const now = new Date();
  return bookmarks.value.filter((item) => {
    const date = new Date(item.bookmarkedAt);
    return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth();
  }).length;
});
const categoryOptions = computed(() => Array.from(new Set(bookmarks.value.map((item) => item.category))).filter(Boolean));
const visibleBookmarks = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  const filtered = bookmarks.value.filter((item) => {
    const matchesKeyword = !keyword || [
      item.title,
      item.summary,
      item.category,
      item.tags.join(" ")
    ].join(" ").toLowerCase().includes(keyword);
    const matchesCategory = !categoryFilter.value || item.category === categoryFilter.value;

    return matchesKeyword && matchesCategory;
  });

  return [...filtered].sort((left, right) => {
    if (sortMode.value === "published") {
      return new Date(right.publishedAt).getTime() - new Date(left.publishedAt).getTime();
    }
    if (sortMode.value === "views") {
      return right.viewCount - left.viewCount;
    }
    return new Date(right.bookmarkedAt).getTime() - new Date(left.bookmarkedAt).getTime();
  });
});

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getMyBookmarks();
    bookmarks.value = response.items;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "收藏列表加载失败";
  } finally {
    loading.value = false;
  }
}

async function removeBookmark(slug: string) {
  try {
    await setBookmark(slug, false);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "取消收藏失败";
  }
}

function formatDate(value: string) {
  return formatDateTime(value);
}
</script>

<template>
  <AccountLayout title="我的收藏" description="整理值得反复阅读的文章，并支持按专题、分类和时间筛选。">
    <template #actions>
      <RouterLink class="button-secondary" to="/archive">继续浏览</RouterLink>
    </template>

    <section class="stats-grid" aria-label="收藏统计">
      <div class="stat-card"><span>收藏文章</span><strong>{{ bookmarks.length }}</strong></div>
      <div class="stat-card"><span>本月新增</span><strong>{{ thisMonthCount }}</strong></div>
      <div class="stat-card"><span>工程实践</span><strong>{{ engineeringCount }}</strong></div>
      <div class="stat-card"><span>当前显示</span><strong>{{ visibleBookmarks.length }}</strong></div>
    </section>

    <section class="panel">
      <form class="archive-toolbar" @submit.prevent="load">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索收藏文章" aria-label="搜索收藏">
        <select v-model="categoryFilter" class="input" aria-label="收藏分类">
          <option value="">全部分类</option>
          <option v-for="item in categoryOptions" :key="item" :value="item">{{ item }}</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序">
          <option value="bookmarked">最近收藏</option>
          <option value="published">最近发布</option>
          <option value="views">阅读最多</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载收藏...</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <div v-else class="article-list">
        <article v-for="item in visibleBookmarks" :key="item.slug" class="article-card">
          <img :src="item.coverImage" :alt="item.title">
          <div class="article-card-body">
            <div class="meta-row"><span class="tag">{{ item.category }}</span><span>收藏于 {{ formatDate(item.bookmarkedAt) }}</span></div>
            <h3><RouterLink :to="`/posts/${item.slug}`">{{ item.title }}</RouterLink></h3>
            <p>{{ item.summary }}</p>
            <div class="meta-row">
              <span>{{ item.readingTime }} 分钟阅读</span>
              <span>{{ item.viewCount }} 次阅读</span>
              <button class="button-secondary" type="button" @click="removeBookmark(item.slug)">取消收藏</button>
            </div>
          </div>
        </article>
        <p v-if="visibleBookmarks.length === 0" class="muted">没有匹配的收藏文章。</p>
      </div>
    </section>
  </AccountLayout>
</template>
