<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";

import SiteBacktop from "../components/SiteBacktop.vue";
import SiteFootbar from "../components/SiteFootbar.vue";
import SiteSearch from "../components/SiteSearch.vue";
import { getCategories, getSiteNavigation, type Category, type NavItem, type OperationsNavigation } from "../shared/api";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const auth = useAuthStore();
const searchOpen = ref(false);
const themeMode = ref<"light" | "dark">("light");
const categories = ref<Category[]>([]);
const navigation = ref<OperationsNavigation | null>(null);
const showChrome = computed(() => !route.meta.hideChrome);
const navCategories = computed(() => categories.value.slice(0, 4));
const defaultTopItems: NavItem[] = [
  { id: "nav_default_home", label: "首页", url: "/", order: 1 },
  { id: "nav_default_archive", label: "归档", url: "/archive", order: 2 },
  { id: "nav_default_topics", label: "专题", url: "/topics", order: 3 },
  { id: "nav_default_submit", label: "投稿", url: "/submit", order: 4 }
];
const topNavItems = computed(() => orderedNavItems(navigation.value?.topItems ?? defaultTopItems));
const showLoginEntry = computed(() => navigation.value?.showLoginEntry ?? true);
const externalLinksNewWindow = computed(() => navigation.value?.externalLinksNewWindow ?? true);

onMounted(() => {
  initializeTheme();
  void auth.loadMe();
  void loadCategories();
  void loadNavigation();
});

function logout() {
  void auth.logout();
}

function initializeTheme() {
  const stored = window.localStorage.getItem("site:theme");
  const prefersDark = typeof window.matchMedia === "function" && window.matchMedia("(prefers-color-scheme: dark)").matches;
  applyTheme(stored === "dark" || (!stored && prefersDark) ? "dark" : "light");
}

function toggleTheme() {
  applyTheme(themeMode.value === "dark" ? "light" : "dark");
}

function applyTheme(mode: "light" | "dark") {
  themeMode.value = mode;
  document.documentElement.dataset.theme = mode;
  window.localStorage.setItem("site:theme", mode);
}

async function loadCategories() {
  try {
    categories.value = (await getCategories()).items;
  } catch {
    categories.value = [];
  }
}

async function loadNavigation() {
  try {
    navigation.value = await getSiteNavigation();
  } catch {
    navigation.value = null;
  }
}

function orderedNavItems(items: NavItem[]) {
  return [...items]
    .filter((item) => item.label.trim() && item.url.trim())
    .sort((left, right) => left.order - right.order);
}

function isExternalUrl(url: string) {
  return /^(https?:)?\/\//.test(url) || url.startsWith("mailto:") || url.startsWith("tel:");
}

function isRouterUrl(url: string) {
  return url.startsWith("/") && !url.startsWith("//") && !/\.[a-z0-9]+($|[?#])/i.test(url);
}

function isActiveNav(url: string) {
  if (!isRouterUrl(url)) {
    return false;
  }

  const path = (url.split("?")[0] || "/").replace(/\/+$/, "") || "/";
  if (path === "/") {
    return route.path === "/";
  }
  return route.path === path || route.path.startsWith(`${path}/`);
}
</script>

<template>
  <header v-if="showChrome" class="site-header">
    <div class="nav">
      <RouterLink class="brand" to="/" aria-label="云间笔记首页">
        <span class="brand-mark">云</span>
        <span>云间笔记</span>
      </RouterLink>
      <nav class="nav-links" aria-label="主导航">
        <template v-for="item in topNavItems" :key="item.id">
          <div v-if="item.url === '/archive'" class="nav-menu-item">
            <RouterLink :class="{ active: isActiveNav(item.url) }" class="nav-parent" :to="item.url">
              {{ item.label }} <span class="nav-caret">⌄</span>
            </RouterLink>
            <div class="nav-submenu">
              <RouterLink to="/archive">全部文章</RouterLink>
              <RouterLink v-for="category in navCategories" :key="category.id" :to="`/archive?category=${encodeURIComponent(category.name)}`">
                {{ category.name }}
              </RouterLink>
            </div>
          </div>
          <div v-else-if="item.url === '/topics'" class="nav-menu-item">
            <RouterLink :class="{ active: isActiveNav(item.url) }" class="nav-parent" :to="item.url">
              {{ item.label }} <span class="nav-caret">⌄</span>
            </RouterLink>
            <div class="nav-submenu">
              <RouterLink to="/topics?topic=blog-system">博客系统</RouterLink>
              <RouterLink to="/topics?topic=vue3-content">Vue3 内容站</RouterLink>
              <RouterLink to="/topics?topic=writing-workflow">写作工作流</RouterLink>
              <RouterLink to="/topics?topic=resource-list">资源清单</RouterLink>
            </div>
          </div>
          <RouterLink v-else-if="isRouterUrl(item.url)" :class="{ active: isActiveNav(item.url) }" :to="item.url">
            {{ item.label }}
          </RouterLink>
          <a
            v-else
            :href="item.url"
            :target="isExternalUrl(item.url) && externalLinksNewWindow ? '_blank' : undefined"
            :rel="isExternalUrl(item.url) && externalLinksNewWindow ? 'noreferrer' : undefined"
          >
            {{ item.label }}
          </a>
        </template>
        <RouterLink v-if="auth.user" :class="{ active: route.path.startsWith('/account') }" to="/account">我的</RouterLink>
        <RouterLink v-if="auth.user?.role === 'admin'" :class="{ active: route.path.startsWith('/admin') }" to="/admin">后台</RouterLink>
      </nav>
      <div class="header-actions">
        <button class="icon-button" type="button" aria-label="搜索" @click="searchOpen = true">⌕</button>
        <button
          class="icon-button"
          :class="{ active: themeMode === 'dark' }"
          type="button"
          :aria-label="themeMode === 'dark' ? '切换浅色模式' : '切换深色模式'"
          @click="toggleTheme"
        >
          {{ themeMode === "dark" ? "☀" : "◐" }}
        </button>
        <template v-if="showLoginEntry">
          <template v-if="auth.user">
            <RouterLink class="icon-button" to="/account/messages" aria-label="站内信">信</RouterLink>
            <RouterLink class="button-secondary" to="/account">{{ auth.user.displayName }}</RouterLink>
            <button class="button-secondary" type="button" @click="logout">退出</button>
          </template>
          <RouterLink v-else class="button-secondary" to="/login">登录</RouterLink>
        </template>
      </div>
    </div>
  </header>

  <RouterView />
  <SiteFootbar v-if="showChrome" :navigation="navigation" />
  <SiteSearch v-if="showChrome" v-model:open="searchOpen" />
  <SiteBacktop v-if="showChrome" />
</template>
