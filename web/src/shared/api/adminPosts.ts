/** Domain API: adminPosts */

import { ApiError, request, toQuery } from "./client";

export type AdminPostStatus = "draft" | "review" | "scheduled" | "published" | "archived";
export type AdminPostVisibility = "public" | "private" | "members";
export interface AdminPost {
  id: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
  visibility: AdminPostVisibility;
  category: string;
  tags: string[];
  coverImage: string;
  authorName: string;
  readingTime: number;
  viewCount: number;
  commentCount: number;
  seoTitle: string;
  seoDescription: string;
  version: number;
  publishedPostSlug?: string;
  scheduledAt?: string;
  publishedAt?: string;
  updatedAt: string;
}
export interface AdminPostRevision {
  id: string;
  version: number;
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
  visibility: AdminPostVisibility;
  category: string;
  tags: string[];
  coverImage: string;
  seoTitle: string;
  seoDescription: string;
  authorName: string;
  createdAt: string;
}
export interface AdminPostStats {
  published: number;
  draft: number;
  review: number;
  scheduled: number;
  monthlyViews: string;
  total: number;
}
export interface AdminPostListResponse {
  items: AdminPost[];
  page: number;
  pageSize: number;
  total: number;
  stats: AdminPostStats;
}
export interface AdminPostRevisionListResponse {
  items: AdminPostRevision[];
  total: number;
}
export interface AdminPostPreview {
  previewUrl: string;
  token: string;
  expiresAt: string;
}
export interface AdminPostPayload {
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
  visibility: AdminPostVisibility;
  scheduledAt?: string;
  category: string;
  tags: string[];
  coverImage: string;
  seoTitle: string;
  seoDescription: string;
}
export interface AdminPostParams {
  q?: string;
  status?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}
export async function getAdminPosts(params: AdminPostParams = {}): Promise<AdminPostListResponse> {
  const query = toQuery(params);
  return request<AdminPostListResponse>(`/admin/posts${query}`);
}
export async function getAdminPost(id: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}`);
}
export async function getAdminPostRevisions(id: string): Promise<AdminPostRevisionListResponse> {
  return request<AdminPostRevisionListResponse>(`/admin/posts/${encodeURIComponent(id)}/revisions`);
}
export async function createAdminPost(payload: AdminPostPayload): Promise<AdminPost> {
  return request<AdminPost>("/admin/posts", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function updateAdminPost(id: string, payload: AdminPostPayload): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function deleteAdminPost(id: string): Promise<{ ok: boolean; post: AdminPost }> {
  return request<{ ok: boolean; post: AdminPost }>(`/admin/posts/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
export async function publishAdminPost(id: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/publish`, {
    method: "POST"
  });
}
export async function archiveAdminPost(id: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/archive`, {
    method: "POST"
  });
}
export async function createAdminPostPreview(id: string): Promise<AdminPostPreview> {
  return request<AdminPostPreview>(`/admin/posts/${encodeURIComponent(id)}/preview`, {
    method: "POST"
  });
}
export async function getPreviewPost(token: string): Promise<AdminPost> {
  return request<AdminPost>(`/preview/${encodeURIComponent(token)}`);
}
export async function restoreAdminPostRevision(id: string, revisionId: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/revisions/${encodeURIComponent(revisionId)}/restore`, {
    method: "POST"
  });
}
export async function deleteAdminPostRevision(id: string, revisionId: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/revisions/${encodeURIComponent(revisionId)}`, {
    method: "DELETE"
  });
}
