<script setup lang="ts">
import { useToastStore } from "../stores/toast";

const toast = useToastStore();

function toneLabel(tone: string) {
  if (tone === "success") return "成功";
  if (tone === "error") return "失败";
  if (tone === "warning") return "注意";
  return "提示";
}
</script>

<template>
  <Teleport to="body">
    <div class="toast-viewport" aria-live="polite" aria-label="操作反馈">
      <TransitionGroup name="toast">
        <article v-for="item in toast.items" :key="item.id" class="toast-item" :class="item.tone" role="status">
          <span class="toast-dot" aria-hidden="true"></span>
          <div>
            <div class="toast-eyebrow">{{ toneLabel(item.tone) }}</div>
            <strong>{{ item.title }}</strong>
            <p v-if="item.message">{{ item.message }}</p>
          </div>
          <button type="button" aria-label="关闭提示" @click="toast.remove(item.id)">×</button>
        </article>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
