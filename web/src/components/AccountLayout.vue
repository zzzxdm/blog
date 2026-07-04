<script setup lang="ts">
import { RouterLink, useRoute } from "vue-router";

defineProps<{
  title: string;
  description: string;
}>();

const route = useRoute();

const navItems = [
  { label: "概览", to: "/account" },
  { label: "我的评论", to: "/account/comments" },
  { label: "我的收藏", to: "/account/bookmarks" },
  { label: "我的投稿", to: "/account/submissions" },
  { label: "站内信", to: "/account/messages" },
  { label: "账号设置", to: "/account/settings" }
];
</script>

<template>
  <main class="page">
    <section class="section-heading">
      <div>
        <h1>{{ title }}</h1>
        <p>{{ description }}</p>
      </div>
      <slot name="actions" />
    </section>

    <section class="account-layout">
      <aside class="panel">
        <div class="profile-card">
          <div class="profile-hero">
            <span class="avatar">林</span>
            <div>
              <strong>林一</strong>
              <div class="meta-row">
                <span>已验证邮箱</span>
                <span>普通注册用户</span>
              </div>
            </div>
          </div>
          <nav class="account-nav" aria-label="个人中心导航">
            <RouterLink
              v-for="item in navItems"
              :key="item.to"
              :class="{ active: route.path === item.to }"
              :to="item.to"
            >
              {{ item.label }}
            </RouterLink>
          </nav>
        </div>
      </aside>

      <div class="settings-stack">
        <slot />
      </div>
    </section>
  </main>
</template>
