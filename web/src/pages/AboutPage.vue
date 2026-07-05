<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";

import { getCategories, getSiteSettings, getSiteStats, type Category, type SiteSettings, type SiteStats } from "../shared/api";

const settings = ref<SiteSettings | null>(null);
const stats = ref<SiteStats | null>(null);
const categories = ref<Category[]>([]);
const loading = ref(true);

const siteName = computed(() => settings.value?.siteName.trim() || "云间笔记");
const tagline = computed(() => settings.value?.siteDescription.trim() || "记录内容产品、工程实践和长期写作。");

onMounted(async () => {
  try {
    const [siteSettings, siteStats, categoryList] = await Promise.all([
      getSiteSettings(),
      getSiteStats(),
      getCategories()
    ]);
    settings.value = siteSettings;
    stats.value = siteStats;
    categories.value = categoryList.items;
  } finally {
    loading.value = false;
  }
});

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>关于{{ siteName }}</h1>
        <p>{{ tagline }}</p>
      </div>
      <div class="meta-row">
        <span>{{ stats ? `${formatNumber(stats.postCount)} 篇文章` : "内容持续更新" }}</span>
        <span>{{ stats ? `${formatNumber(stats.wordCount)} 字` : "长期维护" }}</span>
      </div>
    </section>

    <section class="content-shell">
      <article class="article-body">
        <p>{{ siteName }} 关注内容系统、前端工程、后端架构、搜索、SEO 和写作流程。这里的文章以可复用经验为主，优先记录长期仍值得回看的设计取舍。</p>
        <p>站点支持归档、专题、标签、评论、收藏、投稿审核、站内信和后台运营能力。公开内容可以直接阅读，登录后可以参与评论、收藏文章并提交投稿。</p>
      </article>
    </section>

    <section class="compact-grid archive-view" aria-label="站点概览">
      <article class="metric-card">
        <strong>{{ loading || !stats ? "-" : formatNumber(stats.postCount) }}</strong>
        <span>公开文章</span>
      </article>
      <article class="metric-card">
        <strong>{{ loading || !stats ? "-" : formatNumber(stats.viewCount) }}</strong>
        <span>累计阅读</span>
      </article>
      <article class="metric-card">
        <strong>{{ categories.length }}</strong>
        <span>内容分类</span>
      </article>
    </section>

    <section class="archive-list archive-view" aria-label="主要分类">
      <article v-for="category in categories" :key="category.id" class="archive-list-item">
        <div class="archive-list-main">
          <div class="meta-row"><span class="tag">{{ category.postCount }} 篇</span></div>
          <h3><RouterLink :to="`/archive?category=${encodeURIComponent(category.name)}`">{{ category.name }}</RouterLink></h3>
          <p>{{ category.description || "围绕这个主题整理长期内容。" }}</p>
        </div>
        <div class="archive-list-side">
          <RouterLink :to="`/archive?category=${encodeURIComponent(category.name)}`">查看</RouterLink>
        </div>
      </article>
    </section>
  </main>
</template>
