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
