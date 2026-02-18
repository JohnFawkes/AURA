import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppStatusResponse } from "@/types/config/response-status";

export interface ReloadAppConfig_Response {
  status: AppStatusResponse;
}

export const ReloadAppConfig = async (): Promise<APIResponse<ReloadAppConfig_Response>> => {
  try {
    const response = await apiClient.patch<APIResponse<ReloadAppConfig_Response>>(`/config`);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching onboarding status");
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Config",
      "Reload App Config",
      `Failed to reload app config: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<ReloadAppConfig_Response>(error);
  }
};
