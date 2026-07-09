<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";

const visible = ref(false);

function updateVisible() {
  visible.value = window.scrollY > 360;
}

function scrollTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(() => {
  updateVisible();
  window.addEventListener("scroll", updateVisible, { passive: true });
});

onUnmounted(() => {
  window.removeEventListener("scroll", updateVisible);
});
</script>

<template>
  <button class="backtop" :class="{ visible }" type="button" aria-label="回到顶部" @click="scrollTop">↑</button>
</template>
