import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export interface Login_Request {
  password: string;
}

export interface Login_Response {
  token: string;
}

export const AttemptLogin = async (password: string): Promise<APIResponse<Login_Response>> => {
  try {
    const req: Login_Request = { password };
    const resp = await apiClient.post<APIResponse<Login_Response>>(`/login`, req);
    const token =
      resp.data?.data?.token ??
      (typeof resp.data === "object" &&
      resp.data !== null &&
      "token" in resp.data &&
      typeof (resp.data as Record<string, unknown>).token === "string"
        ? (resp.data as Record<string, unknown>).token
        : undefined);
    if (token && typeof window !== "undefined") {
      localStorage.setItem("aura-auth-token", String(token));
    }
    if (resp.data.status === "error") {
      localStorage.removeItem("aura-auth-token");
      throw new Error(resp.data.error?.message || "Unknown error during login");
    } else {
      log(
        "INFO",
        "Auth",
        "Login",
        "Login successful",
        // Last 10 characters from resp.data
        resp.data?.data?.token?.slice(-10)
      );
    }
    return resp.data;
  } catch (error) {
    log(
      "ERROR",
      "Auth",
      "Login",
      `Failed to login: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    localStorage.removeItem("aura-auth-token");
    return ReturnErrorMessage<Login_Response>(error);
  }
};

export function getAuthToken(): string | null {
  return typeof window !== "undefined" ? localStorage.getItem("aura-auth-token") : null;
}

export async function authFetch(input: RequestInfo | URL, init: RequestInit = {}) {
  const token = getAuthToken();
  const headers = new Headers(init.headers || {});
  if (token) headers.set("Authorization", `Bearer ${token}`);
  return fetch(input, { ...init, headers });
}
