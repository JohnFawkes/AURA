import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse, LogErrorInfo } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { MediuxUserInfo } from "@/types/mediux/mediux-user-follow-hide";

interface SearchQueryResultsResponse {
	search_query: string;
	media_items: MediaItem[];
	mediux_usernames: MediuxUserInfo[];
	saved_sets: DBMediaItemWithPosterSets[];
	error: LogErrorInfo | null;
}

interface SearchResultQueryProps {
	searchQuery: string;
	searchMediaItems: boolean;
	searchMediuxUsers: boolean;
	searchSavedSets: boolean;
}

export const fetchSearchResults = async (
	props: SearchResultQueryProps
): Promise<APIResponse<SearchQueryResultsResponse>> => {
	log("INFO", "API - Search", "Fetch Search Results", `Fetching search results for query: ${props.searchQuery}`);
	try {
		const response = await apiClient.get<APIResponse<SearchQueryResultsResponse>>(`/search`, {
			params: {
				query: props.searchQuery,
				search_media_items: props.searchMediaItems,
				search_mediux_users: props.searchMediuxUsers,
				search_saved_sets: props.searchSavedSets,
			},
		});
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
		return ReturnErrorMessage<SearchQueryResultsResponse>(error);
	}
};
