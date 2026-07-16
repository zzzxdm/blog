/** Domain API: comments */

import { ApiError, request, toQuery } from "./client";

export interface Comment {
  id: string;
  postSlug: string;
  postTitle?: string;
  parentId?: string;
  authorId: string;
  authorName: string;
  avatarText: string;
  body: string;
  status: "approved" | "pending" | "rejected" | "spam" | "deleted";
  likeCount: number;
  replyCount?: number;
  riskLevel?: string;
  isMine: boolean;
  isAuthor: boolean;
  liked: boolean;
  createdAt: string;
}
export interface CommentListResponse {
  items: Comment[];
  total: number;
}
export interface CommentStats {
  total: number;
  pending: number;
  approved: number;
  rejected: number;
  spam: number;
  deleted: number;
  likes: number;
  replies: number;
}
export interface CommentManageListResponse {
  items: Comment[];
  page: number;
  pageSize: number;
  total: number;
  stats: CommentStats;
}
export interface CommentReport {
  id: string;
  commentId: string;
  reporterId: string;
  reason: string;
  status: string;
  createdAt: string;
}
export interface CommentReportListResponse {
  items: CommentReport[];
  total: number;
}
export interface AdminCommentsExport {
  scope: "comments";
  exportedAt: string;
  items: Comment[];
  total: number;
  stats: CommentStats;
}
export interface ManageListParams {
  status?: string;
  q?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}
export async function getComments(postSlug: string): Promise<CommentListResponse> {
  return request<CommentListResponse>(`/posts/${encodeURIComponent(postSlug)}/comments`);
}
export async function createComment(postSlug: string, body: string, parentId = ""): Promise<Comment> {
  return request<Comment>(`/posts/${encodeURIComponent(postSlug)}/comments`, {
    method: "POST",
    body: JSON.stringify({ body, parentId })
  });
}
export async function toggleCommentLike(id: string): Promise<Comment> {
  return request<Comment>(`/comments/${encodeURIComponent(id)}/like`, {
    method: "PUT"
  });
}
export async function reportComment(id: string, reason: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/comments/${encodeURIComponent(id)}/report`, {
    method: "POST",
    body: JSON.stringify({ reason })
  });
}
export async function getMyComments(params: ManageListParams | string = {}): Promise<CommentManageListResponse> {
  const query = toQuery(typeof params === "string" ? { status: params } : params);
  return request<CommentManageListResponse>(`/comments/mine${query}`);
}
export async function getAdminComments(params: ManageListParams | string = {}): Promise<CommentManageListResponse> {
  const query = toQuery(typeof params === "string" ? { status: params } : params);
  return request<CommentManageListResponse>(`/admin/comments${query}`);
}
export async function exportAdminComments(status = ""): Promise<AdminCommentsExport> {
  const query = toQuery({ status });
  return request<AdminCommentsExport>(`/admin/comments/export${query}`);
}
export async function updateCommentStatus(id: string, status: Comment["status"]): Promise<Comment> {
  return request<Comment>(`/admin/comments/${encodeURIComponent(id)}/status`, {
    method: "PUT",
    body: JSON.stringify({ status })
  });
}
export async function deleteAdminComment(id: string): Promise<{ ok: boolean; comment: Comment }> {
  return request<{ ok: boolean; comment: Comment }>(`/admin/comments/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
export async function getAdminCommentReports(status = ""): Promise<CommentReportListResponse> {
  const query = toQuery({ status });
  return request<CommentReportListResponse>(`/admin/comment-reports${query}`);
}
export async function updateAdminCommentReportStatus(id: string, status: string): Promise<CommentReport> {
  return request<CommentReport>(`/admin/comment-reports/${encodeURIComponent(id)}/status`, {
    method: "PUT",
    body: JSON.stringify({ status })
  });
}
