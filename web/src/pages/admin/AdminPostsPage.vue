<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  createAdminPost,
  getAdminPosts,
  type AdminPost,
  type AdminPostPayload,
  type AdminPostStats,
  type AdminPostVisibility
} from "../../shared/api";

const router = useRouter();
const posts = ref<AdminPost[]>([]);
const stats = ref<AdminPostStats>({ published: 0, draft: 0, review: 0, monthlyViews: "0", total: 0 });
const loading = ref(false);
const importing = ref(false);
const error = ref("");
const message = ref("");
const importInput = ref<HTMLInputElement | null>(null);
const searchQuery = ref("");
const statusFilter = ref("");
const sortMode = ref("updated");

const visiblePosts = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  const filtered = posts.value.filter((post) => {
    const matchesKeyword = !keyword || [
      post.title,
      post.summary,
      post.authorName,
      post.category,
      post.slug,
      visibilityText(post.visibility),
      post.tags.join(" ")
    ].join(" ").toLowerCase().includes(keyword);
    const matchesStatus = statusFilter.value === "" || post.status === statusFilter.value;

    return matchesKeyword && matchesStatus;
  });

  return [...filtered].sort((left, right) => {
    if (sortMode.value === "views") {
      return right.viewCount - left.viewCount;
    }
    if (sortMode.value === "scheduled") {
      const leftTime = left.scheduledAt ? new Date(left.scheduledAt).getTime() : Number.MAX_SAFE_INTEGER;
      const rightTime = right.scheduledAt ? new Date(right.scheduledAt).getTime() : Number.MAX_SAFE_INTEGER;
      return leftTime - rightTime;
    }
    return new Date(right.updatedAt).getTime() - new Date(left.updatedAt).getTime();
  });
});

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminPosts();
    posts.value = response.items;
    stats.value = response.stats;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章列表加载失败";
  } finally {
    loading.value = false;
  }
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
    await router.push(`/admin/editor?id=${post.id}`);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Markdown 导入失败";
  } finally {
    importing.value = false;
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
  return new Date(value).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}
</script>

<template>
  <AdminLayout title="文章管理" description="管理草稿、审核、定时发布和已发布内容。" mobile-title="文章管理" primary-action="新建">
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
      <div class="stat-card"><span>本月阅读</span><strong>{{ stats.monthlyViews }}</strong></div>
    </section>

    <p v-if="message" class="muted">{{ message }}</p>

    <section class="table-panel" aria-label="文章列表">
      <form class="table-toolbar" @submit.prevent="load">
        <input v-model="searchQuery" class="input" type="search" placeholder="搜索标题、作者、标签" aria-label="搜索文章">
        <select v-model="statusFilter" class="input" aria-label="文章状态">
          <option value="">全部状态</option>
          <option value="published">已发布</option>
          <option value="draft">草稿</option>
          <option value="review">待审核</option>
          <option value="scheduled">待发布</option>
          <option value="archived">已归档</option>
        </select>
        <select v-model="sortMode" class="input" aria-label="排序">
          <option value="updated">最近更新</option>
          <option value="views">最多阅读</option>
          <option value="scheduled">定时发布</option>
        </select>
      </form>

      <p v-if="loading" class="muted">正在加载文章...</p>
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
          <tr v-for="post in visiblePosts" :key="post.id">
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
            <td><RouterLink class="button-secondary" :to="`/admin/editor?id=${post.id}`">编辑</RouterLink></td>
          </tr>
          <tr v-if="visiblePosts.length === 0">
            <td colspan="7" class="muted">没有匹配的文章。</td>
          </tr>
        </tbody>
      </table>
    </section>
  </AdminLayout>
</template>
