import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

/**
 * Fetches items from a specific media server library section.
 *
 * Logs the fetch operation, sends a GET request to the `/mediaserver/sections/items` endpoint,
 * and returns the API response containing the library section data.
 * Handles errors by logging and returning a standardized error message.
 *
 * @param librarySection - The library section from which to fetch items.
 * @param sectionStartIndex - The starting index for fetching items (used for pagination).
 * @returns A promise that resolves to an APIResponse containing the LibrarySection data.
 */
export const fetchMediaServerLibrarySectionItems = async (
	librarySection: LibrarySection,
	sectionStartIndex: number
): Promise<APIResponse<LibrarySection>> => {
	const logMessage =
		sectionStartIndex === 0
			? `Fetching items for '${librarySection.Title}'...`
			: `Fetching items for '${librarySection.Title}' (index: ${sectionStartIndex})`;
	log(`api.mediaserver - ${logMessage}`);
	try {
		const response = await apiClient.get<APIResponse<LibrarySection>>(`/mediaserver/sections/items`, {
			params: {
				sectionID: librarySection.ID,
				sectionTitle: librarySection.Title,
				sectionType: librarySection.Type,
				sectionStartIndex: sectionStartIndex,
			},
		});
		log(`api.mediaserver - Fetched items for '${librarySection.Title}' successfully.`);
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching items for '${librarySection.Title}' failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<LibrarySection>(error);
	}
};
