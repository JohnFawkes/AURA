import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediuxUserFollowHide } from "@/types/mediux/mediux-user-follow-hide";

export const fetchMediuxUserFollowHides = async (): Promise<APIResponse<MediuxUserFollowHide>> => {
	log(`api.mediux - Fetching Mediux user follow/hide data started`);
	try {
		const response = await apiClient.get<APIResponse<MediuxUserFollowHide>>(`/mediux/user/following_hiding`);
		log(`api.mediux - Fetching Mediux user follow/hide data completed`);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching Mediux user follow/hide data failed with error: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<MediuxUserFollowHide>(error);
	}
};
