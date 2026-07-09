<script setup lang="ts">
import { nextTick, ref, watch } from "vue";

import { useConfirmStore } from "../stores/confirm";

const confirmDialog = useConfirmStore();
const confirmButton = ref<HTMLButtonElement | null>(null);

watch(() => confirmDialog.current?.id, async (id) => {
  if (!id) {
    return;
  }

  await nextTick();
  confirmButton.value?.focus();
});

function onKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    confirmDialog.cancel();
  }
}

function toneLabel(tone: string) {
  if (tone === "danger") return "高风险操作";
  if (tone === "success") return "恢复确认";
  return "操作确认";
}
</script>

<template>
  <Teleport to="body">
    <Transition name="confirm-dialog">
      <div
        v-if="confirmDialog.current"
        class="confirm-overlay"
        role="presentation"
        @click.self="confirmDialog.cancel"
        @keydown="onKeydown"
      >
        <section
          class="confirm-panel"
          :class="confirmDialog.current.tone"
          role="dialog"
          aria-modal="true"
          aria-labelledby="confirm-dialog-title"
          aria-describedby="confirm-dialog-message"
        >
          <div class="confirm-mark" aria-hidden="true">{{ confirmDialog.current.tone === "danger" ? "!" : "✓" }}</div>
          <div class="confirm-copy">
            <div class="confirm-eyebrow">{{ toneLabel(confirmDialog.current.tone) }}</div>
            <h2 id="confirm-dialog-title">{{ confirmDialog.current.title }}</h2>
            <p id="confirm-dialog-message">{{ confirmDialog.current.message }}</p>
          </div>
          <div class="confirm-actions">
            <button class="button-secondary" type="button" @click="confirmDialog.cancel">{{ confirmDialog.current.cancelText }}</button>
            <button
              ref="confirmButton"
              class="button"
              :class="{ 'button-danger': confirmDialog.current.tone === 'danger', 'button-success': confirmDialog.current.tone === 'success' }"
              type="button"
              @click="confirmDialog.confirm"
            >
              {{ confirmDialog.current.confirmText }}
            </button>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>
