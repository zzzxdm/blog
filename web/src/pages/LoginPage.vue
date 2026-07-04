<script setup lang="ts">
import { ref } from "vue";
import { RouterLink } from "vue-router";
import { useRouter } from "vue-router";

import { useAuthStore } from "../stores/auth";

const router = useRouter();
const auth = useAuthStore();

const mode = ref<"login" | "register">("login");
const email = ref("linyi@example.com");
const password = ref("password");
const displayName = ref("林一");
const message = ref("");

async function submit() {
  message.value = "";

  try {
    if (mode.value === "login") {
      await auth.login(email.value, password.value);
    } else {
      await auth.register(email.value, password.value, displayName.value);
    }

    await router.push("/account");
  } catch {
    message.value = auth.error || "操作失败";
  }
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
          <span class="brand-mark">云</span>
          <span>云间笔记</span>
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
        <h2>{{ mode === "login" ? "欢迎回来" : "创建账号" }}</h2>
        <p>{{ mode === "login" ? "登录后可继续评论、查看收藏和接收回复通知。" : "注册后可以投稿、评论、收藏并接收站内信。" }}</p>

        <form class="settings-stack" @submit.prevent="submit">
          <div v-if="mode === 'register'" class="field">
            <label for="displayName">昵称</label>
            <input v-model="displayName" class="input" id="displayName" type="text">
          </div>
          <div class="field">
            <label for="email">邮箱</label>
            <input v-model="email" class="input" id="email" type="email" autocomplete="email">
          </div>
          <div class="field">
            <label for="password">密码</label>
            <input v-model="password" class="input" id="password" type="password" autocomplete="current-password">
          </div>
          <button class="button" type="submit" :disabled="auth.loading">{{ auth.loading ? "处理中..." : mode === "login" ? "登录" : "注册" }}</button>
          <button class="button-secondary" type="button">使用 GitHub 登录</button>
          <p v-if="message" class="error">{{ message }}</p>
        </form>

        <div class="section-heading" style="margin: 20px 0 0;">
          <p>没有账号时，注册后需要先完成邮箱验证。忘记密码可通过邮箱重置。</p>
        </div>
      </div>
    </section>
  </main>
</template>
