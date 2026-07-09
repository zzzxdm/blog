<script setup lang="ts">
import { markdownCodeThemes, markdownPreviewThemes } from "../shared/markdownPreview";

defineProps<{
  previewTheme: string;
  codeTheme: string;
}>();

defineEmits<{
  "update:previewTheme": [value: string];
  "update:codeTheme": [value: string];
}>();
</script>

<template>
  <div class="markdown-theme-switcher" aria-label="Markdown 展示主题">
    <label>
      <span>主题选择</span>
    </label>
    <label>
      正文
      <select
        class="input"
        :value="previewTheme"
        @change="$emit('update:previewTheme', ($event.target as HTMLSelectElement).value)"
      >
        <option v-for="item in markdownPreviewThemes" :key="item.value" :value="item.value">{{ item.label }}</option>
      </select>
    </label>
    <label>
      代码
      <select
        class="input"
        :value="codeTheme"
        @change="$emit('update:codeTheme', ($event.target as HTMLSelectElement).value)"
      >
        <option v-for="item in markdownCodeThemes" :key="item.value" :value="item.value">{{ item.label }}</option>
      </select>
    </label>
  </div>
</template>
