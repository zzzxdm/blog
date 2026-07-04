<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";

import { forgotPassword, getSiteSettings, resetPassword, type SiteSettings } from "../shared/api";
import { useAuthStore } from "../stores/auth";

const router = useRouter();
const route = useRoute();
const auth = useAuthStore();

const mode = ref<"login" | "register" | "forgot" | "reset">("login");
const email = ref("");
const password = ref("");
const displayName = ref("");
const resetToken = ref("");
const newPassword = ref("");
const message = ref("");
const siteSettings = ref<SiteSettings | null>(null);
const siteName = computed(() => siteSettings.value?.siteName.trim() || "云间笔记");
const brandMark = computed(() => siteName.value.slice(0, 1) || "云");

onMounted(() => {
  void loadSiteSettings();
});

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

async function submit() {
  message.value = "";

  try {
    if (mode.value === "login") {
      await auth.login(email.value, password.value);
      await router.push(String(route.query.redirect || "/account"));
      return;
    }

    if (mode.value === "register") {
      await auth.register(email.value, password.value, displayName.value);
      await router.push(String(route.query.redirect || "/account/settings"));
      return;
    }

    if (mode.value === "forgot") {
      const response = await forgotPassword(email.value);
      resetToken.value = response.resetToken || "";
      mode.value = "reset";
      message.value = response.resetToken ? `重置 token：${response.resetToken}` : "如果邮箱存在，重置入口会发送到该邮箱。";
      return;
    }

    await resetPassword(resetToken.value, newPassword.value);
    mode.value = "login";
    password.value = "";
    newPassword.value = "";
    message.value = "密码已重置，请使用新密码登录。";
  } catch {
    message.value = auth.error || "操作失败";
  }
}

function title() {
  if (mode.value === "register") return "创建账号";
  if (mode.value === "forgot") return "找回密码";
  if (mode.value === "reset") return "重置密码";
  return "欢迎回来";
}

function description() {
  if (mode.value === "register") return "注册后可以投稿、评论、收藏并接收站内信。";
  if (mode.value === "forgot") return "输入账号邮箱，系统会生成一次性重置入口。";
  if (mode.value === "reset") return "输入重置 token 和新密码完成更新。";
  return "登录后可继续评论、查看收藏和接收回复通知。";
}

function buttonText() {
  if (auth.loading) return "处理中...";
  if (mode.value === "register") return "注册";
  if (mode.value === "forgot") return "发送重置入口";
  if (mode.value === "reset") return "重置密码";
  return "登录";
}
</script>

<template>
  <main class="auth-shell">
    <section class="auth-visual" aria-label="登录背景">
      <img
        src="https://images.unsplash.com/photo-1455390582262-044cdead277a?auto=format&fit=crop&w=1400&q=80"
        alt="桌面上的笔记本和写作草稿"
      >
      <div class="auth-copy">
        <RouterLink class="brand" to="/">
          <span class="brand-mark">{{ brandMark }}</span>
          <span>{{ siteName }}</span>
        </RouterLink>
        <h1>登录后参与讨论，收藏值得反复看的文章。</h1>
        <p>用户账号用于评论、回复、点赞、收藏和接收通知。访客可以阅读内容，评论默认需要登录。</p>
      </div>
    </section>

    <section class="auth-panel-wrap">
      <div class="auth-panel">
        <div class="auth-tabs">
          <a :class="{ active: mode === 'login' }" href="#login" @click.prevent="mode = 'login'">登录</a>
          <a :class="{ active: mode === 'register' }" href="#register" @click.prevent="mode = 'register'">注册</a>
        </div>
        <h2>{{ title() }}</h2>
        <p>{{ description() }}</p>

        <form class="settings-stack" @submit.prevent="submit">
          <div v-if="mode === 'register'" class="field">
            <label for="displayName">昵称</label>
            <input v-model="displayName" class="input" id="displayName" type="text">
          </div>
          <div v-if="mode !== 'reset'" class="field">
            <label for="email">邮箱</label>
            <input v-model="email" class="input" id="email" type="email" autocomplete="email">
          </div>
          <div v-if="mode === 'login' || mode === 'register'" class="field">
            <label for="password">密码</label>
            <input v-model="password" class="input" id="password" type="password" :autocomplete="mode === 'register' ? 'new-password' : 'current-password'">
          </div>
          <div v-if="mode === 'reset'" class="field">
            <label for="reset-token">重置 token</label>
            <input v-model="resetToken" class="input" id="reset-token">
          </div>
          <div v-if="mode === 'reset'" class="field">
            <label for="new-password">新密码</label>
            <input v-model="newPassword" class="input" id="new-password" type="password" autocomplete="new-password">
          </div>
          <button class="button" type="submit" :disabled="auth.loading">{{ buttonText() }}</button>
          <button v-if="mode === 'login'" class="button-secondary" type="button" @click="mode = 'forgot'">忘记密码</button>
          <button v-if="mode === 'forgot' || mode === 'reset'" class="button-secondary" type="button" @click="mode = 'login'">返回登录</button>
          <p v-if="message" class="muted">{{ message }}</p>
        </form>

        <div class="section-heading" style="margin: 20px 0 0;">
          <p>没有账号时，注册后需要先完成邮箱验证。忘记密码可通过邮箱重置。</p>
        </div>
      </div>
    </section>
  </main>
</template>
