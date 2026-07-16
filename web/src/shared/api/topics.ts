/** Domain API: topics */

import { request, toQuery } from "./client";
import type { ListResponse, Post, PostListParams } from "./posts";

export type TopicTone = "" | "rust" | "amber" | "gray";
export type TopicStatus = "active" | "draft";
export interface Topic {
  id: string;
  slug: string;
  title: string;
  summary: string;
  coverImage: string;
  imageAlt: string;
  tone: TopicTone;
  status: TopicStatus;
  featured: boolean;
  sortOrder: number;
  categories: string[];
  tags: string[];
  postCount: number;
  latestPostAt?: string;
  createdAt: string;
  updatedAt: string;
}
export interface TopicPayload {
  slug: string;
  title: string;
  summary: string;
  coverImage: string;
  imageAlt: string;
  tone: TopicTone;
  status: TopicStatus;
  featured: boolean;
  sortOrder: number;
  categories: string[];
  tags: string[];
}
export interface TopicListParams {
  page?: number;
  pageSize?: number;
  q?: string;
  status?: string;
  featured?: boolean;
  all?: boolean;
}
export async function getTopics(params: TopicListParams = {}): Promise<ListResponse<Topic>> {
  const query = toQuery(params);
  return request<ListResponse<Topic>>(`/topics${query}`);
}
export async function getTopic(slug: string): Promise<Topic> {
  return request<Topic>(`/topics/${encodeURIComponent(slug)}`);
}
export async function getTopicPosts(slug: string, params: PostListParams = {}): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/topics/${encodeURIComponent(slug)}/posts${query}`);
}
export async function getAdminTopics(params: TopicListParams = {}): Promise<ListResponse<Topic>> {
  const query = toQuery(params);
  return request<ListResponse<Topic>>(`/admin/topics${query}`);
}
export async function createAdminTopic(payload: TopicPayload): Promise<Topic> {
  return request<Topic>("/admin/topics", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function updateAdminTopic(id: string, payload: TopicPayload): Promise<Topic> {
  return request<Topic>(`/admin/topics/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function deleteAdminTopic(id: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/admin/topics/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
