import { ApiError, type ApiErrorBody } from "./types";

const BASE_URL = "/api";

async function handleResponse<T>(response: Response): Promise<T> {
  if (response.status === 204) {
    return undefined as T;
  }

  const body = await response.json().catch(() => null);

  if (!response.ok) {
    const errorBody: ApiErrorBody = body?.error ?? {
      code: "internal_error",
      message: "Unexpected server error.",
    };
    throw new ApiError(response.status, errorBody);
  }

  return body as T;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...init?.headers,
      },
    });
  } catch {
    throw new ApiError(0, {
      code: "network_error",
      message: "Could not reach the server. Check your connection and try again.",
    });
  }

  return handleResponse<T>(response);
}

async function requestForm<T>(path: string, formData: FormData): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${BASE_URL}${path}`, { method: "POST", body: formData });
  } catch {
    throw new ApiError(0, {
      code: "network_error",
      message: "Could not reach the server. Check your connection and try again.",
    });
  }

  return handleResponse<T>(response);
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),
  post: <T>(path: string, data: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(data) }),
  put: <T>(path: string, data: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(data) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
  postForm: <T>(path: string, formData: FormData) => requestForm<T>(path, formData),
};
