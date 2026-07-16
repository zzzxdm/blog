/** Shared HTTP client, CSRF helpers, and API error types. */

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";
const CSRF_COOKIE_NAME = "blog_csrf";
const CSRF_HEADER_NAME = "X-CSRF-Token";
function readCookie(name: string): string {
  if (typeof document === "undefined") {
    return "";
  }

  const prefix = `${encodeURIComponent(name)}=`;
  const parts = document.cookie.split("; ");
  for (const part of parts) {
    if (part.startsWith(prefix)) {
      return decodeURIComponent(part.slice(prefix.length));
    }
    // Also handle unencoded cookie names.
    if (part.startsWith(`${name}=`)) {
      return decodeURIComponent(part.slice(name.length + 1));
    }
  }
  return "";
}
export function csrfHeaders(): Record<string, string> {
  const token = readCookie(CSRF_COOKIE_NAME);
  return token ? { [CSRF_HEADER_NAME]: token } : {};
}
function isWriteMethod(method?: string): boolean {
  const value = (method || "GET").toUpperCase();
  return value === "POST" || value === "PUT" || value === "PATCH" || value === "DELETE";
}
let csrfBootstrap: Promise<void> | null = null;
export async function ensureCsrfCookie(): Promise<void> {
  if (readCookie(CSRF_COOKIE_NAME)) {
    return;
  }

  if (!csrfBootstrap) {
    csrfBootstrap = fetch(`${API_BASE_URL}/health`, {
      credentials: "include"
    })
      .then(() => undefined)
      .catch(() => undefined)
      .finally(() => {
        csrfBootstrap = null;
      });
  }

  await csrfBootstrap;
}
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
export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}
export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let response: Response;

  try {
    if (isWriteMethod(init.method)) {
      await ensureCsrfCookie();
    }
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(isWriteMethod(init.method) ? csrfHeaders() : {})
    };
    const extraHeaders = init.headers;
    if (extraHeaders instanceof Headers) {
      extraHeaders.forEach((value, key) => {
        headers[key] = value;
      });
    } else if (Array.isArray(extraHeaders)) {
      for (const [key, value] of extraHeaders) {
        headers[key] = value;
      }
    } else if (extraHeaders) {
      Object.assign(headers, extraHeaders as Record<string, string>);
    }

    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      credentials: "include",
      headers
    });
  } catch (err) {
    throw apiErrorFromFetch(err);
  }

  if (!response.ok) {
    throw await apiErrorFromResponse(response);
  }

  return response.json() as Promise<T>;
}
export async function apiErrorFromResponse(response: Response): Promise<ApiError> {
  const serverMessage = await responseErrorMessage(response);
  return new ApiError(response.status, apiErrorMessage(response.status, serverMessage, response.headers.get("Retry-After")));
}
function apiErrorMessage(status: number, serverMessage?: string, retryAfter?: string | null) {
  const detail = cleanErrorMessage(serverMessage);

  if (status === 429) {
    return detail || rateLimitMessage(retryAfter);
  }

  if (detail) return detail;
  if (status === 400) return "请求参数有误，请检查填写内容后重试。";
  if (status === 401) return "登录状态已失效，请重新登录。";
  if (status === 403) return "没有权限执行该操作。";
  if (status === 404) return "请求的资源不存在或已被删除。";
  if (status === 409) return "当前数据状态已变化，请刷新后重试。";
  if (status === 422) return "提交内容未通过校验，请检查后重试。";
  if (status >= 500) return `服务器处理失败，请稍后再试或联系管理员。`;

  return `请求失败（${status}），请稍后再试。`;
}
export function apiErrorFromFetch(err: unknown): ApiError {
  const message = err instanceof Error ? err.message : String(err || "");
  if (err instanceof DOMException && err.name === "AbortError") {
    return new ApiError(0, "请求已取消，请重试。");
  }

  if (isFetchFailureMessage(message)) {
    return new ApiError(0, "无法连接到服务，请检查网络或确认后端服务已启动。");
  }

  const detail = cleanErrorMessage(message);
  return new ApiError(0, detail ? `网络请求失败：${detail}` : "网络请求失败，请稍后重试。");
}
async function responseErrorMessage(response: Response): Promise<string> {
  const contentType = response.headers.get("Content-Type") || "";
  if (contentType.includes("application/json")) {
    const payload = await response.json().catch(() => null) as { error?: string; message?: string; detail?: string } | null;
    return payload?.error || payload?.message || payload?.detail || "";
  }

  return response.text().catch(() => "");
}
function cleanErrorMessage(message?: string) {
  const value = message?.trim();
  if (!value || isGenericErrorMessage(value)) {
    return "";
  }

  return value;
}
function isGenericErrorMessage(message: string) {
  return /^request failed:?\s*\d+$/i.test(message)
    || /^failed to (fetch|load)( resource)?/i.test(message)
    || /^load failed$/i.test(message)
    || /^networkerror/i.test(message);
}
function isFetchFailureMessage(message: string) {
  return /^failed to (fetch|load)( resource)?/i.test(message)
    || /^load failed$/i.test(message)
    || /^networkerror/i.test(message);
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
export function toQuery(params: object): string {
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
