import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const postSendTestNotification = async (): Promise<APIResponse<string>> => {
	log("INFO", "API - Settings", "Notifications", "Sending test notification");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/health/test/notification`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error sending test notification");
		} else {
			log("INFO", "API - Settings", "Notifications", "Sent test notification successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Notifications",
			`Failed to send test notification: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
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
