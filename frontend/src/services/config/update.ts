import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config";
import { AppStatusResponse } from "@/types/config/response-status";

export interface UpdateAppConfig_Request {
  config: AppConfig;
}

export interface UpdateAppConfig_Response {
  status: AppStatusResponse;
}

export const UpdateAppConfig = async (newConfig: AppConfig): Promise<APIResponse<UpdateAppConfig_Response>> => {
  try {
    const requestBody: UpdateAppConfig_Request = {
      config: newConfig,
    };
    const response = await apiClient.post<APIResponse<UpdateAppConfig_Response>>(`/config`, requestBody);
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
    return ReturnErrorMessage<UpdateAppConfig_Response>(error);
  }
};
