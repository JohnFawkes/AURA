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
	log(
		"INFO",
		"API - DB",
		"Recheck",
		`Forcing recheck for auto-download for item with ID ${saveItem.MediaItemID}`,
		saveItem
	);
	try {
		const response = await apiClient.post<APIResponse<AutodownloadResult>>(`/db/force/recheck`, {
			Item: saveItem,
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error forcing recheck for auto-download");
		} else {
			log(
				"INFO",
				"API - DB",
				"Recheck",
				`Forcing recheck for auto-download for item with ID ${saveItem.MediaItemID} succeeded`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Recheck",
			`Failed to force recheck for auto-download for item with ID ${saveItem.MediaItemID}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<AutodownloadResult>(error);
	}
};
