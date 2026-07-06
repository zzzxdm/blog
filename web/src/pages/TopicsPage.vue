<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";

import PaginationControls from "../components/PaginationControls.vue";
import { getPosts, type Post } from "../shared/api";

type TopicTone = "" | "rust" | "amber";

interface Topic {
  slug: string;
  title: string;
  summary: string;
  image: string;
  imageAlt: string;
  tone: TopicTone;
  tags: string[];
  categories: string[];
}

const route = useRoute();
const loading = ref(false);
const error = ref("");
const posts = ref<Post[]>([]);
const total = ref(0);
const topicPage = ref(1);
const topicPageSize = ref(4);

const topics: Topic[] = [
  {
    slug: "blog-system",
    title: "现代化博客系统",
    summary: "从产品功能、技术架构、用户系统、评论、搜索和后台管理完整设计一个博客系统。",
    image: "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=900&q=80",
    imageAlt: "代码编辑器和开发设备",
    tone: "",
    tags: ["博客系统", "架构", "内容治理", "评论"],
    categories: ["工程实践", "产品设计", "用户系统", "内容治理"]
  },
  {
    slug: "vue3-content",
    title: "Vue3 内容站",
    summary: "路由、状态管理、接口缓存、SEO meta、图片优化和部署策略。",
    image: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=900&q=80",
    imageAlt: "代码编辑器中的程序文件",
    tone: "rust",
    tags: ["Vue3", "SEO", "缓存"],
    categories: ["Vue3"]
  },
  {
    slug: "writing-workflow",
    title: "写作工作流",
    summary: "草稿、版本历史、编辑器、发布审批和长期内容维护。",
    image: "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=900&q=80",
    imageAlt: "笔记本和写作草稿",
    tone: "amber",
    tags: ["工作流", "写作工作流", "Markdown"],
    categories: ["写作工作流"]
  },
  {
    slug: "resource-list",
    title: "资源清单",
    summary: "把工具、部署、数据库和内容运营资料整理成可持续更新的阅读路线。",
    image: "https://images.unsplash.com/photo-1484480974693-6ca0a78fb36b?auto=format&fit=crop&w=900&q=80",
    imageAlt: "桌面上的计划清单和电脑",
    tone: "",
    tags: ["PostgreSQL", "Redis", "全文搜索", "SEO"],
    categories: ["架构", "运营"]
  }
];

const currentTopic = computed(() => {
  const topicSlug = stringQuery(route.query.topic);
  return topics.find((topic) => topic.slug === topicSlug) ?? topics[0];
});

const allCurrentTopicPosts = computed(() => topicPosts(currentTopic.value));
const currentTopicPosts = computed(() => {
  const start = (topicPage.value - 1) * topicPageSize.value;
  return allCurrentTopicPosts.value.slice(start, start + topicPageSize.value);
});

const articleCountText = computed(() => `${total.value || posts.value.length} 篇文章`);

watch(() => currentTopic.value.slug, () => {
  topicPage.value = 1;
});

onMounted(async () => {
  loading.value = true;
  error.value = "";

  try {
    const response = await getPosts({ page: 1, pageSize: 50 });
    posts.value = response.items;
    total.value = response.total;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "专题内容加载失败";
  } finally {
    loading.value = false;
  }
});

function topicPosts(topic: Topic) {
  return posts.value.filter((post) => matchesTopic(post, topic));
}

function matchesTopic(post: Post, topic: Topic) {
  const postTags = post.tags.map((tag) => tag.toLowerCase());
  const topicTags = topic.tags.map((tag) => tag.toLowerCase());

  return (
    topic.categories.includes(post.category) ||
    topicTags.some((tag) => postTags.includes(tag)) ||
    topic.tags.some((tag) => post.title.includes(tag) || post.summary.includes(tag))
  );
}

function topicArticleCount(topic: Topic) {
  return topicPosts(topic).length;
}

function topicLatestLabel(topic: Topic) {
  const latest = topicPosts(topic)
    .map((post) => post.publishedAt)
    .sort((left, right) => new Date(right).getTime() - new Date(left).getTime())[0];

  return latest ? formatDate(latest) : "暂无更新";
}

function topicStatus(index: number) {
  if (index === 0) {
    return { className: "published", label: "推荐阅读" };
  }

  if (index === 1) {
    return { className: "review", label: "进阶阅读" };
  }

  return { className: "draft", label: "延伸阅读" };
}

function topicPostIndex(index: number) {
  return (topicPage.value - 1) * topicPageSize.value + index;
}

function topicLink(topic: Topic) {
  return { path: "/topics", query: { topic: topic.slug } };
}

function topicReadingLink(topic: Topic) {
  return { ...topicLink(topic), hash: "#topic-reading" };
}

function setTopicPage(page: number) {
  topicPage.value = page;
}

function setTopicPageSize(pageSize: number) {
  topicPageSize.value = pageSize;
  topicPage.value = 1;
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

function stringQuery(value: unknown) {
  return Array.isArray(value) ? String(value[0] ?? "") : String(value ?? "");
}
</script>

<template>
  <main class="page">
    <section class="topic-hero">
      <div class="topic-lead">
        <div class="meta-row">
          <span class="tag amber">专题</span>
          <span>{{ topics.length }} 个重点专题</span>
          <span>{{ articleCountText }}</span>
        </div>
        <h1>围绕一个问题持续写，而不是只发布零散文章。</h1>
        <p>专题用于组织系列内容，比如博客系统设计、Vue3 内容站、内容运营和写作工作流。读者可以按主题连续阅读。</p>
        <RouterLink class="button" to="/archive">查看全部文章</RouterLink>
      </div>

      <aside class="panel">
        <div class="panel-title">
          <h2>热门专题</h2>
        </div>
        <ol class="rank-list">
          <li v-for="(topic, index) in topics.slice(0, 3)" :key="topic.slug">
            <span class="rank-number">{{ index + 1 }}</span>
            <RouterLink :to="topicReadingLink(topic)">
              <strong>{{ topic.title }}</strong>
              <span>{{ topicArticleCount(topic) }} 篇文章 · 最近 {{ topicLatestLabel(topic) }}</span>
            </RouterLink>
          </li>
        </ol>
      </aside>
    </section>

    <section class="section-heading">
      <div>
        <h2>全部专题</h2>
        <p>按长期维护的内容方向组织，适合系统阅读。</p>
      </div>
    </section>

    <section class="compact-grid" aria-label="专题列表">
      <article v-for="topic in topics" :key="topic.slug" class="topic-card">
        <img :src="topic.image" :alt="topic.imageAlt">
        <div class="topic-card-body">
          <div class="meta-row">
            <span class="tag" :class="topic.tone">{{ topicArticleCount(topic) }} 篇文章</span>
            <span>最近更新：{{ topicLatestLabel(topic) }}</span>
          </div>
          <h3>
            <RouterLink :to="topicLink(topic)">{{ topic.title }}</RouterLink>
          </h3>
          <p>{{ topic.summary }}</p>
          <RouterLink class="button-secondary" :to="topicReadingLink(topic)">继续阅读</RouterLink>
        </div>
      </article>
    </section>

    <section id="topic-reading" class="article-layout">
      <div>
        <div class="section-heading">
          <div>
            <h2>{{ currentTopic.title }}</h2>
            <p>当前重点专题，按推荐阅读顺序排列。</p>
          </div>
        </div>

        <p v-if="loading" class="muted">正在加载专题文章...</p>
        <p v-else-if="error" class="error">{{ error }}</p>
        <div v-else-if="currentTopicPosts.length" class="topic-list">
          <article v-for="(post, index) in currentTopicPosts" :key="post.id" class="topic-list-item">
            <img :src="post.coverImage" :alt="post.title">
            <div>
              <strong>
                <RouterLink :to="`/posts/${post.slug}`">{{ post.title }}</RouterLink>
              </strong>
              <div class="meta-row">
                <span>{{ post.category }}</span>
                <span>{{ post.readingTime }} 分钟阅读</span>
                <span>{{ formatDate(post.publishedAt) }}</span>
              </div>
            </div>
            <span class="status" :class="topicStatus(topicPostIndex(index)).className">{{ topicStatus(topicPostIndex(index)).label }}</span>
            <RouterLink class="button-secondary" :to="`/posts/${post.slug}`">继续阅读</RouterLink>
          </article>
          <PaginationControls
            :page="topicPage"
            :page-size="topicPageSize"
            :total="allCurrentTopicPosts.length"
            item-label="篇专题文章"
            show-page-size
            :page-size-options="[4, 8, 12, 20]"
            @update:page="setTopicPage"
            @update:page-size="setTopicPageSize"
          />
        </div>
        <p v-else class="muted">这个专题暂无文章。</p>
      </div>

      <aside class="sidebar">
        <section class="panel">
          <div class="panel-title">
            <h2>专题筛选</h2>
          </div>
          <div class="tag-cloud">
            <RouterLink class="tag" to="/archive?category=工程实践">工程实践</RouterLink>
            <RouterLink class="tag rust" to="/archive?category=架构">架构</RouterLink>
            <RouterLink class="tag amber" to="/archive?tag=写作工作流">写作</RouterLink>
            <RouterLink class="tag" to="/archive?tag=SEO">SEO</RouterLink>
            <RouterLink class="tag rust" to="/archive?category=产品设计">产品设计</RouterLink>
            <RouterLink class="tag amber" to="/archive?category=运营">运营</RouterLink>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
