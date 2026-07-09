<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";

import MarkdownPreview from "../components/MarkdownPreview.vue";
import MarkdownThemeSwitcher from "../components/MarkdownThemeSwitcher.vue";
import { getPreviewPost, type AdminPost } from "../shared/api";
import { formatDateTime } from "../shared/datetime";
import { useMarkdownPreviewTheme } from "../shared/markdownPreview";

const route = useRoute();
const post = ref<AdminPost | null>(null);
const loading = ref(false);
const error = ref("");
const { selectedPreviewTheme, selectedCodeTheme } = useMarkdownPreviewTheme();

const token = computed(() => String(route.params.token ?? ""));
const avatarText = computed(() => post.value?.authorName.slice(0, 1) || "预");

watch(token, () => void load(), { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";

  try {
    post.value = await getPreviewPost(token.value);
  } catch (err) {
    post.value = null;
    error.value = err instanceof Error ? err.message : "预览链接无效或已过期";
  } finally {
    loading.value = false;
  }
}

function formatDate(value: string) {
  return formatDateTime(value);
}
</script>

<template>
  <main class="article-shell">
    <p v-if="loading" class="muted">正在加载预览...</p>
    <section v-else-if="error" class="content-shell">
      <article class="article-body">
        <h1>预览不可用</h1>
        <p>{{ error }}</p>
        <p><RouterLink to="/">返回首页</RouterLink></p>
      </article>
    </section>

    <template v-else-if="post">
      <article>
        <header class="article-hero">
          <div class="article-breadcrumb-row">
            <RouterLink class="button-secondary" to="/admin/posts">返回后台</RouterLink>
            <nav class="breadcrumb" aria-label="当前位置">
              <RouterLink to="/">首页</RouterLink>
              <span class="breadcrumb-separator">/</span>
              <span>预览</span>
            </nav>
          </div>
          <div class="meta-row">
            <span class="tag rust">预览</span>
            <span>{{ post.status }}</span>
            <span>{{ formatDate(post.updatedAt) }}</span>
          </div>
          <h1>{{ post.title || "未命名文章" }}</h1>
          <p class="dek">{{ post.summary || "暂无摘要" }}</p>
          <div class="author-row">
            <span class="avatar">{{ avatarText }}</span>
            <div>
              <strong>{{ post.authorName }}</strong>
              <div class="meta-row">
                <span>{{ post.category }}</span>
                <span>{{ post.readingTime }} 分钟阅读</span>
                <span>版本 {{ post.version }}</span>
              </div>
            </div>
          </div>
        </header>

        <figure v-if="post.coverImage" class="article-cover">
          <img :src="post.coverImage" :alt="post.title">
        </figure>

        <MarkdownThemeSwitcher v-model:preview-theme="selectedPreviewTheme" v-model:code-theme="selectedCodeTheme" />

        <section class="article-markdown">
          <MarkdownPreview
            :content="post.content"
            :preview-id="`preview-${post.id}`"
            :preview-theme="selectedPreviewTheme"
            :code-theme="selectedCodeTheme"
          />
        </section>
      </article>
    </template>
  </main>
</template>
