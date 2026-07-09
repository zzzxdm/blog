<script setup lang="ts">
import { Message } from "@element-plus/icons-vue";
import { computed } from "vue";
import { RouterLink, useRoute } from "vue-router";

import { useAuthStore } from "../stores/auth";
import { useMessageStore } from "../stores/messages";

defineProps<{
  title: string;
  description: string;
}>();

const route = useRoute();
const auth = useAuthStore();
const messages = useMessageStore();
const displayName = computed(() => auth.user?.displayName || "用户");
const avatarText = computed(() => auth.user?.avatarText || Array.from(displayName.value)[0] || "用");
const emailStatus = computed(() => auth.user?.emailVerified ? "已验证邮箱" : "邮箱未验证");
const roleText = computed(() => {
  if (auth.user?.role === "admin") return "管理员";
  if (auth.user?.role === "editor") return "编辑";
  if (auth.user?.role === "author") return "作者";
  return "注册用户";
});

const navItems = [
  { label: "概览", to: "/account" },
  { label: "我的评论", to: "/account/comments" },
  { label: "我的收藏", to: "/account/bookmarks" },
  { label: "私密文章", to: "/account/private-posts" },
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
            <span class="avatar">{{ avatarText }}</span>
            <div>
              <strong>{{ displayName }}</strong>
              <div class="meta-row">
                <span>{{ emailStatus }}</span>
                <span>{{ roleText }}</span>
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
              <span>{{ item.label }}</span>
              <span v-if="item.to === '/account/messages' && messages.unread" class="nav-count">
                <Message class="nav-count-icon" aria-hidden="true" />
                {{ messages.unread > 99 ? "99+" : messages.unread }}
              </span>
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
