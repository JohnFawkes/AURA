import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { LibrarySection, MediaItem } from "../types/mediaItem";
import { PosterSet } from "../types/posterSets";

export const fetchPlexSections = async (): Promise<
	APIResponse<LibrarySection[]>
> => {
	try {
		const response = await apiClient.get<APIResponse<LibrarySection[]>>(
			`/plex/sections/all/`
		);
		return response.data;
	} catch {
		return {
			status: "error",
			message: "Failed to fetch data from Plex API",
		};
	}
};

export const fetchPlexItem = async (
	ratingKey: string
): Promise<APIResponse<MediaItem>> => {
	try {
		const response = await apiClient.get<APIResponse<MediaItem>>(
			`/plex/item/${ratingKey}`
		);
		return response.data;
	} catch {
		return {
			status: "error",
			message: "Failed to fetch data from Plex API",
		};
	}
};

export const postSendSetToAPI = async (sendData: {
	Set: PosterSet;
	SelectedTypes: string[];
	Plex: MediaItem;
	AutoDownload: boolean;
}): Promise<APIResponse<null>> => {
	try {
		const response = await apiClient.post<APIResponse<null>>(
			`/plex/update/send`,
			sendData
		);
		return response.data;
	} catch {
		return {
			status: "error",
			message: "Failed to send set to Plex API",
		};
	}
};

export const fetchPlexImageData = async (
	ratingKey: string,
	type: string
): Promise<string> => {
	try {
		const API_URL = `/plex/image/${ratingKey}/${type}`;
		const response = await apiClient.get<APIResponse<null>>(API_URL);
		if (response.status !== 200) {
			throw new Error("Failed to fetch image data");
		}
		return "/api" + API_URL;
	} catch {
		return "/logo.png"; // Fallback image
	}
};
