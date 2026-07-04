<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { RouterLink } from "vue-router";

import type { NavItem, OperationsNavigation } from "../shared/api";

const props = withDefaults(defineProps<{
  navigation?: OperationsNavigation | null;
}>(), {
  navigation: null
});

const siteStartDate = new Date(2020, 7, 21, 0, 0, 0);
const currentYear = new Date().getFullYear();
const runtimeText = ref("");
const defaultFooterItems: NavItem[] = [
  { id: "footer_default_home", label: "首页", url: "/", order: 1 },
  { id: "footer_default_archive", label: "归档", url: "/archive", order: 2 },
  { id: "footer_default_topics", label: "专题", url: "/topics", order: 3 }
];
const footerItems = computed(() => orderedNavItems(props.navigation?.footerItems ?? defaultFooterItems));
const socialLinks = computed(() => {
  const githubUrl = props.navigation ? props.navigation.githubUrl.trim() : "https://github.com/";
  const contactEmail = props.navigation ? props.navigation.contactEmail.trim() : "hello@example.com";
  const links: Array<{ id: string; label: string; href: string; title: string }> = [];

  if (githubUrl) {
    links.push({ id: "github", label: "GH", href: githubUrl, title: "GitHub" });
  }
  if (contactEmail) {
    links.push({ id: "email", label: "@", href: `mailto:${contactEmail}`, title: "邮箱" });
  }

  return links;
});

let timer: number | undefined;

function getRuntimeText(now = new Date()) {
  let years = now.getFullYear() - siteStartDate.getFullYear();
  let anniversary = new Date(siteStartDate);
  anniversary.setFullYear(siteStartDate.getFullYear() + years);

  if (anniversary > now) {
    years -= 1;
    anniversary = new Date(siteStartDate);
    anniversary.setFullYear(siteStartDate.getFullYear() + years);
  }

  const totalSeconds = Math.max(0, Math.floor((now.getTime() - anniversary.getTime()) / 1000));
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  return `本站已运行 ${years} 年 ${days} 天 ${hours} 小时 ${minutes} 分钟 ${seconds} 秒`;
}

function updateRuntime() {
  runtimeText.value = getRuntimeText();
}

function orderedNavItems(items: NavItem[]) {
  return [...items]
    .filter((item) => item.label.trim() && item.url.trim())
    .sort((left, right) => left.order - right.order);
}

function isRouterUrl(url: string) {
  return url.startsWith("/") && !url.startsWith("//") && !/\.[a-z0-9]+($|[?#])/i.test(url);
}

function opensNewWindow(url: string) {
  if (!(props.navigation?.externalLinksNewWindow ?? true)) {
    return false;
  }

  return /^(https?:)?\/\//.test(url);
}

onMounted(() => {
  updateRuntime();
  timer = window.setInterval(updateRuntime, 1000);
});

onUnmounted(() => {
  if (timer) {
    window.clearInterval(timer);
  }
});
</script>

<template>
  <footer class="footbar">
    <div class="footbar-inner">
      <div class="footbar-copy">
        <div>Copyright © {{ currentYear }} <strong>云间笔记</strong>.</div>
        <div>{{ runtimeText }}</div>
        <div>总访问量: 47087 人次 <span>|</span> 访客人数: 28294 人 <span>|</span> 字数统计: 46.8k 字</div>
      </div>
      <div class="footbar-tools">
        <nav v-if="footerItems.length" class="footer-links" aria-label="底部导航">
          <template v-for="item in footerItems" :key="item.id">
            <RouterLink v-if="isRouterUrl(item.url)" :to="item.url">{{ item.label }}</RouterLink>
            <a
              v-else
              :href="item.url"
              :target="opensNewWindow(item.url) ? '_blank' : undefined"
              :rel="opensNewWindow(item.url) ? 'noreferrer' : undefined"
            >
              {{ item.label }}
            </a>
          </template>
        </nav>
        <nav v-if="socialLinks.length" class="social-links" aria-label="社交链接">
          <a
            v-for="link in socialLinks"
            :key="link.id"
            class="social-link"
            :href="link.href"
            :aria-label="link.title"
            :target="opensNewWindow(link.href) ? '_blank' : undefined"
            :rel="opensNewWindow(link.href) ? 'noreferrer' : undefined"
          >
            {{ link.label }}
          </a>
        </nav>
      </div>
    </div>
  </footer>
</template>
