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

export interface User {
  id: string;
  email: string;
  displayName: string;
  role: string;
  status: string;
  avatarText: string;
}

export interface AuthResponse {
  user: User;
}

export interface Comment {
  id: string;
  postSlug: string;
  parentId?: string;
  authorId: string;
  authorName: string;
  avatarText: string;
  body: string;
  status: "approved" | "pending" | "rejected" | "spam" | "deleted";
  likeCount: number;
  isMine: boolean;
  isAuthor: boolean;
  createdAt: string;
}

export interface CommentListResponse {
  items: Comment[];
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

export async function getMe(): Promise<AuthResponse> {
  return request<AuthResponse>("/me");
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

function toQuery(params: PostListParams): string {
  const search = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }

    search.set(key, String(value));
  });

  const query = search.toString();
  return query ? `?${query}` : "";
}
