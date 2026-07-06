import { defineStore } from "pinia";

export type ToastTone = "success" | "error" | "info" | "warning";

export interface ToastItem {
  id: number;
  tone: ToastTone;
  title: string;
  message: string;
}

interface ToastState {
  items: ToastItem[];
  nextId: number;
}

export const useToastStore = defineStore("toast", {
  state: (): ToastState => ({
    items: [],
    nextId: 1
  }),
  actions: {
    push(toast: Omit<ToastItem, "id">, timeout = 3600) {
      const id = this.nextId++;
      this.items.push({ id, ...toast });
      window.setTimeout(() => {
        this.remove(id);
      }, timeout);
      return id;
    },
    success(title: string, message = "") {
      return this.push({ tone: "success", title, message });
    },
    error(title: string, message = "") {
      return this.push({ tone: "error", title, message }, 5200);
    },
    warning(title: string, message = "") {
      return this.push({ tone: "warning", title, message }, 4600);
    },
    info(title: string, message = "") {
      return this.push({ tone: "info", title, message });
    },
    remove(id: number) {
      this.items = this.items.filter((item) => item.id !== id);
    }
  }
});
