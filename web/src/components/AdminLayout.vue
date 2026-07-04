<script setup lang="ts">
import { RouterLink, useRoute } from "vue-router";

defineProps<{
  title: string;
  description: string;
  mobileTitle?: string;
  primaryAction?: string;
}>();

const route = useRoute();

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
      <span class="brand-mark">云</span>
      <span>{{ mobileTitle || title }}</span>
    </RouterLink>
    <RouterLink v-if="primaryAction" class="button" to="/admin/editor">{{ primaryAction }}</RouterLink>
  </div>

  <div class="admin-shell">
    <aside class="admin-sidebar">
      <RouterLink class="admin-brand" to="/admin">
        <span class="brand-mark">云</span>
        <span>云间笔记后台</span>
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
