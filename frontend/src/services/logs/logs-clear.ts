import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const clearLogFile = async (clearCurrent: boolean = false): Promise<APIResponse<void>> => {
    log("INFO", "API - Logs", "Clear Old Logs", `Clearing old logs, clearCurrent=${clearCurrent}`);
    try {
        const response = await apiClient.delete<APIResponse<void>>(`/logs`, {
            params: { option: clearCurrent ? "current" : "old" },
        });
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
        return ReturnErrorMessage<void>(error);
    }
};
