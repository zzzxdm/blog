<script setup lang="ts">
import { onMounted, ref } from "vue";

import AccountLayout from "../../components/AccountLayout.vue";
import {
  changePassword,
  getAccountSettings,
  requestEmailVerification,
  updateAccountSettings,
  verifyEmail,
  type AccountSettings
} from "../../shared/api";
import { useAuthStore } from "../../stores/auth";

const auth = useAuthStore();
const settings = ref<AccountSettings | null>(null);
const loading = ref(false);
const saving = ref(false);
const passwordSaving = ref(false);
const error = ref("");
const message = ref("");
const currentPassword = ref("");
const newPassword = ref("");
const verificationToken = ref("");
const verificationSaving = ref(false);

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    settings.value = await getAccountSettings();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "账号设置加载失败";
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
    const saved = await updateAccountSettings(settings.value);
    settings.value = saved;
    if (auth.user) {
      auth.user.displayName = saved.displayName;
      auth.user.avatarText = saved.avatarText;
    }
    message.value = "账号设置已保存。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "账号设置保存失败";
  } finally {
    saving.value = false;
  }
}

async function savePassword() {
  if (!currentPassword.value || !newPassword.value) {
    error.value = "请输入当前密码和新密码";
    return;
  }

  passwordSaving.value = true;
  error.value = "";
  message.value = "";

  try {
    await changePassword(currentPassword.value, newPassword.value);
    currentPassword.value = "";
    newPassword.value = "";
    message.value = "密码已更新。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "密码修改失败";
  } finally {
    passwordSaving.value = false;
  }
}

async function sendVerification() {
  verificationSaving.value = true;
  error.value = "";
  message.value = "";

  try {
    const response = await requestEmailVerification();
    verificationToken.value = response.verificationToken || "";
    message.value = response.verificationToken ? `验证 token：${response.verificationToken}` : "验证入口已发送。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "发送验证失败";
  } finally {
    verificationSaving.value = false;
  }
}

async function confirmVerification() {
  if (!verificationToken.value) {
    error.value = "请输入邮箱验证 token";
    return;
  }

  verificationSaving.value = true;
  error.value = "";
  message.value = "";

  try {
    const response = await verifyEmail(verificationToken.value);
    auth.user = response.user;
    verificationToken.value = "";
    message.value = "邮箱已验证。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "邮箱验证失败";
  } finally {
    verificationSaving.value = false;
  }
}
</script>

<template>
  <AccountLayout title="账号设置" description="管理公开资料、登录安全、消息偏好和账号数据。">
    <template #actions>
      <button class="button" type="button" :disabled="saving || !settings" @click="save">{{ saving ? "保存中..." : "保存设置" }}</button>
    </template>

    <p v-if="loading" class="muted">正在加载账号设置...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <template v-if="settings">
      <section class="stats-grid" aria-label="账号状态">
        <div class="stat-card"><span>安全等级</span><strong>{{ settings.securityLevel }}</strong></div>
        <div class="stat-card"><span>登录设备</span><strong>{{ settings.loginDeviceCount }}</strong></div>
        <div class="stat-card"><span>公开文章</span><strong>{{ settings.publicPostCount }}</strong></div>
        <div class="stat-card"><span>资料完整度</span><strong>{{ settings.profileCompleteness }}%</strong></div>
      </section>

      <section class="settings-grid">
        <section class="panel">
          <div class="panel-title"><h2>公开资料</h2></div>
          <div class="settings-stack">
            <div class="profile-hero"><span class="avatar">{{ settings.avatarText }}</span><div class="header-actions"><button class="button-secondary" type="button">更换头像</button><button class="button-secondary" type="button">移除</button></div></div>
            <div class="field"><label for="display-name">昵称</label><input v-model="settings.displayName" class="input" id="display-name"></div>
            <div class="field"><label for="username">用户名</label><input v-model="settings.username" class="input" id="username"></div>
            <div class="field"><label for="bio">个人简介</label><textarea v-model="settings.bio" class="input" id="bio"></textarea></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>登录安全</h2></div>
          <div class="settings-stack">
            <div class="field"><label for="email">登录邮箱</label><input v-model="settings.email" class="input" id="email" readonly></div>
            <div class="review-note">
              <strong>{{ auth.user?.emailVerified ? "邮箱已验证" : "邮箱未验证" }}</strong>
              <p>邮箱用于找回密码、重要账号事件提醒和后续两步验证。</p>
            </div>
            <div v-if="!auth.user?.emailVerified" class="field">
              <label for="verification-token">邮箱验证 token</label>
              <input v-model="verificationToken" class="input" id="verification-token">
            </div>
            <div v-if="!auth.user?.emailVerified" class="header-actions">
              <button class="button-secondary" type="button" :disabled="verificationSaving" @click="sendVerification">{{ verificationSaving ? "发送中..." : "发送验证" }}</button>
              <button class="button-secondary" type="button" :disabled="verificationSaving || !verificationToken" @click="confirmVerification">确认验证</button>
            </div>
            <div class="field"><label for="current-password">当前密码</label><input v-model="currentPassword" class="input" id="current-password" type="password" autocomplete="current-password"></div>
            <div class="field"><label for="new-password">新密码</label><input v-model="newPassword" class="input" id="new-password" type="password" autocomplete="new-password" placeholder="至少 6 位"></div>
            <button class="button-secondary" type="button" :disabled="passwordSaving || !currentPassword || !newPassword" @click="savePassword">{{ passwordSaving ? "更新中..." : "更新密码" }}</button>
            <label class="setting-row"><div><strong>两步验证</strong><div class="meta-row"><span>登录时需要邮箱验证码</span></div></div><input v-model="settings.twoFactor" type="checkbox"></label>
            <label class="setting-row"><div><strong>异常登录提醒</strong><div class="meta-row"><span>新设备登录时发送站内信</span></div></div><input v-model="settings.loginAlert" type="checkbox"></label>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>消息偏好</h2><RouterLink class="button-secondary" to="/account/messages">查看站内信</RouterLink></div>
          <div class="settings-stack">
            <label class="setting-row"><div><strong>投稿审核结果</strong><div class="meta-row"><span>通过、退回、拒绝都会发送站内信</span></div></div><input v-model="settings.notifyReview" type="checkbox"></label>
            <label class="setting-row"><div><strong>评论回复</strong><div class="meta-row"><span>有人回复或点赞你的评论时提醒</span></div></div><input v-model="settings.notifyComment" type="checkbox"></label>
            <label class="setting-row"><div><strong>站点公告</strong><div class="meta-row"><span>接收维护、规则和功能更新</span></div></div><input v-model="settings.notifyAnnouncement" type="checkbox"></label>
            <label class="setting-row"><div><strong>邮件提醒</strong><div class="meta-row"><span>重要账号事件同步发送邮件</span></div></div><input v-model="settings.emailNotification" type="checkbox"></label>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>隐私与展示</h2></div>
          <div class="settings-stack">
            <label class="setting-row"><div><strong>展示个人主页</strong><div class="meta-row"><span>公开显示投稿、简介和评论摘要</span></div></div><input v-model="settings.publicProfile" type="checkbox"></label>
            <label class="setting-row"><div><strong>展示收藏列表</strong><div class="meta-row"><span>其他用户可以查看公开收藏</span></div></div><input v-model="settings.publicBookmarks" type="checkbox"></label>
            <div class="field"><label for="profile-url">个人主页地址</label><input v-model="settings.profileUrl" class="input" id="profile-url"></div>
            <div class="field"><label for="timezone">时区</label><select v-model="settings.timezone" class="input" id="timezone"><option>Asia/Shanghai</option><option>UTC</option><option>Asia/Tokyo</option></select></div>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>登录设备</h2></div>
          <div class="timeline">
            <article class="timeline-item"><strong>Windows Chrome</strong><p>{{ settings.currentDeviceDescription }}</p><div class="meta-row"><span class="status published">已信任</span></div></article>
            <article class="timeline-item"><strong>iPhone Safari</strong><p>{{ settings.lastDeviceDescription }}</p><div class="meta-row"><button class="button-secondary" type="button">移除设备</button></div></article>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>账号数据</h2></div>
          <div class="settings-stack">
            <div class="review-note"><strong>注销前请先导出数据</strong><p>账号注销后，未发布草稿、收藏和个人设置将不可恢复；已发布内容会按站点规则保留署名记录。</p></div>
            <button class="button-secondary" type="button">导出我的数据</button>
            <button class="button-secondary" type="button">申请注销账号</button>
          </div>
        </section>
      </section>
    </template>
  </AccountLayout>
</template>
