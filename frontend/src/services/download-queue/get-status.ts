import apiClient from "@/services/api-client";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

// api.Util_Response_SendJSON(w, ld,
// 		map[string]any{
// 			"time":     api.DOWNLOAD_QUEUE_LATEST_INFO.Time,
// 			"status":   api.DOWNLOAD_QUEUE_LATEST_INFO.Status,
// 			"message":  api.DOWNLOAD_QUEUE_LATEST_INFO.Message,
// 			"warnings": api.DOWNLOAD_QUEUE_LATEST_INFO.Warnings,
// 			"errors":   api.DOWNLOAD_QUEUE_LATEST_INFO.Errors,
// 		})

export interface DownloadQueueStatus {
	time: string;
	status: string;
	message: string;
	warnings: string[];
	errors: string[];
}

export const fetchDownloadQueueStatus = async (): Promise<APIResponse<DownloadQueueStatus>> => {
	try {
		const response = await apiClient.get<APIResponse<DownloadQueueStatus>>(`/download-queue/status`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching download queue status");
		}
		return response.data;
	} catch (error) {
		log("ERROR", "API - Download Queue", "Fetch", "Error fetching download queue status", { error });
		throw error;
	}
};
