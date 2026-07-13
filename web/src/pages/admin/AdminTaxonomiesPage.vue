<script setup lang="ts">
import { Search } from "@element-plus/icons-vue";
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  createAdminCategory,
  createAdminTag,
  deleteAdminCategory,
  deleteAdminTag,
  getCategories,
  getTags,
  updateAdminCategory,
  updateAdminTag,
  type Category,
  type Tag
} from "../../shared/api";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";

const confirmDialog = useConfirmStore();
const toast = useToastStore();
const categories = ref<Category[]>([]);
const tags = ref<Tag[]>([]);
const loading = ref(false);
const saving = ref(false);
const actingId = ref("");
const error = ref("");
const message = ref("");

const categoryId = ref("");
const categoryName = ref("");
const categorySlug = ref("");
const categoryDescription = ref("");
const categorySortOrder = ref(10);

const tagId = ref("");
const tagName = ref("");
const tagSlug = ref("");
const categoryPage = ref(1);
const categoryPageSize = ref(10);
const categoryTotal = ref(0);
const tagPage = ref(1);
const tagPageSize = ref(10);
const tagTotal = ref(0);
const categorySearch = ref("");
const tagSearch = ref("");

const categoryPostTotal = computed(() => categories.value.reduce((sum, item) => sum + item.postCount, 0));
const tagPostTotal = computed(() => tags.value.reduce((sum, item) => sum + item.postCount, 0));
const unusedCategories = computed(() => categories.value.filter((item) => item.postCount === 0).length);
const unusedTags = computed(() => tags.value.filter((item) => item.postCount === 0).length);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const [categoryResult, tagResult] = await Promise.all([
      getCategories({ page: categoryPage.value, pageSize: categoryPageSize.value, q: categorySearch.value.trim() }),
      getTags({ page: tagPage.value, pageSize: tagPageSize.value, q: tagSearch.value.trim() })
    ]);
    categories.value = categoryResult.items;
    tags.value = tagResult.items;
    categoryTotal.value = categoryResult.total;
    categoryPage.value = categoryResult.page;
    categoryPageSize.value = categoryResult.pageSize;
    tagTotal.value = tagResult.total;
    tagPage.value = tagResult.page;
    tagPageSize.value = tagResult.pageSize;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "分类标签加载失败";
  } finally {
    loading.value = false;
  }
}

async function searchCategories() {
  categoryPage.value = 1;
  await load();
}

async function searchTags() {
  tagPage.value = 1;
  await load();
}

async function setCategoryPage(value: number) {
  categoryPage.value = value;
  await load();
}

async function setCategoryPageSize(value: number) {
  categoryPageSize.value = value;
  categoryPage.value = 1;
  await load();
}

async function setTagPage(value: number) {
  tagPage.value = value;
  await load();
}

async function setTagPageSize(value: number) {
  tagPageSize.value = value;
  tagPage.value = 1;
  await load();
}

function resetCategoryForm() {
  categoryId.value = "";
  categoryName.value = "";
  categorySlug.value = "";
  categoryDescription.value = "";
  categorySortOrder.value = nextCategorySortOrder();
}

function editCategory(item: Category) {
  categoryId.value = item.id;
  categoryName.value = item.name;
  categorySlug.value = item.slug;
  categoryDescription.value = item.description;
  categorySortOrder.value = item.sortOrder;
}

async function saveCategory() {
  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    const payload = {
      name: categoryName.value,
      slug: categorySlug.value,
      description: categoryDescription.value,
      sortOrder: categorySortOrder.value
    };

    if (categoryId.value) {
      await updateAdminCategory(categoryId.value, payload);
      message.value = "分类已更新。";
    } else {
      await createAdminCategory(payload);
      message.value = "分类已创建。";
    }

    resetCategoryForm();
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "分类保存失败";
  } finally {
    saving.value = false;
  }
}

async function removeCategory(item: Category) {
  if (item.postCount > 0) {
    const text = `分类「${item.name}」仍被 ${item.postCount} 篇文章引用，请先调整这些文章的分类后再删除。`;
    message.value = text;
    error.value = "";
    toast.warning("分类正在使用中", text);
    return;
  }

  const confirmed = await confirmDialog.open({
    title: `删除分类「${item.name}」`,
    message: "删除后不可在文章筛选和归档中继续使用该分类。",
    confirmText: "删除分类",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  actingId.value = item.id;
  error.value = "";
  message.value = "";

  try {
    await deleteAdminCategory(item.id);
    if (categoryId.value === item.id) {
      resetCategoryForm();
    }
    message.value = "分类已删除。";
    toast.success("分类已删除", item.name);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "分类删除失败";
    toast.error("分类删除失败", error.value);
  } finally {
    actingId.value = "";
  }
}

function resetTagForm() {
  tagId.value = "";
  tagName.value = "";
  tagSlug.value = "";
}

function editTag(item: Tag) {
  tagId.value = item.id;
  tagName.value = item.name;
  tagSlug.value = item.slug;
}

async function saveTag() {
  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    const payload = {
      name: tagName.value,
      slug: tagSlug.value
    };

    if (tagId.value) {
      await updateAdminTag(tagId.value, payload);
      message.value = "标签已更新。";
    } else {
      await createAdminTag(payload);
      message.value = "标签已创建。";
    }

    resetTagForm();
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "标签保存失败";
  } finally {
    saving.value = false;
  }
}

async function removeTag(item: Tag) {
  if (item.postCount > 0) {
    const text = `标签「${item.name}」仍被 ${item.postCount} 篇文章引用，请先从这些文章中移除或替换该标签。`;
    message.value = text;
    error.value = "";
    toast.warning("标签正在使用中", text);
    return;
  }

  const confirmed = await confirmDialog.open({
    title: `删除标签「${item.name}」`,
    message: "删除后不可在文章筛选和标签页中继续使用该标签。",
    confirmText: "删除标签",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  actingId.value = item.id;
  error.value = "";
  message.value = "";

  try {
    await deleteAdminTag(item.id);
    if (tagId.value === item.id) {
      resetTagForm();
    }
    message.value = "标签已删除。";
    toast.success("标签已删除", item.name);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "标签删除失败";
    toast.error("标签删除失败", error.value);
  } finally {
    actingId.value = "";
  }
}

function nextCategorySortOrder() {
  const maxOrder = Math.max(0, ...categories.value.map((item) => item.sortOrder));
  return maxOrder + 10;
}
</script>

<template>
  <AdminLayout title="分类标签" description="管理文章分类、标签、排序和引用关系，保持内容结构稳定。" mobile-title="分类标签" primary-action="写作" primary-action-to="/admin/editor">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" :disabled="loading" @click="load">刷新</button>
        <RouterLink class="button" to="/admin/editor">写文章</RouterLink>
      </div>
    </template>

    <section class="stats-grid" aria-label="分类标签统计">
      <div class="stat-card"><span>分类</span><strong>{{ categoryTotal }}</strong></div>
      <div class="stat-card"><span>标签</span><strong>{{ tagTotal }}</strong></div>
      <div class="stat-card"><span>分类引用</span><strong>{{ categoryPostTotal }}</strong></div>
      <div class="stat-card"><span>可清理</span><strong>{{ unusedCategories + unusedTags }}</strong></div>
    </section>

    <LoadingState v-if="loading" variant="table" text="正在加载分类标签..." :rows="4" />
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="admin-grid-2">
      <section class="table-panel" aria-label="分类列表">
        <div class="panel-title" style="padding: 16px 16px 0;">
          <h2>分类</h2>
          <button class="button-secondary" type="button" @click="resetCategoryForm">新建分类</button>
        </div>
        <form class="table-toolbar taxonomy-table-toolbar" @submit.prevent="searchCategories">
          <input v-model="categorySearch" class="input" type="search" placeholder="搜索分类名称、Slug、描述" aria-label="搜索分类">
          <button class="button" type="submit" :disabled="loading">
            <Search class="button-icon" aria-hidden="true" />
            搜索
          </button>
        </form>
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>Slug</th>
              <th>排序</th>
              <th>文章</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in categories" :key="item.id">
              <td>
                <strong>{{ item.name }}</strong>
                <div class="meta-row"><span>{{ item.description || "无描述" }}</span></div>
              </td>
              <td>{{ item.slug }}</td>
              <td>{{ item.sortOrder }}</td>
              <td>{{ item.postCount }}</td>
              <td>
                <div class="header-actions">
                  <button class="button-secondary" type="button" @click="editCategory(item)">编辑</button>
                  <button class="button-secondary" type="button" :disabled="actingId === item.id" @click="removeCategory(item)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <PaginationControls
          :page="categoryPage"
          :page-size="categoryPageSize"
          :total="categoryTotal"
          :loading="loading"
          item-label="个分类"
          show-page-size
          :page-size-options="[5, 10, 20, 50]"
          @update:page="setCategoryPage"
          @update:page-size="setCategoryPageSize"
        />
      </section>

      <aside class="panel">
        <div class="panel-title">
          <h2>{{ categoryId ? "编辑分类" : "新建分类" }}</h2>
          <span class="tag">{{ unusedCategories }} 个未使用</span>
        </div>
        <form class="settings-stack" @submit.prevent="saveCategory">
          <div class="field"><label for="category-name">名称</label><input v-model="categoryName" class="input" id="category-name"></div>
          <div class="field"><label for="category-slug">Slug</label><input v-model="categorySlug" class="input" id="category-slug"></div>
          <div class="field"><label for="category-description">描述</label><textarea v-model="categoryDescription" class="input" id="category-description"></textarea></div>
          <div class="field"><label for="category-order">排序</label><input v-model.number="categorySortOrder" class="input" id="category-order" type="number"></div>
          <div class="header-actions">
            <button class="button" type="submit" :disabled="saving || !categoryName">{{ saving ? "保存中..." : "保存分类" }}</button>
            <button class="button-secondary" type="button" @click="resetCategoryForm">清空</button>
          </div>
        </form>
      </aside>
    </section>

    <section class="admin-grid-2" style="margin-top: 20px;">
      <section class="table-panel" aria-label="标签列表">
        <div class="panel-title" style="padding: 16px 16px 0;">
          <h2>标签</h2>
          <button class="button-secondary" type="button" @click="resetTagForm">新建标签</button>
        </div>
        <form class="table-toolbar taxonomy-table-toolbar" @submit.prevent="searchTags">
          <input v-model="tagSearch" class="input" type="search" placeholder="搜索标签名称或 Slug" aria-label="搜索标签">
          <button class="button" type="submit" :disabled="loading">
            <Search class="button-icon" aria-hidden="true" />
            搜索
          </button>
        </form>
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>Slug</th>
              <th>文章</th>
              <th>热度</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in tags" :key="item.id">
              <td><strong>{{ item.name }}</strong></td>
              <td>{{ item.slug }}</td>
              <td>{{ item.postCount }}</td>
              <td><span class="status" :class="item.postCount > 0 ? 'published' : 'muted'">{{ item.postCount > 0 ? "使用中" : "未使用" }}</span></td>
              <td>
                <div class="header-actions">
                  <button class="button-secondary" type="button" @click="editTag(item)">编辑</button>
                  <button class="button-secondary" type="button" :disabled="actingId === item.id" @click="removeTag(item)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <PaginationControls
          :page="tagPage"
          :page-size="tagPageSize"
          :total="tagTotal"
          :loading="loading"
          item-label="个标签"
          show-page-size
          :page-size-options="[5, 10, 20, 50]"
          @update:page="setTagPage"
          @update:page-size="setTagPageSize"
        />
      </section>

      <aside class="panel">
        <div class="panel-title">
          <h2>{{ tagId ? "编辑标签" : "新建标签" }}</h2>
          <span class="tag rust">{{ tagPostTotal }} 次引用</span>
        </div>
        <form class="settings-stack" @submit.prevent="saveTag">
          <div class="field"><label for="tag-name">名称</label><input v-model="tagName" class="input" id="tag-name"></div>
          <div class="field"><label for="tag-slug">Slug</label><input v-model="tagSlug" class="input" id="tag-slug"></div>
          <div class="review-note">
            <strong>标签用于跨分类检索</strong>
            <p>删除前需要先确认没有文章引用；合并标签可通过编辑文章标签完成。</p>
          </div>
          <div class="header-actions">
            <button class="button" type="submit" :disabled="saving || !tagName">{{ saving ? "保存中..." : "保存标签" }}</button>
            <button class="button-secondary" type="button" @click="resetTagForm">清空</button>
          </div>
        </form>
      </aside>
    </section>
  </AdminLayout>
</template>
