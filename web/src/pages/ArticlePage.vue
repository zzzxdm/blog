<script setup lang="ts">
import { onMounted, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";

import { usePostsStore } from "../stores/posts";

const route = useRoute();
const posts = usePostsStore();

function load() {
  const slug = String(route.params.slug || "");
  if (slug) {
    void posts.loadBySlug(slug);
  }
}

onMounted(load);
watch(() => route.params.slug, load);
</script>

<template>
  <main class="article-page">
    <p v-if="posts.loading" class="muted">正在加载文章...</p>
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>

    <article v-else-if="posts.current" class="article-shell">
      <nav class="breadcrumb" aria-label="当前位置">
        <RouterLink to="/">首页</RouterLink>
        <span>/</span>
        <RouterLink to="/archive">归档</RouterLink>
        <span>/</span>
        <span>{{ posts.current.title }}</span>
      </nav>

      <header class="article-hero">
        <div class="meta-row">
          <span class="tag">{{ posts.current.category }}</span>
          <span>{{ posts.current.readingTime }} 分钟阅读</span>
          <span>{{ new Date(posts.current.publishedAt).toLocaleDateString() }}</span>
        </div>
        <h1>{{ posts.current.title }}</h1>
        <p>{{ posts.current.summary }}</p>
        <div class="meta-row">
          <span>{{ posts.current.authorName }}</span>
          <span>{{ posts.current.viewCount }} 阅读</span>
          <span>{{ posts.current.likeCount }} 赞</span>
          <span>{{ posts.current.dislikeCount }} 踩</span>
          <span>{{ posts.current.commentCount }} 评论</span>
        </div>
      </header>

      <img class="article-cover" :src="posts.current.coverImage" :alt="posts.current.title">

      <section class="article-body">
        <p>{{ posts.current.content }}</p>
        <p>
          后续这里会接入 Markdown 渲染、目录提取、代码块复制和 XSS 清理后的 HTML。
        </p>
      </section>
    </article>
  </main>
</template>
