<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminMedia,
  type MediaAsset
} from "../../shared/api";

const assets = ref<MediaAsset[]>([]);
const selectedId = ref("");
const loading = ref(false);
const error = ref("");

const selected = computed(() => assets.value.find((item) => item.id === selectedId.value) || assets.value[0]);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminMedia();
    assets.value = response.items;
    selectedId.value = assets.value[0]?.id || "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "媒体库加载失败";
  } finally {
    loading.value = false;
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}
</script>

<template>
  <AdminLayout title="媒体库" description="管理文章封面、正文图片、附件和图片替代文本。" mobile-title="媒体库" primary-action="上传">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button">批量选择</button>
        <button class="button" type="button">上传文件</button>
      </div>
    </template>

    <section class="admin-grid-2">
      <div>
        <form class="media-toolbar" @submit.prevent="load">
          <input class="input" type="search" placeholder="搜索文件名、alt 文本、上传人" aria-label="搜索媒体">
          <select class="input" aria-label="文件类型">
            <option>全部类型</option>
            <option>图片</option>
            <option>文档</option>
            <option>视频</option>
          </select>
          <select class="input" aria-label="排序">
            <option>最近上传</option>
            <option>文件最大</option>
            <option>使用最多</option>
          </select>
        </form>

        <p v-if="loading" class="muted">正在加载媒体资源...</p>
        <p v-else-if="error" class="error">{{ error }}</p>

        <section v-else class="media-grid" aria-label="媒体资源">
          <article v-for="asset in assets" :key="asset.id" class="media-card" @click="selectedId = asset.id">
            <img :src="asset.url" :alt="asset.alt">
            <div class="media-card-body"><strong>{{ asset.fileName }}</strong><div class="meta-row"><span>{{ asset.sizeLabel }}</span><span>{{ asset.usageCount ? `已使用 ${asset.usageCount} 次` : "未使用" }}</span></div><span class="tag">{{ asset.category }}</span></div>
          </article>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="upload-zone">
          <div>
            <strong>拖拽文件到这里上传</strong>
            <p>支持 JPG、PNG、WebP、GIF 和 PDF。图片会自动生成响应式尺寸。</p>
            <button class="button" type="button">选择文件</button>
          </div>
        </section>

        <section v-if="selected" class="panel">
          <div class="panel-title">
            <h2>选中资源</h2>
            <span class="tag">图片</span>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="filename">文件名</label><input class="input" id="filename" :value="selected.fileName" readonly></div>
            <div class="field"><label for="alt">Alt 文本</label><input class="input" id="alt" :value="selected.alt" readonly></div>
            <div class="field"><label for="asset-url">资源地址</label><input class="input" id="asset-url" :value="selected.url" readonly></div>
            <div class="meta-row"><span>尺寸 {{ selected.width }} x {{ selected.height }}</span><span>上传于 {{ formatDate(selected.uploadedAt) }}</span></div>
            <button class="button-secondary" type="button">复制地址</button>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
