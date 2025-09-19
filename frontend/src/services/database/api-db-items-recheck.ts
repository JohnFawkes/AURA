import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export interface AutodownloadResult {
	MediaItemTitle: string;
	Sets: AutodownloadSetResult[];
	OverAllResult: "Error" | "Warn" | "Success" | "Skipped";
	OverAllResultMessage: string;
}

export interface AutodownloadSetResult {
	PosterSetID: string;
	Result: "Success" | "Skipped" | "Error";
	Reason: string;
}

export const postForceRecheckDBItemForAutoDownload = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<AutodownloadResult>> => {
	log(`api.db - Forcing recheck for auto-download for item with ID ${saveItem.MediaItemID} started`);
	try {
		const response = await apiClient.post<APIResponse<AutodownloadResult>>(`/db/force/recheck`, {
			Item: saveItem,
		});
		log(`api.db - Forcing recheck for auto-download for item with ID ${saveItem.MediaItemID} succeeded`);
		return response.data;
	} catch (error) {
		log(
			`api.db - Forcing recheck for auto-download for item with ID ${
				saveItem.MediaItemID
			} failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<AutodownloadResult>(error);
	}
};
