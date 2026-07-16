/** Domain API: auth */

import { ApiError, request, toQuery } from "./client";

export interface User {
  id: string;
  email: string;
  displayName: string;
  role: string;
  status: string;
  avatarText: string;
  emailVerified: boolean;
}
export interface AuthResponse {
  user: User;
  verificationToken?: string;
  delivery?: string;
  warning?: string;
}
export interface TokenResponse {
  ok: boolean;
  verificationToken?: string;
  resetToken?: string;
  delivery?: string;
  warning?: string;
}
export interface SessionInfo {
  id: string;
  device: string;
  current: boolean;
  createdAt: string;
  expiresAt: string;
}
export interface SessionListResponse {
  items: SessionInfo[];
  total: number;
}
export interface ExportData {
  user: User;
  sessions: SessionInfo[];
  commentCount: number;
  bookmarkCount: number;
  exportedAt: string;
}
export interface AccountSettings {
  displayName: string;
  username: string;
  email: string;
  emailVerified: boolean;
  avatarText: string;
  bio: string;
  twoFactor: boolean;
  loginAlert: boolean;
  notifyReview: boolean;
  notifyComment: boolean;
  notifyAnnouncement: boolean;
  emailNotification: boolean;
  publicProfile: boolean;
  publicBookmarks: boolean;
  profileUrl: string;
  timezone: string;
  securityLevel: string;
  loginDeviceCount: number;
  publicPostCount: number;
  profileCompleteness: number;
  currentDeviceDescription: string;
  lastDeviceDescription: string;
  updatedAt: string;
}
export async function login(email: string, password: string, turnstileToken = ""): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password, turnstileToken })
  });
}
export async function register(email: string, password: string, displayName: string, turnstileToken = ""): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, displayName, turnstileToken })
  });
}
export async function logout(): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>("/auth/logout", { method: "POST" });
}
export async function requestEmailVerification(): Promise<TokenResponse> {
  return request<TokenResponse>("/auth/email-verification", { method: "POST" });
}
export async function verifyEmail(token: string): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/verify-email", {
    method: "POST",
    body: JSON.stringify({ token })
  });
}
export async function forgotPassword(email: string): Promise<TokenResponse> {
  return request<TokenResponse>("/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify({ email })
  });
}
export async function resetPassword(token: string, newPassword: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>("/auth/reset-password", {
    method: "POST",
    body: JSON.stringify({ token, newPassword })
  });
}
export async function getMe(): Promise<AuthResponse> {
  return request<AuthResponse>("/me");
}
export async function getMySessions(): Promise<SessionListResponse> {
  return request<SessionListResponse>("/me/sessions");
}
export async function deleteMySession(id: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>(`/me/sessions/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
export async function exportMyData(): Promise<ExportData> {
  return request<ExportData>("/me/export", { method: "POST" });
}
export async function deleteMyAccount(): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>("/me", { method: "DELETE" });
}
export async function getAccountSettings(): Promise<AccountSettings> {
  return request<AccountSettings>("/account/settings");
}
export async function updateAccountSettings(payload: AccountSettings): Promise<AccountSettings> {
  return request<AccountSettings>("/account/settings", {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function changePassword(currentPassword: string, newPassword: string): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>("/me/password", {
    method: "PUT",
    body: JSON.stringify({ currentPassword, newPassword })
  });
}
