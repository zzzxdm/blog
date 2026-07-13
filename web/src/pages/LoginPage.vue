<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";

import { ApiError, forgotPassword, getSiteSettings, resetPassword, verifyEmail, type SiteSettings } from "../shared/api";
import { useAuthStore } from "../stores/auth";
import { useToastStore } from "../stores/toast";

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string;
      reset: (widgetId: string) => void;
      remove: (widgetId: string) => void;
    };
  }
}

const router = useRouter();
const route = useRoute();
const auth = useAuthStore();
const toast = useToastStore();

type AuthMode = "login" | "register" | "forgot" | "reset" | "verify";

const mode = ref<AuthMode>("login");
const email = ref("");
const password = ref("");
const displayName = ref("");
const resetToken = ref("");
const verificationToken = ref("");
const turnstileEl = ref<HTMLElement | null>(null);
const turnstileWidgetId = ref("");
const turnstileToken = ref("");
const turnstileError = ref("");
const newPassword = ref("");
const verifying = ref(false);
const message = ref("");
const errorMessage = ref("");
const errorTitle = ref("操作失败");
const siteSettings = ref<SiteSettings | null>(null);
const siteName = computed(() => siteSettings.value?.siteName.trim() || "云间笔记");
const brandMark = computed(() => siteName.value.slice(0, 1) || "云");
const turnstileRequired = computed(() => Boolean(
  siteSettings.value?.turnstileEnabled &&
  ((mode.value === "login" && siteSettings.value.turnstileLogin) ||
    (mode.value === "register" && siteSettings.value.turnstileRegister)) &&
  siteSettings.value.turnstileSiteKey
));

onMounted(() => {
  applyAuthQuery();
  void loadSiteSettings();
});

onBeforeUnmount(() => {
  removeTurnstile();
});

watch(turnstileRequired, (required) => {
  if (required) {
    void renderTurnstile();
    return;
  }

  removeTurnstile();
});

watch(() => route.query, applyAuthQuery);

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

async function submit() {
  message.value = "";
  errorMessage.value = "";
  const submittedMode = mode.value;

  try {
    if (submittedMode === "login") {
      if (turnstileRequired.value && !turnstileToken.value) {
        errorTitle.value = failureTitle(submittedMode);
        errorMessage.value = turnstileError.value || "请先完成人机验证。";
        return;
      }

      await auth.login(email.value, password.value, turnstileToken.value);
      resetTurnstile();
      toast.success("登录成功", "正在进入个人中心。");
      await router.push(String(route.query.redirect || "/account"));
      return;
    }

    if (submittedMode === "register") {
      if (turnstileRequired.value && !turnstileToken.value) {
        errorTitle.value = "注册失败";
        errorMessage.value = turnstileError.value || "请先完成人机验证。";
        return;
      }

      const response = await auth.register(email.value, password.value, displayName.value, turnstileToken.value);
      verificationToken.value = response.verificationToken || "";
      resetTurnstile();
      password.value = "";
      mode.value = "verify";
      if (response.delivery === "email-failed") {
        message.value = "账号已创建，但验证邮件发送失败。请登录后到账号设置里重新发送验证邮件。";
        toast.warning("账号已创建", "验证邮件发送失败，请登录后重新发送。");
        return;
      }
      message.value = response.verificationToken ? `验证 token：${response.verificationToken}` : "验证入口已发送到邮箱，请完成验证后再投稿或互动。";
      toast.success("注册成功", response.verificationToken ? "请使用页面上的验证 token 完成验证。" : "验证入口已发送到邮箱。");
      return;
    }

    if (submittedMode === "verify") {
      verifying.value = true;
      const response = await verifyEmail(verificationToken.value);
      auth.user = response.user;
      verificationToken.value = "";
      message.value = "";
      toast.success("邮箱已验证", "现在可以投稿、评论和收藏。");
      await router.push(String(route.query.redirect || "/account/settings"));
      return;
    }

    if (submittedMode === "forgot") {
      const response = await forgotPassword(email.value);
      resetToken.value = response.resetToken || "";
      if (response.resetToken) {
        mode.value = "reset";
        message.value = `重置 token：${response.resetToken}`;
        toast.info("重置 token 已生成", "请使用页面上的 token 设置新密码。");
        return;
      }

      message.value = "重置入口已发送到该邮箱，请前往邮箱查看。";
      toast.success("重置入口已发送", "请前往邮箱查看。");
      return;
    }

    await resetPassword(resetToken.value, newPassword.value);
    mode.value = "login";
    password.value = "";
    newPassword.value = "";
    resetToken.value = "";
    message.value = "密码已重置，请使用新密码登录。";
    toast.success("密码已重置", "请使用新密码登录。");
    await router.replace({ name: "login", query: redirectQuery() });
  } catch (error) {
    errorTitle.value = failureTitle(submittedMode, error);
    errorMessage.value = friendlyErrorMessage(error, submittedMode);
    if (submittedMode === "login" || submittedMode === "register") {
      resetTurnstile();
    }
  } finally {
    verifying.value = false;
  }
}

function failureTitle(failedMode: AuthMode, error?: unknown) {
  if (error instanceof ApiError && error.status === 429) return "请求过于频繁";
  if (failedMode === "register") return "注册失败";
  if (failedMode === "forgot") return "找回密码失败";
  if (failedMode === "reset") return "重置密码失败";
  if (failedMode === "verify") return "邮箱验证失败";
  return "登录失败";
}

function friendlyErrorMessage(error: unknown, failedMode: AuthMode) {
  if (error instanceof ApiError) {
    if (error.status === 429) return error.message;

    if (failedMode === "login") {
      if (error.status === 410) return "该账号已被删除，不能继续登录。请联系管理员处理。";
      if (error.status === 403) return "该账号已被封禁，不能登录。请联系管理员处理。";
      if (error.status === 401) return "邮箱或密码不正确。";
      if (error.message.includes("turnstile")) return "人机验证未通过，请重新验证后再登录。";
      if (error.status === 400) return "请输入有效的邮箱和密码。";
      return "登录失败，请稍后再试。";
    }

    if (failedMode === "register") {
      if (error.status === 410) return "该邮箱对应的账号已被删除，不能直接重新注册。请联系管理员处理。";
      if (error.status === 409) return "该邮箱已注册，请直接登录或找回密码。";
      if (error.message.includes("turnstile")) return "人机验证未通过，请重新验证后再注册。";
      if (error.message.includes("email verification")) return "验证邮件发送失败，请检查邮箱服务配置后再试。";
      if (error.status === 400) return "请输入邮箱和至少 6 位密码。";
      return "注册失败，请稍后再试。";
    }

    if (failedMode === "reset") {
      if (error.status === 400) return "重置 token 无效、已过期，或新密码不足 6 位。";
      return "密码重置失败，请稍后再试。";
    }

    if (failedMode === "verify") {
      if (error.status === 400) return "验证 token 无效或已过期，请重新获取验证入口。";
      return "邮箱验证失败，请稍后再试。";
    }

    if (failedMode === "forgot") {
      if (error.message.includes("password reset email")) return "重置邮件发送失败，请稍后再试或联系管理员。";
      if (error.status === 404) return "该邮箱未注册，请检查邮箱地址或先注册账号。";
      if (error.status === 400) return "请输入有效的邮箱地址。";
      return "找回密码请求失败，请稍后再试。";
    }

    if (error.status === 400) return "请输入有效的邮箱地址。";
    return "找回密码失败，请稍后再试。";
  }

  return error instanceof Error ? error.message : "操作失败，请稍后再试。";
}

function closeError() {
  errorMessage.value = "";
}

function applyAuthQuery() {
  const queryMode = queryString(route.query.mode);
  const token = queryString(route.query.token);

  if (queryMode === "reset") {
    mode.value = "reset";
    if (token) {
      resetToken.value = token;
      message.value = "请设置新密码完成账号启用。";
    }
    return;
  }

  if (queryMode === "verify") {
    mode.value = "verify";
    if (token) {
      verificationToken.value = token;
      message.value = "请点击完成验证，验证后即可参与互动和投稿。";
    }
  }
}

function queryString(value: unknown) {
  if (Array.isArray(value)) {
    return String(value[0] || "");
  }
  return typeof value === "string" ? value : "";
}

function redirectQuery() {
  const redirect = queryString(route.query.redirect);
  return redirect ? { redirect } : {};
}

function title() {
  if (mode.value === "register") return "创建账号";
  if (mode.value === "forgot") return "找回密码";
  if (mode.value === "reset") return "重置密码";
  if (mode.value === "verify") return "验证邮箱";
  return "欢迎回来";
}

function description() {
  if (mode.value === "register") return "注册后需要先验证邮箱，再投稿、评论和收藏。";
  if (mode.value === "forgot") return "输入账号邮箱，系统会生成一次性重置入口。";
  if (mode.value === "reset") return "输入重置 token 和新密码完成更新。";
  if (mode.value === "verify") return "输入邮箱验证 token，完成后即可参与互动和投稿。";
  return "登录后可继续评论、查看收藏和接收回复通知。";
}

function buttonText() {
  if (auth.loading || verifying.value) return "处理中...";
  if (mode.value === "register") return "注册";
  if (mode.value === "forgot") return "发送重置入口";
  if (mode.value === "reset") return "重置密码";
  if (mode.value === "verify") return "完成验证";
  return "登录";
}

async function renderTurnstile() {
  await nextTick();
  if (!turnstileRequired.value || !turnstileEl.value || turnstileWidgetId.value) {
    return;
  }

  try {
    await loadTurnstileScript();
  } catch {
    turnstileError.value = "人机验证脚本加载失败，请检查浏览器是否能访问 challenges.cloudflare.com。";
    return;
  }

  if (!window.turnstile || !turnstileEl.value || !siteSettings.value?.turnstileSiteKey) {
    turnstileError.value = "人机验证加载失败，请刷新页面后重试。";
    return;
  }

  turnstileError.value = "";
  turnstileWidgetId.value = window.turnstile.render(turnstileEl.value, {
    sitekey: siteSettings.value.turnstileSiteKey,
    callback: (token: string) => {
      turnstileToken.value = token;
      turnstileError.value = "";
    },
    "expired-callback": () => {
      turnstileToken.value = "";
    },
    "error-callback": () => {
      turnstileToken.value = "";
      turnstileError.value = "人机验证无法连接，请确认 Site Key 允许当前域名（localhost/127.0.0.1），或改用 Cloudflare Turnstile 本地测试 Key。";
    }
  });
}

function loadTurnstileScript() {
  if (window.turnstile) {
    return Promise.resolve();
  }

  return new Promise<void>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>("script[data-turnstile]");
    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener("error", () => reject(new Error("turnstile script failed")), { once: true });
      return;
    }

    const script = document.createElement("script");
    script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";
    script.async = true;
    script.defer = true;
    script.dataset.turnstile = "true";
    script.addEventListener("load", () => resolve(), { once: true });
    script.addEventListener("error", () => reject(new Error("turnstile script failed")), { once: true });
    document.head.appendChild(script);
  });
}

function resetTurnstile() {
  turnstileToken.value = "";
  turnstileError.value = "";
  if (turnstileWidgetId.value && window.turnstile) {
    window.turnstile.reset(turnstileWidgetId.value);
  }
}

function removeTurnstile() {
  turnstileToken.value = "";
  turnstileError.value = "";
  if (turnstileWidgetId.value && window.turnstile) {
    window.turnstile.remove(turnstileWidgetId.value);
  }
  turnstileWidgetId.value = "";
}
</script>

<template>
  <main class="auth-shell">
    <Transition name="auth-alert">
      <div v-if="errorMessage" class="auth-alert" role="alert" aria-live="assertive">
        <strong>{{ errorTitle }}</strong>
        <span>{{ errorMessage }}</span>
        <button class="auth-alert-close" type="button" aria-label="关闭提示" @click="closeError">×</button>
      </div>
    </Transition>

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
          <a :class="{ active: mode === 'register' || mode === 'verify' }" href="#register" @click.prevent="mode = 'register'">注册</a>
        </div>
        <h2>{{ title() }}</h2>
        <p>{{ description() }}</p>

        <form class="settings-stack" @submit.prevent="submit">
          <div v-if="mode === 'register'" class="field">
            <label for="displayName">昵称</label>
            <input v-model="displayName" class="input" id="displayName" type="text">
          </div>
          <div v-if="mode !== 'reset' && mode !== 'verify'" class="field">
            <label for="email">邮箱</label>
            <input v-model="email" class="input" id="email" type="email" autocomplete="email">
          </div>
          <div v-if="mode === 'login' || mode === 'register'" class="field">
            <label for="password">密码</label>
            <input v-model="password" class="input" id="password" type="password" :autocomplete="mode === 'register' ? 'new-password' : 'current-password'">
          </div>
          <div v-if="turnstileRequired" class="field">
            <label>人机验证</label>
            <div ref="turnstileEl"></div>
            <p v-if="turnstileError" class="auth-form-error" role="alert">{{ turnstileError }}</p>
          </div>
          <div v-if="mode === 'reset'" class="field">
            <label for="reset-token">重置 token</label>
            <input v-model="resetToken" class="input" id="reset-token">
          </div>
          <div v-if="mode === 'reset'" class="field">
            <label for="new-password">新密码</label>
            <input v-model="newPassword" class="input" id="new-password" type="password" autocomplete="new-password">
          </div>
          <div v-if="mode === 'verify'" class="field">
            <label for="verification-token">邮箱验证 token</label>
            <input v-model="verificationToken" class="input" id="verification-token" autocomplete="one-time-code">
          </div>
          <button class="button" type="submit" :disabled="auth.loading || verifying">{{ buttonText() }}</button>
          <button v-if="mode === 'login'" class="button-secondary" type="button" @click="mode = 'forgot'">忘记密码</button>
          <button v-if="mode === 'forgot' || mode === 'reset' || mode === 'verify'" class="button-secondary" type="button" @click="mode = 'login'">返回登录</button>
          <p v-if="message" class="muted">{{ message }}</p>
        </form>

        <div class="section-heading" style="margin: 20px 0 0;">
          <p>没有账号时，注册后需要先完成邮箱验证。忘记密码可通过邮箱重置。</p>
        </div>
      </div>
    </section>
  </main>
</template>
