<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import {
  createAdminBackup,
  getAdminSettings,
  sendAdminTestMail,
  updateAdminSettings,
  type OperationsSettings
} from "../../shared/api";

const settings = ref<OperationsSettings | null>(null);
const blockedWordsText = ref("");
const loading = ref(false);
const saving = ref(false);
const testingMail = ref(false);
const runningBackup = ref(false);
const error = ref("");
const message = ref("");

onMounted(load);

watch(settings, (value) => {
  blockedWordsText.value = value?.blockedWords.join(", ") || "";
});

async function load() {
  loading.value = true;
  error.value = "";

  try {
    settings.value = await getAdminSettings();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "设置加载失败";
  } finally {
    loading.value = false;
  }
}

async function save() {
  if (!settings.value) {
    return;
  }

  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    const payload = {
      ...settings.value,
      blockedWords: blockedWordsText.value.split(/[,，\n]/).map((item) => item.trim()).filter(Boolean)
    };
    settings.value = await updateAdminSettings(payload);
    message.value = "设置已保存。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "设置保存失败";
  } finally {
    saving.value = false;
  }
}

async function testMail() {
  testingMail.value = true;
  error.value = "";
  message.value = "";

  try {
    const result = await sendAdminTestMail();
    message.value = result.message;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "测试邮件生成失败";
  } finally {
    testingMail.value = false;
  }
}

async function runBackup() {
  runningBackup.value = true;
  error.value = "";
  message.value = "";

  try {
    const result = await createAdminBackup();
    settings.value = result.settings;
    message.value = `${result.message} 文件：${result.fileName}`;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "备份失败";
  } finally {
    runningBackup.value = false;
  }
}
</script>

<template>
  <AdminLayout title="系统设置" description="配置站点基础信息、评论策略、邮件服务、安全和备份。" mobile-title="系统设置" primary-action="保存">
    <template #mobile-action>
      <button class="button" type="button" :disabled="saving || !settings" @click="save">{{ saving ? "保存中..." : "保存" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" @click="load">重新加载</button>
        <button class="button" type="button" :disabled="saving || !settings" @click="save">{{ saving ? "保存中..." : "保存设置" }}</button>
      </div>
    </template>

    <p v-if="loading" class="muted">正在加载设置...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section v-if="settings" class="settings-grid">
      <section class="panel">
        <div class="panel-title"><h2>站点信息</h2></div>
        <div class="settings-stack">
          <div class="field"><label for="site-name">站点名称</label><input v-model="settings.siteName" class="input" id="site-name"></div>
          <div class="field"><label for="site-desc">站点描述</label><textarea v-model="settings.siteDescription" class="input" id="site-desc"></textarea></div>
          <div class="field"><label for="site-url">站点域名</label><input v-model="settings.siteUrl" class="input" id="site-url"></div>
          <div class="field"><label for="beian">备案信息</label><input v-model="settings.beian" class="input" id="beian"></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>主题外观</h2></div>
        <div class="settings-stack">
          <div class="field"><label>主色</label><div class="swatches"><button v-for="color in ['#295b4b', '#b95f2d', '#e3b45d']" :key="color" class="swatch" :style="{ background: color }" type="button" :aria-label="color" @click="settings.themePrimary = color"></button></div></div>
          <div class="field"><label for="homepage-layout">首页布局</label><select v-model="settings.homepageLayout" class="input" id="homepage-layout"><option>精选文章 + 最新列表</option><option>专题优先</option><option>极简文章流</option></select></div>
          <label class="setting-row"><div><strong>深色模式</strong><div class="meta-row"><span>允许读者切换浅色和深色</span></div></div><input v-model="settings.darkModeEnabled" type="checkbox"></label>
          <label class="setting-row"><div><strong>显示阅读进度</strong><div class="meta-row"><span>文章页顶部展示阅读进度条</span></div></div><input v-model="settings.readingProgressEnabled" type="checkbox"></label>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>评论策略</h2></div>
        <div class="settings-stack">
          <label class="setting-row"><div><strong>开启评论</strong><div class="meta-row"><span>文章默认允许评论</span></div></div><input v-model="settings.commentsEnabled" type="checkbox"></label>
          <label class="setting-row"><div><strong>仅登录用户评论</strong><div class="meta-row"><span>访客只读，降低垃圾评论</span></div></div><input v-model="settings.loginRequiredForComment" type="checkbox"></label>
          <label class="setting-row"><div><strong>评论自动通过</strong><div class="meta-row"><span>开启后新评论直接展示，关闭后进入待审核</span></div></div><input v-model="settings.autoApproveComments" type="checkbox"></label>
          <div class="field"><label for="blocked-words">屏蔽关键词</label><textarea v-model="blockedWordsText" class="input" id="blocked-words"></textarea></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>投稿策略</h2></div>
        <div class="settings-stack">
          <label class="setting-row"><div><strong>开放用户投稿</strong><div class="meta-row"><span>登录用户可以提交文章审核</span></div></div><input v-model="settings.submissionsEnabled" type="checkbox"></label>
          <label class="setting-row"><div><strong>固定人工审核</strong><div class="meta-row"><span>通过后才进入公开文章列表</span></div></div><input v-model="settings.submissionManualReview" type="checkbox" disabled></label>
          <div class="field"><label for="submission-limit">投稿频率限制</label><select v-model="settings.submissionLimit" class="input" id="submission-limit"><option>每天最多 3 篇</option><option>每天最多 1 篇</option><option>每周最多 3 篇</option></select></div>
          <div class="field"><label for="submission-guide">投稿说明</label><textarea v-model="settings.submissionGuide" class="input" id="submission-guide"></textarea></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>邮件与 RSS</h2></div>
        <div class="settings-stack">
          <label class="setting-row"><div><strong>邮件推送策略预留</strong><div class="meta-row"><span>发布推送接入后启用</span></div></div><input v-model="settings.mailEnabled" type="checkbox"></label>
          <div class="field"><label for="mail-provider">邮件服务</label><select v-model="settings.mailProvider" class="input" id="mail-provider"><option>Resend</option><option>SendGrid</option><option>SMTP</option></select></div>
          <div class="field"><label for="from-email">发件邮箱</label><input v-model="settings.fromEmail" class="input" id="from-email"></div>
          <button class="button-secondary" type="button" :disabled="testingMail" @click="testMail">{{ testingMail ? "生成中..." : "生成测试邮件" }}</button>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>安全</h2></div>
        <div class="settings-stack">
          <label class="setting-row"><div><strong>管理员 2FA 策略预留</strong><div class="meta-row"><span>完整登录挑战接入后再强制启用</span></div></div><input v-model="settings.adminTwoFactorRequired" type="checkbox"></label>
          <label class="setting-row"><div><strong>登录失败锁定</strong><div class="meta-row"><span>连续失败后临时锁定账号</span></div></div><input v-model="settings.loginFailureLock" type="checkbox"></label>
          <div class="field"><label for="session-days">会话有效期</label><select v-model.number="settings.sessionDays" class="input" id="session-days"><option :value="7">7 天</option><option :value="14">14 天</option><option :value="30">30 天</option></select></div>
        </div>
      </section>

      <section class="panel">
        <div class="panel-title"><h2>备份</h2><span class="status published">正常</span></div>
        <div class="settings-stack">
          <div class="field"><label for="backup-cycle">备份计划频率</label><select v-model="settings.backupCycle" class="input" id="backup-cycle"><option>每日全量备份</option><option>每周全量备份</option><option>手动备份</option></select></div>
          <div class="meta-row"><span>上次备份：{{ new Date(settings.lastBackupAt).toLocaleString("zh-CN") }}</span><span>保留 {{ settings.backupRetentionDays }} 天</span></div>
          <button class="button-secondary" type="button" :disabled="runningBackup" @click="runBackup">{{ runningBackup ? "生成中..." : "生成备份记录" }}</button>
        </div>
      </section>
    </section>
  </AdminLayout>
</template>
