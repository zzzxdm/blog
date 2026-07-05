<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";

import { usePostsStore } from "../stores/posts";
import {
  getCategories,
  getSiteSettings,
  getTags,
  type Category,
  type Post,
  type SiteSettings,
  type Tag
} from "../shared/api";

type TopicLink = {
  key: string;
  label: string;
  count: number;
  tone: "" | "rust" | "amber";
  to: {
    path: string;
    query: Record<string, string>;
  };
};

const posts = usePostsStore();
const siteSettings = ref<SiteSettings | null>(null);
const categories = ref<Category[]>([]);
const tags = ref<Tag[]>([]);
const topicLinkLimit = 6;

const allPosts = computed(() => posts.list?.items ?? []);
const featuredPosts = computed(() => allPosts.value.slice(0, 4));
const featureIndex = ref(0);
const featurePost = computed(() => featuredPosts.value[featureIndex.value] ?? featuredPosts.value[0] ?? null);
const weeklyPosts = computed(() => allPosts.value.filter((_, index) => index !== featureIndex.value).slice(0, 3));
const latestPosts = computed(() => allPosts.value.slice(1, 5));
const homepageLayout = computed(() => siteSettings.value?.homepageLayout || "精选文章 + 最新列表");
const topicFirstLayout = computed(() => homepageLayout.value === "专题优先");
const minimalLayout = computed(() => homepageLayout.value === "极简文章流");
const streamPosts = computed(() => minimalLayout.value ? allPosts.value : latestPosts.value);
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
  const categoryNames = categories.value
    .filter((item) => item.name)
    .sort((left, right) => right.postCount - left.postCount || left.sortOrder - right.sortOrder)
    .slice(0, 3)
    .map((item) => item.name);

  if (categoryNames.length) {
    return categoryNames.join("、");
  }

  const names = [...new Set(allPosts.value.map((post) => post.category).filter(Boolean))];
  return names.slice(0, 3).join("、") || "持续更新";
});
const totalCommentCount = computed(() => allPosts.value.reduce((sum, post) => sum + post.commentCount, 0));
const topicLinks = computed<TopicLink[]>(() => {
  const taxonomyLinks = [
    ...categories.value
      .filter((item) => item.name)
      .map((item) => ({
        key: `category-${item.id}`,
        label: item.name,
        count: item.postCount,
        to: { path: "/archive", query: { category: item.name } }
      })),
    ...tags.value
      .filter((item) => item.name)
      .map((item) => ({
        key: `tag-${item.id}`,
        label: item.name,
        count: item.postCount,
        to: { path: "/archive", query: { tag: item.name } }
      }))
  ];

  const sources = taxonomyLinks.length ? taxonomyLinks : postTopicLinks();
  return sources
    .sort((left, right) => right.count - left.count || left.label.localeCompare(right.label, "zh-CN"))
    .slice(0, topicLinkLimit)
    .map((item, index) => ({ ...item, tone: topicTone(index) }));
});
const featuredTopicCount = computed(() => categories.value.length + tags.value.length || topicLinks.value.length);

onMounted(() => {
  void posts.loadList({ page: 1, pageSize: 12 });
  void loadSiteSettings();
  void loadTaxonomies();
  startFeatureCarousel();
});

onUnmounted(stopFeatureCarousel);

watch(() => featuredPosts.value.length, (length) => {
  if (featureIndex.value >= length) {
    featureIndex.value = 0;
  }
  restartFeatureCarousel();
});

let featureTimer: number | undefined;

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

async function loadTaxonomies() {
  try {
    const [categoryResult, tagResult] = await Promise.all([getCategories(), getTags()]);
    categories.value = categoryResult.items;
    tags.value = tagResult.items;
  } catch {
    categories.value = [];
    tags.value = [];
  }
}

function postTopicLinks() {
  const grouped = new Map<string, Omit<TopicLink, "tone">>();

  allPosts.value.forEach((post) => {
    addTopic(grouped, `category:${post.category}`, post.category, { category: post.category });

    post.tags.forEach((tag) => {
      addTopic(grouped, `tag:${tag}`, tag, { tag });
    });
  });

  return [...grouped.values()];
}

function addTopic(
  grouped: Map<string, Omit<TopicLink, "tone">>,
  key: string,
  label: string,
  query: Record<string, string>
) {
  if (!label) {
    return;
  }

  const current = grouped.get(key);
  if (current) {
    current.count += 1;
    return;
  }

  grouped.set(key, {
    key,
    label,
    count: 1,
    to: { path: "/archive", query }
  });
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

function topicTone(index: number): "" | "rust" | "amber" {
  if (index % 3 === 1) {
    return "rust";
  }

  if (index % 3 === 2) {
    return "amber";
  }

  return "";
}

function startFeatureCarousel() {
  if (featureTimer || featuredPosts.value.length <= 1) {
    return;
  }

  featureTimer = window.setInterval(() => {
    nextFeature(false);
  }, 6000);
}

function stopFeatureCarousel() {
  if (featureTimer) {
    window.clearInterval(featureTimer);
    featureTimer = undefined;
  }
}

function restartFeatureCarousel() {
  stopFeatureCarousel();
  startFeatureCarousel();
}

function showFeature(index: number, resetTimer = true) {
  const total = featuredPosts.value.length;
  if (total <= 1) {
    featureIndex.value = 0;
    return;
  }

  featureIndex.value = (index + total) % total;
  if (resetTimer) {
    restartFeatureCarousel();
  }
}

function nextFeature(resetTimer = true) {
  showFeature(featureIndex.value + 1, resetTimer);
}

function previousFeature() {
  showFeature(featureIndex.value - 1);
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
      <section v-if="topicFirstLayout && topicLinks.length" class="topic-strip" aria-label="专题导航">
        <div class="section-heading">
          <div>
            <h2>专题导航</h2>
            <p>先从主题进入，再选择具体文章。</p>
          </div>
          <RouterLink class="button-secondary" to="/topics">查看全部专题</RouterLink>
        </div>

        <div class="topic-strip-grid">
          <RouterLink
            v-for="topic in topicLinks"
            :key="topic.key"
            class="topic-strip-card"
            :class="topic.tone"
            :to="topic.to"
          >
            <strong>{{ topic.label }}</strong>
            <span>{{ topic.count }} 篇内容</span>
          </RouterLink>
        </div>
      </section>

      <section v-if="!minimalLayout && featurePost" class="hero-grid" aria-label="精选内容">
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
            <div class="feature-actions">
              <RouterLink class="button" :to="`/posts/${featurePost.slug}`">阅读全文</RouterLink>
              <div v-if="featuredPosts.length > 1" class="feature-controls" aria-label="精选内容轮播">
                <button class="icon-button" type="button" aria-label="上一篇" @click="previousFeature">‹</button>
                <div class="feature-dots">
                  <button
                    v-for="(post, index) in featuredPosts"
                    :key="post.id"
                    type="button"
                    :class="{ active: index === featureIndex }"
                    :aria-label="`切换到第 ${index + 1} 篇`"
                    :aria-current="index === featureIndex ? 'true' : undefined"
                    @click="showFeature(index)"
                  ></button>
                </div>
                <button class="icon-button" type="button" aria-label="下一篇" @click="nextFeature()">›</button>
              </div>
            </div>
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

      <section class="article-layout home-latest-layout" :class="{ 'home-minimal-layout': minimalLayout }" :aria-label="minimalLayout ? '文章流' : '最新文章'">
        <div>
          <div class="section-heading">
            <div>
              <h2>{{ minimalLayout ? "文章流" : "最新文章" }}</h2>
              <p>{{ minimalLayout ? "保留最少干扰，按发布时间浏览全部文章。" : "按发布时间排序，适合快速浏览最近更新。" }}</p>
            </div>
            <RouterLink class="button-secondary" to="/archive">查看归档</RouterLink>
          </div>

          <div class="article-list">
            <article v-for="(post, index) in streamPosts" :key="post.id" class="article-card">
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

        <aside v-if="!minimalLayout" class="sidebar">
          <section v-if="!topicFirstLayout" class="panel">
            <div class="panel-title">
              <h2>专题</h2>
            </div>
            <div class="tag-cloud">
              <RouterLink
                v-for="topic in topicLinks"
                :key="topic.key"
                class="tag"
                :class="topic.tone"
                :to="topic.to"
              >
                {{ topic.label }}
              </RouterLink>
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
