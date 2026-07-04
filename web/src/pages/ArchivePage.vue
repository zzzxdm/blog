<script setup lang="ts">
import { onMounted } from "vue";
import { RouterLink } from "vue-router";

import { usePostsStore } from "../stores/posts";

const posts = usePostsStore();

onMounted(() => {
  void posts.loadList({ page: 1, pageSize: 12 });
});
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <span class="eyebrow">Archive</span>
        <h1>文章归档</h1>
        <p>按发布时间倒序展示文章，后续会继续接入筛选、分页和视图记忆。</p>
      </div>
    </section>

    <p v-if="posts.loading" class="muted">正在加载文章...</p>
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>

    <section v-else class="archive-list" aria-label="文章列表">
      <article v-for="post in posts.list?.items" :key="post.id" class="archive-item">
        <div>
          <div class="meta-row">
            <span class="tag">{{ post.category }}</span>
            <span>{{ new Date(post.publishedAt).toLocaleDateString() }}</span>
            <span>{{ post.readingTime }} 分钟阅读</span>
          </div>
          <h2>
            <RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink>
          </h2>
          <p>{{ post.summary }}</p>
        </div>
        <div class="archive-stats">
          <span>{{ post.viewCount }} 阅读</span>
          <span>{{ post.likeCount }} 赞</span>
          <span>{{ post.commentCount }} 评论</span>
        </div>
      </article>
    </section>
  </main>
</template>
