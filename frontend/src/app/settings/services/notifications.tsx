import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";

export const postSendTestNotification = async (): Promise<APIResponse<string>> => {
	log("api.settings - Sending test notification started");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/health/test/notification`);
		log("api.settings - Sending test notification succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Sending test notification failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const sendTestNotification = async (): Promise<{ ok: boolean; message: string }> => {
	try {
		const response = await postSendTestNotification();
		if (response.status === "error") {
			toast.error(response.error?.Message || "Error sending test notifications. Check logs.");
			return { ok: false, message: response.error?.Message || "Error sending test notifications. Check logs." };
		}

		toast.success(`Successfully sent test notifications`, { duration: 1000 });
		return { ok: true, message: "Successfully sent test notifications" };
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		toast.error(errorResponse.error?.Message || "Error sending test notifications. Check logs.");
		return { ok: false, message: errorResponse.error?.Message || "Error sending test notifications. Check logs." };
	}
};
