import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { DBSavedItem } from "@/types/database/db-poster-set";
import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import type { MediuxUserInfo } from "@/types/mediux/mediux-user-follow-hide";

interface HandleSearch_Response {
  search_query: string;
  media_items: MediaItem[];
  media_items_last_full_update: number;
  mediux_usernames: MediuxUserInfo[];
  mediux_usernames_last_full_update: number;
  saved_sets: DBSavedItem[];
}

interface HandleSearch_QueryProps {
  searchQuery: string;
  searchMediaItems: boolean;
  searchMediuxUsers: boolean;
  searchSavedSets: boolean;
}

export const HandleSearch = async (props: HandleSearch_QueryProps): Promise<APIResponse<HandleSearch_Response>> => {
  log("INFO", "API - Search", "Fetch Search Results", `Fetching search results for query: ${props.searchQuery}`);
  try {
    const params = {
      query: props.searchQuery,
      search_media_items: props.searchMediaItems,
      search_mediux_users: props.searchMediuxUsers,
      search_saved_sets: props.searchSavedSets,
    };
    const response = await apiClient.get<APIResponse<HandleSearch_Response>>(`/search`, { params });
    if (response.data.status === "error") {
      throw new Error(
        response.data.error?.message || `Unknown error fetching search results for query: ${props.searchQuery}`
      );
    } else {
      log(
        "INFO",
        "API - Search",
        "Fetch Search Results",
        `Fetched search results for query: ${props.searchQuery} successfully`,
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Search",
      "Fetch Search Results",
      `Failed to fetch search results for query: ${props.searchQuery}: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<HandleSearch_Response>(error);
  }
};
