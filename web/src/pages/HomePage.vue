<script setup lang="ts">
import { onMounted } from "vue";

import { useHealthStore } from "../stores/health";

const health = useHealthStore();

onMounted(() => {
  void health.load();
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

    <section class="content-grid" aria-label="开发入口">
      <article class="panel">
        <h2>前台</h2>
        <p>首页、归档、专题和文章详情会从当前骨架继续接入真实接口。</p>
      </article>
      <article class="panel">
        <h2>后台</h2>
        <p>文章管理、媒体库、评论审核和投稿审核会按模块逐步实现。</p>
      </article>
      <article class="panel">
        <h2>搜索</h2>
        <p>搜索使用 PostgreSQL 全文搜索，不引入专用搜索中间件。</p>
      </article>
    </section>
  </main>
</template>
