<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  getAdminPosts,
  getMyPrivatePosts,
  getMySubmissions,
  type AdminPost,
  type Post,
  type Submission
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";

const auth = useAuthStore();
const toast = useToastStore();
const posts = ref<Post[]>([]);
const loading = ref(false);
const error = ref("");
const searchQuery = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
/** 私密投稿：publishedPostSlug 与 slug 均可命中 posts.slug */
const submissionsBySlug = ref<Record<string, Submission>>({});
/** 后台发布的私密文：admin_posts 通常无 submissions 行 */
const adminPostsBySlug = ref<Record<string, AdminPost>>({});

const currentCount = computed(() => posts.value.length);
const isAdmin = computed(() => auth.user?.role === "admin");

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const [response, submissionResponse] = await Promise.all([
      getMyPrivatePosts({
        q: searchQuery.value,
        page: page.value,
        pageSize: pageSize.value
      }),
      getMySubmissions({ status: "published", all: true })
    ]);
    posts.value = response.items;
    submissionsBySlug.value = indexSubmissions(submissionResponse.items);
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;

    if (isAdmin.value) {
      adminPostsBySlug.value = await loadAdminPrivatePosts();
    } else {
      adminPostsBySlug.value = {};
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "私密文章加载失败";
    toast.error("私密文章加载失败", error.value);
  } finally {
    loading.value = false;
  }
}

function indexSubmissions(items: Submission[]): Record<string, Submission> {
  const map: Record<string, Submission> = {};
  for (const item of items) {
    if (item.visibility !== "private") continue;
    if (item.publishedPostSlug) {
      map[item.publishedPostSlug] = item;
    }
    if (item.slug) {
      map[item.slug] = item;
    }
  }
  return map;
}

async function loadAdminPrivatePosts(): Promise<Record<string, AdminPost>> {
  try {
    const response = await getAdminPosts({
      status: "published",
      all: true
    });
    const map: Record<string, AdminPost> = {};
    for (const item of response.items) {
      if (item.visibility !== "private") continue;
      if (item.publishedPostSlug) {
        map[item.publishedPostSlug] = item;
      }
      if (item.slug) {
        map[item.slug] = item;
      }
    }
    return map;
  } catch {
    // 非管理员或接口失败时静默忽略
    return {};
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

function adminPostForPost(post: Post) {
  return adminPostsBySlug.value[post.slug];
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
              <template v-if="submissionForPost(item)">
                <RouterLink class="button-secondary" :to="`/submit?id=${encodeURIComponent(submissionForPost(item)!.id)}`">编辑私密文章</RouterLink>
                <RouterLink class="button-secondary" :to="`/submit?id=${encodeURIComponent(submissionForPost(item)!.id)}&visibility=public`">转公开投稿</RouterLink>
              </template>
              <RouterLink
                v-else-if="adminPostForPost(item)"
                class="button-secondary"
                :to="`/admin/editor?id=${encodeURIComponent(adminPostForPost(item)!.id)}`"
              >后台编辑</RouterLink>
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
