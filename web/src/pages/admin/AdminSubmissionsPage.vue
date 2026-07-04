<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminSubmissions,
  reviewSubmission,
  updateAdminSubmission,
  updateAdminUserRole,
  type Submission,
  type SubmissionPayload,
  type SubmissionStats
} from "../../shared/api";

const submissions = ref<Submission[]>([]);
const stats = ref<SubmissionStats>({ draft: 0, submitted: 0, returned: 0, rejected: 0, published: 0, total: 0 });
const selectedId = ref("");
const filterStatus = ref("submitted");
const loading = ref(false);
const acting = ref(false);
const editing = ref(false);
const savingEdit = ref(false);
const upgradingAuthor = ref(false);
const error = ref("");
const message = ref("");
const reviewNote = ref("内容结构清楚，可以发布。建议把标题和摘要再压缩一点。");
const publishSlug = ref("");
const publishCategory = ref("工程实践");
const editTitle = ref("");
const editSummary = ref("");
const editContent = ref("");
const editCategory = ref("");
const editTags = ref("");
const editCoverImage = ref("");
const editSlug = ref("");
const searchQuery = ref("");
const sortMode = ref("latest");

const selected = computed(() => submissions.value.find((item) => item.id === selectedId.value) || submissions.value[0]);
const previewParagraphs = computed(() => selected.value?.content.split(/\n+/).map((item) => item.trim()).filter(Boolean) || []);
const visibleSubmissions = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  const riskRank: Record<string, number> = { 高: 3, 中: 2, 低: 1 };
  const filtered = submissions.value.filter((item) => {
    if (!keyword) {
      return true;
    }

    return [
      item.title,
      item.summary,
      item.authorName,
      item.authorId,
      item.category,
      item.slug,
      item.tags.join(" "),
      item.riskLevel
    ].join(" ").toLowerCase().includes(keyword);
  });

  return [...filtered].sort((left, right) => {
    if (sortMode.value === "risk") {
      return (riskRank[right.riskLevel] || 0) - (riskRank[left.riskLevel] || 0);
    }
    if (sortMode.value === "quality") {
      return right.wordCount - left.wordCount;
    }

    return submissionTime(right) - submissionTime(left);
  });
});

onMounted(load);

watch(selected, (item) => {
  if (!item) {
    editing.value = false;
    return;
  }
  reviewNote.value = item.reviewNote || "内容结构清楚，可以发布。建议把标题和摘要再压缩一点。";
  publishSlug.value = item.slug;
  publishCategory.value = item.category;
  seedEdit(item);
  editing.value = false;
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminSubmissions(filterStatus.value);
    submissions.value = response.items;
    stats.value = response.stats;
    if (!submissions.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = submissions.value[0]?.id || "";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿审核列表加载失败";
  } finally {
    loading.value = false;
  }
}

async function review(action: "approve" | "return" | "reject") {
  if (!selected.value) {
    return;
  }

  acting.value = true;
  error.value = "";
  message.value = "";

  try {
    const updated = await reviewSubmission(selected.value.id, {
      action,
      note: reviewNote.value,
      slug: publishSlug.value,
      category: publishCategory.value
    });
    message.value = action === "approve" ? `已发布为 /posts/${updated.publishedPostSlug}` : "审核结果已发送给投稿人。";
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "审核操作失败";
  } finally {
    acting.value = false;
  }
}

function seedEdit(item: Submission) {
  editTitle.value = item.title;
  editSummary.value = item.summary;
  editContent.value = item.content;
  editCategory.value = item.category;
  editTags.value = item.tags.join(", ");
  editCoverImage.value = item.coverImage;
  editSlug.value = item.slug;
}

function beginEdit() {
  if (!selected.value) {
    return;
  }

  seedEdit(selected.value);
  editing.value = true;
}

function cancelEdit() {
  if (selected.value) {
    seedEdit(selected.value);
  }
  editing.value = false;
}

function editPayload(): SubmissionPayload {
  return {
    title: editTitle.value,
    summary: editSummary.value,
    content: editContent.value,
    category: editCategory.value,
    tags: editTags.value.split(/[,，]/).map((item) => item.trim()).filter(Boolean),
    coverImage: editCoverImage.value,
    slug: editSlug.value,
    submit: false
  };
}

async function saveEdit() {
  if (!selected.value) {
    return;
  }

  savingEdit.value = true;
  error.value = "";
  message.value = "";

  try {
    const updated = await updateAdminSubmission(selected.value.id, editPayload());
    selectedId.value = updated.id;
    publishSlug.value = updated.slug;
    publishCategory.value = updated.category;
    message.value = `已保存《${updated.title}》的审核修订。`;
    editing.value = false;
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿修订保存失败";
  } finally {
    savingEdit.value = false;
  }
}

async function upgradeAuthor() {
  if (!selected.value) {
    return;
  }

  upgradingAuthor.value = true;
  error.value = "";
  message.value = "";

  try {
    const user = await updateAdminUserRole(selected.value.authorId, "author");
    message.value = `已将 ${user.displayName} 升级为作者。`;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "作者升级失败";
  } finally {
    upgradingAuthor.value = false;
  }
}

function formatDate(value?: string) {
  if (!value) {
    return "未提交";
  }

  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function submissionTime(item: Submission) {
  return new Date(item.submittedAt || item.updatedAt || item.createdAt).getTime();
}

function statusText(value: Submission["status"]) {
  if (value === "submitted") {
    return "待审核";
  }
  if (value === "returned") {
    return "待复审";
  }
  if (value === "rejected") {
    return "已拒绝";
  }
  if (value === "published") {
    return "已发布";
  }
  return "草稿";
}

function statusClass(value: Submission["status"]) {
  if (value === "submitted" || value === "returned") {
    return "review";
  }
  if (value === "rejected") {
    return "rejected";
  }
  if (value === "published") {
    return "published";
  }
  return "draft";
}
</script>

<template>
  <AdminLayout title="投稿审核" description="审核登录用户提交的文章，确认质量后发布到正式内容库。" mobile-title="投稿审核" primary-action="通过发布">
    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" :disabled="acting || !selected" @click="review('return')">退回修改</button>
        <button class="button" type="button" :disabled="acting || !selected" @click="review('approve')">通过并发布</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="投稿统计">
      <div class="stat-card"><span>待审核</span><strong>{{ stats.submitted }}</strong></div>
      <div class="stat-card"><span>今日提交</span><strong>{{ submissions.length }}</strong></div>
      <div class="stat-card"><span>退回修改</span><strong>{{ stats.returned }}</strong></div>
      <div class="stat-card"><span>已发布</span><strong>{{ stats.published }}</strong></div>
    </section>

    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="admin-grid-2">
      <div class="settings-stack">
        <section class="table-panel">
          <form class="table-toolbar" @submit.prevent="load">
            <input v-model="searchQuery" class="input" type="search" placeholder="搜索投稿标题、投稿人、标签" aria-label="搜索投稿">
            <select v-model="filterStatus" class="input" aria-label="投稿状态" @change="load">
              <option value="">全部状态</option>
              <option value="submitted">待审核</option>
              <option value="returned">退回修改</option>
              <option value="published">已发布</option>
              <option value="rejected">已拒绝</option>
            </select>
            <select v-model="sortMode" class="input" aria-label="排序">
              <option value="latest">最近提交</option>
              <option value="risk">高风险优先</option>
              <option value="quality">高质量优先</option>
            </select>
          </form>

          <p v-if="loading" class="muted">正在加载投稿...</p>
          <table v-else>
            <thead>
              <tr>
                <th>投稿</th>
                <th>投稿人</th>
                <th>状态</th>
                <th>风险</th>
                <th>提交时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in visibleSubmissions" :key="item.id">
                <td>
                  <strong>{{ item.title }}</strong>
                  <div class="meta-row"><span>{{ item.category }}</span><span>{{ item.wordCount }} 字</span></div>
                </td>
                <td>{{ item.authorName }}<div class="meta-row"><span>版本 {{ item.version }}</span></div></td>
                <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
                <td>{{ item.riskLevel }}</td>
                <td>{{ formatDate(item.submittedAt) }}</td>
                <td><button class="button-secondary" type="button" @click="selectedId = item.id">查看</button></td>
              </tr>
              <tr v-if="visibleSubmissions.length === 0">
                <td colspan="6" class="muted">没有匹配的投稿。</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section v-if="selected" class="editor-panel">
          <div class="editor-toolbar">
            <div class="meta-row">
              <span class="tag">投稿预览</span>
              <span>{{ selected.category }}</span>
              <span>{{ selected.wordCount }} 字</span>
            </div>
            <button class="button-secondary" type="button" :disabled="savingEdit" @click="editing ? cancelEdit() : beginEdit()">{{ editing ? "退出编辑" : "编辑内容" }}</button>
          </div>
          <form v-if="editing" class="settings-stack" @submit.prevent="saveEdit">
            <div class="admin-grid-2">
              <div class="field"><label for="edit-title">标题</label><input v-model="editTitle" class="input" id="edit-title"></div>
              <div class="field"><label for="edit-slug">Slug</label><input v-model="editSlug" class="input" id="edit-slug"></div>
            </div>
            <div class="field"><label for="edit-summary">摘要</label><textarea v-model="editSummary" class="input" id="edit-summary"></textarea></div>
            <div class="field"><label for="edit-content">正文</label><textarea v-model="editContent" class="input" id="edit-content" style="min-height: 260px;"></textarea></div>
            <div class="admin-grid-2">
              <div class="field"><label for="edit-category">分类</label><input v-model="editCategory" class="input" id="edit-category"></div>
              <div class="field"><label for="edit-tags">标签</label><input v-model="editTags" class="input" id="edit-tags"></div>
            </div>
            <div class="field"><label for="edit-cover">封面图 URL</label><input v-model="editCoverImage" class="input" id="edit-cover"></div>
            <div class="header-actions">
              <button class="button" type="submit" :disabled="savingEdit || !editTitle">{{ savingEdit ? "保存中..." : "保存修订" }}</button>
              <button class="button-secondary" type="button" :disabled="savingEdit" @click="cancelEdit">取消</button>
            </div>
          </form>
          <article v-else class="preview-area" style="min-height: 420px;">
            <h1>{{ selected.title }}</h1>
            <p>{{ selected.summary }}</p>
            <p v-for="paragraph in previewParagraphs" :key="paragraph">{{ paragraph }}</p>
          </article>
        </section>
      </div>

      <aside v-if="selected" class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>审核动作</h2>
            <span class="status" :class="statusClass(selected.status)">{{ statusText(selected.status) }}</span>
          </div>
          <div class="settings-stack">
            <div class="field">
              <label for="review-note">审核意见</label>
              <textarea v-model="reviewNote" class="input" id="review-note"></textarea>
            </div>
            <button class="button" type="button" :disabled="acting" @click="review('approve')">通过并发布</button>
            <button class="button-secondary" type="button" :disabled="acting" @click="review('return')">退回修改</button>
            <button class="button-secondary" type="button" :disabled="acting" @click="review('reject')">拒绝投稿</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>投稿人</h2>
          </div>
          <div class="profile-hero">
            <span class="avatar">{{ selected.authorAvatar }}</span>
            <div>
              <strong>{{ selected.authorName }}</strong>
              <div class="meta-row"><span>注册用户</span><span>{{ selected.authorId }}</span></div>
            </div>
          </div>
          <div class="settings-stack" style="margin-top: 16px;">
            <div class="setting-row">
              <div>
                <strong>历史投稿</strong>
                <div class="meta-row"><span>统计会在用户模块完成后接入</span></div>
              </div>
            </div>
            <div class="setting-row">
              <div>
                <strong>评论质量</strong>
                <div class="meta-row"><span>评论审核模块会补充质量指标</span></div>
              </div>
            </div>
            <button class="button-secondary" type="button" :disabled="upgradingAuthor" @click="upgradeAuthor">{{ upgradingAuthor ? "升级中..." : "升级为作者" }}</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>发布设置</h2>
          </div>
          <div class="settings-stack">
            <div class="field"><label for="slug">Slug</label><input v-model="publishSlug" class="input" id="slug"></div>
            <div class="field"><label for="category">分类</label><select v-model="publishCategory" class="input" id="category"><option>用户系统</option><option>内容治理</option><option>工程实践</option><option>写作工作流</option></select></div>
            <div class="field"><label for="publish-time">发布时间</label><input class="input" id="publish-time" type="datetime-local"></div>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
