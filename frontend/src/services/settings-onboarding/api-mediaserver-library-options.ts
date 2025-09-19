import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediaServer } from "@/types/config/config-app";

export async function postMediaServerLibraryOptions(
	mediaServerInfo: AppConfigMediaServer
): Promise<APIResponse<string>> {
	log("api.settings - Posting media server info to get library options started");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/config/get/mediaserver/sections`, mediaServerInfo);
		log("api.settings - Posting media server info to get library options succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Posting media server info to get library options failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const fetchMediaServerLibraryOptions = async (
	mediaServerInfo: AppConfigMediaServer
): Promise<{ ok: boolean; data: string[] }> => {
	try {
		const loadingToast = toast.loading(`Fetching library options for ${mediaServerInfo.Type}...`);
		const response = await postMediaServerLibraryOptions(mediaServerInfo);
		if (response.status === "error") {
			toast.dismiss(loadingToast);
			return {
				ok: false,
				data: [],
			};
		}
		toast.dismiss(loadingToast);
		toast.success(`Successfully fetched library options for ${mediaServerInfo.Type}`, { duration: 1000 });
		const data: string[] = Array.isArray(response.data)
			? response.data
			: typeof response.data === "string"
				? [response.data]
				: [];
		return { ok: true, data };
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		toast.error(errorResponse.error?.Message || "Couldn't connect to media server. Check the URL and Token", {
			duration: 1000,
		});
		return {
			ok: false,
			data: [],
		};
	}
};
