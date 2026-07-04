import { defineStore } from "pinia";

import {
  getPostBySlug,
  getPosts,
  type ListResponse,
  type Post,
  type PostListParams
} from "../shared/api";

interface PostsState {
  list: ListResponse<Post> | null;
  current: Post | null;
  loading: boolean;
  error: string | null;
}

export const usePostsStore = defineStore("posts", {
  state: (): PostsState => ({
    list: null,
    current: null,
    loading: false,
    error: null
  }),
  actions: {
    async loadList(params: PostListParams = {}) {
      this.loading = true;
      this.error = null;

      try {
        this.list = await getPosts(params);
      } catch (error) {
        this.error = error instanceof Error ? error.message : "Unknown error";
      } finally {
        this.loading = false;
      }
    },
    async loadBySlug(slug: string) {
      this.loading = true;
      this.error = null;
      this.current = null;

      try {
        this.current = await getPostBySlug(slug);
      } catch (error) {
        this.error = error instanceof Error ? error.message : "Unknown error";
      } finally {
        this.loading = false;
      }
    }
  }
});

