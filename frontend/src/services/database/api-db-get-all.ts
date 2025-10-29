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
	searchTMDBID: string,
	searchLibrary: string,
	searchYear: number,
	searchTitle: string,
	librarySections: string[],
	filteredTypes: string[],
	filterAutodownload: "all" | "on" | "off",
	multisetOnly: boolean,
	filteredUsernames: string[],
	itemsPerPage: number,
	pageNumber: number,
	sortOption: string,
	sortOrder: "asc" | "desc"
): Promise<APIResponse<DBGetAllItemsWithFiltersResponse>> => {
	try {
		log("INFO", "API - DB", "Fetch", "Fetching saved sets with filters", {
			searchTMDBID,
			searchLibrary,
			searchYear,
			searchTitle,
			librarySections,
			filteredTypes,
			filterAutodownload,
			multisetOnly,
			filteredUsernames,
			itemsPerPage,
			pageNumber,
			sortOption,
			sortOrder,
		});
		const params = {
			searchTMDBID,
			searchLibrary,
			searchYear,
			searchTitle,
			librarySections: librarySections.join(","),
			filteredTypes: filteredTypes.join(","),
			filterAutodownload,
			multisetOnly,
			filteredUsernames: filteredUsernames.join(","),
			itemsPerPage,
			pageNumber,
			sortOption,
			sortOrder,
		};
		const response = await apiClient.get<APIResponse<DBGetAllItemsWithFiltersResponse>>(`/db/get-all`, {
			params,
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching saved sets with filters");
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
