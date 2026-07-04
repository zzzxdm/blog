const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";

export interface HealthResponse {
  status: string;
  env: string;
  time: string;
}

export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch(`${API_BASE_URL}/health`, {
    credentials: "include"
  });

  if (!response.ok) {
    throw new Error(`Health check failed: ${response.status}`);
  }

  return response.json() as Promise<HealthResponse>;
}

export interface Post {
  id: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  category: string;
  tags: string[];
  coverImage: string;
  authorName: string;
  readingTime: number;
  viewCount: number;
  likeCount: number;
  dislikeCount: number;
  commentCount: number;
  publishedAt: string;
}

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
}

export interface TokenResponse {
  ok: boolean;
  verificationToken?: string;
  resetToken?: string;
  delivery?: string;
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
  total: number;
  stats: CommentStats;
}

export interface ReactionSummary {
  postSlug: string;
  likeCount: number;
  dislikeCount: number;
  bookmarkCount: number;
  myReaction: "" | "like" | "dislike";
  bookmarked: boolean;
}

export type SubmissionStatus = "draft" | "submitted" | "returned" | "rejected" | "published";

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
  total: number;
}

export interface SubmissionListResponse {
  items: Submission[];
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
  submit?: boolean;
}

export interface ReviewPayload {
  action: "approve" | "return" | "reject";
  note: string;
  slug?: string;
  category?: string;
}

export type MessageType = "review" | "comment" | "system" | "admin" | "account";
export type MessageStatus = "unread" | "read" | "archived";

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
  createdAt: string;
}

export interface MessageStats {
  unread: number;
  review: number;
  admin: number;
  archived: number;
  total: number;
}

export interface MessageListResponse {
  items: StationMessage[];
  total: number;
  stats: MessageStats;
}

export type BookmarkItem = Post & {
  bookmarkedAt: string;
};

export interface BookmarkListResponse {
  items: BookmarkItem[];
  total: number;
}

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
  adminTwoFactorRequired: boolean;
  loginFailureLock: boolean;
  sessionDays: number;
  backupCycle: string;
  lastBackupAt: string;
  backupRetentionDays: number;
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
  rssRate: string;
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

export interface AdminCommentsExport {
  scope: "comments";
  exportedAt: string;
  items: Comment[];
  total: number;
  stats: CommentStats;
}

export interface AdminMessagesExport {
  scope: "messages";
  exportedAt: string;
  items: StationMessage[];
  total: number;
  stats: MessageStats;
}

export interface AdminUsersExport {
  scope: "users";
  exportedAt: string;
  items: ManagedUser[];
  total: number;
  stats: UserStats;
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
  total: number;
  stats: UserStats;
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
  resetToken?: string;
  delivery: string;
}

export interface AccountSettings {
  displayName: string;
  username: string;
  email: string;
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

export type AdminPostStatus = "draft" | "review" | "scheduled" | "published" | "archived";

export interface AdminPost {
  id: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
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
  monthlyViews: string;
  total: number;
}

export interface AdminPostListResponse {
  items: AdminPost[];
  total: number;
  stats: AdminPostStats;
}

export interface AdminPostRevisionListResponse {
  items: AdminPostRevision[];
  total: number;
}

export interface AdminPostPayload {
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
  category: string;
  tags: string[];
  coverImage: string;
  seoTitle: string;
  seoDescription: string;
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
  sort?: "views" | "comments" | "likes";
}

export async function getPosts(params: PostListParams = {}): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/posts${query}`);
}

export async function getPostBySlug(slug: string): Promise<Post> {
  return request<Post>(`/posts/${encodeURIComponent(slug)}`);
}

export async function searchPosts(params: PostListParams): Promise<ListResponse<Post>> {
  const query = toQuery(params);
  return request<ListResponse<Post>>(`/search${query}`);
}

export async function getCategories(): Promise<TaxonomyListResponse<Category>> {
  return request<TaxonomyListResponse<Category>>("/categories");
}

export async function getTags(): Promise<TaxonomyListResponse<Tag>> {
  return request<TaxonomyListResponse<Tag>>("/tags");
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

export async function login(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function register(email: string, password: string, displayName: string): Promise<AuthResponse> {
  return request<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password, displayName })
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

export async function getMyComments(status = ""): Promise<CommentManageListResponse> {
  const query = toQuery({ status });
  return request<CommentManageListResponse>(`/comments/mine${query}`);
}

export async function getAdminComments(status = ""): Promise<CommentManageListResponse> {
  const query = toQuery({ status });
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

export async function getReaction(postSlug: string): Promise<ReactionSummary> {
  return request<ReactionSummary>(`/posts/${encodeURIComponent(postSlug)}/reaction`);
}

export async function setPostReaction(postSlug: string, type: "like" | "dislike" | ""): Promise<ReactionSummary> {
  return request<ReactionSummary>(`/posts/${encodeURIComponent(postSlug)}/reaction`, {
    method: "PUT",
    body: JSON.stringify({ type })
  });
}

export async function setBookmark(postSlug: string, bookmarked: boolean): Promise<ReactionSummary> {
  return request<ReactionSummary>(`/posts/${encodeURIComponent(postSlug)}/bookmark`, {
    method: "PUT",
    body: JSON.stringify({ bookmarked })
  });
}

export async function getMyBookmarks(): Promise<BookmarkListResponse> {
  return request<BookmarkListResponse>("/bookmarks/mine");
}

export async function getMySubmissions(status = ""): Promise<SubmissionListResponse> {
  const query = toQuery({ status });
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

export async function submitExistingSubmission(id: string): Promise<Submission> {
  return request<Submission>(`/submissions/${encodeURIComponent(id)}/submit`, {
    method: "POST"
  });
}

export async function getAdminSubmissions(status = ""): Promise<SubmissionListResponse> {
  const query = toQuery({ status });
  return request<SubmissionListResponse>(`/admin/submissions${query}`);
}

export async function reviewSubmission(id: string, payload: ReviewPayload): Promise<Submission> {
  return request<Submission>(`/admin/submissions/${encodeURIComponent(id)}/review`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function getMessages(params: { status?: string; type?: string } = {}): Promise<MessageListResponse> {
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

export async function getAdminMessages(params: { status?: string; type?: string } = {}): Promise<MessageListResponse> {
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
}): Promise<StationMessage> {
  return request<StationMessage>("/admin/messages", {
    method: "POST",
    body: JSON.stringify(payload)
  });
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

export async function getAdminNavigation(): Promise<OperationsNavigation> {
  return request<OperationsNavigation>("/admin/navigation");
}

export async function updateAdminNavigation(payload: OperationsNavigation): Promise<OperationsNavigation> {
  return request<OperationsNavigation>("/admin/navigation", {
    method: "PUT",
    body: JSON.stringify(payload)
  });
}

export async function getAdminMedia(): Promise<MediaListResponse> {
  return request<MediaListResponse>("/admin/media");
}

export async function getAdminMediaAsset(id: string): Promise<MediaAsset> {
  return request<MediaAsset>(`/admin/media/${encodeURIComponent(id)}`);
}

export async function uploadAdminMedia(file: File, payload: { alt?: string; category?: string } = {}): Promise<MediaAsset> {
  const form = new FormData();
  form.set("file", file);

  if (payload.alt) {
    form.set("alt", payload.alt);
  }
  if (payload.category) {
    form.set("category", payload.category);
  }

  const response = await fetch(`${API_BASE_URL}/admin/media`, {
    method: "POST",
    credentials: "include",
    body: form
  });

  if (!response.ok) {
    const error = await response.json().catch(() => null) as { error?: string } | null;
    throw new ApiError(response.status, error?.error || `Request failed: ${response.status}`);
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

export async function getAdminStats(): Promise<AdminStats> {
  return request<AdminStats>("/admin/stats");
}

export async function exportAdminStats(): Promise<AdminStatsExport> {
  return request<AdminStatsExport>("/admin/stats/export");
}

export async function getAdminAuditLogs(params: AuditLogParams = {}): Promise<AuditLogListResponse> {
  const query = toQuery(params);
  return request<AuditLogListResponse>(`/admin/audit-logs${query}`);
}

export async function getAdminUsers(): Promise<UserListResponse> {
  return request<UserListResponse>("/admin/users");
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

export async function updateAdminUserStatus(id: string, status: ManagedUser["status"]): Promise<ManagedUser> {
  return request<ManagedUser>(`/admin/users/${encodeURIComponent(id)}/status`, {
    method: "PUT",
    body: JSON.stringify({ status })
  });
}

export async function requestAdminUserPasswordReset(id: string): Promise<AdminPasswordResetResponse> {
  return request<AdminPasswordResetResponse>(`/admin/users/${encodeURIComponent(id)}/password-reset`, {
    method: "POST"
  });
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

export async function getAdminPosts(): Promise<AdminPostListResponse> {
  return request<AdminPostListResponse>("/admin/posts");
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

export async function publishAdminPost(id: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/publish`, {
    method: "POST"
  });
}

export async function restoreAdminPostRevision(id: string, revisionId: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/revisions/${encodeURIComponent(revisionId)}/restore`, {
    method: "POST"
  });
}

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init.headers
    },
    ...init
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { error?: string } | null;
    throw new ApiError(response.status, payload?.error || `Request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

function toQuery(params: object): string {
  const search = new URLSearchParams();

  Object.entries(params as Record<string, unknown>).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }

    search.set(key, String(value));
  });

  const query = search.toString();
  return query ? `?${query}` : "";
}
