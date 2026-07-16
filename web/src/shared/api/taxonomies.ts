/** Domain API: taxonomies */

import { ApiError, request, toQuery } from "./client";

export interface Category {
  id: string;
  slug: string;
  name: string;
  description: string;
  sortOrder: number;
  postCount: number;
}
export interface Tag {
  id: string;
  slug: string;
  name: string;
  postCount: number;
}
export interface TaxonomyListResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
}
export interface SaveCategoryPayload {
  slug: string;
  name: string;
  description: string;
  sortOrder: number;
}
export interface SaveTagPayload {
  slug: string;
  name: string;
}
export interface TaxonomyListParams {
  page?: number;
  pageSize?: number;
  q?: string;
}
export async function getCategories(params: TaxonomyListParams = {}): Promise<TaxonomyListResponse<Category>> {
  const query = toQuery(params);
  return request<TaxonomyListResponse<Category>>(`/categories${query}`);
}
export async function getTags(params: TaxonomyListParams = {}): Promise<TaxonomyListResponse<Tag>> {
  const query = toQuery(params);
  return request<TaxonomyListResponse<Tag>>(`/tags${query}`);
}
export async function createAdminCategory(payload: SaveCategoryPayload): Promise<Category> {
  return request<Category>("/admin/categories", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function updateAdminCategory(id: string, payload: SaveCategoryPayload): Promise<Category> {
  return request<Category>(`/admin/categories/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function deleteAdminCategory(id: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/admin/categories/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
export async function createAdminTag(payload: SaveTagPayload): Promise<Tag> {
  return request<Tag>("/admin/tags", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function updateAdminTag(id: string, payload: SaveTagPayload): Promise<Tag> {
  return request<Tag>(`/admin/tags/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function deleteAdminTag(id: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/admin/tags/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
