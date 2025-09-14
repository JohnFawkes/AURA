import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";
import { AppConfigMediux } from "@/types/config";

export async function postMediuxNewTokenStatus(mediuxInfo: AppConfigMediux): Promise<APIResponse<string>> {
	log("api.settings - Posting mediux new token status started");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/config/validate/mediux`, mediuxInfo);
		log("api.settings - Posting mediux new token status succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Posting mediux new token status failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const checkMediuxNewTokenStatusResult = async (
	mediuxInfo: AppConfigMediux
): Promise<{ ok: boolean; message: string }> => {
	try {
		const response = await postMediuxNewTokenStatus(mediuxInfo);
		if (response.status === "error") {
			toast.error(response.error?.Message || "Couldn't connect to MediUX. Check the Token");
			return { ok: false, message: response.error?.Message || "Token invalid" };
		}

		toast.success(`Successfully connected to MediUX`, { duration: 1000 });
		return { ok: true, message: "Successfully connected to MediUX" };
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		toast.error(errorResponse.error?.Message || "Couldn't connect to MediUX. Check the Token");
		return { ok: false, message: errorResponse.error?.Message || "Couldn't connect to MediUX. Check the Token" };
	}
};
