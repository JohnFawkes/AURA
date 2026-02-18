import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediux } from "@/types/config/config";

interface ValidateMediuxInfo_Request {
  mediux_info: AppConfigMediux;
}
interface ValidateMediuxInfo_Response {
  valid: boolean;
  message: string;
}

export const ValidateMediuxInfo = async (
  mediuxInfo: AppConfigMediux,
  showToast = true
): Promise<ValidateMediuxInfo_Response> => {
  try {
    const req: ValidateMediuxInfo_Request = { mediux_info: mediuxInfo };
    const response = await apiClient.post<APIResponse<ValidateMediuxInfo_Response>>(`/validate/mediux`, req);

    if (response.data.status === "error") {
      const msg = response.data.error?.message || "Couldn't connect to MediUX. Check the Token";
      if (showToast) toast.error(msg);
      return { valid: false, message: msg };
    }

    const data = response.data.data;
    if (!data) {
      const msg = "Couldn't connect to MediUX. Check the Token";
      if (showToast) toast.error(msg);
      return { valid: false, message: msg };
    }

    if (showToast) toast.success(data.message || `Successfully connected to MediUX`, { duration: 1000 });
    return data;
  } catch (error) {
    const errorResponse = ReturnErrorMessage<ValidateMediuxInfo_Response>(error);
    const msg = errorResponse.error?.message || "Couldn't connect to MediUX. Check the Token";
    if (showToast) toast.error(msg);
    return { valid: false, message: msg };
  }
};
