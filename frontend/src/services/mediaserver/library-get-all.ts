import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediaServer } from "@/types/config/config";
import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

/**
 * Fetches all library sections from the media server that are available as options.
 *
 * Initiates a GET request to the `/api/mediaserver/libraries/options` endpoint and returns the response data.
 *
 * @returns {Promise<APIResponse<LibrarySection[]>>} A promise that resolves to the API response containing an array of library sections.
 */
export const getLibrarySectionOptions = async (
    config: AppConfigMediaServer
): Promise<APIResponse<LibrarySection[]>> => {
    log("INFO", "API - Media Server", "Fetch Library Section Options", "Fetching all library section options");
    try {
        const response = await apiClient.post<APIResponse<LibrarySection[]>>(`/mediaserver/libraries/options`, config);
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error fetching all library section options");
        } else {
            log(
                "INFO",
                "API - Media Server",
                "Fetch Library Section Options",
                `Fetched all library section options successfully`
            );
        }
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - Media Server",
            "Fetch Library Section Options",
            `Failed to fetch all library section options: ${error instanceof Error ? error.message : "Unknown error"}`,
            error
        );
        return ReturnErrorMessage<LibrarySection[]>(error);
    }
};
