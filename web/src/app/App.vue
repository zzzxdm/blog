<script setup lang="ts">
import { Close, Menu as MenuIcon, Message, Moon, Search, Sunny } from "@element-plus/icons-vue";
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";

import ConfirmDialogViewport from "../components/ConfirmDialogViewport.vue";
import SiteBacktop from "../components/SiteBacktop.vue";
import SiteFootbar from "../components/SiteFootbar.vue";
import SiteSearch from "../components/SiteSearch.vue";
import ToastViewport from "../components/ToastViewport.vue";
import { getCategories, getSiteNavigation, getSiteSettings, type Category, type NavItem, type OperationsNavigation, type SiteSettings } from "../shared/api";
import { applyPrimaryColor, applyThemeMode, getInitialThemeMode, type ThemeMode } from "../shared/theme";
import { useAuthStore } from "../stores/auth";
import { useMessageStore } from "../stores/messages";

const route = useRoute();
const auth = useAuthStore();
const messages = useMessageStore();
const navOpen = ref(false);
const mobileSubmenus = ref<Record<string, boolean>>({});
const searchOpen = ref(false);
const themeMode = ref<ThemeMode>("light");
const categories = ref<Category[]>([]);
const navigation = ref<OperationsNavigation | null>(null);
const siteSettings = ref<SiteSettings | null>(null);
const showChrome = computed(() => !route.meta.hideChrome);
const navCategories = computed(() => categories.value.slice(0, 4));
const siteName = computed(() => siteSettings.value?.siteName.trim() || "云间笔记");
const siteBeian = computed(() => siteSettings.value?.beian.trim() || "");
const brandMark = computed(() => siteName.value.slice(0, 1) || "云");
const defaultTopItems: NavItem[] = [
  { id: "nav_default_home", label: "首页", url: "/", order: 1 },
  { id: "nav_default_archive", label: "归档", url: "/archive", order: 2 },
  { id: "nav_default_topics", label: "专题", url: "/topics", order: 3 },
  { id: "nav_default_submit", label: "投稿", url: "/submit", order: 4 },
  { id: "nav_default_about", label: "关于", url: "/about", order: 5 }
];
const topNavItems = computed(() => orderedNavItems(navigation.value?.topItems ?? defaultTopItems));
const showLoginEntry = computed(() => navigation.value?.showLoginEntry ?? true);
const externalLinksNewWindow = computed(() => navigation.value?.externalLinksNewWindow ?? true);
const mobileCollapse = computed(() => navigation.value?.mobileCollapse ?? true);
const darkModeEnabled = computed(() => siteSettings.value?.darkModeEnabled ?? true);

onMounted(() => {
  initializeTheme();
  void loadSiteSettings();
  void auth.loadMe();
  void loadCategories();
  void loadNavigation();
});

watch(() => route.fullPath, () => {
  navOpen.value = false;
  mobileSubmenus.value = {};
  if (auth.user) {
    void messages.refreshUnread();
  }
});

watch(
  () => auth.user?.id,
  (userID) => {
    if (userID) {
      void messages.refreshUnread();
    } else {
      messages.clear();
    }
  }
);

function logout() {
  void auth.logout();
}

function initializeTheme() {
  applyTheme(getInitialThemeMode());
}

function toggleTheme() {
  applyTheme(themeMode.value === "dark" ? "light" : "dark");
}

function applyTheme(mode: ThemeMode) {
  themeMode.value = mode;
  applyThemeMode(mode);
}

async function loadCategories() {
  try {
    categories.value = (await getCategories()).items;
  } catch {
    categories.value = [];
  }
}

async function loadSiteSettings() {
  try {
    const settings = await getSiteSettings();
    siteSettings.value = settings;
    applyPrimaryColor(settings.themePrimary);
    if (!settings.darkModeEnabled) {
      applyTheme("light");
    }
  } catch {
    siteSettings.value = null;
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

function isMobileSubmenuOpen(url: string) {
  return Boolean(mobileSubmenus.value[url]);
}

function toggleMobileSubmenu(url: string) {
  mobileSubmenus.value = {
    ...mobileSubmenus.value,
    [url]: !mobileSubmenus.value[url]
  };
}
</script>

<template>
  <header v-if="showChrome" class="site-header">
    <div class="nav">
      <div class="nav-main-row">
        <RouterLink class="brand" to="/" :aria-label="`${siteName}首页`">
          <span class="brand-mark">{{ brandMark }}</span>
          <span>{{ siteName }}</span>
        </RouterLink>
        <button
          v-if="mobileCollapse"
          class="icon-button nav-toggle"
          type="button"
          aria-controls="site-nav"
          :aria-expanded="navOpen"
          :aria-label="navOpen ? '收起导航' : '展开导航'"
          @click="navOpen = !navOpen"
        >
          <Close v-if="navOpen" class="button-icon" aria-hidden="true" />
          <MenuIcon v-else class="button-icon" aria-hidden="true" />
        </button>
      </div>
      <nav id="site-nav" class="nav-links" :class="{ 'is-collapsible': mobileCollapse, 'is-open': navOpen }" aria-label="主导航">
        <template v-for="item in topNavItems" :key="item.id">
          <div v-if="item.url === '/archive'" class="nav-menu-item" :class="{ 'is-mobile-open': isMobileSubmenuOpen(item.url) }">
            <div class="nav-parent-row">
              <RouterLink :class="{ active: isActiveNav(item.url) }" class="nav-parent" :to="item.url">
                {{ item.label }} <span class="nav-caret nav-desktop-caret">⌄</span>
              </RouterLink>
              <button
                class="nav-submenu-toggle"
                type="button"
                :aria-controls="`nav-submenu-${item.id}`"
                :aria-expanded="isMobileSubmenuOpen(item.url)"
                :aria-label="`${isMobileSubmenuOpen(item.url) ? '收起' : '展开'}${item.label}`"
                @click="toggleMobileSubmenu(item.url)"
              >
                <span class="nav-caret" aria-hidden="true">⌄</span>
              </button>
            </div>
            <div class="nav-submenu" :id="`nav-submenu-${item.id}`">
              <RouterLink to="/archive">全部文章</RouterLink>
              <RouterLink v-for="category in navCategories" :key="category.id" :to="`/archive?category=${encodeURIComponent(category.name)}`">
                {{ category.name }}
              </RouterLink>
            </div>
          </div>
          <div v-else-if="item.url === '/topics'" class="nav-menu-item" :class="{ 'is-mobile-open': isMobileSubmenuOpen(item.url) }">
            <div class="nav-parent-row">
              <RouterLink :class="{ active: isActiveNav(item.url) }" class="nav-parent" :to="item.url">
                {{ item.label }} <span class="nav-caret nav-desktop-caret">⌄</span>
              </RouterLink>
              <button
                class="nav-submenu-toggle"
                type="button"
                :aria-controls="`nav-submenu-${item.id}`"
                :aria-expanded="isMobileSubmenuOpen(item.url)"
                :aria-label="`${isMobileSubmenuOpen(item.url) ? '收起' : '展开'}${item.label}`"
                @click="toggleMobileSubmenu(item.url)"
              >
                <span class="nav-caret" aria-hidden="true">⌄</span>
              </button>
            </div>
            <div class="nav-submenu" :id="`nav-submenu-${item.id}`">
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
        <button class="icon-button" type="button" aria-label="搜索" title="搜索" @click="searchOpen = true">
          <Search class="button-icon" aria-hidden="true" />
        </button>
        <button
          v-if="darkModeEnabled"
          class="icon-button"
          :class="{ active: themeMode === 'dark' }"
          type="button"
          :aria-label="themeMode === 'dark' ? '切换浅色模式' : '切换深色模式'"
          @click="toggleTheme"
        >
          <Sunny v-if="themeMode === 'dark'" class="button-icon" aria-hidden="true" />
          <Moon v-else class="button-icon" aria-hidden="true" />
        </button>
        <template v-if="showLoginEntry">
          <template v-if="auth.user">
            <RouterLink class="icon-button notification-button" to="/account/messages" aria-label="站内信" title="站内信">
              <Message class="button-icon" aria-hidden="true" />
              <span v-if="messages.unread" class="notification-badge" aria-label="未读站内信">
                {{ messages.unread > 99 ? "99+" : messages.unread }}
              </span>
            </RouterLink>
            <RouterLink class="button-secondary" to="/account">{{ auth.user.displayName }}</RouterLink>
            <button class="button-secondary" type="button" @click="logout">退出</button>
          </template>
          <RouterLink v-else class="button-secondary" to="/login">登录</RouterLink>
        </template>
      </div>
    </div>
  </header>

  <RouterView />
  <SiteFootbar v-if="showChrome" :navigation="navigation" :site-name="siteName" :beian="siteBeian" />
  <SiteSearch v-if="showChrome" v-model:open="searchOpen" />
  <SiteBacktop v-if="showChrome" />
  <ConfirmDialogViewport />
  <ToastViewport />
</template>
