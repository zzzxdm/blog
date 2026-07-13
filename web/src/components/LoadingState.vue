<script setup lang="ts">
import { computed } from "vue";
import { ElIcon, ElSkeleton, ElSkeletonItem } from "element-plus";
import { Loading as LoadingIcon } from "@element-plus/icons-vue";

import "element-plus/es/components/icon/style/css";
import "element-plus/es/components/skeleton/style/css";
import "element-plus/es/components/skeleton-item/style/css";

const props = withDefaults(defineProps<{
  text?: string;
  rows?: number;
  compact?: boolean;
  variant?: "default" | "table" | "page" | "card";
}>(), {
  text: "正在加载...",
  rows: 3,
  compact: false,
  variant: "default"
});

const rowCount = computed(() => Math.max(1, props.rows));
const cardCount = computed(() => Math.min(rowCount.value, props.variant === "page" ? 4 : 6));
const tableColumns = computed(() => props.compact ? 3 : 5);
</script>

<template>
  <section class="loading-state" :class="[`loading-state-${variant}`, { 'loading-state-compact': compact }]" role="status" aria-live="polite">
    <div class="loading-state-header">
      <span class="loading-state-orb" aria-hidden="true">
        <ElIcon class="loading-state-icon">
          <LoadingIcon />
        </ElIcon>
      </span>
      <div class="loading-state-copy">
        <strong>{{ text }}</strong>
        <span v-if="!compact" class="loading-state-hint">请稍候，内容马上呈现</span>
      </div>
    </div>

    <ElSkeleton class="loading-state-skeleton" animated>
      <template #template>
        <div v-if="variant === 'table'" class="loading-skeleton-table">
          <div class="loading-skeleton-table-head">
            <ElSkeletonItem v-for="column in tableColumns" :key="`head-${column}`" variant="text" />
          </div>
          <div v-for="row in rowCount" :key="`row-${row}`" class="loading-skeleton-table-row">
            <ElSkeletonItem v-for="column in tableColumns" :key="`row-${row}-${column}`" variant="text" />
          </div>
        </div>

        <div v-else-if="variant === 'page' || variant === 'card'" class="loading-skeleton-page">
          <div v-if="variant === 'page'" class="loading-skeleton-hero">
            <ElSkeletonItem variant="h1" />
            <ElSkeletonItem variant="text" />
            <ElSkeletonItem variant="text" />
          </div>
          <div class="loading-skeleton-grid">
            <div v-for="row in cardCount" :key="`card-${row}`" class="loading-skeleton-card">
              <ElSkeletonItem variant="image" />
              <ElSkeletonItem variant="h3" />
              <ElSkeletonItem variant="text" />
            </div>
          </div>
        </div>

        <div v-else class="loading-skeleton-list">
          <div v-for="row in rowCount" :key="`list-${row}`" class="loading-skeleton-list-row">
            <ElSkeletonItem variant="circle" />
            <div>
              <ElSkeletonItem variant="h3" />
              <ElSkeletonItem variant="text" />
            </div>
          </div>
        </div>
      </template>
    </ElSkeleton>
  </section>
</template>
