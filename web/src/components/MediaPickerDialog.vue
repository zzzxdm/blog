<script setup lang="ts">
import { ref, watch } from "vue";

import LoadingState from "./LoadingState.vue";
import PaginationControls from "./PaginationControls.vue";
import { getAdminMedia, type MediaAsset } from "../shared/api";

const props = withDefaults(defineProps<{
  open: boolean;
  title?: string;
  type?: string;
}>(), {
  title: "选择媒体",
  type: "image"
});

const emit = defineEmits<{
  close: [];
  select: [asset: MediaAsset];
}>();

const items = ref<MediaAsset[]>([]);
const selectedId = ref("");
const loading = ref(false);
const error = ref("");
const searchQuery = ref("");
const page = ref(1);
const pageSize = ref(12);
const total = ref(0);

watch(() => props.open, (open) => {
  if (open) {
    void load();
  }
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminMedia({
      q: searchQuery.value,
      type: props.type,
      page: page.value,
      pageSize: pageSize.value
    });
    items.value = response.items;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
    if (!items.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = items.value[0]?.id || "";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "媒体资源加载失败";
  } finally {
    loading.value = false;
  }
}

async function applySearch() {
  page.value = 1;
  await load();
}

async function setPage(value: number) {
  page.value = value;
  await load();
}

function choose(asset: MediaAsset) {
  selectedId.value = asset.id;
}

function confirmSelection() {
  const asset = items.value.find((item) => item.id === selectedId.value);
  if (asset) {
    emit("select", asset);
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="confirm-dialog">
      <section v-if="open" class="media-picker-overlay" role="dialog" aria-modal="true" :aria-label="title" @click.self="emit('close')">
        <div class="media-picker-panel">
          <header class="media-picker-header">
            <h2>{{ title }}</h2>
            <button type="button" aria-label="关闭" @click="emit('close')">×</button>
          </header>

          <form class="media-picker-toolbar" @submit.prevent="applySearch">
            <input v-model="searchQuery" class="input" type="search" placeholder="搜索文件名、alt 文本、分类" aria-label="搜索媒体">
            <button class="button-secondary" type="submit">搜索</button>
          </form>

          <LoadingState v-if="loading" variant="card" text="正在加载媒体资源..." :rows="3" />
          <p v-else-if="error" class="error">{{ error }}</p>
          <section v-else class="media-picker-grid" aria-label="可选媒体">
            <button
              v-for="item in items"
              :key="item.id"
              class="media-picker-card"
              :class="{ selected: selectedId === item.id }"
              type="button"
              @click="choose(item)"
            >
              <img :src="item.url" :alt="item.alt || item.fileName">
              <span>{{ item.fileName }}</span>
            </button>
            <p v-if="items.length === 0" class="muted">没有匹配的媒体资源。</p>
          </section>

          <PaginationControls
            v-if="!loading && !error"
            :page="page"
            :page-size="pageSize"
            :total="total"
            :loading="loading"
            item-label="个资源"
            @update:page="setPage"
          />

          <footer class="media-picker-actions">
            <button class="button-secondary" type="button" @click="emit('close')">取消</button>
            <button class="button" type="button" :disabled="!selectedId" @click="confirmSelection">使用选中图片</button>
          </footer>
        </div>
      </section>
    </Transition>
  </Teleport>
</template>
