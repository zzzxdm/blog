<script setup lang="ts">
import { MdPreview } from "md-editor-v3";
import "md-editor-v3/lib/style.css";

import { markdownPreviewOptions } from "../shared/markdownPreview";

withDefaults(defineProps<{
  content: string;
  previewId?: string;
  previewTheme?: string;
  codeTheme?: string;
}>(), {
  previewId: "markdown-preview",
  previewTheme: markdownPreviewOptions.previewTheme,
  codeTheme: markdownPreviewOptions.codeTheme
});
</script>

<template>
  <MdPreview
    class="markdown-preview"
    :model-value="content"
    :id="previewId"
    :theme="markdownPreviewOptions.theme"
    :preview-theme="previewTheme"
    :code-theme="codeTheme"
    :language="markdownPreviewOptions.language"
    :no-img-zoom-in="markdownPreviewOptions.noImgZoomIn"
    :no-katex="markdownPreviewOptions.noKatex"
    :no-mermaid="markdownPreviewOptions.noMermaid"
  />
</template>
