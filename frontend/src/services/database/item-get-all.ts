import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export const getAllItemsFromDB = async (
    searchTMDBID: string,
    searchLibrary: string,
    searchYear: number,
    searchTitle: string,
    libraryTitles: string[],
    filteredTypes: string[],
    filterAutodownload: "all" | "on" | "off",
    multisetOnly: boolean,
    filteredUsernames: string[],
    itemsPerPage: number,
    pageNumber: number,
    sortOption: string,
    sortOrder: "asc" | "desc"
): Promise<
    APIResponse<{
        items: DBSavedItem[];
        total_items: number;
        unique_users: string[];
    }>
> => {
    try {
        const response = await apiClient.get<
            APIResponse<{ items: DBSavedItem[]; total_items: number; unique_users: string[] }>
        >(`/db`, {
            params: {
                item_tmdb_id: searchTMDBID,
                item_library_title: searchLibrary,
                item_year: searchYear,
                item_title: searchTitle,
                library_titles: libraryTitles.join(","),
                filtered_types: filteredTypes.join(","),
                filter_autodownload: filterAutodownload,
                multiset_only: multisetOnly,
                filtered_usernames: filteredUsernames.join(","),
                items_per_page: itemsPerPage,
                page_number: pageNumber,
                sort_option: sortOption,
                sort_order: sortOrder,
            },
        });
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error fetching all DB items");
        } else {
            log("INFO", "API - DB", "Get All DB Items", `Fetched ${response.data?.data?.items.length} items from DB`);
        }
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - DB",
            "Get All DB Items",
            `Failed to fetch all DB items: ${error instanceof Error ? error.message : "Unknown error"}`,
            error
        );
        return ReturnErrorMessage<{ items: DBSavedItem[]; total_items: number; unique_users: string[] }>(error);
    }
};
