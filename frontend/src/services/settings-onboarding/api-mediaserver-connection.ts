import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediaServer } from "@/types/config/config-app";

export async function postMediaServerNewInfoConnectionStatus(
	mediaServerInfo: AppConfigMediaServer
): Promise<APIResponse<string>> {
	log("INFO", "API - Settings", "Media Server", "Posting media server new info to check connection status");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/config/validate/mediaserver`, mediaServerInfo);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error posting media server new info");
		} else {
			log("INFO", "API - Settings", "Media Server", "Posted media server new info successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Media Server",
			`Failed to post media server new info: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const checkMediaServerNewInfoConnectionStatus = async (
	mediaServerInfo: AppConfigMediaServer,
	showToast = true
): Promise<{ ok: boolean; message: string; data: AppConfigMediaServer | null }> => {
	try {
		let loadingToast: string | number | undefined;
		if (showToast) {
			loadingToast = toast.loading(`Checking connection to ${mediaServerInfo.Type}...`);
		}
		const response = await postMediaServerNewInfoConnectionStatus(mediaServerInfo);
		if (response.status === "error") {
			if (showToast && loadingToast) toast.dismiss(loadingToast);
			if (showToast) {
				toast.error(response.error?.message || "Couldn't connect to media server. Check the URL and Token", {
					duration: 1000,
				});
			}
			return {
				ok: false,
				message: response.error?.message || "Couldn't connect to media server. Check the URL and Token",
				data: null,
			};
		}
		if (showToast && loadingToast) toast.dismiss(loadingToast);
		if (showToast) {
			toast.success(`Successfully connected to ${mediaServerInfo.Type}`, { duration: 1000 });
		}
		return {
			ok: true,
			message: `Successfully connected to ${mediaServerInfo.Type}`,
			data: response.data as unknown as AppConfigMediaServer | null,
		};
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		if (showToast) {
			toast.dismiss();
			toast.error(errorResponse.error?.message || "Couldn't connect to media server. Check the URL and Token", {
				duration: 1000,
			});
		}
		return {
			ok: false,
			message: errorResponse.error?.message || "Couldn't connect to media server. Check the URL and Token",
			data: null,
		};
	}
};
