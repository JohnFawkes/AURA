import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse, LogErrorInfo } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { MediuxUserInfo } from "@/types/mediux/mediux-user-follow-hide";

interface SearchQueryResultsResponse {
	search_query: string;
	media_items: MediaItem[];
	mediux_usernames: MediuxUserInfo[];
	error: LogErrorInfo | null;
}

export const fetchSearchResults = async (searchQuery: string): Promise<APIResponse<SearchQueryResultsResponse>> => {
	log("INFO", "API - Search", "Fetch Search Results", `Fetching search results for query: ${searchQuery}`);
	try {
		const response = await apiClient.get<APIResponse<SearchQueryResultsResponse>>(`/search`, {
			params: { query: searchQuery },
		});
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.message || `Unknown error fetching search results for query: ${searchQuery}`
			);
		} else {
			log(
				"INFO",
				"API - Search",
				"Fetch Search Results",
				`Fetched search results for query: ${searchQuery} successfully`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Search",
			"Fetch Search Results",
			`Failed to fetch search results for query: ${searchQuery}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<SearchQueryResultsResponse>(error);
	}
};
