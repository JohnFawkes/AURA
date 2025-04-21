import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { PosterSets } from "../types/posterSets";
import { ReturnErrorMessage } from "./api.shared";

export const fetchMediuxSets = async (
	tmdbID: string,
	itemType: string
): Promise<APIResponse<PosterSets>> => {
	try {
		const response = await apiClient.get<APIResponse<PosterSets>>(
			`/mediux/sets/get/${itemType}/${tmdbID}`
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<PosterSets>(error);
	}
};

export const fetchMediuxImageData = async (
	assetID: string,
	modifiedDate: string
): Promise<string> => {
	try {
		const API_URL = `/mediux/image/${assetID}?modifiedDate=${modifiedDate}`;
		const response = await apiClient.get(API_URL);
		if (response.status !== 200) {
			throw new Error("Failed to fetch image data");
		}
		return "/api" + API_URL;
	} catch {
		return "";
	}
};
