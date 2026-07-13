<script setup lang="ts">
import { Search } from "@element-plus/icons-vue";
import { computed, onMounted, ref, watch } from "vue";
import { ElOption, ElSelect } from "element-plus";
import "element-plus/es/components/select/style/css";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  createAdminTopic,
  deleteAdminTopic,
  getCategories,
  getAdminTopics,
  getTags,
  updateAdminTopic,
  type Category,
  type Tag,
  type Topic,
  type TopicPayload,
  type TopicStatus,
  type TopicTone
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";

type SelectOption = {
  value: string;
  label: string;
  meta: string;
};

const confirmDialog = useConfirmStore();
const toast = useToastStore();
const topics = ref<Topic[]>([]);
const loading = ref(false);
const saving = ref(false);
const categoryOptionsLoading = ref(false);
const tagOptionsLoading = ref(false);
const actingId = ref("");
const error = ref("");
const message = ref("");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const statusFilter = ref("");
const searchQuery = ref("");
const titleInput = ref<HTMLInputElement | null>(null);
const submitAttempted = ref(false);

const topicId = ref("");
const title = ref("");
const slug = ref("");
const summary = ref("");
const coverImage = ref("");
const imageAlt = ref("");
const tone = ref<TopicTone>("");
const status = ref<TopicStatus>("active");
const featured = ref(true);
const sortOrder = ref(10);
const selectedCategories = ref<string[]>([]);
const selectedTags = ref<string[]>([]);
const categoryOptions = ref<SelectOption[]>([]);
const tagOptions = ref<SelectOption[]>([]);
const titleRequiredError = "请输入专题标题后再保存。";
const titleRequiredFieldError = "请输入专题标题";

const activeCount = computed(() => topics.value.filter((topic) => topic.status === "active").length);
const draftCount = computed(() => topics.value.filter((topic) => topic.status === "draft").length);
const referencedPostCount = computed(() => topics.value.reduce((sum, topic) => sum + topic.postCount, 0));
const titleError = computed(() => submitAttempted.value && !title.value.trim() ? titleRequiredFieldError : "");

onMounted(() => {
  void load();
  void loadTaxonomyOptions();
});

watch(statusFilter, () => {
  page.value = 1;
  void load();
});

watch(title, (value) => {
  if (value.trim() && error.value === titleRequiredError) {
    error.value = "";
  }
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminTopics({
      page: page.value,
      pageSize: pageSize.value,
      q: searchQuery.value.trim(),
      status: statusFilter.value,
      all: true
    });
    topics.value = response.items;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "专题加载失败";
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

function resetForm() {
  topicId.value = "";
  title.value = "";
  slug.value = "";
  summary.value = "";
  coverImage.value = "";
  imageAlt.value = "";
  tone.value = "";
  status.value = "active";
  featured.value = true;
  sortOrder.value = nextSortOrder();
  selectedCategories.value = [];
  selectedTags.value = [];
  submitAttempted.value = false;
}

function editTopic(topic: Topic) {
  topicId.value = topic.id;
  title.value = topic.title;
  slug.value = topic.slug;
  summary.value = topic.summary;
  coverImage.value = topic.coverImage;
  imageAlt.value = topic.imageAlt;
  tone.value = topic.tone;
  status.value = topic.status;
  featured.value = topic.featured;
  sortOrder.value = topic.sortOrder;
  selectedCategories.value = [...topic.categories];
  selectedTags.value = [...topic.tags];
  categoryOptions.value = mergeOptions(categoryOptions.value, selectedCategories.value);
  tagOptions.value = mergeOptions(tagOptions.value, selectedTags.value);
  submitAttempted.value = false;
}

async function saveTopic() {
  submitAttempted.value = true;
  error.value = "";
  message.value = "";

  if (!title.value.trim()) {
    error.value = titleRequiredError;
    titleInput.value?.focus();
    toast.warning("专题标题不能为空", "请输入专题标题后再保存。");
    return;
  }

  saving.value = true;

  const payload: TopicPayload = {
    slug: slug.value,
    title: title.value,
    summary: summary.value,
    coverImage: coverImage.value,
    imageAlt: imageAlt.value,
    tone: tone.value,
    status: status.value,
    featured: featured.value,
    sortOrder: sortOrder.value,
    categories: [...selectedCategories.value],
    tags: [...selectedTags.value]
  };

  try {
    if (topicId.value) {
      await updateAdminTopic(topicId.value, payload);
      message.value = "专题已更新。";
      toast.success("专题已更新", title.value.trim());
    } else {
      await createAdminTopic(payload);
      message.value = "专题已创建。";
      toast.success("专题已创建", title.value.trim());
    }
    resetForm();
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "专题保存失败";
    toast.error("专题保存失败", error.value);
  } finally {
    saving.value = false;
  }
}

async function removeTopic(topic: Topic) {
  const confirmed = await confirmDialog.open({
    title: `删除专题「${topic.title}」`,
    message: "公开页面将不再展示这个专题，关联文章不会被删除。",
    confirmText: "删除专题",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  actingId.value = topic.id;
  error.value = "";
  message.value = "";

  try {
    await deleteAdminTopic(topic.id);
    if (topicId.value === topic.id) {
      resetForm();
    }
    message.value = "专题已删除。";
    toast.success("专题已删除", topic.title);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "专题删除失败";
    toast.error("专题删除失败", error.value);
  } finally {
    actingId.value = "";
  }
}

async function loadTaxonomyOptions() {
  await Promise.all([loadCategoryOptions(), loadTagOptions()]);
}

async function loadCategoryOptions(query = "") {
  categoryOptionsLoading.value = true;
  try {
    const response = await getCategories({ page: 1, pageSize: 100, q: query.trim() });
    categoryOptions.value = mergeOptions(response.items.map(categoryOption), selectedCategories.value);
  } catch {
    categoryOptions.value = mergeOptions(categoryOptions.value, selectedCategories.value);
  } finally {
    categoryOptionsLoading.value = false;
  }
}

async function loadTagOptions(query = "") {
  tagOptionsLoading.value = true;
  try {
    const response = await getTags({ page: 1, pageSize: 100, q: query.trim() });
    tagOptions.value = mergeOptions(response.items.map(tagOption), selectedTags.value);
  } catch {
    tagOptions.value = mergeOptions(tagOptions.value, selectedTags.value);
  } finally {
    tagOptionsLoading.value = false;
  }
}

function searchCategoryOptions(query: string) {
  void loadCategoryOptions(query);
}

function searchTagOptions(query: string) {
  void loadTagOptions(query);
}

function categoryOption(item: Category): SelectOption {
  return {
    value: item.name,
    label: item.name,
    meta: `${item.slug} · ${item.postCount} 篇`
  };
}

function tagOption(item: Tag): SelectOption {
  return {
    value: item.name,
    label: item.name,
    meta: `${item.slug} · ${item.postCount} 篇`
  };
}

function mergeOptions(options: SelectOption[], selectedValues: string[]) {
  const merged = new Map<string, SelectOption>();
  options.forEach((item) => merged.set(item.value, item));
  selectedValues.forEach((value) => {
    if (!merged.has(value)) {
      merged.set(value, { value, label: value, meta: "已选" });
    }
  });
  return [...merged.values()];
}

function nextSortOrder() {
  const maxOrder = Math.max(0, ...topics.value.map((topic) => topic.sortOrder));
  return maxOrder + 10;
}

function statusText(value: TopicStatus) {
  return value === "draft" ? "草稿" : "启用";
}

function statusClass(value: TopicStatus) {
  return value === "draft" ? "draft" : "published";
}

function toneText(value: TopicTone) {
  if (value === "rust") return "赤陶";
  if (value === "amber") return "琥珀";
  if (value === "gray") return "灰色";
  return "默认";
}

function formatDate(value?: string) {
  return formatDateTime(value, "暂无更新");
}
</script>

<template>
  <AdminLayout title="专题管理" description="管理公开专题、文章匹配规则、排序和专题卡片展示。" mobile-title="专题管理" primary-action="新建专题">
    <template #mobile-action>
      <button class="button" type="button" @click="resetForm">新建</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <select v-model="statusFilter" class="input">
          <option value="">全部状态</option>
          <option value="active">启用</option>
          <option value="draft">草稿</option>
        </select>
        <button class="button-secondary" type="button" :disabled="loading" @click="load">刷新</button>
        <button class="button" type="button" @click="resetForm">新建专题</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="专题统计">
      <div class="stat-card"><span>专题</span><strong>{{ total }}</strong></div>
      <div class="stat-card"><span>当前页启用</span><strong>{{ activeCount }}</strong></div>
      <div class="stat-card"><span>当前页草稿</span><strong>{{ draftCount }}</strong></div>
      <div class="stat-card"><span>文章引用</span><strong>{{ referencedPostCount }}</strong></div>
    </section>

    <p v-if="loading" class="muted">正在加载专题...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="admin-grid-2">
      <section class="table-panel" aria-label="专题列表">
        <div class="panel-title" style="padding: 16px 16px 0;">
          <h2>专题</h2>
          <span class="tag">{{ total }} 个专题</span>
        </div>
        <form class="table-toolbar topic-table-toolbar" @submit.prevent="applyFilters">
          <input v-model="searchQuery" class="input" type="search" placeholder="搜索标题、Slug、摘要、分类、标签" aria-label="搜索专题">
          <button class="button" type="submit" :disabled="loading">
            <Search class="button-icon" aria-hidden="true" />
            搜索
          </button>
        </form>
        <table>
          <thead>
            <tr>
              <th>专题</th>
              <th>状态</th>
              <th>排序</th>
              <th>文章</th>
              <th>最近更新</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="topic in topics" :key="topic.id">
              <td>
                <strong>{{ topic.title }}</strong>
                <div class="meta-row">
                  <span>{{ topic.slug }}</span>
                  <span>{{ toneText(topic.tone) }}</span>
                  <span>{{ topic.featured ? "推荐" : "普通" }}</span>
                </div>
                <div class="meta-row">
                  <span>{{ topic.categories.join("、") || "未匹配分类" }}</span>
                  <span>{{ topic.tags.join("、") || "未匹配标签" }}</span>
                </div>
              </td>
              <td><span class="status" :class="statusClass(topic.status)">{{ statusText(topic.status) }}</span></td>
              <td>{{ topic.sortOrder }}</td>
              <td>{{ topic.postCount }}</td>
              <td>{{ formatDate(topic.latestPostAt) }}</td>
              <td>
                <div class="header-actions">
                  <button class="button-secondary" type="button" @click="editTopic(topic)">编辑</button>
                  <button class="button-secondary button-danger" type="button" :disabled="actingId === topic.id" @click="removeTopic(topic)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="!topics.length && !loading">
              <td colspan="6">暂无专题。</td>
            </tr>
          </tbody>
        </table>
        <PaginationControls
          :page="page"
          :page-size="pageSize"
          :total="total"
          :loading="loading"
          item-label="个专题"
          show-page-size
          :page-size-options="[5, 10, 20, 50]"
          @update:page="setPage"
          @update:page-size="setPageSize"
        />
      </section>

      <aside class="panel">
        <div class="panel-title">
          <h2>{{ topicId ? "编辑专题" : "新建专题" }}</h2>
          <span class="tag" :class="tone">{{ toneText(tone) }}</span>
        </div>
        <form class="settings-stack" @submit.prevent="saveTopic">
          <div class="field">
            <label for="topic-title">标题（必填）</label>
            <input
              v-model="title"
              ref="titleInput"
              class="input"
              :class="{ 'input-invalid': titleError }"
              id="topic-title"
              aria-required="true"
              :aria-invalid="Boolean(titleError)"
              aria-describedby="topic-title-error"
            >
            <span v-if="titleError" id="topic-title-error" class="field-error">{{ titleError }}</span>
          </div>
          <div class="field"><label for="topic-slug">Slug</label><input v-model="slug" class="input" id="topic-slug" placeholder="留空时按标题生成"></div>
          <div class="field"><label for="topic-summary">摘要</label><textarea v-model="summary" class="input" id="topic-summary"></textarea></div>
          <div class="field"><label for="topic-cover">封面图</label><input v-model="coverImage" class="input" id="topic-cover"></div>
          <div class="field"><label for="topic-alt">图片 Alt</label><input v-model="imageAlt" class="input" id="topic-alt"></div>
          <div class="field">
            <label for="topic-categories">匹配分类</label>
            <ElSelect
              v-model="selectedCategories"
              class="element-select"
              input-id="topic-categories"
              multiple
              filterable
              remote
              reserve-keyword
              clearable
              :loading="categoryOptionsLoading"
              :remote-method="searchCategoryOptions"
              placeholder="选择分类"
            >
              <ElOption v-for="item in categoryOptions" :key="item.value" :label="item.label" :value="item.value">
                <span>{{ item.label }}</span>
                <span class="select-option-meta">{{ item.meta }}</span>
              </ElOption>
            </ElSelect>
          </div>
          <div class="field">
            <label for="topic-tags">匹配标签</label>
            <ElSelect
              v-model="selectedTags"
              class="element-select"
              input-id="topic-tags"
              multiple
              filterable
              remote
              reserve-keyword
              clearable
              :loading="tagOptionsLoading"
              :remote-method="searchTagOptions"
              placeholder="选择标签"
            >
              <ElOption v-for="item in tagOptions" :key="item.value" :label="item.label" :value="item.value">
                <span>{{ item.label }}</span>
                <span class="select-option-meta">{{ item.meta }}</span>
              </ElOption>
            </ElSelect>
          </div>
          <div class="field"><label for="topic-tone">展示色</label>
            <select v-model="tone" class="input" id="topic-tone">
              <option value="">默认</option>
              <option value="rust">赤陶</option>
              <option value="amber">琥珀</option>
              <option value="gray">灰色</option>
            </select>
          </div>
          <div class="field"><label for="topic-status">状态</label>
            <select v-model="status" class="input" id="topic-status">
              <option value="active">启用</option>
              <option value="draft">草稿</option>
            </select>
          </div>
          <div class="field"><label for="topic-order">排序</label><input v-model.number="sortOrder" class="input" id="topic-order" type="number"></div>
          <label class="checkbox-line">
            <input v-model="featured" type="checkbox">
            <span>推荐到首页和热门专题</span>
          </label>
          <div class="header-actions">
            <button class="button" type="button" :disabled="saving" @click="saveTopic">{{ saving ? "保存中..." : "保存专题" }}</button>
            <button class="button-secondary" type="button" @click="resetForm">清空</button>
          </div>
        </form>
      </aside>
    </section>
  </AdminLayout>
</template>
