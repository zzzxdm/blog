<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";

import { getPosts, type ListResponse, type Post } from "../shared/api";

const route = useRoute();
const posts = ref<ListResponse<Post> | null>(null);
const loading = ref(false);
const error = ref("");

const authorId = computed(() => String(route.params.id ?? ""));
const authorName = computed(() => posts.value?.items[0]?.authorName || authorId.value);
const totalViews = computed(() => posts.value?.items.reduce((sum, post) => sum + post.viewCount, 0) ?? 0);
const totalComments = computed(() => posts.value?.items.reduce((sum, post) => sum + post.commentCount, 0) ?? 0);

watch(authorId, () => void load(), { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";

  try {
    posts.value = await getPosts({ author: authorId.value, page: 1, pageSize: 30 });
  } catch (err) {
    posts.value = null;
    error.value = err instanceof Error ? err.message : "作者文章加载失败";
  } finally {
    loading.value = false;
  }
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
        <h1>{{ authorName }}</h1>
        <p>作者主页聚合该作者已发布的公开文章。</p>
      </div>
      <div class="meta-row">
        <span>{{ posts?.total ?? 0 }} 篇文章</span>
        <span>{{ formatNumber(totalViews) }} 阅读</span>
        <span>{{ totalComments }} 评论</span>
      </div>
    </section>

    <p v-if="loading" class="muted">正在加载作者文章...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <section v-else-if="posts?.items.length" class="archive-list archive-view" aria-label="作者文章">
      <article v-for="post in posts.items" :key="post.id" class="archive-list-item">
        <div class="archive-list-main">
          <div class="meta-row">
            <span class="tag">{{ post.category }}</span>
            <span>{{ formatDate(post.publishedAt) }}</span>
            <span>{{ post.readingTime }} 分钟阅读</span>
          </div>
          <h3><RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink></h3>
          <p>{{ post.summary }}</p>
          <div class="meta-row">
            <RouterLink v-for="tag in post.tags" :key="tag" :to="`/archive?tag=${encodeURIComponent(tag)}`">#{{ tag }}</RouterLink>
          </div>
        </div>
        <div class="archive-list-side">
          <span>{{ formatNumber(post.viewCount) }} 阅读</span>
          <span>{{ post.commentCount }} 评论</span>
        </div>
      </article>
    </section>
    <section v-else class="content-shell">
      <article class="article-body">
        <p>没有找到该作者的公开文章。</p>
        <p><RouterLink to="/archive">返回归档查看全部内容</RouterLink></p>
      </article>
    </section>
  </main>
</template>
