import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export interface ClearLogFiles_Response {
  message: string;
}

export const ClearLogFiles = async (clearCurrent: boolean = false): Promise<APIResponse<ClearLogFiles_Response>> => {
  log("INFO", "API - Logs", "Clear Old Logs", `Clearing old logs, clearCurrent=${clearCurrent}`);
  try {
    const params = { option: clearCurrent ? "current" : "old" };
    const response = await apiClient.delete<APIResponse<ClearLogFiles_Response>>(`/logs`, { params });
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error clearing old logs");
    } else {
      log("INFO", "API - Logs", "Clear Old Logs", "Cleared old logs successfully", response.data);
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Logs",
      "Clear Old Logs",
      `Failed to clear old logs: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<ClearLogFiles_Response>(error);
  }
};
