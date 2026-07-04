import { defineStore } from "pinia";

import {
  ApiError,
  getMe,
  login,
  logout,
  register,
  type User
} from "../shared/api";

interface AuthState {
  user: User | null;
  loading: boolean;
  error: string | null;
}

export const useAuthStore = defineStore("auth", {
  state: (): AuthState => ({
    user: null,
    loading: false,
    error: null
  }),
  actions: {
    async loadMe() {
      this.loading = true;
      this.error = null;

      try {
        const response = await getMe();
        this.user = response.user;
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) {
          this.user = null;
          return;
        }

        this.error = error instanceof Error ? error.message : "加载用户失败";
      } finally {
        this.loading = false;
      }
    },
    async login(email: string, password: string) {
      this.loading = true;
      this.error = null;

      try {
        const response = await login(email, password);
        this.user = response.user;
      } catch (error) {
        this.error = error instanceof Error ? error.message : "登录失败";
        throw error;
      } finally {
        this.loading = false;
      }
    },
    async register(email: string, password: string, displayName: string) {
      this.loading = true;
      this.error = null;

      try {
        const response = await register(email, password, displayName);
        this.user = response.user;
      } catch (error) {
        this.error = error instanceof Error ? error.message : "注册失败";
        throw error;
      } finally {
        this.loading = false;
      }
    },
    async logout() {
      this.loading = true;
      this.error = null;

      try {
        await logout();
      } finally {
        this.user = null;
        this.loading = false;
      }
    }
  }
});
