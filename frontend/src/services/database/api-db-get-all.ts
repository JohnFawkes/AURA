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
	sortOrder: "asc" | "desc",
	filteredTypes: string[],
	filterMultiSetOnly: boolean
): Promise<APIResponse<DBGetAllItemsWithFiltersResponse>> => {
	try {
		log("INFO", "API - DB", "Fetch", "Fetching saved sets with filters", {
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
			filteredTypes,
			filterMultiSetOnly,
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
				filteredTypes,
				filterMultiSetOnly,
			},
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error fetching saved sets with filters");
		} else {
			log("INFO", "API - DB", "Fetch", "Fetched saved sets with filters", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Fetch",
			`Failed to get all saved sets with filters: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<DBGetAllItemsWithFiltersResponse>(error);
	}
};
