import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";

import { ReturnErrorMessage } from "./api.shared";
import apiClient from "./apiClient";

export const postLogin = async (password: string): Promise<APIResponse<{ token: string }>> => {
	log("api.auth - Login started");
	try {
		const response = await apiClient.post<APIResponse<{ token: string }>>(`/login`, { password });

		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const token = response.data?.data?.token || (response.data as any)?.token;
		if (token && typeof window !== "undefined") {
			localStorage.setItem("aura-auth-token", token);
		}

		log("api.auth - Login succeeded");
		return response.data;
	} catch (error) {
		log(`api.auth - Login failed: ${error instanceof Error ? error.message : "Unknown error"}`);
		return ReturnErrorMessage<{ token: string }>(error);
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
