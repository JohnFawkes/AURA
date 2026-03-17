import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { AppConfigMediaServer } from "@/types/config/config";

export interface ValidateMediaServerInfo_Request {
  media_server: AppConfigMediaServer;
}

export interface ValidateMediaServerInfo_Response {
  valid: boolean;
  media_server: AppConfigMediaServer | undefined;
  message: string;
}

export const ValidateMediaServerInfo = async (
  mediaServerInfo: AppConfigMediaServer,
  showToast = true
): Promise<ValidateMediaServerInfo_Response> => {
  log("INFO", "API - Settings", "Media Server", "Validating media server connection info");

  let loadingToast: string | number | undefined;
  try {
    if (showToast) {
      loadingToast = toast.loading(`Checking connection to ${mediaServerInfo.type}...`);
    }

    const req: ValidateMediaServerInfo_Request = {
      media_server: mediaServerInfo,
    };
    const response = await apiClient.post<APIResponse<ValidateMediaServerInfo_Response>>(`/validate/mediaserver`, req);

    if (showToast && loadingToast) toast.dismiss(loadingToast);

    // API-level error wrapper
    if (response.data.status === "error") {
      const msg = response.data.error?.message || "Couldn't connect to media server. Check the URL and Token";
      if (showToast) toast.error(msg, { duration: 1000 });
      return { valid: false, message: msg, media_server: undefined };
    }

    // Success wrapper but missing/invalid payload guard
    const data = response.data.data;
    if (!data) {
      const msg = "Couldn't connect to media server. Check the URL and Token";
      if (showToast) toast.error(msg, { duration: 1000 });
      return { valid: false, message: msg, media_server: undefined };
    }

    // Backend decides valid/message
    if (showToast) {
      if (data.valid)
        toast.success(data.message || `Successfully connected to ${mediaServerInfo.type}`, { duration: 1500 });
      else
        toast.error(data.message || "Couldn't connect to media server. Check the URL and Token", {
          duration: 1000,
        });
    }

    log("INFO", "API - Settings", "Media Server", "Validation response received", data);
    return data;
  } catch (error) {
    if (showToast) {
      if (loadingToast) toast.dismiss(loadingToast);
      else toast.dismiss();
    }

    const errorResponse = ReturnErrorMessage<ValidateMediaServerInfo_Response>(error);
    const msg = errorResponse.error?.message || "Couldn't connect to media server. Check the URL and Token";

    log(
      "ERROR",
      "API - Settings",
      "Media Server",
      `Validation request failed: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );

    if (showToast) toast.error(msg, { duration: 1000 });
    return { valid: false, message: msg, media_server: undefined };
  }
};
