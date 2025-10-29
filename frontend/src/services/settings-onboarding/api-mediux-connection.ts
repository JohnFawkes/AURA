import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediux } from "@/types/config/config-app";

export async function postMediuxNewTokenStatus(mediuxInfo: AppConfigMediux): Promise<APIResponse<string>> {
	log("INFO", "API - Settings", "MediUX", "Posting mediux new token to check connection status");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/config/validate/mediux`, mediuxInfo);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error posting mediux new token");
		} else {
			log("INFO", "API - Settings", "MediUX", "Posted mediux new token successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"MediUX",
			`Failed to post mediux new token: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const checkMediuxNewTokenStatusResult = async (
	mediuxInfo: AppConfigMediux,
	showToast = true
): Promise<{ ok: boolean; message: string }> => {
	try {
		const response = await postMediuxNewTokenStatus(mediuxInfo);
		if (response.status === "error") {
			if (showToast) toast.error(response.error?.message || "Couldn't connect to MediUX. Check the Token");
			return { ok: false, message: response.error?.message || "Token invalid" };
		}

		if (showToast) toast.success(`Successfully connected to MediUX`, { duration: 1000 });
		return { ok: true, message: "Successfully connected to MediUX" };
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		if (showToast) toast.error(errorResponse.error?.message || "Couldn't connect to MediUX. Check the Token");
		return { ok: false, message: errorResponse.error?.message || "Couldn't connect to MediUX. Check the Token" };
	}
};
