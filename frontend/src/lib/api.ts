const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface RequestOptions {
  method?: string;
  body?: unknown;
}

const MAX_RETRIES = 2;
const RETRY_DELAY_MS = 500;

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = "GET", body } = opts;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  let lastError: ApiError | null = null;

  for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
    if (attempt > 0) {
      await new Promise((r) => setTimeout(r, RETRY_DELAY_MS * attempt));
    }

    const res = await fetch(`${API_URL}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
      credentials: "include",
    });

    if (res.ok) {
      return res.json();
    }

    const error = await res.json().catch(() => ({ message: res.statusText }));
    lastError = new ApiError(res.status, error.message || res.statusText);

    // Only retry on 5xx server errors
    if (res.status < 500) {
      throw lastError;
    }
  }

  throw lastError!;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export const api = {
  get: <T>(path: string) =>
    request<T>(path),

  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "POST", body }),

  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body }),

  del: <T>(path: string) =>
    request<T>(path, { method: "DELETE" }),
};
