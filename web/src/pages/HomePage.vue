<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";

import { usePostsStore } from "../stores/posts";
import { getSiteSettings, type Post, type SiteSettings } from "../shared/api";

const posts = usePostsStore();
const siteSettings = ref<SiteSettings | null>(null);
const featuredTopicCount = 4;

const allPosts = computed(() => posts.list?.items ?? []);
const featurePost = computed(() => allPosts.value[0] ?? null);
const weeklyPosts = computed(() => allPosts.value.slice(1, 4));
const latestPosts = computed(() => allPosts.value.slice(1, 5));
const siteDescription = computed(() => siteSettings.value?.siteDescription || "技术、产品、工程实践和长期写作的沉淀。");
const submissionsEnabled = computed(() => siteSettings.value?.submissionsEnabled ?? true);
const submissionGuide = computed(() => siteSettings.value?.submissionGuide || "登录用户可以提交原创文章，审核通过后发布到站点。");
const monthlyPostCount = computed(() => {
  const now = new Date();
  const count = allPosts.value.filter((post) => {
    const publishedAt = new Date(post.publishedAt);
    return publishedAt.getFullYear() === now.getFullYear() && publishedAt.getMonth() === now.getMonth();
  }).length;

  return count || Math.min(allPosts.value.length, posts.list?.total ?? 0);
});
const categorySummary = computed(() => {
  const names = [...new Set(allPosts.value.map((post) => post.category).filter(Boolean))];
  return names.slice(0, 3).join("、") || "持续更新";
});
const totalCommentCount = computed(() => allPosts.value.reduce((sum, post) => sum + post.commentCount, 0));

onMounted(() => {
  void posts.loadList({ page: 1, pageSize: 12 });
  void loadSiteSettings();
});

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}

function tagTone(post: Post, index = 0) {
  if (post.category === "架构" || index % 3 === 1) {
    return "rust";
  }

  if (post.category === "运营" || index % 3 === 2) {
    return "amber";
  }

  return "";
}
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>今天值得读</h1>
        <p>{{ siteDescription }}</p>
      </div>
      <div class="meta-row">
        <span>{{ posts.list?.total ?? 0 }} 篇文章</span>
        <span>{{ featuredTopicCount }} 个专题</span>
        <span>每周更新</span>
      </div>
    </section>

    <p v-if="posts.loading" class="muted">正在加载精选内容...</p>
    <p v-else-if="posts.error" class="error">{{ posts.error }}</p>

    <template v-else>
      <section v-if="featurePost" class="hero-grid" aria-label="精选内容">
        <article class="feature">
          <img :src="featurePost.coverImage" :alt="featurePost.title">
          <div class="feature-content">
            <div class="meta-row">
              <span class="tag">{{ featurePost.category }}</span>
              <span>{{ featurePost.readingTime }} 分钟阅读</span>
              <span>{{ formatDate(featurePost.publishedAt) }}</span>
            </div>
            <h1>
              <RouterLink :to="`/posts/${featurePost.slug}`">{{ featurePost.title }}</RouterLink>
            </h1>
            <p>{{ featurePost.summary }}</p>
            <RouterLink class="button" :to="`/posts/${featurePost.slug}`">阅读全文</RouterLink>
          </div>
        </article>

        <aside class="side-stack" aria-label="本周精选">
          <section class="panel">
            <div class="panel-title">
              <h2>本周精选</h2>
              <RouterLink class="tag amber" to="/archive">更多</RouterLink>
            </div>
            <ol class="rank-list">
              <li v-for="(post, index) in weeklyPosts" :key="post.id">
                <span class="rank-number">{{ index + 1 }}</span>
                <RouterLink :to="`/posts/${post.slug}`">
                  <strong>{{ post.title }}</strong>
                  <span>{{ post.category }} · {{ post.readingTime }} 分钟</span>
                </RouterLink>
              </li>
            </ol>
          </section>
        </aside>
      </section>

      <section class="article-layout home-latest-layout" aria-label="最新文章">
        <div>
          <div class="section-heading">
            <div>
              <h2>最新文章</h2>
              <p>按发布时间排序，适合快速浏览最近更新。</p>
            </div>
            <RouterLink class="button-secondary" to="/archive">查看归档</RouterLink>
          </div>

          <div class="article-list">
            <article v-for="(post, index) in latestPosts" :key="post.id" class="article-card">
              <img :src="post.coverImage" :alt="post.title">
              <div class="article-card-body">
                <div class="meta-row">
                  <span class="tag" :class="tagTone(post, index)">{{ post.category }}</span>
                  <span>{{ post.readingTime }} 分钟阅读</span>
                </div>
                <h3>
                  <RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink>
                </h3>
                <p>{{ post.summary }}</p>
                <div class="meta-row">
                  <span>{{ formatDate(post.publishedAt) }}</span>
                  <span>{{ formatNumber(post.viewCount) }} 次阅读</span>
                  <span>{{ post.commentCount }} 条评论</span>
                </div>
              </div>
            </article>
          </div>
        </div>

        <aside class="sidebar">
          <section class="panel">
            <div class="panel-title">
              <h2>专题</h2>
            </div>
            <div class="tag-cloud">
              <RouterLink class="tag" to="/topics?topic=vue3-content">Vue3</RouterLink>
              <RouterLink class="tag rust" to="/topics?topic=blog-system">系统架构</RouterLink>
              <RouterLink class="tag amber" to="/topics?topic=writing-workflow">写作工作流</RouterLink>
              <RouterLink class="tag" to="/archive?tag=SEO">SEO</RouterLink>
              <RouterLink class="tag rust" to="/topics?topic=resource-list">数据库</RouterLink>
              <RouterLink class="tag amber" to="/archive?category=运营">内容运营</RouterLink>
            </div>
          </section>

          <section v-if="submissionsEnabled" class="panel">
            <div class="panel-title">
              <h2>开放投稿</h2>
              <span class="tag">审核制</span>
            </div>
            <p style="margin: 0 0 14px; color: var(--muted);">{{ submissionGuide }}</p>
            <RouterLink class="button" to="/submit">开始投稿</RouterLink>
          </section>

          <section class="panel">
            <div class="panel-title">
              <h2>站点状态</h2>
            </div>
            <ul class="link-list">
              <li>
                <strong>本月更新</strong>
                <span>{{ monthlyPostCount }} 篇文章 · {{ featuredTopicCount }} 个专题</span>
              </li>
              <li>
                <strong>热门分类</strong>
                <span>{{ categorySummary }}</span>
              </li>
              <li>
                <strong>读者反馈</strong>
                <span>{{ totalCommentCount }} 条评论互动</span>
              </li>
            </ul>
          </section>
        </aside>
      </section>
    </template>
  </main>
</template>
