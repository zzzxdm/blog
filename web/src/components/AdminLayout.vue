<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
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
const siteName = computed(() => siteSettings.value?.siteName.trim() || "云间笔记");
const brandMark = computed(() => siteName.value.slice(0, 1) || "云");

const navItems = [
  { label: "查看站点", to: "/" },
  { label: "概览", to: "/admin" },
  { label: "文章", to: "/admin/posts" },
  { label: "投稿", to: "/admin/submissions" },
  { label: "写作", to: "/admin/editor" },
  { label: "分类标签", to: "/admin/taxonomies" },
  { label: "评论", to: "/admin/comments" },
  { label: "用户", to: "/admin/users" },
  { label: "站内信", to: "/admin/messages" },
  { label: "媒体库", to: "/admin/media" },
  { label: "导航", to: "/admin/navigation" },
  { label: "统计", to: "/admin/stats" },
  { label: "日志", to: "/admin/audit" },
  { label: "设置", to: "/admin/settings" }
];

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

function isActive(to: string) {
  if (to === "/") {
    return false;
  }

  return route.path === to;
}
</script>

<template>
  <div class="mobile-admin-bar">
    <RouterLink class="brand" to="/admin">
      <span class="brand-mark">{{ brandMark }}</span>
      <span>{{ mobileTitle || title }}</span>
    </RouterLink>
    <slot name="mobile-action">
      <RouterLink v-if="primaryAction && primaryActionTo" class="button" :to="primaryActionTo">{{ primaryAction }}</RouterLink>
    </slot>
  </div>

  <div class="admin-shell">
    <aside class="admin-sidebar">
      <RouterLink class="admin-brand" to="/admin">
        <span class="brand-mark">{{ brandMark }}</span>
        <span>{{ siteName }}后台</span>
      </RouterLink>
      <nav class="admin-nav" aria-label="后台导航">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :class="{ active: isActive(item.to) }"
          :to="item.to"
        >
          {{ item.label }}
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
