import { APIResponse } from "@/types/apiResponse";
import { LibrarySection, MediaItem } from "@/types/mediaItem";
import apiClient from "./apiClient";
import { ReturnErrorMessage } from "./api.shared";
import { ClientMessage } from "@/types/clientMessage";

export const fetchMediaServerLibraryItems = async (): Promise<
	APIResponse<LibrarySection[]>
> => {
	try {
		const response = await apiClient.get<APIResponse<LibrarySection[]>>(
			`/mediaserver/sections/all/`
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<LibrarySection[]>(error);
	}
};

export const fetchMediaServerItemContent = async (
	ratingKey: string
): Promise<APIResponse<MediaItem>> => {
	try {
		console.log("Fetching Media Item Content:", ratingKey);
		const response = await apiClient.get<APIResponse<MediaItem>>(
			`/mediaserver/item/${ratingKey}`
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<MediaItem>(error);
	}
};

export const postSendSetToAPI = async (
	sendData: ClientMessage
): Promise<APIResponse<null>> => {
	try {
		const response = await apiClient.post<APIResponse<null>>(
			`/mediaserver/update/send`,
			sendData
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<null>(error);
	}
};
