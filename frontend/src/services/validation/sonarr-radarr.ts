import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigSonarrRadarrApp } from "@/types/config/config";

export async function postSonarrRadarrNewAPIKeyStatus(srApp: AppConfigSonarrRadarrApp): Promise<APIResponse<string>> {
  log("INFO", "API - Settings", srApp.type, `Posting new ${srApp.type} info to check connection status`);
  try {
    const response = await apiClient.post<APIResponse<string>>(`/validate/${srApp.type.toLowerCase()}`, srApp);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || `Unknown error posting ${srApp.type} new info`);
    } else {
      log("INFO", "API - Settings", srApp.type, `Posted ${srApp.type} new info successfully`, response.data);
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Settings",
      srApp.type,
      `Failed to post ${srApp.type} new info: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<string>(error);
  }
}

export const checkSonarrRadarrNewAPIKeyStatusResult = async (
  srApp: AppConfigSonarrRadarrApp,
  showToast = true
): Promise<{ ok: boolean; message: string }> => {
  try {
    const response = await postSonarrRadarrNewAPIKeyStatus(srApp);
    if (response.status === "error") {
      if (showToast) toast.error(response.error?.message || "Couldn't connect. Check the API Key and URL");
      return { ok: false, message: response.error?.message || "API Key invalid" };
    }

    if (showToast) toast.success(`Successfully connected to ${srApp.type} (${srApp.library})`, { duration: 1000 });
    return { ok: true, message: `Successfully connected to ${srApp.type} (${srApp.library})` };
  } catch (error) {
    const errorResponse = ReturnErrorMessage<string>(error);
    if (showToast)
      toast.error(errorResponse.error?.message || `Couldn't connect to ${srApp.type}. Check the API Key and URL`);
    return {
      ok: false,
      message: errorResponse.error?.message || `Couldn't connect to ${srApp.type}. Check the API Key and URL`,
    };
  }
};
