<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  deleteAdminMedia,
  getAdminMedia,
  getAdminMediaAsset,
  updateAdminMedia,
  uploadAdminMedia,
  type MediaAsset
} from "../../shared/api";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";
import { formatDateTime } from "../../shared/datetime";

const confirmDialog = useConfirmStore();
const toast = useToastStore();
const assets = ref<MediaAsset[]>([]);
const selectedId = ref("");
const loading = ref(false);
const uploading = ref(false);
const savingMetadata = ref(false);
const deleting = ref(false);
const batchMode = ref(false);
const batchDeleting = ref(false);
const error = ref("");
const uploadError = ref("");
const notice = ref("");
const fileInput = ref<HTMLInputElement | null>(null);
const editAlt = ref("");
const editCategory = ref("");
const selectedAssetIds = ref<string[]>([]);
const searchQuery = ref("");
const typeFilter = ref("");
const sortMode = ref("latest");
const page = ref(1);
const pageSize = ref(12);
const total = ref(0);

const selected = computed(() => assets.value.find((item) => item.id === selectedId.value) || assets.value[0]);
const selectedBatchAssets = computed(() => assets.value.filter((item) => selectedAssetIds.value.includes(item.id)));
const deletableBatchAssets = computed(() => selectedBatchAssets.value.filter((item) => item.usageCount === 0));
const blockedBatchCount = computed(() => selectedBatchAssets.value.length - deletableBatchAssets.value.length);

onMounted(load);
watch(selected, (asset) => {
  editAlt.value = asset?.alt || "";
  editCategory.value = asset?.category || "";
}, { immediate: true });

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminMedia({
      q: searchQuery.value,
      type: typeFilter.value,
      sort: sortMode.value,
      page: page.value,
      pageSize: pageSize.value
    });
    assets.value = response.items;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
    if (!assets.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = assets.value[0]?.id || "";
    }
    selectedAssetIds.value = selectedAssetIds.value.filter((id) => assets.value.some((item) => item.id === id));
  } catch (err) {
    error.value = err instanceof Error ? err.message : "媒体库加载失败";
    toast.error("媒体库加载失败", error.value);
  } finally {
    loading.value = false;
  }
}

async function applyFilters() {
  page.value = 1;
  await load();
}

async function setPage(value: number) {
  page.value = value;
  await load();
}

async function setPageSize(value: number) {
  pageSize.value = value;
  page.value = 1;
  await load();
}

async function selectAsset(id: string) {
  selectedId.value = id;
  uploadError.value = "";
  notice.value = "";

  try {
    const asset = await getAdminMediaAsset(id);
    assets.value = assets.value.map((item) => (item.id === id ? asset : item));
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : "资源详情加载失败";
    toast.error("资源详情加载失败", uploadError.value);
  }
}

function openPicker() {
  fileInput.value?.click();
}

async function handleInputChange(event: Event) {
  const input = event.target as HTMLInputElement;
  await uploadFiles(input.files);
  input.value = "";
}

async function handleDrop(event: DragEvent) {
  await uploadFiles(event.dataTransfer?.files ?? null);
}

async function uploadFiles(fileList: FileList | null) {
  const files = Array.from(fileList ?? []);
  if (!files.length || uploading.value) {
    return;
  }

  uploading.value = true;
  uploadError.value = "";
  notice.value = "";

  try {
    let lastUploaded: MediaAsset | null = null;
    for (const file of files) {
      lastUploaded = await uploadAdminMedia(file, {
        alt: defaultAlt(file.name),
        category: file.type === "application/pdf" ? "文档" : "上传"
      });
    }

    await load();
    if (lastUploaded) {
      selectedId.value = lastUploaded.id;
    }
    notice.value = files.length > 1 ? `已上传 ${files.length} 个文件` : "文件已上传";
    toast.success("文件已上传", files.length > 1 ? `已上传 ${files.length} 个文件。` : lastUploaded?.fileName);
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : "上传失败";
    toast.error("上传失败", uploadError.value);
  } finally {
    uploading.value = false;
  }
}

async function saveMetadata() {
  if (!selected.value) {
    return;
  }

  savingMetadata.value = true;
  uploadError.value = "";
  notice.value = "";

  try {
    const updated = await updateAdminMedia(selected.value.id, {
      alt: editAlt.value,
      category: editCategory.value
    });
    assets.value = assets.value.map((item) => (item.id === updated.id ? updated : item));
    notice.value = "资源信息已保存";
    toast.success("资源信息已保存", updated.fileName);
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : "资源信息保存失败";
    toast.error("资源信息保存失败", uploadError.value);
  } finally {
    savingMetadata.value = false;
  }
}

async function deleteSelected() {
  if (!selected.value) {
    return;
  }
  if (selected.value.usageCount > 0) {
    toast.warning("资源正在使用中", "请先移除内容引用后再删除。");
    return;
  }

  const confirmed = await confirmDialog.open({
    title: `删除 ${selected.value.fileName}`,
    message: "该资源会从媒体库移除，未保存的引用不会自动修复。",
    confirmText: "删除资源",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  const deletedId = selected.value.id;
  const deletedFileName = selected.value.fileName;
  deleting.value = true;
  uploadError.value = "";
  notice.value = "";

  try {
    await deleteAdminMedia(deletedId);
    assets.value = assets.value.filter((item) => item.id !== deletedId);
    selectedId.value = assets.value[0]?.id || "";
    notice.value = "资源已删除";
    toast.success("资源已删除", deletedFileName);
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : "资源删除失败";
    toast.error("资源删除失败", uploadError.value);
  } finally {
    deleting.value = false;
  }
}

function toggleBatchMode() {
  batchMode.value = !batchMode.value;
  selectedAssetIds.value = [];
  notice.value = batchMode.value ? "已进入批量选择模式。" : "";
  uploadError.value = "";
  toast.info(batchMode.value ? "已进入批量选择" : "已退出批量选择");
}

function isBatchSelected(id: string) {
  return selectedAssetIds.value.includes(id);
}

function toggleBatchAsset(asset: MediaAsset) {
  if (!batchMode.value) {
    return;
  }

  if (isBatchSelected(asset.id)) {
    selectedAssetIds.value = selectedAssetIds.value.filter((id) => id !== asset.id);
    return;
  }

  selectedAssetIds.value = [...selectedAssetIds.value, asset.id];
}

async function deleteBatchSelected() {
  if (!deletableBatchAssets.value.length) {
    toast.warning("没有可删除资源", selectedBatchAssets.value.length ? "选中的资源正在被内容引用。" : "请先选择未使用资源。");
    return;
  }

  const confirmed = await confirmDialog.open({
    title: `删除 ${deletableBatchAssets.value.length} 个未使用资源`,
    message: blockedBatchCount.value > 0
      ? `${blockedBatchCount.value} 个已被引用的资源会保留，只删除未使用资源。`
      : "这些资源会从媒体库移除。",
    confirmText: "批量删除",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  batchDeleting.value = true;
  uploadError.value = "";
  notice.value = "";

  try {
    for (const asset of deletableBatchAssets.value) {
      await deleteAdminMedia(asset.id);
    }
    const deletedIds = new Set(deletableBatchAssets.value.map((item) => item.id));
    assets.value = assets.value.filter((item) => !deletedIds.has(item.id));
    selectedAssetIds.value = selectedAssetIds.value.filter((id) => !deletedIds.has(id));
    if (!assets.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = assets.value[0]?.id || "";
    }
    notice.value = blockedBatchCount.value > 0
      ? `已删除 ${deletedIds.size} 个未使用资源，${blockedBatchCount.value} 个资源仍被引用。`
      : `已删除 ${deletedIds.size} 个资源。`;
    toast.success("批量删除完成", notice.value);
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : "批量删除失败";
    toast.error("批量删除失败", uploadError.value);
  } finally {
    batchDeleting.value = false;
  }
}

async function copySelectedUrl() {
  if (!selected.value) {
    return;
  }

  try {
    if (!navigator.clipboard) {
      throw new Error("当前浏览器不支持剪贴板写入");
    }
    await navigator.clipboard.writeText(selected.value.url);
    notice.value = "资源地址已复制";
    toast.success("资源地址已复制", selected.value.fileName);
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : "复制失败";
    toast.error("复制失败", uploadError.value);
  }
}

function defaultAlt(fileName: string) {
  return fileName.replace(/\.[^.]+$/, "");
}

function typeLabel(type: string) {
  return type === "document" ? "文档" : "图片";
}

function formatDate(value: string) {
  return formatDateTime(value);
}

</script>

<template>
  <AdminLayout title="媒体库" description="管理文章封面、正文图片、附件和图片替代文本。" mobile-title="媒体库" primary-action="上传">
    <template #mobile-action>
      <button class="button" type="button" :disabled="uploading" @click="openPicker">{{ uploading ? "上传中..." : "上传" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" @click="toggleBatchMode">{{ batchMode ? "退出批量" : "批量选择" }}</button>
        <button v-if="batchMode" class="button-secondary" type="button" :disabled="batchDeleting || !deletableBatchAssets.length" @click="deleteBatchSelected">
          {{ batchDeleting ? "删除中..." : `删除选中 ${deletableBatchAssets.length}` }}
        </button>
        <button class="button" type="button" :disabled="uploading" @click="openPicker">{{ uploading ? "上传中..." : "上传文件" }}</button>
        <input ref="fileInput" hidden type="file" accept="image/jpeg,image/png,image/webp,image/gif,application/pdf" multiple @change="handleInputChange">
      </div>
    </template>

    <section class="admin-grid-2">
      <div>
        <form class="media-toolbar" @submit.prevent="applyFilters">
          <input v-model="searchQuery" class="input" type="search" placeholder="搜索文件名、alt 文本、上传人" aria-label="搜索媒体">
          <select v-model="typeFilter" class="input" aria-label="文件类型" @change="applyFilters">
            <option value="">全部类型</option>
            <option value="image">图片</option>
            <option value="document">文档</option>
          </select>
          <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
            <option value="latest">最近上传</option>
            <option value="size">文件最大</option>
            <option value="usage">使用最多</option>
          </select>
        </form>

        <LoadingState v-if="loading" variant="table" text="正在加载媒体资源..." :rows="4" />
        <p v-else-if="error" class="error">{{ error }}</p>
        <template v-else>
          <p v-if="uploadError" class="error">{{ uploadError }}</p>
          <p v-else-if="notice" class="muted">{{ notice }}</p>
          <p v-if="batchMode && selectedBatchAssets.length" class="muted">
            已选择 {{ selectedBatchAssets.length }} 个资源；{{ deletableBatchAssets.length }} 个可删除，{{ blockedBatchCount }} 个正在被引用。
          </p>

          <section class="media-grid" aria-label="媒体资源">
            <article v-for="asset in assets" :key="asset.id" class="media-card" :class="{ 'is-selected': isBatchSelected(asset.id) }" @click="batchMode ? toggleBatchAsset(asset) : selectAsset(asset.id)">
              <label v-if="batchMode" class="media-card-checkbox" @click.stop>
                <input type="checkbox" :checked="isBatchSelected(asset.id)" @change="toggleBatchAsset(asset)">
              </label>
              <img v-if="asset.type === 'image'" :src="asset.url" :alt="asset.alt">
              <div v-else class="media-card-file">{{ typeLabel(asset.type) }}</div>
              <div class="media-card-body"><strong>{{ asset.fileName }}</strong><div class="meta-row"><span>{{ asset.sizeLabel }}</span><span>{{ asset.usageCount ? `已使用 ${asset.usageCount} 次` : "未使用" }}</span></div><span class="tag">{{ asset.category }}</span></div>
            </article>
            <p v-if="assets.length === 0" class="muted">没有匹配的媒体资源。</p>
          </section>
          <PaginationControls
            :page="page"
            :page-size="pageSize"
            :total="total"
            :loading="loading"
            item-label="个资源"
            show-page-size
            :page-size-options="[6, 12, 24, 48, 96]"
            @update:page="setPage"
            @update:page-size="setPageSize"
          />
        </template>
      </div>

      <aside class="settings-stack">
        <section class="upload-zone" :aria-busy="uploading" @dragover.prevent @drop.prevent="handleDrop">
          <div>
            <strong>{{ uploading ? "正在上传文件" : "拖拽文件到这里上传" }}</strong>
            <p>支持 JPG、PNG、WebP、GIF 和 PDF。图片会记录宽高和替代文本。</p>
            <button class="button" type="button" :disabled="uploading" @click="openPicker">选择文件</button>
          </div>
        </section>

        <section v-if="selected" class="panel">
          <div class="panel-title">
            <h2>选中资源</h2>
            <span class="tag">{{ typeLabel(selected.type) }}</span>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="filename">文件名</label><input class="input" id="filename" :value="selected.fileName" readonly></div>
            <div class="field"><label for="alt">Alt 文本</label><input v-model="editAlt" class="input" id="alt"></div>
            <div class="field"><label for="media-category">分类</label><input v-model="editCategory" class="input" id="media-category"></div>
            <div class="field"><label for="asset-url">资源地址</label><input class="input" id="asset-url" :value="selected.url" readonly></div>
            <div class="meta-row"><span>尺寸 {{ selected.width }} x {{ selected.height }}</span><span>上传于 {{ formatDate(selected.uploadedAt) }}</span></div>
            <div class="header-actions">
              <button class="button" type="button" :disabled="savingMetadata" @click="saveMetadata">{{ savingMetadata ? "保存中..." : "保存信息" }}</button>
              <button class="button-secondary" type="button" @click="copySelectedUrl">复制地址</button>
              <button class="button-secondary" type="button" :disabled="deleting || selected.usageCount > 0" @click="deleteSelected">{{ deleting ? "删除中..." : "删除" }}</button>
            </div>
            <p v-if="selected.usageCount > 0" class="muted">该资源正在被内容引用，不能直接删除。</p>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
