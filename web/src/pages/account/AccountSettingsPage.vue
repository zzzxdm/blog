<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";

import AccountLayout from "../../components/AccountLayout.vue";
import {
  changePassword,
  deleteMyAccount,
  deleteMySession,
  exportMyData,
  getAccountSettings,
  getMySessions,
  requestEmailVerification,
  updateAccountSettings,
  verifyEmail,
  type AccountSettings,
  type SessionInfo
} from "../../shared/api";
import { useAuthStore } from "../../stores/auth";
import { useToastStore } from "../../stores/toast";

const router = useRouter();
const auth = useAuthStore();
const toast = useToastStore();
const settings = ref<AccountSettings | null>(null);
const sessions = ref<SessionInfo[]>([]);
const loading = ref(false);
const saving = ref(false);
const passwordSaving = ref(false);
const sessionActingId = ref("");
const exporting = ref(false);
const deletingAccount = ref(false);
const error = ref("");
const message = ref("");
const currentPassword = ref("");
const newPassword = ref("");
const verificationToken = ref("");
const verificationSaving = ref(false);
const persistentMessage = computed(() => Boolean(verificationToken.value && message.value));

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    const [accountSettings, sessionResult] = await Promise.all([getAccountSettings(), getMySessions()]);
    settings.value = accountSettings;
    sessions.value = sessionResult.items;
    if (auth.user) {
      auth.user.emailVerified = accountSettings.emailVerified;
    }
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
  settings.value.avatarText = normalizeAvatarText(settings.value.avatarText, settings.value.displayName);

  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    const saved = await updateAccountSettings(settings.value);
    settings.value = saved;
    if (auth.user) {
      auth.user.displayName = saved.displayName;
      auth.user.avatarText = saved.avatarText;
      auth.user.emailVerified = saved.emailVerified;
    }
    message.value = "账号设置已保存。";
    toast.success("账号设置已保存", "公开资料和偏好已更新。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "账号设置保存失败";
  } finally {
    saving.value = false;
  }
}

function resetAvatarText() {
  if (!settings.value) {
    return;
  }

  settings.value.avatarText = normalizeAvatarText("", settings.value.displayName);
  message.value = "头像已重置为昵称首字，保存设置后生效。";
  error.value = "";
  toast.info("头像已重置", "保存设置后生效。");
}

function normalizeAvatarText(value: string, displayName: string) {
  const text = Array.from(value.trim()).slice(0, 2).join("");
  if (text) {
    return text;
  }

  return Array.from(displayName.trim() || "用")[0];
}

function trimAvatarText() {
  if (!settings.value) {
    return;
  }

  settings.value.avatarText = Array.from(settings.value.avatarText.trim()).slice(0, 2).join("");
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
    toast.success("密码已更新", "下次登录请使用新密码。");
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
    if (!response.verificationToken) {
      toast.success("验证入口已发送", "请检查邮箱收件箱。");
    }
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
    if (settings.value) {
      settings.value.emailVerified = response.user.emailVerified;
    }
    verificationToken.value = "";
    message.value = "邮箱已验证。";
    toast.success("邮箱已验证", "现在可以投稿、评论和收藏。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "邮箱验证失败";
  } finally {
    verificationSaving.value = false;
  }
}

async function removeSession(session: SessionInfo) {
  if (session.current || !window.confirm("移除这个登录设备？")) {
    return;
  }

  sessionActingId.value = session.id;
  error.value = "";
  message.value = "";

  try {
    await deleteMySession(session.id);
    message.value = "登录设备已移除。";
    toast.success("登录设备已移除", session.device);
    await load();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "登录设备移除失败";
  } finally {
    sessionActingId.value = "";
  }
}

async function exportData() {
  exporting.value = true;
  error.value = "";
  message.value = "";

  try {
    const data = await exportMyData();
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `blog-account-export-${new Date().toISOString().slice(0, 10)}.json`;
    link.click();
    URL.revokeObjectURL(url);
    message.value = "账号数据已导出。";
    toast.success("账号数据已导出", "下载文件已生成。");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "账号数据导出失败";
  } finally {
    exporting.value = false;
  }
}

async function deleteAccount() {
  if (!window.confirm("确认申请注销账号？该操作会退出当前账号。")) {
    return;
  }

  deletingAccount.value = true;
  error.value = "";
  message.value = "";

  try {
    await deleteMyAccount();
    auth.user = null;
    toast.success("账号已注销", "已退出当前账号。");
    await router.push("/");
  } catch (err) {
    error.value = err instanceof Error ? err.message : "账号注销失败";
  } finally {
    deletingAccount.value = false;
  }
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
  <AccountLayout title="账号设置" description="管理公开资料、登录安全、消息偏好和账号数据。">
    <template #actions>
      <button class="button" type="button" :disabled="saving || !settings" @click="save">{{ saving ? "保存中..." : "保存设置" }}</button>
    </template>

    <p v-if="loading" class="muted">正在加载账号设置...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="persistentMessage" class="muted">{{ message }}</p>

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
            <div class="profile-hero"><span class="avatar">{{ settings.avatarText || normalizeAvatarText("", settings.displayName) }}</span><div class="header-actions"><button class="button-secondary" type="button" @click="resetAvatarText">重置头像</button></div></div>
            <div class="field"><label for="avatar-text">头像文字</label><input v-model="settings.avatarText" class="input" id="avatar-text" maxlength="2" @input="trimAvatarText"></div>
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
              <strong>{{ settings.emailVerified ? "邮箱已验证" : "邮箱未验证" }}</strong>
              <p>邮箱用于找回密码、重要账号事件提醒和后续两步验证。</p>
            </div>
            <div v-if="!settings.emailVerified" class="field">
              <label for="verification-token">邮箱验证 token</label>
              <input v-model="verificationToken" class="input" id="verification-token">
            </div>
            <div v-if="!settings.emailVerified" class="header-actions">
              <button class="button-secondary" type="button" :disabled="verificationSaving" @click="sendVerification">{{ verificationSaving ? "发送中..." : "发送验证" }}</button>
              <button class="button-secondary" type="button" :disabled="verificationSaving || !verificationToken" @click="confirmVerification">确认验证</button>
            </div>
            <div class="field"><label for="current-password">当前密码</label><input v-model="currentPassword" class="input" id="current-password" type="password" autocomplete="current-password"></div>
            <div class="field"><label for="new-password">新密码</label><input v-model="newPassword" class="input" id="new-password" type="password" autocomplete="new-password" placeholder="至少 6 位"></div>
            <button class="button-secondary" type="button" :disabled="passwordSaving || !currentPassword || !newPassword" @click="savePassword">{{ passwordSaving ? "更新中..." : "更新密码" }}</button>
            <label class="setting-row"><div><strong>两步验证资料预留</strong><div class="meta-row"><span>登录验证码挑战接入后启用</span></div></div><input v-model="settings.twoFactor" type="checkbox" disabled></label>
            <label class="setting-row"><div><strong>异常登录提醒预留</strong><div class="meta-row"><span>新设备提醒接入后发送站内信</span></div></div><input v-model="settings.loginAlert" type="checkbox" disabled></label>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>消息偏好</h2><RouterLink class="button-secondary" to="/account/messages">查看站内信</RouterLink></div>
          <div class="settings-stack">
            <label class="setting-row"><div><strong>固定接收投稿审核结果</strong><div class="meta-row"><span>通过、退回、拒绝都会发送站内信</span></div></div><input v-model="settings.notifyReview" type="checkbox" disabled></label>
            <label class="setting-row"><div><strong>评论回复提醒预留</strong><div class="meta-row"><span>回复和点赞通知接入后启用</span></div></div><input v-model="settings.notifyComment" type="checkbox" disabled></label>
            <label class="setting-row"><div><strong>固定接收站点公告</strong><div class="meta-row"><span>维护、规则和功能更新通过站内信发送</span></div></div><input v-model="settings.notifyAnnouncement" type="checkbox" disabled></label>
            <label class="setting-row"><div><strong>邮件提醒偏好预留</strong><div class="meta-row"><span>邮件服务接入后同步重要账号事件</span></div></div><input v-model="settings.emailNotification" type="checkbox" disabled></label>
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
            <article v-for="session in sessions" :key="session.id" class="timeline-item">
              <strong>{{ session.device }}</strong>
              <p>{{ session.current ? settings.currentDeviceDescription : settings.lastDeviceDescription }}</p>
              <div class="meta-row">
                <span :class="['status', session.current ? 'published' : 'muted']">{{ session.current ? "当前设备" : "其他设备" }}</span>
                <span>登录于 {{ formatDate(session.createdAt) }}</span>
                <span>有效至 {{ formatDate(session.expiresAt) }}</span>
                <button class="button-secondary" type="button" :disabled="session.current || sessionActingId === session.id" @click="removeSession(session)">
                  {{ sessionActingId === session.id ? "移除中..." : "移除设备" }}
                </button>
              </div>
            </article>
            <p v-if="!sessions.length" class="muted">暂无有效登录设备。</p>
          </div>
        </section>

        <section class="panel">
          <div class="panel-title"><h2>账号数据</h2></div>
          <div class="settings-stack">
            <div class="review-note"><strong>注销前请先导出数据</strong><p>账号注销后，未发布草稿、收藏和个人设置将不可恢复；已发布内容会按站点规则保留署名记录。</p></div>
            <button class="button-secondary" type="button" :disabled="exporting" @click="exportData">{{ exporting ? "导出中..." : "导出我的数据" }}</button>
            <button class="button-secondary" type="button" :disabled="deletingAccount" @click="deleteAccount">{{ deletingAccount ? "处理中..." : "申请注销账号" }}</button>
          </div>
        </section>
      </section>
    </template>
  </AccountLayout>
</template>
