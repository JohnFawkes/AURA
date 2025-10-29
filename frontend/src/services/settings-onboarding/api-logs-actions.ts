import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const fetchLogContents = async (): Promise<APIResponse<string>> => {
	log("INFO", "API - Logs", "Fetch Log Contents", "Fetching log contents");
	try {
		const response = await apiClient.get<APIResponse<string>>(`/log`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching log contents");
		} else {
			log("INFO", "API - Logs", "Fetch Log Contents", "Fetched log contents successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Logs",
			"Fetch Log Contents",
			`Failed to fetch log contents: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const postClearOldLogs = async (clearToday: boolean = false): Promise<APIResponse<void>> => {
	log("INFO", "API - Logs", "Clear Old Logs", `Clearing old logs, clearToday=${clearToday}`);
	try {
		const response = await apiClient.post<APIResponse<void>>(`/log/clear`, { clearToday });
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error clearing old logs");
		} else {
			log("INFO", "API - Logs", "Clear Old Logs", "Cleared old logs successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Logs",
			"Clear Old Logs",
			`Failed to clear old logs: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<void>(error);
	}
};
