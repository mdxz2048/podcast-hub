export interface ApiError extends Error {
  code: string;
  status: number;
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
  const baseURL = import.meta.env.VITE_API_BASE_URL ?? "http://127.0.0.1:8080";
  let response: Response;
  try {
    response = await fetch(`${baseURL}${path}`, {
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(init.headers ?? {})
      },
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
    try {
      const payload = await response.json() as { error?: { code?: string; message?: string } };
      code = payload.error?.code ?? code;
      message = payload.error?.message ?? message;
    } catch {
      // keep generic
    }
    const err = new Error(message) as ApiError;
    err.code = code;
    err.status = response.status;
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
