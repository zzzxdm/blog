<script setup lang="ts">
import { RouterLink, RouterView } from "vue-router";
</script>

<template>
  <header class="site-header">
    <RouterLink class="brand" to="/">
      <span class="brand-mark">云</span>
      <span>云间笔记</span>
    </RouterLink>
    <nav class="nav-links" aria-label="主导航">
      <RouterLink to="/">首页</RouterLink>
      <RouterLink to="/archive">归档</RouterLink>
      <RouterLink to="/admin">后台</RouterLink>
    </nav>
  </header>

  <RouterView />
</template>

