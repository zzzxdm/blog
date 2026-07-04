<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  getAdminAuditLogs,
  type AuditLog
} from "../../shared/api";

const logs = ref<AuditLog[]>([]);
const loading = ref(false);
const error = ref("");
const page = ref(1);
const pageSize = 12;
const total = ref(0);
const action = ref("");
const resourceType = ref("");

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)));
const successCount = computed(() => logs.value.filter((item) => item.status === "success").length);
const blockedCount = computed(() => logs.value.filter((item) => item.status !== "success").length);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const response = await getAdminAuditLogs({
      page: page.value,
      pageSize,
      action: action.value,
      resourceType: resourceType.value
    });
    logs.value = response.items;
    total.value = response.total;
    page.value = response.page;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "操作日志加载失败";
  } finally {
    loading.value = false;
  }
}

async function applyFilters() {
  page.value = 1;
  await load();
}

async function nextPage() {
  if (page.value >= totalPages.value) {
    return;
  }
  page.value++;
  await load();
}

async function prevPage() {
  if (page.value <= 1) {
    return;
  }
  page.value--;
  await load();
}

function actionText(value: string) {
  const labels: Record<string, string> = {
    "settings.update": "更新设置",
    "settings.test_mail": "发送测试邮件",
    "backup.create": "创建备份",
    "stats.export": "导出统计",
    "comments.export": "导出评论",
    "messages.export": "导出站内信",
    "users.export": "导出用户",
    "navigation.update": "更新导航",
    "media.create": "上传媒体",
    "media.update": "更新媒体",
    "media.delete": "删除媒体",
    "post.create": "创建文章",
    "post.update": "更新文章",
    "post.publish": "发布文章",
    "post.restore": "恢复版本",
    "submission.review": "审核投稿",
    "comment.moderate": "处理评论",
    "user.update": "更新用户",
    "message.send": "发送站内信",
    "taxonomy.update": "更新分类标签"
  };

  return labels[value] || value;
}

function resourceText(value: string) {
  const labels: Record<string, string> = {
    settings: "设置",
    backup: "备份",
    stat: "统计",
    navigation: "导航",
    media: "媒体",
    post: "文章",
    submission: "投稿",
    comment: "评论",
    user: "用户",
    message: "站内信",
    taxonomy: "分类标签"
  };

  return labels[value] || value;
}

function statusText(value: AuditLog["status"]) {
  if (value === "blocked") return "已拦截";
  if (value === "error") return "失败";
  return "成功";
}

function statusClass(value: AuditLog["status"]) {
  if (value === "success") return "published";
  if (value === "blocked") return "muted";
  return "banned";
}

function formatDate(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}
</script>

<template>
  <AdminLayout title="操作日志" description="追踪后台关键写操作、操作者、资源、访问来源和处理结果。" mobile-title="操作日志" primary-action="刷新">
    <template #actions>
      <div class="header-actions">
        <select v-model="resourceType" class="input" aria-label="资源类型" @change="applyFilters">
          <option value="">全部资源</option>
          <option value="settings">设置</option>
          <option value="backup">备份</option>
          <option value="stat">统计</option>
          <option value="navigation">导航</option>
          <option value="post">文章</option>
          <option value="submission">投稿</option>
          <option value="comment">评论</option>
          <option value="media">媒体</option>
          <option value="user">用户</option>
          <option value="message">站内信</option>
          <option value="taxonomy">分类标签</option>
        </select>
        <select v-model="action" class="input" aria-label="动作类型" @change="applyFilters">
          <option value="">全部动作</option>
          <option value="settings.update">更新设置</option>
          <option value="settings.test_mail">发送测试邮件</option>
          <option value="backup.create">创建备份</option>
          <option value="stats.export">导出统计</option>
          <option value="comments.export">导出评论</option>
          <option value="messages.export">导出站内信</option>
          <option value="users.export">导出用户</option>
          <option value="post.publish">发布文章</option>
          <option value="post.restore">恢复版本</option>
          <option value="submission.review">审核投稿</option>
          <option value="comment.moderate">处理评论</option>
          <option value="media.create">上传媒体</option>
          <option value="user.update">更新用户</option>
        </select>
        <button class="button-secondary" type="button" :disabled="loading" @click="load">刷新</button>
      </div>
    </template>

    <section class="stats-grid" aria-label="操作日志概览">
      <div class="stat-card"><span>日志总数</span><strong>{{ total }}</strong><div class="meta-row"><span>当前筛选结果</span></div></div>
      <div class="stat-card"><span>本页成功</span><strong>{{ successCount }}</strong><div class="meta-row"><span>状态为成功</span></div></div>
      <div class="stat-card"><span>本页异常</span><strong>{{ blockedCount }}</strong><div class="meta-row"><span>拦截或失败</span></div></div>
    </section>

    <p v-if="loading" class="muted">正在加载操作日志...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <section v-else class="table-panel">
      <table>
        <thead>
          <tr>
            <th>时间</th>
            <th>操作者</th>
            <th>动作</th>
            <th>资源</th>
            <th>状态</th>
            <th>来源</th>
            <th>详情</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in logs" :key="item.id">
            <td>{{ formatDate(item.createdAt) }}</td>
            <td><strong>{{ item.actorName }}</strong><div class="meta-row"><span>{{ item.actorId || "未登录" }}</span></div></td>
            <td>{{ actionText(item.action) }}</td>
            <td><strong>{{ resourceText(item.resourceType) }}</strong><div class="meta-row"><span>{{ item.resourceTitle || item.resourceId || "-" }}</span></div></td>
            <td><span class="status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
            <td>{{ item.ip }}</td>
            <td><div class="meta-row"><span>{{ item.detail }}</span></div></td>
          </tr>
          <tr v-if="!logs.length">
            <td colspan="7">暂无操作日志</td>
          </tr>
        </tbody>
      </table>

      <div class="pagination">
        <button class="button-secondary" type="button" :disabled="page <= 1 || loading" @click="prevPage">上一页</button>
        <span>第 {{ page }} / {{ totalPages }} 页</span>
        <button class="button-secondary" type="button" :disabled="page >= totalPages || loading" @click="nextPage">下一页</button>
      </div>
    </section>
  </AdminLayout>
</template>
