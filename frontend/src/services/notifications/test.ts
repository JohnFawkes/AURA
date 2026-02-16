import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigNotificationProviders } from "@/types/config/config";

export async function postSendTestNotification(
    nProvider: AppConfigNotificationProviders
): Promise<APIResponse<string>> {
    log(
        "INFO",
        "API - Settings",
        `Notification ${nProvider.provider}`,
        `Posting new ${nProvider.provider} info to check connection status`
    );
    try {
        const response = await apiClient.post<APIResponse<string>>(`/validate/notifications`, nProvider);
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || `Unknown error posting ${nProvider.provider} new info`);
        } else {
            log(
                "INFO",
                "API - Settings",
                nProvider.provider,
                `Posted ${nProvider.provider} new info successfully`,
                response.data
            );
        }
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - Settings",
            nProvider.provider,
            `Failed to post ${nProvider.provider} new info: ${error instanceof Error ? error.message : "Unknown error"}`,
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
            if (showToast) toast.error(response.error?.message || "Couldn't connect. Check the connection details");
            return { ok: false, message: response.error?.message || "API Key invalid" };
        }

        if (showToast) toast.success(`Successfully tested ${nProvider.provider}`, { duration: 1000 });
        return { ok: true, message: `Successfully tested ${nProvider.provider}` };
    } catch (error) {
        const errorResponse = ReturnErrorMessage<string>(error);
        if (showToast)
            toast.error(
                errorResponse.error?.message ||
                    `Couldn't connect to ${nProvider.provider}. Check the connection details`
            );
        return {
            ok: false,
            message:
                errorResponse.error?.message ||
                `Couldn't connect to ${nProvider.provider}. Check the connection details`,
        };
    }
};
