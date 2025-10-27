import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigNotificationProviders } from "@/types/config/config-app";

export async function postSendTestNotification(
	nProvider: AppConfigNotificationProviders
): Promise<APIResponse<string>> {
	log(
		"INFO",
		"API - Settings",
		`Notification ${nProvider.Provider}`,
		`Posting new ${nProvider.Provider} info to check connection status`
	);
	try {
		const response = await apiClient.post<APIResponse<string>>(
			`/config/validate/notification/${nProvider.Provider.toLowerCase()}`,
			nProvider
		);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || `Unknown error posting ${nProvider.Provider} new info`);
		} else {
			log(
				"INFO",
				"API - Settings",
				nProvider.Provider,
				`Posted ${nProvider.Provider} new info successfully`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			nProvider.Provider,
			`Failed to post ${nProvider.Provider} new info: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const sendTestNotification = async (
	nProvider: AppConfigNotificationProviders,
	showToast = true
): Promise<{ ok: boolean; message: string }> => {
	try {
		const response = await postSendTestNotification(nProvider);
		if (response.status === "error") {
			if (showToast) toast.error(response.error?.Message || "Couldn't connect. Check the connection details");
			return { ok: false, message: response.error?.Message || "API Key invalid" };
		}

		if (showToast) toast.success(`Successfully tested ${nProvider.Provider}`, { duration: 1000 });
		return { ok: true, message: `Successfully tested ${nProvider.Provider}` };
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		if (showToast)
			toast.error(
				errorResponse.error?.Message ||
					`Couldn't connect to ${nProvider.Provider}. Check the connection details`
			);
		return {
			ok: false,
			message:
				errorResponse.error?.Message ||
				`Couldn't connect to ${nProvider.Provider}. Check the connection details`,
		};
	}
};
