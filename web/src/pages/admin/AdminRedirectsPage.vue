<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AdminLayout from "../../components/AdminLayout.vue";
import { getAdminRedirects, replaceAdminRedirects, type RedirectRule } from "../../shared/api";

const redirects = ref<RedirectRule[]>([]);
const loading = ref(false);
const saving = ref(false);
const error = ref("");
const message = ref("");
const searchQuery = ref("");

const visibleRedirects = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) {
    return redirects.value;
  }

  return redirects.value.filter((item) => `${item.from} ${item.to} ${item.code}`.toLowerCase().includes(keyword));
});

onMounted(load);

async function load() {
  loading.value = true;
  error.value = "";

  try {
    redirects.value = (await getAdminRedirects()).items;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "重定向规则加载失败";
  } finally {
    loading.value = false;
  }
}

function addRedirect() {
  redirects.value.unshift({
    from: "/old-path",
    to: "/new-path",
    code: 301
  });
}

function removeRedirect(redirect: RedirectRule) {
  redirects.value = redirects.value.filter((item) => item !== redirect);
}

async function save() {
  saving.value = true;
  error.value = "";
  message.value = "";

  try {
    redirects.value = (await replaceAdminRedirects(normalizedRedirects(redirects.value))).items;
    message.value = "重定向规则已保存。";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "重定向规则保存失败";
  } finally {
    saving.value = false;
  }
}

function normalizedRedirects(items: RedirectRule[]) {
  return items
    .filter((item) => item.from.trim() && item.to.trim())
    .map((item) => ({
      from: normalizePath(item.from),
      to: normalizeTarget(item.to),
      code: [301, 302, 307, 308].includes(Number(item.code)) ? Number(item.code) : 301
    }));
}

function normalizePath(value: string) {
  const path = value.trim();
  if (!path) {
    return "";
  }
  return path.startsWith("/") ? path : `/${path}`;
}

function normalizeTarget(value: string) {
  const target = value.trim();
  if (/^(https?:)?\/\//.test(target)) {
    return target;
  }
  return target.startsWith("/") ? target : `/${target}`;
}
</script>

<template>
  <AdminLayout title="重定向管理" description="维护旧地址到新地址的跳转规则，减少迁移和改版带来的断链。" mobile-title="重定向">
    <template #mobile-action>
      <button class="button" type="button" :disabled="saving" @click="save">{{ saving ? "保存中..." : "保存" }}</button>
    </template>

    <template #actions>
      <div class="header-actions">
        <button class="button-secondary" type="button" @click="addRedirect">添加规则</button>
        <button class="button" type="button" :disabled="saving" @click="save">{{ saving ? "保存中..." : "保存规则" }}</button>
      </div>
    </template>

    <p v-if="loading" class="muted">正在加载重定向规则...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-if="message" class="muted">{{ message }}</p>

    <section class="panel">
      <div class="panel-title">
        <h2>规则列表</h2>
        <span class="tag">{{ redirects.length }} 条</span>
      </div>
      <input v-model="searchQuery" class="input" type="search" placeholder="搜索旧地址、新地址或状态码" aria-label="搜索重定向规则">

      <div class="nav-builder">
        <article v-for="(redirect, index) in visibleRedirects" :key="`${redirect.from}-${index}`" class="nav-item redirect-item">
          <input v-model="redirect.from" class="input" aria-label="旧地址" placeholder="/old-path">
          <input v-model="redirect.to" class="input" aria-label="新地址" placeholder="/new-path">
          <select v-model.number="redirect.code" class="input" aria-label="状态码">
            <option :value="301">301 永久</option>
            <option :value="302">302 临时</option>
            <option :value="307">307 临时</option>
            <option :value="308">308 永久</option>
          </select>
          <button class="button-secondary" type="button" @click="removeRedirect(redirect)">删除</button>
        </article>
      </div>

      <p v-if="!visibleRedirects.length" class="muted">暂无匹配规则。</p>
    </section>
  </AdminLayout>
</template>
