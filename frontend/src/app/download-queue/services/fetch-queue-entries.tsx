import apiClient from "@/services/api-client";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const fetchDownloadQueueEntries = async (): Promise<
	APIResponse<{
		in_progress_entries: DBMediaItemWithPosterSets[];
		error_entries: DBMediaItemWithPosterSets[];
		warning_entries: DBMediaItemWithPosterSets[];
	}>
> => {
	try {
		log("INFO", "API - Download Queue", "Fetch", "Fetching download queue entries");
		const response = await apiClient.get<
			APIResponse<{
				in_progress_entries: DBMediaItemWithPosterSets[];
				error_entries: DBMediaItemWithPosterSets[];
				warning_entries: DBMediaItemWithPosterSets[];
			}>
		>(`/download-queue/get-results`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching download queue entries");
		} else {
			log("INFO", "API - Download Queue", "Fetch", "Fetched download queue entries successfully", {
				in_progress_entries: response.data.data?.in_progress_entries,
				error_entries: response.data.data?.error_entries,
				warning_entries: response.data.data?.warning_entries,
			});
		}
		return response.data;
	} catch (error) {
		log("ERROR", "API - Download Queue", "Fetch", "Error fetching download queue entries", { error });
		throw error;
	}
};
