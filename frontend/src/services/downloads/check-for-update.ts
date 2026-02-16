import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export interface AutodownloadResult {
    item: string;
    sets: AutodownloadSetResult[];
    overall_result: "error" | "warn" | "success" | "skipped";
    overall_message: string;
}

export interface AutodownloadSetResult {
    id: string;
    title: string;
    user_created: string;
    result: "success" | "skipped" | "error";
    reason: string;
}

export const checkSavedDBItemForUpdates = async (saveItem: DBSavedItem): Promise<APIResponse<AutodownloadResult>> => {
    let complete = true;
    let size = JSON.stringify(saveItem).length / 1024 / 1024;
    if (size > 5) {
        complete = false;
        saveItem.poster_sets.forEach((posterSet) => {
            posterSet.images = [];
        });
    }
    for (const set of saveItem.poster_sets) {
        set.last_downloaded = new Date().toISOString();
    }
    log(
        "INFO",
        "API - DB",
        "Recheck",
        `Forcing recheck for auto-download for ${saveItem.media_item.title} (${saveItem.media_item.tmdb_id} | ${saveItem.media_item.library_title})`,
        { saveItem, complete, size }
    );
    try {
        const response = await apiClient.post<APIResponse<AutodownloadResult>>(`/db/force-check`, {
            Item: saveItem,
            Complete: complete,
        });
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error forcing recheck for auto-download");
        } else {
            log(
                "INFO",
                "API - DB",
                "Recheck",
                `Forcing recheck for auto-download for ${saveItem.media_item.title} (${saveItem.media_item.tmdb_id} | ${saveItem.media_item.library_title}) succeeded`,
                response.data
            );
        }
        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - DB",
            "Recheck",
            `Failed to force recheck for auto-download for ${saveItem.media_item.title} (${saveItem.media_item.tmdb_id} | ${saveItem.media_item.library_title}): ${
                error instanceof Error ? error.message : "Unknown error"
            }`,
            error
        );
        return ReturnErrorMessage<AutodownloadResult>(error);
    }
};
