import apiClient from "@/services/api-client";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export interface DownloadQueueStatus {
	time: string;
	status: string;
	message: string;
	warnings: string[];
	errors: string[];
}

export const getDownloadQueueStatus = async (): Promise<APIResponse<DownloadQueueStatus>> => {
	try {
		const response = await apiClient.get<APIResponse<DownloadQueueStatus>>(`/download/queue`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching download queue status");
		}
		return response.data;
	} catch (error) {
		log("ERROR", "API - Download Queue", "Fetch", "Error fetching download queue status", { error });
		throw error;
	}
};
