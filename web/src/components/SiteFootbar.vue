<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { RouterLink } from "vue-router";

import { getSiteStats, type NavItem, type OperationsNavigation, type SiteStats } from "../shared/api";

const props = withDefaults(defineProps<{
  navigation?: OperationsNavigation | null;
  siteName?: string;
  beian?: string;
}>(), {
  navigation: null,
  siteName: "云间笔记",
  beian: ""
});

const siteStartDate = new Date(2020, 7, 21, 0, 0, 0);
const currentYear = new Date().getFullYear();
const runtimeText = ref("");
const siteStats = ref<SiteStats | null>(null);
const footerItems = computed(() => orderedNavItems(props.navigation?.footerItems ?? []));
const socialLinks = computed(() => {
  const githubUrl = props.navigation ? props.navigation.githubUrl.trim() : "https://github.com/zzzxdm/blog";
  const contactEmail = props.navigation ? props.navigation.contactEmail.trim() : "admin@jecyai.com";
  const links: Array<{ id: string; label: string; href: string; title: string; icon?: "linuxdo" }> = [];

  if (githubUrl) {
    links.push({ id: "github", label: "GH", href: githubUrl, title: "GitHub" });
  }
  links.push({ id: "linuxdo", label: "LINUX DO", href: "https://linux.do/u/zzzxdm/summary", title: "Linux.do", icon: "linuxdo" });
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

async function loadSiteStats() {
  try {
    siteStats.value = await getSiteStats();
  } catch {
    siteStats.value = null;
  }
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN").format(value);
}

function formatWords(value: number) {
  if (value >= 10000) {
    return `${(value / 10000).toFixed(1)}w`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}k`;
  }

  return String(value);
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
  void loadSiteStats();
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
        <div>Copyright © {{ currentYear }} <strong>{{ siteName }}</strong>.</div>
        <div>{{ runtimeText }}</div>
        <div v-if="beian">{{ beian }}</div>
        <div v-if="siteStats">
          总阅读量: {{ formatNumber(siteStats.viewCount) }} 次 <span>|</span>
          文章数: {{ formatNumber(siteStats.postCount) }} 篇 <span>|</span>
          字数统计: {{ formatWords(siteStats.wordCount) }} 字
        </div>
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
            <svg
              v-if="link.icon === 'linuxdo'"
              class="linuxdo-icon"
              viewBox="0 0 32 32"
              aria-hidden="true"
              focusable="false"
            >
              <rect class="linuxdo-icon-back" x="4" y="4" width="24" height="24" rx="7" />
              <clipPath id="linuxdoIconClip">
                <circle cx="16" cy="16" r="12" />
              </clipPath>
              <g clip-path="url(#linuxdoIconClip)">
                <rect class="linuxdo-icon-top" x="4" y="4" width="24" height="8" />
                <rect class="linuxdo-icon-middle" x="4" y="12" width="24" height="10" />
                <rect class="linuxdo-icon-bottom" x="4" y="22" width="24" height="6" />
              </g>
              <circle class="linuxdo-icon-ring" cx="16" cy="16" r="12" />
            </svg>
            <span v-else>{{ link.label }}</span>
          </a>
        </nav>
      </div>
    </div>
  </footer>
</template>
