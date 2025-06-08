import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { AppConfig } from "../types/config";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "@/lib/logger";

export const fetchConfig = async (): Promise<APIResponse<AppConfig>> => {
	log("api.settings - Fetching app configuration started");
	try {
		const response = await apiClient.get<APIResponse<AppConfig>>(`/config`);
		log("api.settings - Fetching app configuration succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Fetching app configuration failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<AppConfig>(error);
	}
};

export const fetchLogContents = async (): Promise<APIResponse<string>> => {
	log("api.settings - Fetching log contents started");
	try {
		const response = await apiClient.get<APIResponse<string>>(`/logs`);
		log("api.settings - Fetching log contents succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Fetching log contents failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const postClearTempImagesFolder = async (): Promise<
	APIResponse<void>
> => {
	log("api.settings - Clearing temp images folder started");
	try {
		const response = await apiClient.post<APIResponse<void>>(
			`/temp-images/clear`
		);
		log("api.settings - Clearing temp images folder succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Clearing temp images folder failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<void>(error);
	}
};

export const fetchMediaServerConnectionStatus = async (): Promise<
	APIResponse<string>
> => {
	log("api.settings - Fetching media server connection status started");
	try {
		const response = await apiClient.get<APIResponse<string>>(
			`/health/status/mediaserver`
		);
		log("api.settings - Fetching media server connection status succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Fetching media server connection status failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const postSendTestNotification = async (): Promise<
	APIResponse<string>
> => {
	log("api.settings - Sending test notification started");
	try {
		const response = await apiClient.post<APIResponse<string>>(
			`/health/status/notification`
		);
		log("api.settings - Sending test notification succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Sending test notification failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};
