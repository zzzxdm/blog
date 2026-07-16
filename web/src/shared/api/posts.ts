/** Domain API: posts */

import { ApiError, request, toQuery } from "./client";

export interface Post {
  id: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  visibility: "public" | "private";
  category: string;
  tags: string[];
  coverImage: string;
  authorId: string;
  authorName: string;
  readingTime: number;
  viewCount: number;
  likeCount: number;
  dislikeCount: number;
  commentCount: number;
  publishedAt: string;
}
export interface SiteStats {
  postCount: number;
  viewCount: number;
  wordCount: number;
}
export interface ListResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
}
export interface PostListParams {
  page?: number;
  pageSize?: number;
  q?: string;
  category?: string;
  tag?: string;
  author?: string;
  sort?: "views" | "comments" | "likes";
}
export async function getPosts(params: PostListParams = {}): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/posts${query}`);
}
export async function getPostBySlug(slug: string): Promise<Post> {
  return request<Post>(`/posts/${encodeURIComponent(slug)}`);
}
export async function getMyPrivatePosts(params: PostListParams = {}): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/me/private-posts${query}`);
}
export async function getSiteStats(): Promise<SiteStats> {
  return request<SiteStats>("/site-stats");
}
export async function searchPosts(params: PostListParams): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/search${query}`);
}
export async function getCategoryPosts(slug: string, params: PostListParams = {}): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/categories/${encodeURIComponent(slug)}/posts${query}`);
}
export async function getTagPosts(slug: string, params: PostListParams = {}): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/tags/${encodeURIComponent(slug)}/posts${query}`);
}
