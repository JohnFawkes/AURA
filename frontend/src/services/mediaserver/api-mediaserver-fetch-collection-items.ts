import { CollectionItem } from "@/app/collections/page";
import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const fetchCollectionItems = async (): Promise<APIResponse<CollectionItem[]>> => {
	log("INFO", "API - Media Server", "Fetch Collection Items", "Fetching all collection items");
	try {
		const response = await apiClient.get<APIResponse<CollectionItem[]>>(`/mediaserver/collection-items`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching all collection items");
		} else {
			log("INFO", "API - Media Server", "Fetch Collection Items", `Fetched all collection items successfully`);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Fetch Collection Items",
			`Failed to fetch all collection items: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<CollectionItem[]>(error);
	}
};
