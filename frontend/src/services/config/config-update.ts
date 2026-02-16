import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config";
import { AppStatusResponse } from "@/types/config/response-status";

export const updateAppConfig = async (newConfig: AppConfig): Promise<APIResponse<AppStatusResponse>> => {
    try {
        const response = await apiClient.post<APIResponse<AppStatusResponse>>(`/config`, newConfig);
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error updating onboarding status");
        }
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - Config",
            "Update Config",
            `Failed to update onboarding status: ${error instanceof Error ? error.message : "Unknown error"}`,
            error
        );
        return ReturnErrorMessage<AppStatusResponse>(error);
    }
};
