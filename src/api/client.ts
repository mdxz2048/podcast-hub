export interface ApiError extends Error {
  code: string;
  status: number;
  validation_issues?: string[];
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const baseURL = import.meta.env.VITE_API_BASE_URL ?? "http://127.0.0.1:8080";
  const headers = new Headers(init.headers ?? {});
  if (!(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  let response: Response;
  try {
    response = await fetch(`${baseURL}${path}`, {
      credentials: "include",
      headers,
      ...init
    });
  } catch {
    const err = new Error("网络连接失败，请检查后重试。") as ApiError;
    err.code = "network_error";
    err.status = 0;
    throw err;
  }
  if (!response.ok) {
    let code = "unknown_error";
    let message = "请求失败，请稍后重试。";
    let validationIssues: string[] | undefined;
    try {
      const payload = await response.json() as { error?: { code?: string; message?: string; validation_issues?: string[] } };
      code = payload.error?.code ?? code;
      message = payload.error?.message ?? message;
      validationIssues = payload.error?.validation_issues;
    } catch {
      // keep generic
    }
    const err = new Error(message) as ApiError;
    err.code = code;
    err.status = response.status;
    err.validation_issues = validationIssues;
    throw err;
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}

export function readCookie(name: string): string | undefined {
  const matched = document.cookie.split("; ").find((entry) => entry.startsWith(`${name}=`));
  if (!matched) return undefined;
  return decodeURIComponent(matched.split("=").slice(1).join("="));
}
