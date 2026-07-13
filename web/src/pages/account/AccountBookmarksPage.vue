<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  getCategories,
  getMyBookmarks,
  setBookmark,
  type BookmarkItem,
  type Category
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";

const bookmarks = ref<BookmarkItem[]>([]);
const categories = ref<Category[]>([]);
const loading = ref(false);
const error = ref("");
const searchQuery = ref("");
const categoryFilter = ref("");
const sortMode = ref("bookmarked");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

const engineeringCount = computed(() => bookmarks.value.filter((item) => item.category === "工程实践").length);
const thisMonthCount = computed(() => {
  const now = new Date();
  return bookmarks.value.filter((item) => {
    const date = new Date(item.bookmarkedAt);
    return date.getFullYear() === now.getFullYear() && date.getMonth() === now.getMonth();
  }).length;
});
const categoryOptions = computed(() => categories.value.map((item) => item.name));

onMounted(() => {
  void load();
  void loadCategories();
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getMyBookmarks({
      q: searchQuery.value,
      category: categoryFilter.value,
      sort: sortMode.value,
      page: page.value,
      pageSize: pageSize.value
    });
    bookmarks.value = response.items;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "收藏列表加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadCategories() {
  try {
    categories.value = (await getCategories({ page: 1, pageSize: 100 })).items;
  } catch {
    categories.value = [];
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
      <div class="stat-card"><span>收藏文章</span><strong>{{ total }}</strong></div>
      <div class="stat-card"><span>本月新增</span><strong>{{ thisMonthCount }}</strong></div>
      <div class="stat-card"><span>工程实践</span><strong>{{ engineeringCount }}</strong></div>
      <div class="stat-card"><span>当前显示</span><strong>{{ bookmarks.length }}</strong></div>
    </section>

    <section class="panel">
      <form class="archive-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索收藏文章" aria-label="搜索收藏">
        <select v-model="categoryFilter" class="input" aria-label="收藏分类" @change="applyFilters">
          <option value="">全部分类</option>
          <option v-for="item in categoryOptions" :key="item" :value="item">{{ item }}</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
          <option value="bookmarked">最近收藏</option>
          <option value="published">最近发布</option>
          <option value="views">阅读最多</option>
        </select>
      </form>

      <LoadingState v-if="loading" variant="table" text="正在加载收藏..." :rows="4" />
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
        <p v-if="bookmarks.length === 0" class="muted">没有匹配的收藏文章。</p>
      </div>
      <PaginationControls
        v-if="!loading && !error"
        :page="page"
        :page-size="pageSize"
        :total="total"
        :loading="loading"
        item-label="篇收藏"
        show-page-size
        :page-size-options="[5, 10, 20, 50, 100]"
        @update:page="setPage"
        @update:page-size="setPageSize"
      />
    </section>
  </AccountLayout>
</template>
