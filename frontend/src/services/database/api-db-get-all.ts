import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export interface DBGetAllItemsWithFiltersResponse {
	items: DBMediaItemWithPosterSets[];
	total_items: number;
	unique_users: string[];
}

export const fetchAllItemFromDBWithFilters = async (
	mediaItemID: string,
	cleanedQuery: string,
	mediaItemLibraryTitles: string[],
	mediaItemYear: number,
	autoDownloadOnly: boolean,
	userNames: string[],
	itemsPerPage: number,
	pageNumber: number,
	sortOption: string,
	sortOrder: "asc" | "desc"
): Promise<APIResponse<DBGetAllItemsWithFiltersResponse>> => {
	log("api.db - Fetching all items from the database with filters started");
	try {
		log("Using filters:", {
			mediaItemID,
			cleanedQuery,
			mediaItemLibraryTitles,
			mediaItemYear,
			autoDownloadOnly,
			userNames,
			itemsPerPage,
			pageNumber,
			sortOption,
			sortOrder,
		});
		const response = await apiClient.get<APIResponse<DBGetAllItemsWithFiltersResponse>>(`/db/get/all`, {
			params: {
				mediaItemID,
				cleanedQuery,
				mediaItemLibraryTitles,
				mediaItemYear,
				autoDownloadOnly,
				userNames,
				itemsPerPage,
				pageNumber,
				sortOption,
				sortOrder,
			},
		});
		log("api.db - Fetching all items from the database with filters succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.db - Fetching all items from the database with filters failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<DBGetAllItemsWithFiltersResponse>(error);
	}
};
