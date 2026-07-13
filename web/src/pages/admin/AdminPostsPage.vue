<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import AdminLayout from "../../components/AdminLayout.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import {
  archiveAdminPost,
  createAdminPost,
  getAdminPosts,
  publishAdminPost,
  type AdminPost,
  type AdminPostPayload,
  type AdminPostStats,
  type AdminPostVisibility
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";

const router = useRouter();
const confirmDialog = useConfirmStore();
const toast = useToastStore();
const posts = ref<AdminPost[]>([]);
const stats = ref<AdminPostStats>({ published: 0, draft: 0, review: 0, scheduled: 0, monthlyViews: "0", total: 0 });
const loading = ref(false);
const importing = ref(false);
const archivingId = ref("");
const restoringId = ref("");
const error = ref("");
const message = ref("");
const importInput = ref<HTMLInputElement | null>(null);
const searchQuery = ref("");
const statusFilter = ref("");
const sortMode = ref("updated");
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminPosts({
      q: searchQuery.value,
      status: statusFilter.value,
      sort: sortMode.value,
      page: page.value,
      pageSize: pageSize.value
    });
    posts.value = response.items;
    stats.value = response.stats;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章列表加载失败";
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

function openImport() {
  importInput.value?.click();
}

async function importMarkdown(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file || importing.value) {
    return;
  }

  importing.value = true;
  error.value = "";
  message.value = "";

  try {
    const content = await file.text();
    const payload = markdownPayload(content, file.name);
    const post = await createAdminPost(payload);
    message.value = `已导入草稿：${post.title}`;
    toast.success("Markdown 已导入", `已创建草稿：${post.title}`);
    await router.push(`/admin/editor?id=${post.id}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Markdown 导入失败";
    toast.error("Markdown 导入失败", error.value);
  } finally {
    importing.value = false;
  }
}

async function archivePost(post: AdminPost) {
  const confirmed = await confirmDialog.open({
    title: "下架已发布文章",
    message: `确认下架《${post.title}》吗？下架后公开列表和搜索中将不再展示。`,
    confirmText: "下架文章",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  archivingId.value = post.id;
  error.value = "";
  message.value = "";
  try {
    const archived = await archiveAdminPost(post.id);
    message.value = `已下架：${archived.title}`;
    toast.success("文章已下架", archived.title);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章下架失败";
    toast.error("文章下架失败", error.value);
  } finally {
    archivingId.value = "";
  }
}

async function restorePost(post: AdminPost) {
  const confirmed = await confirmDialog.open({
    title: "重新上架文章",
    message: `确认重新上架《${post.title}》吗？上架后会重新进入公开访问和搜索范围。`,
    confirmText: "重新上架",
    tone: "success"
  });
  if (!confirmed) {
    return;
  }

  restoringId.value = post.id;
  error.value = "";
  message.value = "";
  try {
    const restored = await publishAdminPost(post.id);
    message.value = `已重新上架：${restored.title}`;
    toast.success("文章已重新上架", restored.title);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章重新上架失败";
    toast.error("重新上架失败", error.value);
  } finally {
    restoringId.value = "";
  }
}

function markdownPayload(markdown: string, fileName: string): AdminPostPayload {
  const parsed = parseFrontMatter(markdown);
  const body = parsed.body.trim();
  const title = parsed.meta.title || firstHeading(body) || fileName.replace(/\.[^.]+$/, "");
  const summary = parsed.meta.summary || parsed.meta.description || firstParagraph(body);
  const tags = parseTags(parsed.meta.tags);

  return {
    title,
    summary,
    content: body || markdown.trim(),
    slug: parsed.meta.slug || slugFrom(fileName) || slugFrom(title),
    status: "draft",
    category: parsed.meta.category || "工程实践",
    tags,
    coverImage: parsed.meta.coverImage || "https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80",
    visibility: visibilityFromMeta(parsed.meta.visibility),
    seoTitle: parsed.meta.seoTitle || title,
    seoDescription: parsed.meta.seoDescription || summary
  };
}

function parseFrontMatter(markdown: string) {
  const normalized = markdown.replace(/^\uFEFF/, "");
  if (!normalized.startsWith("---")) {
    return { meta: {} as Record<string, string>, body: normalized };
  }

  const end = normalized.indexOf("\n---", 3);
  if (end < 0) {
    return { meta: {} as Record<string, string>, body: normalized };
  }

  const raw = normalized.slice(3, end).trim();
  const meta: Record<string, string> = {};
  raw.split(/\r?\n/).forEach((line) => {
    const index = line.indexOf(":");
    if (index < 0) {
      return;
    }
    const key = line.slice(0, index).trim();
    const value = line.slice(index + 1).trim().replace(/^["']|["']$/g, "");
    if (key) {
      meta[key] = value;
    }
  });

  return { meta, body: normalized.slice(end + 4) };
}

function firstHeading(markdown: string) {
  return markdown.split(/\r?\n/).map((line) => line.match(/^#\s+(.+)$/)?.[1]?.trim()).find(Boolean) || "";
}

function firstParagraph(markdown: string) {
  return markdown
    .split(/\r?\n\r?\n/)
    .map((item) => item.replace(/^#+\s+/, "").trim())
    .find((item) => item && !item.startsWith("```")) || "";
}

function parseTags(value = "") {
  return value
    .replace(/^\[|\]$/g, "")
    .split(/[,，]/)
    .map((item) => item.trim().replace(/^["']|["']$/g, ""))
    .filter(Boolean);
}

function slugFrom(value: string) {
  const slug = value
    .toLowerCase()
    .replace(/\.[^.]+$/, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return slug || `import-${Date.now()}`;
}

function visibilityFromMeta(value = ""): AdminPostVisibility {
  const normalized = value.trim().toLowerCase();
  if (normalized === "private" || normalized === "私密") return "private";
  if (normalized === "members" || normalized === "member" || normalized === "会员可见") return "members";
  return "public";
}

function visibilityText(visibility: AdminPostVisibility) {
  if (visibility === "private") return "私密";
  if (visibility === "members") return "会员可见";
  return "公开";
}

function statusText(status: AdminPost["status"]) {
  if (status === "published") return "已发布";
  if (status === "scheduled") return "待发布";
  if (status === "review") return "待审核";
  if (status === "archived") return "已归档";
  return "草稿";
}

function statusClass(status: AdminPost["status"]) {
  if (status === "published") return "published";
  if (status === "draft") return "draft";
  if (status === "archived") return "muted";
  return "review";
}

function formatDate(value: string) {
  return formatDateTime(value);
}
</script>

<template>
  <AdminLayout title="文章管理" description="管理草稿、审核、定时发布和已发布内容。" mobile-title="文章管理" primary-action="新建" primary-action-to="/admin/editor">
    <template #actions>
      <div class="header-actions">
        <input ref="importInput" class="sr-only" type="file" accept=".md,.markdown,text/markdown,text/plain" @change="importMarkdown">
        <button class="button-secondary" type="button" :disabled="importing" @click="openImport">{{ importing ? "导入中..." : "导入" }}</button>
        <RouterLink class="button" to="/admin/editor">新建文章</RouterLink>
      </div>
    </template>

    <section class="stats-grid" aria-label="文章统计">
      <div class="stat-card"><span>已发布</span><strong>{{ stats.published }}</strong></div>
      <div class="stat-card"><span>草稿</span><strong>{{ stats.draft }}</strong></div>
      <div class="stat-card"><span>待审核</span><strong>{{ stats.review }}</strong></div>
      <div class="stat-card"><span>待发布</span><strong>{{ stats.scheduled }}</strong></div>
      <div class="stat-card"><span>本月阅读</span><strong>{{ stats.monthlyViews }}</strong></div>
    </section>

    <p v-if="message" class="muted">{{ message }}</p>

    <section class="table-panel" aria-label="文章列表">
      <form class="table-toolbar" @submit.prevent="applyFilters">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索标题、作者、标签" aria-label="搜索文章">
        <select v-model="statusFilter" class="input" aria-label="文章状态" @change="applyFilters">
          <option value="">全部状态</option>
          <option value="published">已发布</option>
          <option value="draft">草稿</option>
          <option value="review">待审核</option>
          <option value="scheduled">待发布</option>
          <option value="archived">已归档</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
          <option value="updated">最近更新</option>
          <option value="views">最多阅读</option>
          <option value="scheduled">定时发布</option>
        </select>
      </form>

      <LoadingState v-if="loading" variant="table" text="正在加载文章..." :rows="5" />
      <p v-else-if="error" class="error">{{ error }}</p>

      <table v-else>
        <thead>
          <tr>
            <th>标题</th>
            <th>状态</th>
            <th>分类</th>
            <th>阅读</th>
            <th>评论</th>
            <th>更新时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="post in posts" :key="post.id">
            <td>
              <strong>{{ post.title }}</strong>
              <div class="meta-row">
                <span>{{ post.authorName }}</span>
                <span>{{ visibilityText(post.visibility) }}</span>
                <span v-if="post.status === 'scheduled' && post.scheduledAt">定时发布：{{ formatDate(post.scheduledAt) }}</span>
                <span v-else>/posts/{{ post.publishedPostSlug || post.slug }}</span>
              </div>
            </td>
            <td><span class="status" :class="statusClass(post.status)">{{ statusText(post.status) }}</span></td>
            <td>{{ post.category }}</td>
            <td>{{ post.viewCount }}</td>
            <td>{{ post.commentCount }}</td>
            <td>{{ formatDate(post.updatedAt) }}</td>
            <td>
              <div class="header-actions">
                <RouterLink class="button-secondary" :to="`/admin/editor?id=${post.id}`">编辑</RouterLink>
                <button v-if="post.status === 'published'" class="button-secondary" type="button" :disabled="archivingId === post.id" @click="archivePost(post)">
                  {{ archivingId === post.id ? "下架中..." : "下架" }}
                </button>
                <button v-if="post.status === 'archived'" class="button-secondary" type="button" :disabled="restoringId === post.id" @click="restorePost(post)">
                  {{ restoringId === post.id ? "上架中..." : "重新上架" }}
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="posts.length === 0">
            <td colspan="7" class="muted">没有匹配的文章。</td>
          </tr>
        </tbody>
      </table>
      <PaginationControls
        v-if="!loading && !error"
        :page="page"
        :page-size="pageSize"
        :total="total"
        :loading="loading"
        item-label="篇文章"
        show-page-size
        :page-size-options="[5, 10, 20, 50, 100]"
        @update:page="setPage"
        @update:page-size="setPageSize"
      />
    </section>
  </AdminLayout>
</template>
