import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse, LogData } from "@/types/api/api-response";

export interface FetchLogContentsResponse {
	total_log_entries: number;
	possible_actions_paths: Record<string, { label: string; section: string }>;
	log_entries: LogData[];
}

export const fetchLogContents = async (
	filteredLogLevels: string[],
	filteredStatuses: string[],
	filteredActions: string[],
	itemsPerPage: number,
	pageNumber: number
): Promise<APIResponse<FetchLogContentsResponse>> => {
	log("INFO", "API - Logs", "Fetch Log Contents", "Fetching log contents");
	try {
		const response = await apiClient.get<APIResponse<FetchLogContentsResponse>>(`/log`, {
			params: {
				filteredLogLevels: filteredLogLevels.join(","),
				filteredStatuses: filteredStatuses.join(","),
				filteredActions: filteredActions.join(","),
				itemsPerPage,
				pageNumber,
			},
		});
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
		return ReturnErrorMessage<FetchLogContentsResponse>(error);
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
