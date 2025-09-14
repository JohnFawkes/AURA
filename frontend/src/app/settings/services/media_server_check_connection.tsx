import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";
import { AppConfigMediaServer } from "@/types/config";

export const checkMediaServerConnectionStatus = async (mediaServerInfo: AppConfigMediaServer) => {
	try {
		const response = await fetchMediaServerConnectionStatus();

		if (response.status === "error") {
			toast.error(response.error?.Message || "Failed to check media server status");
			return;
		}

		toast.success(`${mediaServerInfo.Type} running with version: ${response.data}`);
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		toast.error(errorResponse.error?.Message || "Failed to check media server status");
	}
};

export async function fetchMediaServerConnectionStatus(): Promise<APIResponse<string>> {
	log("api.settings - Fetching media server connection status started");
	try {
		const response = await apiClient.get<APIResponse<string>>(`/health/status/mediaserver`);
		log("api.settings - Fetching media server connection status succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Fetching media server connection status failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<string>(error);
	}
}

export async function postMediaServerNewInfoConnectionStatus(
	mediaServerInfo: AppConfigMediaServer
): Promise<APIResponse<string>> {
	log("api.settings - Posting media server new info connection status started");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/config/validate/mediaserver`, mediaServerInfo);
		log("api.settings - Posting media server new info connection status succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Posting media server new info connection status failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const checkMediaServerNewInfoConnectionStatus = async (
	mediaServerInfo: AppConfigMediaServer
): Promise<{ ok: boolean; message: string; data: AppConfigMediaServer | null }> => {
	try {
		const loadingToast = toast.loading(`Checking connection to ${mediaServerInfo.Type}...`);
		const response = await postMediaServerNewInfoConnectionStatus(mediaServerInfo);
		if (response.status === "error") {
			toast.dismiss(loadingToast);
			return {
				ok: false,
				message: response.error?.Message || "Couldn't connect to media server. Check the URL and Token",
				data: null,
			};
		}
		toast.dismiss(loadingToast);
		toast.success(`Successfully connected to ${mediaServerInfo.Type}`, { duration: 1000 });
		return {
			ok: true,
			message: `Successfully connected to ${mediaServerInfo.Type}`,
			data: response.data ? mediaServerInfo : null,
		};
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		toast.error(errorResponse.error?.Message || "Couldn't connect to media server. Check the URL and Token", {
			duration: 1000,
		});
		return {
			ok: false,
			message: errorResponse.error?.Message || "Couldn't connect to media server. Check the URL and Token",
			data: null,
		};
	}
};
