/** Domain API: messages */

import { ApiError, request, toQuery } from "./client";

export type MessageType = "review" | "comment" | "system" | "admin" | "account";
export type MessageStatus = "unread" | "read" | "archived" | "scheduled";
export interface StationMessage {
  id: string;
  recipientId: string;
  recipientName: string;
  senderId: string;
  senderName: string;
  type: MessageType;
  priority: string;
  title: string;
  body: string;
  targetType?: string;
  targetId?: string;
  targetTitle?: string;
  status: MessageStatus;
  readAt?: string;
  archivedAt?: string;
  scheduledAt?: string;
  createdAt: string;
}
export interface MessageStats {
  unread: number;
  review: number;
  admin: number;
  archived: number;
  scheduled: number;
  total: number;
}
export interface MessageListResponse {
  items: StationMessage[];
  page: number;
  pageSize: number;
  total: number;
  stats: MessageStats;
}
export interface AdminMessagesExport {
  scope: "messages";
  exportedAt: string;
  items: StationMessage[];
  total: number;
  stats: MessageStats;
}
export interface MessageListParams {
  status?: string;
  type?: string;
  q?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}
export async function getMessages(params: MessageListParams = {}): Promise<MessageListResponse> {
  const query = toQuery(params);
  return request<MessageListResponse>(`/messages${query}`);
}
export async function markMessageRead(id: string): Promise<StationMessage> {
  return request<StationMessage>(`/messages/${encodeURIComponent(id)}/read`, {
    method: "PUT"
  });
}
export async function markAllMessagesRead(): Promise<{ stats: MessageStats }> {
  return request<{ stats: MessageStats }>("/messages/read-all", {
    method: "PUT"
  });
}
export async function archiveMessage(id: string): Promise<StationMessage> {
  return request<StationMessage>(`/messages/${encodeURIComponent(id)}/archive`, {
    method: "PUT"
  });
}
export async function getAdminMessages(params: MessageListParams = {}): Promise<MessageListResponse> {
  const query = toQuery(params);
  return request<MessageListResponse>(`/admin/messages${query}`);
}
export async function exportAdminMessages(params: { status?: string; type?: string } = {}): Promise<AdminMessagesExport> {
  const query = toQuery(params);
  return request<AdminMessagesExport>(`/admin/messages/export${query}`);
}
export async function createAdminMessage(payload: {
  recipientId: string;
  recipientName: string;
  type: MessageType;
  priority: string;
  title: string;
  body: string;
  targetType?: string;
  targetId?: string;
  targetTitle?: string;
  scheduledAt?: string;
}): Promise<StationMessage> {
  return request<StationMessage>("/admin/messages", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function broadcastAdminMessage(payload: {
  recipients: Array<{ id: string; name: string }>;
  type: MessageType;
  priority: string;
  title: string;
  body: string;
  targetType?: string;
  targetId?: string;
  targetTitle?: string;
  scheduledAt?: string;
}): Promise<{ items: StationMessage[]; total: number }> {
  return request<{ items: StationMessage[]; total: number }>("/admin/messages/broadcast", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function revokeAdminMessage(id: string): Promise<{ ok: boolean; message: StationMessage }> {
  return request<{ ok: boolean; message: StationMessage }>(`/admin/messages/${encodeURIComponent(id)}/revoke`, {
    method: "POST"
  });
}
export async function getAdminMessageStatistics(id: string): Promise<Record<string, unknown>> {
  return request<Record<string, unknown>>(`/admin/messages/${encodeURIComponent(id)}/statistics`);
}
