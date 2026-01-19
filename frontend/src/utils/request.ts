import { ApiResponse } from "@/types";

const BASE_URL = "/api";

interface RequestOptions extends RequestInit {
  params?: Record<string, any>;
}

export async function request<T = any>(
  url: string,
  options: RequestOptions = {},
): Promise<T> {
  const { params, ...fetchOptions } = options;

  let fullUrl = `${BASE_URL}${url}`;

  if (params) {
    const queryString = new URLSearchParams(params).toString();
    fullUrl += `?${queryString}`;
  }

  const response = await fetch(fullUrl, {
    ...fetchOptions,
    headers: {
      "Content-Type": "application/json",
      ...fetchOptions.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data: ApiResponse<T> = await response.json();

  if (data.code !== 0) {
    throw new Error(data.message || "Request failed");
  }

  return data.data;
}

export const api = {
  get: <T = any>(url: string, params?: Record<string, any>) =>
    request<T>(url, { method: "GET", params }),

  post: <T = any>(url: string, data?: any) =>
    request<T>(url, { method: "POST", body: JSON.stringify(data) }),

  put: <T = any>(url: string, data?: any) =>
    request<T>(url, { method: "PUT", body: JSON.stringify(data) }),

  delete: <T = any>(url: string) => request<T>(url, { method: "DELETE" }),
};
