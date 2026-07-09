import { defineStore } from "pinia";

import { getHealth, type HealthResponse } from "../shared/api";

interface HealthState {
  data: HealthResponse | null;
  loading: boolean;
  error: string | null;
}

export const useHealthStore = defineStore("health", {
  state: (): HealthState => ({
    data: null,
    loading: false,
    error: null
  }),
  actions: {
    async load() {
      this.loading = true;
      this.error = null;

      try {
        this.data = await getHealth();
      } catch (error) {
        this.error = error instanceof Error ? error.message : "Unknown error";
      } finally {
        this.loading = false;
      }
    }
  }
});
