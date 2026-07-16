/** Domain API: users */

import { request, toQuery } from "./client";
import type { SessionListResponse } from "./auth";

export interface AdminUsersExport {
  scope: "users";
  exportedAt: string;
  items: ManagedUser[];
  total: number;
  stats: UserStats;
}
export interface ManagedUser {
  id: string;
  email: string;
  displayName: string;
  role: string;
  status: "active" | "muted" | "banned" | "deleted";
  avatarText: string;
  emailVerified: boolean;
  twoFactor: boolean;
  commentCount: number;
  bookmarkCount: number;
  lastLoginAt: string;
  registeredAt: string;
  moderationNote: string;
}
export interface UserStats {
  total: number;
  emailVerified: number;
  authors: number;
  muted: number;
  banned: number;
}
export interface UserListResponse {
  items: ManagedUser[];
  page: number;
  pageSize: number;
  total: number;
  stats: UserStats;
}
export interface AdminUserParams {
  page?: number;
  pageSize?: number;
  q?: string;
  status?: string;
  role?: string;
  all?: boolean;
}
export interface AdminPasswordResetResponse {
  ok: boolean;
  user: ManagedUser;
  resetToken?: string;
  delivery: string;
}
export interface AdminInvitationResponse {
  ok: boolean;
  user: ManagedUser;
  initialPassword?: string;
  resetToken?: string;
  delivery: string;
}
export async function getAdminUsers(params: AdminUserParams = {}): Promise<UserListResponse> {
  const query = toQuery(params);
  return request<UserListResponse>(`/admin/users${query}`);
}
export async function getAdminUser(id: string): Promise<ManagedUser> {
  return request<ManagedUser>(`/admin/users/${encodeURIComponent(id)}`);
}
export async function getAdminUserSessions(id: string): Promise<SessionListResponse> {
  return request<SessionListResponse>(`/admin/users/${encodeURIComponent(id)}/sessions`);
}
export async function exportAdminUsers(): Promise<AdminUsersExport> {
  return request<AdminUsersExport>("/admin/users/export");
}
export async function inviteAdminUser(payload: { email: string; displayName: string; role: string }): Promise<AdminInvitationResponse> {
  return request<AdminInvitationResponse>("/admin/users/invitations", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function updateAdminUserRole(id: string, role: string): Promise<ManagedUser> {
  return request<ManagedUser>(`/admin/users/${encodeURIComponent(id)}/role`, {
    method: "PUT",
    body: JSON.stringify({ role })
  });
}
export async function updateAdminUserStatus(id: string, status: ManagedUser["status"]): Promise<ManagedUser> {
  return request<ManagedUser>(`/admin/users/${encodeURIComponent(id)}/status`, {
    method: "PUT",
    body: JSON.stringify({ status })
  });
}
export async function deleteAdminUser(id: string): Promise<ManagedUser> {
  return request<ManagedUser>(`/admin/users/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
export async function restoreAdminUser(id: string): Promise<ManagedUser> {
  return request<ManagedUser>(`/admin/users/${encodeURIComponent(id)}/restore`, {
    method: "POST"
  });
}
export async function requestAdminUserPasswordReset(id: string): Promise<AdminPasswordResetResponse> {
  return request<AdminPasswordResetResponse>(`/admin/users/${encodeURIComponent(id)}/password-reset`, {
    method: "POST"
  });
}
