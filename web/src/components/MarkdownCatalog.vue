<script setup lang="ts">
import { MdCatalog } from "md-editor-v3";
import { computed } from "vue";
import { useMarkdownPreviewTheme } from "../shared/markdownPreview";

const props = withDefaults(defineProps<{
  editorId?: string;
  theme?: string;
}>(), {
  editorId: "markdown-preview",
});

const { selectedPreviewTheme } = useMarkdownPreviewTheme();
const resolvedTheme = computed(() => props.theme || selectedPreviewTheme.value);
</script>

<template>
  <MdCatalog
    :editor-id="editorId"
    :theme="resolvedTheme"
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
