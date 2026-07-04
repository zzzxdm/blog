<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { RouterLink } from "vue-router";

import { searchPosts, type ListResponse, type Post } from "../shared/api";

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const inputRef = ref<HTMLInputElement | null>(null);
const query = ref("");
const results = ref<ListResponse<Post> | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
let searchRun = 0;

function close() {
  emit("update:open", false);
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    close();
  }
}

watch(
  () => props.open,
  async (open) => {
    if (!open) {
      return;
    }

    await nextTick();
    inputRef.value?.focus();
  }
);

watch(query, (value, _oldValue, onCleanup) => {
  const keyword = value.trim();
  const run = ++searchRun;

  if (!keyword) {
    results.value = null;
    error.value = null;
    loading.value = false;
    return;
  }

  const timer = window.setTimeout(async () => {

    loading.value = true;
    error.value = null;

    try {
      const response = await searchPosts({ q: keyword, page: 1, pageSize: 6 });
      if (run === searchRun) {
        results.value = response;
      }
    } catch (searchError) {
      if (run === searchRun) {
        error.value = searchError instanceof Error ? searchError.message : "搜索失败";
      }
    } finally {
      if (run === searchRun) {
        loading.value = false;
      }
    }
  }, 220);

  onCleanup(() => window.clearTimeout(timer));
});

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
});
</script>

<template>
  <div class="search-overlay" :class="{ open }" @click.self="close">
    <section class="search-dialog" role="dialog" aria-modal="true" aria-label="站内搜索">
      <div class="search-dialog-header">
        <input
          ref="inputRef"
          v-model="query"
          class="input"
          type="search"
          placeholder="搜索文章、专题或标签"
          aria-label="搜索关键词"
        >
        <button class="search-close" type="button" aria-label="关闭搜索" @click="close">×</button>
      </div>
      <div class="search-results">
        <div v-if="!query.trim()" class="search-result">
          <strong>输入关键词开始搜索</strong>
          <p>支持标题、摘要、正文、分类和标签。</p>
        </div>
        <div v-else-if="loading" class="search-result">
          <strong>正在搜索</strong>
          <p>从 PostgreSQL 全文搜索接口获取结果。</p>
        </div>
        <div v-else-if="error" class="search-result">
          <strong>搜索失败</strong>
          <p>{{ error }}</p>
        </div>
        <div v-else-if="results && !results.items.length" class="search-result">
          <strong>没有找到结果</strong>
          <p>换个关键词试试。</p>
        </div>
        <template v-else>
          <RouterLink
            v-for="post in results?.items"
            :key="post.id"
            class="search-result"
            :to="`/posts/${post.slug}`"
            @click="close"
          >
            <div class="meta-row"><span class="tag">{{ post.category }}</span></div>
            <strong>{{ post.title }}</strong>
            <p>{{ post.summary }}</p>
          </RouterLink>
          <RouterLink
            v-if="results && results.total > results.items.length"
            class="search-result"
            :to="{ path: '/archive', query: { q: query.trim() } }"
            @click="close"
          >
            <strong>查看全部 {{ results.total }} 条结果</strong>
            <p>在归档页继续筛选分类、标签和排序。</p>
          </RouterLink>
        </template>
      </div>
    </section>
  </div>
</template>
