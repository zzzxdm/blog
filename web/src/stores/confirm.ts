import { defineStore } from "pinia";
import { ref } from "vue";

export type ConfirmTone = "default" | "danger" | "success";

export interface ConfirmOptions {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  tone?: ConfirmTone;
}

export interface ConfirmState extends Required<ConfirmOptions> {
  id: number;
}

export const useConfirmStore = defineStore("confirm", () => {
  const current = ref<ConfirmState | null>(null);
  let nextId = 1;
  let resolveCurrent: ((confirmed: boolean) => void) | null = null;

  function open(options: ConfirmOptions) {
    if (resolveCurrent) {
      resolveCurrent(false);
    }

    current.value = {
      id: nextId++,
      title: options.title,
      message: options.message,
      confirmText: options.confirmText || "确认",
      cancelText: options.cancelText || "取消",
      tone: options.tone || "default"
    };

    return new Promise<boolean>((resolve) => {
      resolveCurrent = resolve;
    });
  }

  function close(confirmed: boolean) {
    if (!resolveCurrent) {
      current.value = null;
      return;
    }

    const resolve = resolveCurrent;
    resolveCurrent = null;
    current.value = null;
    resolve(confirmed);
  }

  function confirm() {
    close(true);
  }

  function cancel() {
    close(false);
  }

  return {
    current,
    open,
    confirm,
    cancel
  };
});
