<script setup lang="ts">
import { computed } from "vue";
import { ElIcon, ElSkeleton } from "element-plus";
import { Loading as LoadingIcon } from "@element-plus/icons-vue";

import "element-plus/es/components/icon/style/css";
import "element-plus/es/components/skeleton/style/css";
import "element-plus/es/components/skeleton-item/style/css";

const props = withDefaults(defineProps<{
  text?: string;
  rows?: number;
  compact?: boolean;
  variant?: "default" | "table" | "page";
}>(), {
  text: "正在加载...",
  rows: 3,
  compact: false,
  variant: "default"
});

const rowCount = computed(() => Math.max(1, props.rows));
</script>

<template>
  <section class="loading-state" :class="[`loading-state-${variant}`, { 'loading-state-compact': compact }]" role="status" aria-live="polite">
    <div class="loading-state-label">
      <ElIcon class="loading-state-icon">
        <LoadingIcon />
      </ElIcon>
      <span>{{ text }}</span>
    </div>
    <ElSkeleton class="loading-state-skeleton" animated :rows="rowCount" />
  </section>
</template>
