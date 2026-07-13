<script setup lang="ts">
import { Loading as LoadingIcon } from "@element-plus/icons-vue";
import { ElIcon } from "element-plus";
import { computed, ref } from "vue";
import { MdEditor, type ExposeParam, type ToolbarNames, type UploadImgCallBack } from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import "element-plus/es/components/icon/style/css";

import MarkdownThemeSwitcher from "./MarkdownThemeSwitcher.vue";
import { uploadMedia } from "../shared/api";
import { markdownPreviewOptions, useMarkdownPreviewTheme } from "../shared/markdownPreview";

const props = withDefaults(defineProps<{
  modelValue: string;
  disabled?: boolean;
  editorId?: string;
  height?: string;
  placeholder?: string;
  uploadCategory?: string;
}>(), {
  disabled: false,
  editorId: "rich-markdown-editor",
  height: "620px",
  placeholder: "开始写作，支持 Markdown、粘贴图片、拖拽图片和实时预览。",
  uploadCategory: "写作插图"
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
  save: [value: string];
  uploadError: [message: string];
}>();

const editorRef = ref<ExposeParam>();
const emojiOpen = ref(false);
const uploading = ref(false);
const { selectedPreviewTheme, selectedCodeTheme } = useMarkdownPreviewTheme();
const editorToolbars: ToolbarNames[] = [
  "bold",
  "underline",
  "italic",
  "strikeThrough",
  "-",
  "title",
  "quote",
  "unorderedList",
  "orderedList",
  "task",
  "-",
  "codeRow",
  "code",
  "link",
  "image",
  "table",
  "-",
  "revoke",
  "next",
  "save",
  "=",
  "preview",
  "previewOnly",
  "fullscreen",
  "catalog"
];
const emojis = ["😀", "😄", "😂", "🙂", "😉", "😍", "🤔", "👍", "👏", "🙏", "🔥", "✨", "✅", "⚠️", "💡", "📌", "📷", "📝", "🚀", "🎉"];

const toolbarState = computed(() => props.disabled || uploading.value);

function updateValue(value: string) {
  emit("update:modelValue", value);
}

function handleSave(value: string) {
  emit("save", value);
}

function insertEmoji(emoji: string) {
  editorRef.value?.insert(() => ({
    targetValue: emoji,
    select: false
  }));
  editorRef.value?.focus();
  emojiOpen.value = false;
}

async function handleUploadImg(files: File[], callback: UploadImgCallBack) {
  if (props.disabled) {
    return;
  }

  const imageFiles = files.filter((file) => file.type.startsWith("image/"));
  if (imageFiles.length !== files.length) {
    emit("uploadError", "只能上传图片文件。");
  }
  if (!imageFiles.length) {
    return;
  }

  uploading.value = true;
  try {
    const assets = await Promise.all(imageFiles.map((file) => uploadMedia(file, {
      alt: defaultAlt(file.name),
      category: props.uploadCategory
    })));
    callback(assets.map((asset) => ({
      url: asset.url,
      alt: asset.alt || defaultAlt(asset.fileName),
      title: asset.fileName
    })));
  } catch (err) {
    emit("uploadError", err instanceof Error ? err.message : "图片上传失败");
  } finally {
    uploading.value = false;
  }
}

function defaultAlt(fileName: string) {
  return fileName.replace(/\.[^.]+$/, "");
}
</script>

<template>
  <div class="rich-markdown-editor" :class="{ disabled }">
    <div class="rich-editor-actions">
      <div class="emoji-picker" :class="{ open: emojiOpen }">
        <button class="button-secondary emoji-trigger" type="button" :disabled="toolbarState" @click="emojiOpen = !emojiOpen">
          表情
        </button>
        <div v-if="emojiOpen" class="emoji-panel" role="menu" aria-label="插入表情">
          <button v-for="emoji in emojis" :key="emoji" type="button" role="menuitem" @click="insertEmoji(emoji)">
            {{ emoji }}
          </button>
        </div>
      </div>
      <MarkdownThemeSwitcher v-model:preview-theme="selectedPreviewTheme" v-model:code-theme="selectedCodeTheme" />
      <span v-if="uploading" class="inline-loading" role="status" aria-live="polite">
        <ElIcon class="inline-loading-icon">
          <LoadingIcon />
        </ElIcon>
        图片上传中...
      </span>
    </div>

    <MdEditor
      ref="editorRef"
      :model-value="modelValue"
      :editor-id="editorId"
      :disabled="disabled"
      :placeholder="placeholder"
      :toolbars="editorToolbars"
      :preview="true"
      :html-preview="false"
      :theme="markdownPreviewOptions.theme"
      :preview-theme="selectedPreviewTheme"
      :code-theme="selectedCodeTheme"
      :language="markdownPreviewOptions.language"
      :no-img-zoom-in="markdownPreviewOptions.noImgZoomIn"
      :no-katex="markdownPreviewOptions.noKatex"
      :no-mermaid="markdownPreviewOptions.noMermaid"
      :style="{ height }"
      @update:model-value="updateValue"
      @on-save="handleSave"
      @on-upload-img="handleUploadImg"
    />
  </div>
</template>
