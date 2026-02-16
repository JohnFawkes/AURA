import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export const addItemToDownloadQueue = async (dbItem: DBSavedItem): Promise<APIResponse<DBSavedItem>> => {
    log(
        "INFO",
        "API - Media Server",
        "Add to Queue",
        `Adding '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id}) to the download queue`
    );
    for (const set of dbItem.poster_sets) {
        set.last_downloaded = new Date().toISOString();
    }
    try {
        const response = await apiClient.post<APIResponse<DBSavedItem>>(`/download/queue/item`, dbItem);
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error while adding to download queue");
        } else {
            log(
                "INFO",
                "API - Media Server",
                "Add to Queue",
                `Added '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id}) to the download queue`,
                response.data
            );
        }
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - Media Server",
            "Add to Queue",
            `Failed to add '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id}) to the download queue: ${
                dbItem.media_item.rating_key
            }: ${error instanceof Error ? error.message : "Unknown error"}`,
            error
        );
        return ReturnErrorMessage<DBSavedItem>(error);
    }
};
