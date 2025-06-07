import { APIResponse } from "@/types/apiResponse";
import { LibrarySection, MediaItem } from "@/types/mediaItem";
import apiClient from "./apiClient";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "@/lib/logger";
import { PosterFile } from "@/types/posterSets";

export const fetchMediaServerType = async (): Promise<
	APIResponse<{ serverType: string }>
> => {
	log("api.mediaserver - Fetching server type started");
	try {
		const response = await apiClient.get<
			APIResponse<{ serverType: string }>
		>(`/mediaserver/type`);
		log("api.mediaserver - Fetching server type succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching server type failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<{ serverType: string }>(error);
	}
};

export const fetchMediaServerLibrarySections = async (): Promise<
	APIResponse<LibrarySection[]>
> => {
	log("api.mediaserver - Fetching all library sections started");
	try {
		const response = await apiClient.get<APIResponse<LibrarySection[]>>(
			`/mediaserver/sections/`
		);
		log("api.mediaserver - Fetching all library sections succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching all library sections failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<LibrarySection[]>(error);
	}
};
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
		const response = await apiClient.get<APIResponse<LibrarySection>>(
			`/mediaserver/sections/items`,
			{
				params: {
					sectionID: librarySection.ID,
					sectionTitle: librarySection.Title,
					sectionType: librarySection.Type,
					sectionStartIndex: sectionStartIndex,
				},
			}
		);
		log(
			`api.mediaserver - Fetched items for '${librarySection.Title}' successfully.`
		);
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching items for '${
				librarySection.Title
			}' failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<LibrarySection>(error);
	}
};

export const fetchMediaServerItemContent = async (
	ratingKey: string,
	sectionTitle: string
): Promise<APIResponse<MediaItem>> => {
	log(
		`api.mediaserver - Fetching item content for ratingKey ${ratingKey} started`
	);
	try {
		// Encode sectionTitle to handle spaces and special characters
		const encodedSectionTitle = encodeURIComponent(sectionTitle);

		const response = await apiClient.get<APIResponse<MediaItem>>(
			`/mediaserver/item/${ratingKey}?sectionTitle=${encodedSectionTitle}`
		);
		log(
			`api.mediaserver - Fetching item content for ratingKey ${ratingKey} succeeded`
		);
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching item content for ratingKey ${ratingKey} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<MediaItem>(error);
	}
};

export const patchDownloadPosterFileAndUpdateMediaServer = async (
	posterFile: PosterFile,
	mediaItem: MediaItem
): Promise<APIResponse<string>> => {
	log(
		`api.mediaserver - Downloading poster file and updating media server for ratingKey ${mediaItem.RatingKey} started`,
		{
			posterFile: posterFile,
			mediaItem: mediaItem,
		}
	);
	try {
		const response = await apiClient.patch<APIResponse<string>>(
			`/mediaserver/download/file`,
			{
				PosterFile: posterFile,
				MediaItem: mediaItem,
			}
		);
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Download and update failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};
