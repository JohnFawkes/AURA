import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";

export const stopIgnoringItemInDB = async (
    tmdbID: string,
    libraryTitle: string
): Promise<
    APIResponse<{
        ignored: boolean;
        tmdb_id: string;
        library_title: string;
    }>
> => {
    log(
        "INFO",
        "API - DB",
        "Stop Ignoring Item",
        `Stopping ignoring item TMDB_ID: ${tmdbID} in library: ${libraryTitle}`
    );
    try {
        const response = await apiClient.patch<
            APIResponse<{
                ignored: boolean;
                tmdb_id: string;
                library_title: string;
            }>
        >(`/db/ignore/stop`, null, {
            params: {
                tmdb_id: tmdbID,
                library_title: libraryTitle,
            },
        });
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error stopping ignoring item in DB");
        } else {
            log(
                "INFO",
                "API - DB",
                "Stop Ignoring Item",
                `Stopped ignoring item TMDB_ID: ${tmdbID} in library: ${libraryTitle} successfully`,
                response.data
            );
        }
        const { updateIgnoreStatus } = useLibrarySectionsStore.getState();
        updateIgnoreStatus(tmdbID, libraryTitle, false, "");
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - DB",
            "Stop Ignoring",
            `Failed to stop ignoring item in DB: ${error instanceof Error ? error.message : "Unknown error"}`,
            error
        );
        return ReturnErrorMessage<{
            ignored: boolean;
            tmdb_id: string;
            library_title: string;
        }>(error);
    }
};

export const addIgnoreItemToDB = async (
    tmdbID: string,
    libraryTitle: string,
    ignoreMode: string
): Promise<
    APIResponse<{
        ignored: boolean;
        tmdb_id: string;
        library_title: string;
        mode: string;
    }>
> => {
    log(
        "INFO",
        "API - DB",
        "Ignore Item",
        `Ignoring item TMDB_ID: ${tmdbID} in library: ${libraryTitle} with mode: ${ignoreMode}`
    );
    try {
        const response = await apiClient.patch<
            APIResponse<{
                ignored: boolean;
                tmdb_id: string;
                library_title: string;
                mode: string;
            }>
        >(`/db/ignore`, null, {
            params: {
                tmdb_id: tmdbID,
                library_title: libraryTitle,
                mode: ignoreMode,
            },
        });
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error ignoring item in DB");
        } else {
            log(
                "INFO",
                "API - DB",
                "Ignore Item",
                `Ignored item TMDB_ID: ${tmdbID} in library: ${libraryTitle} with mode: ${ignoreMode} successfully`,
                response.data
            );
        }
        const { updateIgnoreStatus } = useLibrarySectionsStore.getState();
        updateIgnoreStatus(tmdbID, libraryTitle, true, ignoreMode);
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - DB",
            "Add",
            `Failed to add item to DB: ${error instanceof Error ? error.message : "Unknown error"}`,
            error
        );
        return ReturnErrorMessage<{
            ignored: boolean;
            tmdb_id: string;
            library_title: string;
            mode: string;
        }>(error);
    }
};
