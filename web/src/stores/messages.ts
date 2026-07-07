import { defineStore } from "pinia";

import { getMessages } from "../shared/api";

interface MessageState {
  unread: number;
  loading: boolean;
}

export const useMessageStore = defineStore("messages", {
  state: (): MessageState => ({
    unread: 0,
    loading: false
  }),
  actions: {
    async refreshUnread() {
      this.loading = true;

      try {
        const response = await getMessages({ status: "unread", page: 1, pageSize: 1 });
        this.unread = response.stats.unread;
      } catch {
        this.unread = 0;
      } finally {
        this.loading = false;
      }
    },
    setUnread(value: number) {
      this.unread = Math.max(0, value);
    },
    clear() {
      this.unread = 0;
    }
  }
});
