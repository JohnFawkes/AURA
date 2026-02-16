import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediux } from "@/types/config/config";

interface ValidationResponse {
	valid: boolean;
	message: string;
}

export const validateMediuxInfo = async (
	mediuxInfo: AppConfigMediux,
	showToast = true
): Promise<ValidationResponse> => {
	try {
		const response = await apiClient.post<APIResponse<ValidationResponse>>(`/validate/mediux`, mediuxInfo);

		if (response.data.status === "error") {
			const msg = response.data.error?.message || "Couldn't connect to MediUX. Check the Token";
			if (showToast) toast.error(msg);
			return { valid: false, message: msg };
		}

		const data = response.data.data;
		if (!data) {
			const msg = "Couldn't connect to MediUX. Check the Token";
			if (showToast) toast.error(msg);
			return { valid: false, message: msg };
		}

		if (showToast) toast.success(data.message || `Successfully connected to MediUX`, { duration: 1000 });
		return data;
	} catch (error) {
		const errorResponse = ReturnErrorMessage<ValidationResponse>(error);
		const msg = errorResponse.error?.message || "Couldn't connect to MediUX. Check the Token";
		if (showToast) toast.error(msg);
		return { valid: false, message: msg };
	}
};
