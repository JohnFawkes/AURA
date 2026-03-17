import apiClient from "@/services/api-client";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";

export interface GetDownloadQueueStatus_Response {
  time: string;
  status: string;
  message: string;
  warnings: string[];
  errors: string[];
}

export const GetDownloadQueueStatus = async (): Promise<APIResponse<GetDownloadQueueStatus_Response>> => {
  try {
    const response = await apiClient.get<APIResponse<GetDownloadQueueStatus_Response>>(`/download/queue`);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching download queue status");
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Download Queue", "Fetch", "Error fetching download queue status", { error });
    throw error;
  }
};
