<script setup lang="ts">
import {
  ChatDotSquare,
  Close,
  CollectionTag,
  Document,
  EditPen,
  Files,
  Histogram,
  HomeFilled,
  Link,
  Management,
  Menu as MenuIcon,
  Message,
  Picture,
  Setting,
  Tickets,
  Upload,
  User,
  View
} from "@element-plus/icons-vue";
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";

import { getSiteSettings, type SiteSettings } from "../shared/api";

defineProps<{
  title: string;
  description: string;
  mobileTitle?: string;
  primaryAction?: string;
  primaryActionTo?: string;
}>();

const route = useRoute();
const siteSettings = ref<SiteSettings | null>(null);
const sidebarCollapsed = ref(false);
const mobileSidebarOpen = ref(false);
const siteName = computed(() => siteSettings.value?.siteName.trim() || "云间笔记");
const brandMark = computed(() => siteName.value.slice(0, 1) || "云");

const navItems = [
  { label: "查看站点", to: "/", icon: View },
  { label: "概览", to: "/admin", icon: HomeFilled },
  { label: "文章", to: "/admin/posts", icon: Document },
  { label: "投稿", to: "/admin/submissions", icon: Upload },
  { label: "写作", to: "/admin/editor", icon: EditPen },
  { label: "分类标签", to: "/admin/taxonomies", icon: CollectionTag },
  { label: "专题", to: "/admin/topics", icon: Management },
  { label: "评论", to: "/admin/comments", icon: ChatDotSquare },
  { label: "用户", to: "/admin/users", icon: User },
  { label: "站内信", to: "/admin/messages", icon: Message },
  { label: "媒体库", to: "/admin/media", icon: Picture },
  { label: "导航", to: "/admin/navigation", icon: MenuIcon },
  { label: "重定向", to: "/admin/redirects", icon: Link },
  { label: "统计", to: "/admin/stats", icon: Histogram },
  { label: "导入导出", to: "/admin/import-export", icon: Files },
  { label: "日志", to: "/admin/audit", icon: Tickets },
  { label: "设置", to: "/admin/settings", icon: Setting }
];

onMounted(() => {
  sidebarCollapsed.value = window.localStorage.getItem("admin:sidebar-collapsed") === "true";
  void loadSiteSettings();
});

watch(() => route.fullPath, () => {
  mobileSidebarOpen.value = false;
});

async function loadSiteSettings() {
  try {
    siteSettings.value = await getSiteSettings();
  } catch {
    siteSettings.value = null;
  }
}

function isActive(to: string) {
  if (to === "/") {
    return false;
  }
  if (to === "/admin") {
    return route.path === "/admin";
  }

  return route.path === to || route.path.startsWith(`${to}/`);
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value;
  window.localStorage.setItem("admin:sidebar-collapsed", String(sidebarCollapsed.value));
}

function toggleMobileSidebar() {
  mobileSidebarOpen.value = !mobileSidebarOpen.value;
}

function closeMobileSidebar() {
  mobileSidebarOpen.value = false;
}
</script>

<template>
  <div class="mobile-admin-bar">
    <button
      class="icon-button mobile-admin-menu"
      type="button"
      aria-controls="admin-sidebar"
      :aria-expanded="mobileSidebarOpen"
      :aria-label="mobileSidebarOpen ? '关闭后台菜单' : '打开后台菜单'"
      @click="toggleMobileSidebar"
    >
      <Close v-if="mobileSidebarOpen" class="button-icon" aria-hidden="true" />
      <MenuIcon v-else class="button-icon" aria-hidden="true" />
    </button>
    <RouterLink class="brand" to="/admin" @click="closeMobileSidebar">
      <span class="brand-mark">{{ brandMark }}</span>
      <span>{{ mobileTitle || title }}</span>
    </RouterLink>
    <div class="mobile-admin-actions">
      <slot name="mobile-action">
        <RouterLink v-if="primaryAction && primaryActionTo" class="button" :to="primaryActionTo">{{ primaryAction }}</RouterLink>
      </slot>
    </div>
  </div>

  <div class="admin-shell" :class="{ 'sidebar-collapsed': sidebarCollapsed, 'mobile-sidebar-open': mobileSidebarOpen }">
    <button
      v-if="mobileSidebarOpen"
      class="admin-sidebar-backdrop"
      type="button"
      aria-label="关闭后台菜单"
      @click="closeMobileSidebar"
    ></button>
    <aside id="admin-sidebar" class="admin-sidebar">
      <div class="admin-sidebar-header">
        <RouterLink class="admin-brand" to="/admin" :title="`${siteName}后台`" @click="closeMobileSidebar">
          <span class="brand-mark">{{ brandMark }}</span>
          <span class="admin-brand-text">{{ siteName }}后台</span>
        </RouterLink>
        <button
          class="admin-sidebar-toggle"
          type="button"
          :aria-label="sidebarCollapsed ? '展开后台菜单' : '折叠后台菜单'"
          :title="sidebarCollapsed ? '展开菜单' : '折叠菜单'"
          @click="toggleSidebar"
        >
          {{ sidebarCollapsed ? "›" : "‹" }}
        </button>
        <button
          class="icon-button admin-mobile-close"
          type="button"
          aria-label="关闭后台菜单"
          @click="closeMobileSidebar"
        >
          <Close class="button-icon" aria-hidden="true" />
        </button>
      </div>
      <nav class="admin-nav" aria-label="后台导航">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :class="{ active: isActive(item.to) }"
          :to="item.to"
          :title="item.label"
          @click="closeMobileSidebar"
        >
          <span class="admin-nav-icon" aria-hidden="true"><component :is="item.icon" /></span>
          <span class="admin-nav-label">{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <main class="admin-main">
      <header class="admin-topbar">
        <div>
          <h1>{{ title }}</h1>
          <p>{{ description }}</p>
        </div>
        <slot name="actions" />
      </header>

      <slot />
    </main>
  </div>
</template>
