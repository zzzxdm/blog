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

export interface ReactionSummary {
  postSlug: string;
  likeCount: number;
  dislikeCount: number;
  bookmarkCount: number;
  myReaction: "" | "like" | "dislike";
  bookmarked: boolean;
}

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

export type BookmarkItem = Post & {
  bookmarkedAt: string;
};

export interface BookmarkListResponse {
  items: BookmarkItem[];
  page: number;
  pageSize: number;
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

export type AdminPostStatus = "draft" | "review" | "scheduled" | "published" | "archived";
export type AdminPostVisibility = "public" | "private" | "members";

export interface AdminPost {
  id: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
  visibility: AdminPostVisibility;
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
  visibility: AdminPostVisibility;
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
  scheduled: number;
  monthlyViews: string;
  total: number;
}

export interface AdminPostListResponse {
  items: AdminPost[];
  page: number;
  pageSize: number;
  total: number;
  stats: AdminPostStats;
}

export interface AdminPostRevisionListResponse {
  items: AdminPostRevision[];
  total: number;
}

export interface AdminPostPreview {
  previewUrl: string;
  token: string;
  expiresAt: string;
}

export interface AdminPostPayload {
  slug: string;
  title: string;
  summary: string;
  content: string;
  status: AdminPostStatus;
  visibility: AdminPostVisibility;
  scheduledAt?: string;
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
  author?: string;
  sort?: "views" | "comments" | "likes";
}

export interface TaxonomyListParams {
  page?: number;
  pageSize?: number;
  q?: string;
}

export interface TopicListParams {
  page?: number;
  pageSize?: number;
  q?: string;
  status?: string;
  featured?: boolean;
  all?: boolean;
}

export interface ManageListParams {
  status?: string;
  q?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}

export interface MessageListParams {
  status?: string;
  type?: string;
  q?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}

export interface BookmarkListParams {
  q?: string;
  category?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
}

export interface MediaListParams {
  q?: string;
  type?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
}

export interface AdminPostParams {
  q?: string;
  status?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
  all?: boolean;
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

export async function getCategories(params: TaxonomyListParams = {}): Promise<TaxonomyListResponse<Category>> {
  const query = toQuery(params);
  return request<TaxonomyListResponse<Category>>(`/categories${query}`);
}

export async function getTags(params: TaxonomyListParams = {}): Promise<TaxonomyListResponse<Tag>> {
  const query = toQuery(params);
  return request<TaxonomyListResponse<Tag>>(`/tags${query}`);
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

export async function getMyBookmarks(params: BookmarkListParams = {}): Promise<BookmarkListResponse> {
  return request<BookmarkListResponse>(`/bookmarks/mine${toQuery(params)}`);
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

export async function getAdminPosts(params: AdminPostParams = {}): Promise<AdminPostListResponse> {
  const query = toQuery(params);
  return request<AdminPostListResponse>(`/admin/posts${query}`);
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

export async function deleteAdminPost(id: string): Promise<{ ok: boolean; post: AdminPost }> {
  return request<{ ok: boolean; post: AdminPost }>(`/admin/posts/${encodeURIComponent(id)}`, {
    method: "DELETE"
  });
}

export async function publishAdminPost(id: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/publish`, {
    method: "POST"
  });
}

export async function archiveAdminPost(id: string): Promise<AdminPost> {
  return request<AdminPost>(`/admin/posts/${encodeURIComponent(id)}/archive`, {
    method: "POST"
  });
}

export async function createAdminPostPreview(id: string): Promise<AdminPostPreview> {
  return request<AdminPostPreview>(`/admin/posts/${encodeURIComponent(id)}/preview`, {
    method: "POST"
  });
}

export async function getPreviewPost(token: string): Promise<AdminPost> {
  return request<AdminPost>(`/preview/${encodeURIComponent(token)}`);
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
    throw await apiErrorFromResponse(response);
  }

  return response.json() as Promise<T>;
}

async function apiErrorFromResponse(response: Response): Promise<ApiError> {
  const payload = await response.json().catch(() => null) as { error?: string } | null;
  return new ApiError(response.status, apiErrorMessage(response.status, payload?.error, response.headers.get("Retry-After")));
}

function apiErrorMessage(status: number, serverMessage?: string, retryAfter?: string | null) {
  if (status === 429) {
    return rateLimitMessage(retryAfter);
  }

  return serverMessage?.trim() || `Request failed: ${status}`;
}

function rateLimitMessage(retryAfter?: string | null) {
  const seconds = Number(retryAfter);
  if (Number.isFinite(seconds) && seconds > 0) {
    if (seconds < 60) {
      return `请求过于频繁，请 ${Math.ceil(seconds)} 秒后再试。`;
    }

    return `请求过于频繁，请 ${Math.ceil(seconds / 60)} 分钟后再试。`;
  }

  return "请求过于频繁，请稍后再试。";
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
