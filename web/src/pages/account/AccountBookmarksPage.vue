<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import {
  getMyBookmarks,
  setBookmark,
  type BookmarkItem
} from "../../shared/api";

const bookmarks = ref<BookmarkItem[]>([]);
const loading = ref(false);
const error = ref("");

const engineeringCount = computed(() => bookmarks.value.filter((item) => item.category === "工程实践").length);
const thisMonthCount = computed(() => {
  const now = new Date();
  return bookmarks.value.filter((item) => {
    const date = new Date(item.bookmarkedAt);
    return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth();
  }).length;
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
  return new Date(value).toLocaleDateString("zh-CN", {
    month: "2-digit",
    day: "2-digit"
  });
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
      <div class="stat-card"><span>待读</span><strong>0</strong></div>
    </section>

    <section class="panel">
      <form class="archive-toolbar" @submit.prevent="load">
        <input class="input" type="search" placeholder="搜索收藏文章" aria-label="搜索收藏">
        <select class="input" aria-label="收藏分类"><option>全部分类</option><option>工程实践</option><option>产品设计</option><option>内容治理</option></select>
        <select class="input" aria-label="排序"><option>最近收藏</option><option>最近发布</option><option>阅读最多</option></select>
      </form>

      <p v-if="loading" class="muted">正在加载收藏...</p>
      <p v-else-if="error" class="error">{{ error }}</p>

      <div v-else class="article-list">
        <article v-for="item in bookmarks" :key="item.slug" class="article-card">
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
        <p v-if="bookmarks.length === 0" class="muted">还没有收藏文章。</p>
      </div>
    </section>
  </AccountLayout>
</template>
