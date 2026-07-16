/** Domain API: operations */

import { API_BASE_URL, ApiError, apiErrorFromFetch, apiErrorFromResponse, csrfHeaders, ensureCsrfCookie, request, toQuery } from "./client";

export interface OperationsSettings {
  siteName: string;
  siteDescription: string;
  siteUrl: string;
  beian: string;
  themePrimary: string;
  homepageLayout: string;
  darkModeEnabled: boolean;
  readingProgressEnabled: boolean;
  commentsEnabled: boolean;
  loginRequiredForComment: boolean;
  autoApproveComments: boolean;
  blockedWords: string[];
  submissionsEnabled: boolean;
  submissionManualReview: boolean;
  submissionLimit: string;
  submissionGuide: string;
  mailEnabled: boolean;
  mailProvider: string;
  fromEmail: string;
  turnstileEnabled: boolean;
  turnstileSiteKey: string;
  turnstileSecretKey: string;
  turnstileRegister: boolean;
  turnstileLogin: boolean;
  turnstileSubmission: boolean;
  adminTwoFactorRequired: boolean;
  loginFailureLock: boolean;
  sessionDays: number;
  backupCycle: string;
  lastBackupAt: string;
  backupRetentionDays: number;
  updatedAt: string;
}
export interface SiteSettings {
  siteName: string;
  siteDescription: string;
  siteUrl: string;
  beian: string;
  themePrimary: string;
  homepageLayout: string;
  darkModeEnabled: boolean;
  readingProgressEnabled: boolean;
  commentsEnabled: boolean;
  loginRequiredForComment: boolean;
  submissionsEnabled: boolean;
  submissionLimit: string;
  submissionGuide: string;
  turnstileEnabled: boolean;
  turnstileSiteKey: string;
  turnstileRegister: boolean;
  turnstileLogin: boolean;
  turnstileSubmission: boolean;
  updatedAt: string;
}
export interface TestMailResult {
  ok: boolean;
  provider: string;
  fromEmail: string;
  delivery: string;
  message: string;
  testedAt: string;
}
export interface BackupResult {
  ok: boolean;
  id: string;
  status: string;
  fileName: string;
  sizeLabel: string;
  message: string;
  createdAt: string;
  settings: OperationsSettings;
}
export interface AdminJob {
  id: string;
  type: string;
  scope: string;
  status: string;
  progress: number;
  message: string;
  fileName?: string;
  downloadUrl?: string;
  result?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}
export interface NavItem {
  id: string;
  label: string;
  url: string;
  order: number;
}
export interface RedirectRule {
  from: string;
  to: string;
  code: number;
}
export interface RedirectListResponse {
  items: RedirectRule[];
  total: number;
}
export interface OperationsNavigation {
  topItems: NavItem[];
  footerItems: NavItem[];
  mobileCollapse: boolean;
  externalLinksNewWindow: boolean;
  showLoginEntry: boolean;
  githubUrl: string;
  contactEmail: string;
  rssUrl: string;
  redirects: RedirectRule[];
  updatedAt: string;
}
export interface MediaAsset {
  id: string;
  fileName: string;
  url: string;
  alt: string;
  type: string;
  category: string;
  sizeLabel: string;
  width: number;
  height: number;
  usageCount: number;
  uploadedBy: string;
  uploadedAt: string;
}
export interface MediaListResponse {
  items: MediaAsset[];
  page: number;
  pageSize: number;
  total: number;
}
export interface MediaReference {
  id: string;
  resourceId: string;
  resourceType: string;
  title: string;
  context: string;
  status: string;
  url: string;
  adminUrl: string;
  updatedAt: string;
}
export interface MediaReferenceListResponse {
  items: MediaReference[];
  page: number;
  pageSize: number;
  total: number;
}
export interface StatMetric {
  label: string;
  value: string;
  delta: string;
}
export interface BarPoint {
  label: string;
  value: string;
  percent: number;
  tone?: string;
}
export interface TopPostStat {
  title: string;
  views: string;
  bookmarks: number;
  comments: number;
  engagementRate: string;
}
export interface SearchTermStat {
  term: string;
  count: number;
}
export interface ContentSuggestion {
  title: string;
  body: string;
}
export interface AdminStats {
  range: string;
  rangeLabel: string;
  metrics: StatMetric[];
  trend: BarPoint[];
  topPosts: TopPostStat[];
  sources: BarPoint[];
  searchTerms: SearchTermStat[];
  suggestions: ContentSuggestion[];
}
export interface AdminStatsExport {
  scope: "stats";
  exportedAt: string;
  stats: AdminStats;
}
export interface AuditLog {
  id: string;
  actorId: string;
  actorName: string;
  action: string;
  resourceType: string;
  resourceId: string;
  resourceTitle: string;
  status: "success" | "blocked" | "error";
  ip: string;
  userAgent: string;
  detail: string;
  createdAt: string;
}
export interface AuditLogListResponse {
  items: AuditLog[];
  page: number;
  pageSize: number;
  total: number;
}
export interface AuditLogParams {
  page?: number;
  pageSize?: number;
  action?: string;
  resourceType?: string;
}
export interface MediaListParams {
  q?: string;
  type?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}
export async function getAdminSettings(): Promise<OperationsSettings> {
  return request<OperationsSettings>("/admin/settings");
}
export async function updateAdminSettings(payload: OperationsSettings): Promise<OperationsSettings> {
  return request<OperationsSettings>("/admin/settings", {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function sendAdminTestMail(): Promise<TestMailResult> {
  return request<TestMailResult>("/admin/settings/test-mail", {
    method: "POST"
  });
}
export async function createAdminBackup(): Promise<BackupResult> {
  return request<BackupResult>("/admin/backups", {
    method: "POST"
  });
}
export async function createAdminImportJob(payload: { scope: string; fileName?: string }): Promise<AdminJob> {
  return request<AdminJob>("/admin/import", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function createAdminExportJob(payload: { scope: string; fileName?: string }): Promise<AdminJob> {
  return request<AdminJob>("/admin/export", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function getAdminJob(id: string): Promise<AdminJob> {
  return request<AdminJob>(`/admin/jobs/${encodeURIComponent(id)}`);
}
export async function getSiteSettings(): Promise<SiteSettings> {
  return request<SiteSettings>("/settings");
}
export async function getSiteNavigation(): Promise<OperationsNavigation> {
  return request<OperationsNavigation>("/navigation");
}
export async function getAdminNavigation(): Promise<OperationsNavigation> {
  return request<OperationsNavigation>("/admin/navigation");
}
export async function updateAdminNavigation(payload: OperationsNavigation): Promise<OperationsNavigation> {
  return request<OperationsNavigation>("/admin/navigation", {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}
export async function getAdminRedirects(): Promise<RedirectListResponse> {
  return request<RedirectListResponse>("/admin/redirects");
}
export async function createAdminRedirect(payload: RedirectRule): Promise<RedirectListResponse & { item: RedirectRule }> {
  return request<RedirectListResponse & { item: RedirectRule }>("/admin/redirects", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}
export async function replaceAdminRedirects(items: RedirectRule[]): Promise<RedirectListResponse> {
  return request<RedirectListResponse>("/admin/redirects", {
    method: "PUT",
    body: JSON.stringify({ items })
  });
}
export async function getAdminMedia(params: MediaListParams = {}): Promise<MediaListResponse> {
  const query = toQuery(params);
  return request<MediaListResponse>(`/admin/media${query}`);
}
export async function getAdminMediaAsset(id: string): Promise<MediaAsset> {
  return request<MediaAsset>(`/admin/media/${encodeURIComponent(id)}`);
}
export async function getAdminMediaReferences(id: string, params: { page?: number; pageSize?: number } = {}): Promise<MediaReferenceListResponse> {
  const query = toQuery(params);
  return request<MediaReferenceListResponse>(`/admin/media/${encodeURIComponent(id)}/references${query}`);
}
export async function uploadAdminMedia(file: File, payload: { alt?: string; category?: string } = {}): Promise<MediaAsset> {
  return uploadMediaTo("/admin/media", file, payload, "POST");
}
export async function uploadMedia(file: File, payload: { alt?: string; category?: string } = {}): Promise<MediaAsset> {
  return uploadMediaTo("/media/uploads", file, payload, "POST");
}
export async function replaceAdminMediaFile(id: string, file: File): Promise<MediaAsset> {
  return uploadMediaTo(`/admin/media/${encodeURIComponent(id)}/file`, file, {}, "PUT");
}
async function uploadMediaTo(path: string, file: File, payload: { alt?: string; category?: string } = {}, method: "POST" | "PUT"): Promise<MediaAsset> {
  await ensureCsrfCookie();
  const form = new FormData();
  form.set("file", file);

  if (payload.alt) {
    form.set("alt", payload.alt);
  }
  if (payload.category) {
    form.set("category", payload.category);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    credentials: "include",
    headers: csrfHeaders(),
    body: form
  });

  if (!response.ok) {
    throw await apiErrorFromResponse(response);
  }

  return response.json() as Promise<MediaAsset>;
}
export async function updateAdminMedia(id: string, payload: { alt: string; category: string }): Promise<MediaAsset> {
  return request<MediaAsset>(`/admin/media/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(payload)
  });
}
export async function deleteAdminMedia(id: string): Promise<{ ok: boolean; asset: MediaAsset }> {
  return request<{ ok: boolean; asset: MediaAsset }>(`/admin/media/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}
export async function getAdminStats(range = "30d"): Promise<AdminStats> {
  const query = toQuery({ range });
  return request<AdminStats>(`/admin/stats${query}`);
}
export async function exportAdminStats(range = "30d"): Promise<AdminStatsExport> {
  const query = toQuery({ range });
  return request<AdminStatsExport>(`/admin/stats/export${query}`);
}
export async function getAdminAuditLogs(params: AuditLogParams = {}): Promise<AuditLogListResponse> {
  const query = toQuery(params);
  return request<AuditLogListResponse>(`/admin/audit-logs${query}`);
}
