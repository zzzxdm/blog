<script setup lang="ts">
import { MdCatalog, type Themes } from "md-editor-v3";

import { markdownPreviewOptions } from "../shared/markdownPreview";

withDefaults(defineProps<{
  editorId?: string;
  /** md-editor-v3 的 UI 主题，仅 light/dark（不是 previewTheme 如 github） */
  theme?: Themes;
}>(), {
  editorId: "markdown-preview",
  theme: markdownPreviewOptions.theme
});
</script>

<template>
  <MdCatalog
    :editor-id="editorId"
    :theme="theme"
    class="custom-markdown-catalog"
  />
</template>

<style scoped>
.custom-markdown-catalog {
  --md-catalog-color: var(--color-foreground);
  --md-catalog-hover-color: var(--color-primary);
  --md-catalog-active-color: var(--color-primary);
}
</style>
