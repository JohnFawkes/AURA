import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediaServer } from "@/types/config/config-app";

export async function postMediaServerLibraryOptions(
	mediaServerInfo: AppConfigMediaServer
): Promise<APIResponse<string>> {
	log("INFO", "API - Settings", "Media Server", "Posting media server info to get library options");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/config/get/mediaserver/sections`, mediaServerInfo);
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.Message || "Unknown error posting media server info to get library options"
			);
		} else {
			log(
				"INFO",
				"API - Settings",
				"Media Server",
				"Posted media server info to get library options successfully",
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Media Server",
			`Failed to post media server info to get library options: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
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
