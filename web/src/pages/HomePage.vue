<script setup lang="ts">
import { onMounted } from "vue";
import { RouterLink } from "vue-router";

import { useHealthStore } from "../stores/health";
import { usePostsStore } from "../stores/posts";

const health = useHealthStore();
const posts = usePostsStore();

onMounted(() => {
  void health.load();
  void posts.loadList({ page: 1, pageSize: 3 });
});
</script>

<template>
  <main class="page">
    <section class="hero">
      <div>
        <span class="eyebrow">Vue3 + Go/Gin</span>
        <h1>云间笔记</h1>
        <p>一个面向长期写作、投稿审核、用户评论和站内信的现代化博客系统。</p>
      </div>
      <aside class="status-panel">
        <h2>API 状态</h2>
        <p v-if="health.loading">正在连接后端...</p>
        <p v-else-if="health.error" class="error">{{ health.error }}</p>
        <p v-else-if="health.data">
          {{ health.data.status }} · {{ health.data.env }} · {{ health.data.time }}
        </p>
        <p v-else>等待检测。</p>
      </aside>
    </section>

    <section class="section-heading">
      <div>
        <span class="eyebrow">Latest</span>
        <h2>最新文章</h2>
        <p>数据来自 Go API，当前使用内存 repository，后续会替换为 PostgreSQL。</p>
      </div>
      <RouterLink class="button-secondary" to="/archive">查看归档</RouterLink>
    </section>

    <p v-if="posts.loading" class="muted">正在加载文章...</p>
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>

    <section v-else class="content-grid" aria-label="最新文章">
      <article v-for="post in posts.list?.items" :key="post.id" class="post-card">
        <img :src="post.coverImage" :alt="post.title">
        <div class="post-card-body">
          <div class="meta-row">
            <span class="tag">{{ post.category }}</span>
            <span>{{ post.readingTime }} 分钟阅读</span>
          </div>
          <h2>
            <RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink>
          </h2>
          <p>{{ post.summary }}</p>
          <div class="meta-row">
            <span>{{ post.viewCount }} 次阅读</span>
            <span>{{ post.commentCount }} 条评论</span>
            <span>{{ post.likeCount }} 赞</span>
          </div>
        </div>
      </article>
    </section>
  </main>
</template>
