<script setup lang="ts">
import { Search } from "@element-plus/icons-vue";
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";

import PaginationControls from "../components/PaginationControls.vue";
import { getTopic, getTopicPosts, getTopics, type Post, type Topic } from "../shared/api";

const route = useRoute();
const loading = ref(false);
const postsLoading = ref(false);
const error = ref("");
const postsError = ref("");
const topics = ref<Topic[]>([]);
const hotTopics = ref<Topic[]>([]);
const selectedTopic = ref<Topic | null>(null);
const currentTopicPosts = ref<Post[]>([]);
const topicTotal = ref(0);
const topicPostTotal = ref(0);
const searchQuery = ref("");
const topicListPage = ref(1);
const topicListPageSize = ref(9);
const topicPostPage = ref(1);
const topicPostPageSize = ref(4);

const currentTopic = computed(() => {
  const topicSlug = stringQuery(route.query.topic);
  if (topicSlug) {
    if (searchQuery.value.trim() && !topics.value.some((topic) => topic.slug === topicSlug)) {
      return topics.value[0] ?? null;
    }
    if (selectedTopic.value?.slug === topicSlug) {
      return selectedTopic.value;
    }
    return topics.value.find((topic) => topic.slug === topicSlug) ?? null;
  }

  return topics.value[0] ?? null;
});
const articleCountText = computed(() => {
  const visiblePostCount = topics.value.reduce((sum, topic) => sum + topic.postCount, 0);
  return visiblePostCount ? `当前页 ${visiblePostCount} 篇关联文章` : "专题文章持续更新";
});
const filterLinks = computed(() => {
  const topic = currentTopic.value;
  if (!topic) {
    return [];
  }

  const categoryLinks = topic.categories.map((item) => ({
    key: `category-${item}`,
    label: item,
    to: { path: "/archive", query: { category: item } }
  }));
  const tagLinks = topic.tags.map((item) => ({
    key: `tag-${item}`,
    label: item,
    to: { path: "/archive", query: { tag: item } }
  }));

  return [...categoryLinks, ...tagLinks].filter((item, index, list) =>
    list.findIndex((candidate) => candidate.label === item.label) === index
  );
});

onMounted(loadInitialTopics);

watch(() => stringQuery(route.query.topic), async (slug, previous) => {
  if (slug !== previous) {
    topicPostPage.value = 1;
  }
  await syncSelectedTopic();
  await loadTopicPosts();
});

watch([topicPostPage, topicPostPageSize], () => {
  if (currentTopic.value) {
    void loadTopicPosts();
  }
});

async function loadInitialTopics() {
  loading.value = true;
  error.value = "";

  try {
    const [response, hotResponse] = await Promise.all([
      getTopics({ page: topicListPage.value, pageSize: topicListPageSize.value, q: searchQuery.value.trim() }),
      getTopics({ page: 1, pageSize: 3, featured: true })
    ]);
    topics.value = response.items;
    topicTotal.value = response.total;
    topicListPage.value = response.page;
    topicListPageSize.value = response.pageSize;
    hotTopics.value = hotResponse.items.length ? hotResponse.items : response.items.slice(0, 3);
    await syncSelectedTopic();
    await loadTopicPosts();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "专题内容加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadTopics() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getTopics({ page: topicListPage.value, pageSize: topicListPageSize.value, q: searchQuery.value.trim() });
    topics.value = response.items;
    topicTotal.value = response.total;
    topicListPage.value = response.page;
    topicListPageSize.value = response.pageSize;
    if (!hotTopics.value.length) {
      hotTopics.value = response.items.slice(0, 3);
    }
    await syncSelectedTopic();
    await loadTopicPosts();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "专题内容加载失败";
  } finally {
    loading.value = false;
  }
}

async function syncSelectedTopic() {
  const topicSlug = stringQuery(route.query.topic);
  if (!topicSlug) {
    selectedTopic.value = null;
    return;
  }

  const localTopic = topics.value.find((topic) => topic.slug === topicSlug) ?? hotTopics.value.find((topic) => topic.slug === topicSlug);
  if (localTopic) {
    selectedTopic.value = localTopic;
    return;
  }

  if (searchQuery.value.trim()) {
    selectedTopic.value = null;
    return;
  }

  if (selectedTopic.value?.slug === topicSlug) {
    return;
  }

  try {
    selectedTopic.value = await getTopic(topicSlug);
  } catch {
    selectedTopic.value = null;
  }
}

async function loadTopicPosts() {
  const topic = currentTopic.value;
  if (!topic) {
    currentTopicPosts.value = [];
    topicPostTotal.value = 0;
    return;
  }

  postsLoading.value = true;
  postsError.value = "";

  try {
    const response = await getTopicPosts(topic.slug, { page: topicPostPage.value, pageSize: topicPostPageSize.value });
    currentTopicPosts.value = response.items;
    topicPostTotal.value = response.total;
    topicPostPage.value = response.page;
    topicPostPageSize.value = response.pageSize;
  } catch (err) {
    postsError.value = err instanceof Error ? err.message : "专题文章加载失败";
  } finally {
    postsLoading.value = false;
  }
}

function topicArticleCount(topic: Topic) {
  return topic.postCount;
}

function topicLatestLabel(topic: Topic) {
  return topic.latestPostAt ? formatDate(topic.latestPostAt) : "暂无更新";
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
  return (topicPostPage.value - 1) * topicPostPageSize.value + index;
}

function topicLink(topic: Topic) {
  return { path: "/topics", query: { topic: topic.slug } };
}

function topicReadingLink(topic: Topic) {
  return { ...topicLink(topic), hash: "#topic-reading" };
}

async function applyTopicSearch() {
  topicListPage.value = 1;
  topicPostPage.value = 1;
  await loadTopics();
}

async function setTopicListPage(page: number) {
  topicListPage.value = page;
  await loadTopics();
}

async function setTopicListPageSize(pageSize: number) {
  topicListPageSize.value = pageSize;
  topicListPage.value = 1;
  await loadTopics();
}

async function setTopicPostPage(page: number) {
  topicPostPage.value = page;
}

async function setTopicPostPageSize(pageSize: number) {
  topicPostPageSize.value = pageSize;
  topicPostPage.value = 1;
}

function topicImage(topic: Topic) {
  return topic.coverImage || "https://images.unsplash.com/photo-1498050108023-c5249f4df0856?auto=format&fit=crop&w=900&q=80";
}

function filterTone(index: number) {
  if (index % 3 === 1) return "rust";
  if (index % 3 === 2) return "amber";
  return "";
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
          <span>{{ topicTotal }} 个重点专题</span>
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
        <p v-if="loading" class="muted">正在加载专题...</p>
        <ol v-else-if="hotTopics.length" class="rank-list">
          <li v-for="(topic, index) in hotTopics.slice(0, 3)" :key="topic.slug">
            <span class="rank-number">{{ index + 1 }}</span>
            <RouterLink :to="topicReadingLink(topic)">
              <strong>{{ topic.title }}</strong>
              <span>{{ topicArticleCount(topic) }} 篇文章 · 最近 {{ topicLatestLabel(topic) }}</span>
            </RouterLink>
          </li>
        </ol>
        <p v-else class="muted">暂无专题。</p>
      </aside>
    </section>

    <p v-if="error" class="error">{{ error }}</p>

    <section class="section-heading">
      <div>
        <h2>全部专题</h2>
        <p>按长期维护的内容方向组织，适合系统阅读。</p>
      </div>
    </section>

    <form class="topic-search-toolbar" @submit.prevent="applyTopicSearch">
      <input v-model="searchQuery" class="input" type="search" placeholder="搜索专题标题、摘要、分类、标签" aria-label="搜索专题">
      <button class="button" type="submit" :disabled="loading">
        <Search class="button-icon" aria-hidden="true" />
        搜索
      </button>
    </form>

    <section v-if="topics.length" class="compact-grid" aria-label="专题列表">
      <article v-for="topic in topics" :key="topic.slug" class="topic-card">
        <img :src="topicImage(topic)" :alt="topic.imageAlt || topic.title">
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
    <p v-else-if="!loading" class="muted">暂无专题。</p>
    <PaginationControls
      v-if="topicTotal > 0"
      :page="topicListPage"
      :page-size="topicListPageSize"
      :total="topicTotal"
      :loading="loading"
      item-label="个专题"
      show-page-size
      :page-size-options="[6, 9, 12, 24]"
      @update:page="setTopicListPage"
      @update:page-size="setTopicListPageSize"
    />

    <section v-if="currentTopic" id="topic-reading" class="article-layout topic-reading-layout">
      <div>
        <div class="section-heading">
          <div>
            <h2>{{ currentTopic.title }}</h2>
            <p>当前重点专题，按发布时间顺序排列。</p>
          </div>
        </div>

        <p v-if="postsLoading" class="muted">正在加载专题文章...</p>
        <p v-else-if="postsError" class="error">{{ postsError }}</p>
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
            :page="topicPostPage"
            :page-size="topicPostPageSize"
            :total="topicPostTotal"
            :loading="postsLoading"
            item-label="篇专题文章"
            show-page-size
            :page-size-options="[4, 8, 12, 20]"
            @update:page="setTopicPostPage"
            @update:page-size="setTopicPostPageSize"
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
            <RouterLink
              v-for="(item, index) in filterLinks"
              :key="item.key"
              class="tag"
              :class="filterTone(index)"
              :to="item.to"
            >
              {{ item.label }}
            </RouterLink>
          </div>
        </section>
      </aside>
    </section>
  </main>
</template>
