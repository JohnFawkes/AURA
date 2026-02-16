import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppStatusResponse } from "@/types/config/response-status";

export const getAppConfigStatus = async (reload: boolean): Promise<APIResponse<AppStatusResponse>> => {
  try {
    let response;
    if (!reload) {
      response = await apiClient.get<APIResponse<AppStatusResponse>>(`/config`);
    } else {
      response = await apiClient.patch<APIResponse<AppStatusResponse>>(`/config`);
    }
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching onboarding status");
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Config",
      "Get App Config Status",
      `Failed to fetch onboarding status: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<AppStatusResponse>(error);
  }
};
