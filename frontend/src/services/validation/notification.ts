import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type {
  AppConfigNotificationCustomNotification,
  AppConfigNotificationProviders,
  AppConfigNotificationTemplate,
} from "@/types/config/config";

export interface SendTestNotification_Request {
  provider: AppConfigNotificationProviders;
  template_type: keyof AppConfigNotificationTemplate;
  template: AppConfigNotificationCustomNotification;
}
export interface SendTestNotification_Response {
  message: string;
}

export async function postSendTestNotification(
  nProvider: AppConfigNotificationProviders,
  template_type: keyof AppConfigNotificationTemplate,
  template: AppConfigNotificationCustomNotification
): Promise<APIResponse<SendTestNotification_Response>> {
  log(
    "INFO",
    "API - Settings",
    `Notification ${nProvider.provider}`,
    `Posting new ${nProvider.provider} info to check connection status`
  );
  try {
    const req: SendTestNotification_Request = { provider: nProvider, template_type: template_type, template: template };
    const response = await apiClient.post<APIResponse<SendTestNotification_Response>>(`/validate/notifications`, req);
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
    return ReturnErrorMessage<SendTestNotification_Response>(error);
  }
}

export const SendTestNotification = async (
  nProvider: AppConfigNotificationProviders,
  template_type: keyof AppConfigNotificationTemplate,
  template: AppConfigNotificationCustomNotification,
  showToast = true
): Promise<{ ok: boolean; message: string }> => {
  try {
    const response = await postSendTestNotification(nProvider, template_type, template);
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
        errorResponse.error?.message || `Couldn't connect to ${nProvider.provider}. Check the connection details`
      );
    return {
      ok: false,
      message:
        errorResponse.error?.message || `Couldn't connect to ${nProvider.provider}. Check the connection details`,
    };
  }
};
