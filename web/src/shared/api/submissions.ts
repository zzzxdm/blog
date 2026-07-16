/** Domain API: submissions */

import { ApiError, request, toQuery } from "./client";

import type { ManageListParams } from "./comments";
export type SubmissionStatus = "draft" | "submitted" | "returned" | "rejected" | "published" | "archived";
export type SubmissionVisibility = "public" | "private";
export interface Submission {
  id: string;
  authorId: string;
  authorName: string;
  authorAvatar: string;
  title: string;
  summary: string;
  content: string;
  category: string;
  tags: string[];
  coverImage: string;
  slug: string;
  visibility: SubmissionVisibility;
  status: SubmissionStatus;
  reviewNote: string;
  reviewerId?: string;
  reviewerName?: string;
  publishedPostSlug?: string;
  wordCount: number;
  version: number;
  riskLevel: string;
  createdAt: string;
  updatedAt: string;
  submittedAt?: string;
  reviewedAt?: string;
  publishedAt?: string;
}
export interface SubmissionStats {
  draft: number;
  submitted: number;
  returned: number;
  rejected: number;
  published: number;
  archived: number;
  total: number;
}
export interface SubmissionListResponse {
  items: Submission[];
  page: number;
  pageSize: number;
  total: number;
  stats: SubmissionStats;
}
export interface SubmissionPayload {
  title: string;
  summary: string;
  content: string;
  category: string;
  tags: string[];
  coverImage: string;
  slug: string;
  visibility: SubmissionVisibility;
  submit?: boolean;
  turnstileToken?: string;
}
export interface ReviewPayload {
  action: "approve" | "return" | "reject";
  note: string;
  slug?: string;
  category?: string;
}
export async function getMySubmissions(params: ManageListParams | string = {}): Promise<SubmissionListResponse> {
  const query = toQuery(typeof params === "string" ? { status: params } : params);
  return request<SubmissionListResponse>(`/submissions${query}`);
}
export async function createSubmission(payload: SubmissionPayload): Promise<Submission> {
  return request<Submission>("/submissions", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function updateSubmission(id: string, payload: SubmissionPayload): Promise<Submission> {
  return request<Submission>(`/submissions/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function submitExistingSubmission(id: string, turnstileToken = ""): Promise<Submission> {
  return request<Submission>(`/submissions/${encodeURIComponent(id)}/submit`, {
    method: "POST",
    body: JSON.stringify({ turnstileToken })
  });
}
export async function getAdminSubmissions(params: ManageListParams | string = {}): Promise<SubmissionListResponse> {
  const query = toQuery(typeof params === "string" ? { status: params } : params);
  return request<SubmissionListResponse>(`/admin/submissions${query}`);
}
export async function updateAdminSubmission(id: string, payload: SubmissionPayload): Promise<Submission> {
  return request<Submission>(`/admin/submissions/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function reviewSubmission(id: string, payload: ReviewPayload): Promise<Submission> {
  return request<Submission>(`/admin/submissions/${encodeURIComponent(id)}/review`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function archiveSubmission(id: string): Promise<Submission> {
  return request<Submission>(`/admin/submissions/${encodeURIComponent(id)}/archive`, {
    method: "POST"
  });
}
export async function restoreSubmission(id: string): Promise<Submission> {
  return request<Submission>(`/admin/submissions/${encodeURIComponent(id)}/restore`, {
    method: "POST"
  });
}
