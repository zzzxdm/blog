<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";

const siteStartDate = new Date(2020, 7, 21, 0, 0, 0);
const currentYear = new Date().getFullYear();
const runtimeText = ref("");

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
        <nav class="social-links" aria-label="站点链接">
          <a class="social-link" href="https://github.com/" aria-label="GitHub">GH</a>
          <a class="social-link" href="mailto:hello@example.com" aria-label="邮箱">✉</a>
        </nav>
      </div>
    </div>
  </footer>
</template>
