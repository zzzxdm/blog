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

async function request<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    credentials: "include"
  });

  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`);
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
