<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";

import SiteBacktop from "../components/SiteBacktop.vue";
import SiteFootbar from "../components/SiteFootbar.vue";
import SiteSearch from "../components/SiteSearch.vue";
import { getCategories, type Category } from "../shared/api";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const auth = useAuthStore();
const searchOpen = ref(false);
const categories = ref<Category[]>([]);
const showChrome = computed(() => !route.meta.hideChrome);
const navCategories = computed(() => categories.value.slice(0, 4));

onMounted(() => {
  void auth.loadMe();
  void loadCategories();
});

function logout() {
  void auth.logout();
}

async function loadCategories() {
  try {
    categories.value = (await getCategories()).items;
  } catch {
    categories.value = [];
  }
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
        <RouterLink :class="{ active: route.name === 'home' }" to="/">首页</RouterLink>
        <div class="nav-menu-item">
          <RouterLink :class="{ active: route.name === 'archive' }" class="nav-parent" to="/archive">
            归档 <span class="nav-caret">⌄</span>
          </RouterLink>
          <div class="nav-submenu">
            <RouterLink to="/archive">全部文章</RouterLink>
            <RouterLink v-for="item in navCategories" :key="item.id" :to="`/archive?category=${encodeURIComponent(item.name)}`">
              {{ item.name }}
            </RouterLink>
          </div>
        </div>
        <div class="nav-menu-item">
          <RouterLink :class="{ active: route.name === 'topics' }" class="nav-parent" to="/topics">
            专题 <span class="nav-caret">⌄</span>
          </RouterLink>
          <div class="nav-submenu">
            <RouterLink to="/topics?topic=blog-system">博客系统</RouterLink>
            <RouterLink to="/topics?topic=vue3-content">Vue3 内容站</RouterLink>
            <RouterLink to="/topics?topic=writing-workflow">写作工作流</RouterLink>
            <RouterLink to="/topics?topic=resource-list">资源清单</RouterLink>
          </div>
        </div>
        <RouterLink :class="{ active: route.name === 'submit' }" to="/submit">投稿</RouterLink>
        <RouterLink :class="{ active: route.path.startsWith('/account') }" to="/account">我的</RouterLink>
        <RouterLink v-if="auth.user?.role === 'admin'" :class="{ active: route.path.startsWith('/admin') }" to="/admin">后台</RouterLink>
      </nav>
      <div class="header-actions">
        <button class="icon-button" type="button" aria-label="搜索" @click="searchOpen = true">⌕</button>
        <button class="icon-button" type="button" aria-label="切换深色模式">◐</button>
        <template v-if="auth.user">
          <RouterLink class="icon-button" to="/account/messages" aria-label="站内信">信</RouterLink>
          <RouterLink class="button-secondary" to="/account">{{ auth.user.displayName }}</RouterLink>
          <button class="button-secondary" type="button" @click="logout">退出</button>
        </template>
        <RouterLink v-else class="button-secondary" to="/login">登录</RouterLink>
      </div>
    </div>
  </header>

  <RouterView />
  <SiteFootbar v-if="showChrome" />
  <SiteSearch v-if="showChrome" v-model:open="searchOpen" />
  <SiteBacktop v-if="showChrome" />
</template>
