/** Domain API: reactions */

import { ApiError, request, toQuery } from "./client";

import type { Post } from "./posts";
export interface ReactionSummary {
  postSlug: string;
  likeCount: number;
  dislikeCount: number;
  bookmarkCount: number;
  myReaction: "" | "like" | "dislike";
  bookmarked: boolean;
}
export type BookmarkItem = Post & {
  bookmarkedAt: string;
};
export interface BookmarkListResponse {
  items: BookmarkItem[];
  page: number;
  pageSize: number;
  total: number;
}
export interface BookmarkListParams {
  q?: string;
  category?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
}
export async function getReaction(postSlug: string): Promise<ReactionSummary> {
  return request<ReactionSummary>(`/posts/${encodeURIComponent(postSlug)}/reaction`);
}
export async function setPostReaction(postSlug: string, type: "like" | "dislike" | ""): Promise<ReactionSummary> {
  return request<ReactionSummary>(`/posts/${encodeURIComponent(postSlug)}/reaction`, {
    method: "PUT",
    body: JSON.stringify({ type })
  });
}
export async function setBookmark(postSlug: string, bookmarked: boolean): Promise<ReactionSummary> {
  return request<ReactionSummary>(`/posts/${encodeURIComponent(postSlug)}/bookmark`, {
    method: "PUT",
    body: JSON.stringify({ bookmarked })
  });
}
export async function getMyBookmarks(params: BookmarkListParams = {}): Promise<BookmarkListResponse> {
  return request<BookmarkListResponse>(`/bookmarks/mine${toQuery(params)}`);
}
