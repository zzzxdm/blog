<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import MarkdownPreview from "../../components/MarkdownPreview.vue";
import MarkdownThemeSwitcher from "../../components/MarkdownThemeSwitcher.vue";
import PaginationControls from "../../components/PaginationControls.vue";
import RichMarkdownEditor from "../../components/RichMarkdownEditor.vue";
import {
  getAdminUsers,
  getAdminSubmissions,
  archiveSubmission,
  restoreSubmission,
  reviewSubmission,
  updateAdminSubmission,
  updateAdminUserRole,
  type ManagedUser,
  type Submission,
  type SubmissionPayload,
  type SubmissionStats
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { useMarkdownPreviewTheme } from "../../shared/markdownPreview";
import { useConfirmStore } from "../../stores/confirm";
import { useToastStore } from "../../stores/toast";

const confirmDialog = useConfirmStore();
const toast = useToastStore();

const submissions = ref<Submission[]>([]);
const allSubmissions = ref<Submission[]>([]);
const users = ref<ManagedUser[]>([]);
const stats = ref<SubmissionStats>({ draft: 0, submitted: 0, returned: 0, rejected: 0, published: 0, archived: 0, total: 0 });
const selectedId = ref("");
const filterStatus = ref("submitted");
const loading = ref(false);
const acting = ref(false);
const editing = ref(false);
const savingEdit = ref(false);
const upgradingAuthor = ref(false);
const archiving = ref(false);
const restoring = ref(false);
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
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const { selectedPreviewTheme, selectedCodeTheme } = useMarkdownPreviewTheme();

const selected = computed(() => submissions.value.find((item) => item.id === selectedId.value) || submissions.value[0]);
const selectedAuthorUser = computed(() => users.value.find((user) => user.id === selected.value?.authorId));
const authorSubmissionStats = computed(() => {
  const authorId = selected.value?.authorId || "";
  const result = { total: 0, submitted: 0, returned: 0, rejected: 0, published: 0, archived: 0, draft: 0 };
  allSubmissions.value.forEach((item) => {
    if (item.authorId !== authorId) {
      return;
    }

    result.total++;
    result[item.status]++;
  });
  return result;
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
    const filteredPromise = getAdminSubmissions({
      status: filterStatus.value,
      q: searchQuery.value,
      sort: sortMode.value,
      page: page.value,
      pageSize: pageSize.value
    });
    const allPromise = getAdminSubmissions({ all: true });
    const [response, allResponse, userResponse] = await Promise.all([filteredPromise, allPromise, getAdminUsers({ all: true })]);
    submissions.value = response.items;
    allSubmissions.value = allResponse.items;
    users.value = userResponse.items;
    stats.value = response.stats;
    total.value = response.total;
    page.value = response.page;
    pageSize.value = response.pageSize;
    if (!submissions.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = submissions.value[0]?.id || "";
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿审核列表加载失败";
    toast.error("投稿审核列表加载失败", error.value);
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
    toast.success(action === "approve" ? "投稿已通过" : "审核结果已发送", message.value);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "审核操作失败";
    toast.error("审核操作失败", error.value);
  } finally {
    acting.value = false;
  }
}

async function archiveSelected() {
  if (!selected.value?.publishedPostSlug) {
    return;
  }
  const confirmed = await confirmDialog.open({
    title: "下架已发布文章",
    message: `确认下架《${selected.value.title}》吗？下架后公开列表和搜索中将不再展示。`,
    confirmText: "下架文章",
    tone: "danger"
  });
  if (!confirmed) {
    return;
  }

  archiving.value = true;
  error.value = "";
  message.value = "";
  try {
    await archiveSubmission(selected.value.id);
    message.value = `已下架《${selected.value.title}》。`;
    toast.success("文章已下架", selected.value.title);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章下架失败";
    toast.error("文章下架失败", error.value);
  } finally {
    archiving.value = false;
  }
}

async function restoreSelected() {
  if (!selected.value?.publishedPostSlug) {
    return;
  }
  const confirmed = await confirmDialog.open({
    title: "重新上架文章",
    message: `确认重新上架《${selected.value.title}》吗？上架后会重新进入公开访问和搜索范围。`,
    confirmText: "重新上架",
    tone: "success"
  });
  if (!confirmed) {
    return;
  }

  restoring.value = true;
  error.value = "";
  message.value = "";
  try {
    await restoreSubmission(selected.value.id);
    message.value = `已重新上架《${selected.value.title}》。`;
    toast.success("文章已重新上架", selected.value.title);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "文章重新上架失败";
    toast.error("重新上架失败", error.value);
  } finally {
    restoring.value = false;
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
  toast.info("已进入审核编辑", selected.value.title);
}

function cancelEdit() {
  if (selected.value) {
    seedEdit(selected.value);
  }
  editing.value = false;
  toast.info("已退出审核编辑", selected.value?.title || "当前投稿");
}

function selectSubmission(item: Submission) {
  selectedId.value = item.id;
  toast.info("已打开投稿", item.title);
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
    visibility: selected.value?.visibility || "public",
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
    toast.success("审核修订已保存", updated.title);
    editing.value = false;
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "投稿修订保存失败";
    toast.error("投稿修订保存失败", error.value);
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
    toast.success("作者已升级", user.displayName);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "作者升级失败";
    toast.error("作者升级失败", error.value);
  } finally {
    upgradingAuthor.value = false;
  }
}

function formatDate(value?: string) {
  return formatDateTime(value, "未提交");
}

function visibilityText(value: Submission["visibility"]) {
  return value === "private" ? "私密" : "公开";
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
  if (value === "archived") {
    return "已下架";
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
  if (value === "archived") {
    return "muted";
  }
  return "draft";
}

function userStatusText(value?: ManagedUser["status"]) {
  if (value === "muted") return "已禁言";
  if (value === "banned") return "已封禁";
  if (value === "deleted") return "已删除";
  return "账号正常";
}
</script>

<template>
  <AdminLayout title="投稿审核" description="审核登录用户提交的文章，确认质量后发布到正式内容库。" mobile-title="投稿审核" primary-action="通过发布">
    <template #mobile-action>
      <button class="button" type="button" :disabled="acting || !selected" @click="review('approve')">
        {{ acting ? "发布中..." : "通过发布" }}
      </button>
    </template>

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
          <form class="table-toolbar" @submit.prevent="applyFilters">
            <input v-model="searchQuery" class="input" type="search" placeholder="搜索投稿标题、投稿人、标签" aria-label="搜索投稿">
            <select v-model="filterStatus" class="input" aria-label="投稿状态" @change="applyFilters">
              <option value="">全部状态</option>
              <option value="submitted">待审核</option>
              <option value="returned">退回修改</option>
	              <option value="published">已发布</option>
	              <option value="archived">已下架</option>
              <option value="rejected">已拒绝</option>
            </select>
            <select v-model="sortMode" class="input" aria-label="排序" @change="applyFilters">
              <option value="latest">最近提交</option>
              <option value="risk">高风险优先</option>
              <option value="quality">高质量优先</option>
            </select>
          </form>

          <LoadingState v-if="loading" variant="table" text="正在加载投稿..." :rows="4" />
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
              <tr v-for="item in submissions" :key="item.id">
                <td>
                  <strong>{{ item.title }}</strong>
                  <div class="meta-row"><span>{{ item.category }}</span><span>{{ visibilityText(item.visibility) }}</span><span>{{ item.wordCount }} 字</span></div>
                </td>
                <td>{{ item.authorName }}<div class="meta-row"><span>版本 {{ item.version }}</span></div></td>
                <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
                <td>{{ item.riskLevel }}</td>
                <td>{{ formatDate(item.submittedAt) }}</td>
                <td><button class="button-secondary" type="button" @click="selectSubmission(item)">查看</button></td>
              </tr>
              <tr v-if="submissions.length === 0">
                <td colspan="6" class="muted">没有匹配的投稿。</td>
              </tr>
            </tbody>
          </table>
          <PaginationControls
            v-if="!loading"
            :page="page"
            :page-size="pageSize"
            :total="total"
            :loading="loading"
            item-label="篇投稿"
            show-page-size
            :page-size-options="[5, 10, 20, 50, 100]"
            @update:page="setPage"
            @update:page-size="setPageSize"
          />
        </section>

        <section v-if="selected" class="editor-panel">
          <div class="editor-toolbar">
            <div class="meta-row">
	              <span class="tag">投稿预览</span>
	              <span>{{ selected.category }}</span>
	              <span>{{ visibilityText(selected.visibility) }}</span>
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
            <div class="field">
              <label>正文</label>
              <RichMarkdownEditor
                v-model="editContent"
                editor-id="admin-submission-editor"
                height="420px"
                upload-category="投稿修订插图"
                placeholder="修订投稿正文，支持 Markdown、实时预览、表情、粘贴或拖拽上传图片。"
                @upload-error="(value) => { error = value; message = ''; }"
              />
            </div>
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
            <MarkdownThemeSwitcher v-model:preview-theme="selectedPreviewTheme" v-model:code-theme="selectedCodeTheme" />
            <MarkdownPreview
              :content="selected.content"
              :preview-id="`admin-submission-preview-${selected.id}`"
              :preview-theme="selectedPreviewTheme"
              :code-theme="selectedCodeTheme"
            />
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
            <button v-if="selected.status === 'published' && selected.publishedPostSlug" class="button-secondary" type="button" :disabled="archiving" @click="archiveSelected">{{ archiving ? "下架中..." : "下架文章" }}</button>
            <button v-if="selected.status === 'archived' && selected.publishedPostSlug" class="button-secondary" type="button" :disabled="restoring" @click="restoreSelected">{{ restoring ? "上架中..." : "重新上架" }}</button>
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
                <div class="meta-row">
                  <span>共 {{ authorSubmissionStats.total }} 篇</span>
                  <span>已发布 {{ authorSubmissionStats.published }}</span>
                  <span>退回 {{ authorSubmissionStats.returned }}</span>
                </div>
              </div>
            </div>
            <div class="setting-row">
              <div>
                <strong>评论质量</strong>
                <div class="meta-row">
                  <span>{{ selectedAuthorUser ? `${selectedAuthorUser.commentCount} 条评论` : "暂无评论数据" }}</span>
                  <span>{{ userStatusText(selectedAuthorUser?.status) }}</span>
                </div>
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
