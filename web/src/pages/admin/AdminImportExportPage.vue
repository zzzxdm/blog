<script setup lang="ts">
import { ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  createAdminExportJob,
  createAdminBackup,
  exportAdminComments,
  exportAdminMessages,
  exportAdminStats,
  exportAdminUsers,
  type AdminJob,
  type BackupResult
} from "../../shared/api";
import { formatDateTime } from "../../shared/datetime";
import { downloadJson, exportFileName } from "../../shared/download";
import { useToastStore } from "../../stores/toast";

const toast = useToastStore();
const runningKey = ref("");
const error = ref("");
const message = ref("");
const backup = ref<BackupResult | null>(null);
const jobs = ref<AdminJob[]>([]);
const importScope = ref("posts");
const importFileName = ref("import-posts.json");

async function runExport(key: string) {
  runningKey.value = key;
  error.value = "";
  message.value = "";

  try {
    const job = await createAdminExportJob({ scope: key });
    jobs.value.unshift(job);
    if (key === "users") {
      downloadJson(exportFileName("users"), await exportAdminUsers());
    }
    if (key === "comments") {
      downloadJson(exportFileName("comments"), await exportAdminComments());
    }
    if (key === "messages") {
      downloadJson(exportFileName("messages"), await exportAdminMessages());
    }
    if (key === "stats") {
      downloadJson(exportFileName("stats"), await exportAdminStats("30d"));
    }
    message.value = "导出文件已生成。";
    toast.success("导出文件已生成", exportTitle(key));
  } catch (err) {
    error.value = err instanceof Error ? err.message : "导出失败";
    toast.error("导出失败", error.value);
  } finally {
    runningKey.value = "";
  }
}

async function runBackup() {
  runningKey.value = "backup";
  error.value = "";
  message.value = "";

  try {
    backup.value = await createAdminBackup();
    message.value = backup.value.message;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "备份任务创建失败";
  } finally {
    runningKey.value = "";
  }
}

function createImport() {
  error.value = "";
  message.value = "批量导入任务暂未开放，请使用文章管理中的 Markdown 导入。";
  toast.info("批量导入暂未开放", "文章 Markdown 导入请在文章管理页面使用");
}

function exportTitle(key: string) {
  if (key === "users") return "用户数据";
  if (key === "comments") return "评论数据";
  if (key === "messages") return "站内信数据";
  if (key === "stats") return "统计报表";
  return key;
}
</script>

<template>
  <AdminLayout title="导入导出" description="集中处理运营数据导出和备份记录，便于上线前检查与迁移。" mobile-title="导入导出">
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="admin-grid-2">
      <div class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>数据导出</h2>
            <span class="tag">JSON</span>
          </div>
          <div class="settings-stack">
            <article class="setting-row">
              <div><strong>用户数据</strong><div class="meta-row"><span>用户状态、角色和统计摘要</span></div></div>
              <button class="button-secondary" type="button" :disabled="!!runningKey" @click="runExport('users')">{{ runningKey === "users" ? "导出中..." : "导出" }}</button>
            </article>
            <article class="setting-row">
              <div><strong>评论数据</strong><div class="meta-row"><span>评论审核、作者和文章关联</span></div></div>
              <button class="button-secondary" type="button" :disabled="!!runningKey" @click="runExport('comments')">{{ runningKey === "comments" ? "导出中..." : "导出" }}</button>
            </article>
            <article class="setting-row">
              <div><strong>站内信数据</strong><div class="meta-row"><span>系统通知、审核消息和管理员消息</span></div></div>
              <button class="button-secondary" type="button" :disabled="!!runningKey" @click="runExport('messages')">{{ runningKey === "messages" ? "导出中..." : "导出" }}</button>
            </article>
            <article class="setting-row">
              <div><strong>统计报表</strong><div class="meta-row"><span>近 30 天阅读、互动和热门内容</span></div></div>
              <button class="button-secondary" type="button" :disabled="!!runningKey" @click="runExport('stats')">{{ runningKey === "stats" ? "导出中..." : "导出" }}</button>
            </article>
          </div>
        </section>
      </div>

      <aside class="settings-stack">
        <section class="panel">
          <div class="panel-title">
            <h2>备份任务</h2>
          </div>
          <p class="muted">生成数据库备份记录，生产环境由部署脚本负责实际落盘与恢复。</p>
          <button class="button" type="button" :disabled="!!runningKey" @click="runBackup">{{ runningKey === "backup" ? "生成中..." : "生成备份记录" }}</button>
          <div v-if="backup" class="review-note">
            <strong>{{ backup.fileName }}</strong>
            <p>{{ backup.sizeLabel }} · {{ backup.status }} · {{ formatDateTime(backup.createdAt) }}</p>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>导入入口</h2>
            <span class="tag">暂未开放</span>
          </div>
          <p class="muted">批量导入队列暂未开放；文章 Markdown 导入请在文章管理页面使用。</p>
          <div class="settings-stack">
            <div class="field">
              <label for="import-scope">导入范围</label>
              <select v-model="importScope" class="input" id="import-scope">
                <option value="posts">文章</option>
                <option value="taxonomies">分类标签</option>
                <option value="media">媒体元数据</option>
                <option value="site">站点配置</option>
              </select>
            </div>
            <div class="field"><label for="import-file">文件名</label><input v-model="importFileName" class="input" id="import-file"></div>
            <button class="button-secondary" type="button" :disabled="!!runningKey" @click="createImport">暂未开放</button>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title">
            <h2>任务队列</h2>
            <span class="tag">{{ jobs.length }} 个任务</span>
          </div>
          <div class="timeline">
            <p v-if="!jobs.length" class="muted">暂无导入导出任务。</p>
            <article v-for="job in jobs" :key="job.id" class="timeline-item">
              <strong>{{ job.type }} · {{ job.scope }} · {{ job.status }}</strong>
              <p>{{ job.message }}</p>
              <div class="meta-row">
                <span>{{ job.progress }}%</span>
                <span>{{ job.fileName }}</span>
                <span>{{ formatDateTime(job.updatedAt) }}</span>
              </div>
            </article>
          </div>
        </section>
      </aside>
    </section>
  </AdminLayout>
</template>
