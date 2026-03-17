import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { AppConfig } from "@/types/config/config";

export const finalizeOnboarding = async (
  newConfig: AppConfig
): Promise<APIResponse<{ onboarding_finalized: boolean }>> => {
  log("INFO", "API - Settings", "Onboarding", "Finalizing app configuration", newConfig);
  try {
    const response = await apiClient.post<APIResponse<{ onboarding_finalized: boolean }>>(
      `/onboarding/finalize`,
      newConfig
    );
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error finalizing app configuration");
    } else {
      log("INFO", "API - Settings", "Onboarding", "Finalized app configuration successfully", response.data);
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Settings",
      "Onboarding",
      `Failed to finalize app configuration: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<{ onboarding_finalized: boolean }>(error);
  }
};
