import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { AppConfig } from "../types/config";

export const fetchConfig = async (): Promise<APIResponse<AppConfig>> => {
	try {
		const response = await apiClient.get<APIResponse<AppConfig>>(`/config`);
		return response.data;
	} catch {
		return {
			status: "error",
			message: "Failed to fetch data from Config API",
		};
	}
};

export const fetchLogContents = async (): Promise<APIResponse<string>> => {
	try {
		const response = await apiClient.get<APIResponse<string>>(`/logs`);
		return response.data;
	} catch {
		return {
			status: "error",
			message: "Failed to fetch data from Logs API",
		};
	}
};
