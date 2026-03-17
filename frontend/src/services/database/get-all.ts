import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { DBSavedItem } from "@/types/database/db-poster-set";
import { TYPE_FILTER_AUTO_DOWNLOAD_OPTIONS } from "@/types/ui-options";

export interface GetAllDBItems_Response {
  items: DBSavedItem[];
  total_items: number;
  unique_users: string[];
}

export const getAllItemsFromDB = async (
  searchTMDBID: string,
  searchLibrary: string,
  searchYear: number,
  searchTitle: string,
  libraryTitles: string[],
  filteredTypes: string[],
  filterAutodownload: TYPE_FILTER_AUTO_DOWNLOAD_OPTIONS,
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
    const params = {
      item_tmdb_id: searchTMDBID,
      item_library_title: searchLibrary,
      item_year: searchYear,
      item_title: searchTitle,
      library_titles: libraryTitles.join(","),
      image_types: filteredTypes.join(","),
      autodownload: filterAutodownload,
      multiset_only: multisetOnly,
      usernames: filteredUsernames.join(","),
      items_per_page: itemsPerPage,
      page_number: pageNumber,
      sort_option: sortOption,
      sort_order: sortOrder,
    };
    const response = await apiClient.get<APIResponse<GetAllDBItems_Response>>(`/db`, { params: params });
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
    return ReturnErrorMessage<GetAllDBItems_Response>(error);
  }
};
