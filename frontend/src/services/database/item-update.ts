import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export const updateItemInDB = async (saveItem: DBSavedItem): Promise<APIResponse<DBSavedItem>> => {
    log(
        "INFO",
        "API - DB",
        "Update",
        `Patching ${saveItem.media_item.title} (${saveItem.media_item.tmdb_id} | ${saveItem.media_item.library_title}) in DB`,
        saveItem
    );
    for (const set of saveItem.poster_sets) {
        set.last_downloaded = new Date().toISOString();
    }

    try {
        const response = await apiClient.patch<APIResponse<DBSavedItem>>(`/db`, {
            update_item: saveItem,
            complete: true,
        });
        if (response.data.status === "error") {
            throw new Error(response.data.error?.message || "Unknown error patching item in DB");
        } else {
            log(
                "INFO",
                "API - DB",
                "Update",
                `Patched ${saveItem.media_item.title} (${saveItem.media_item.tmdb_id} | ${saveItem.media_item.library_title}) in DB`,
                response.data
            );
        }
        const { upsertMediaItemSavedSet } = useLibrarySectionsStore.getState();
        for (const savedSet of saveItem.poster_sets) {
            upsertMediaItemSavedSet({
                tmdbID: saveItem.media_item.tmdb_id,
                libraryTitle: saveItem.media_item.library_title,
                setID: savedSet.id,
                setUser: savedSet.user_created,
                selectedTypes: savedSet.selected_types,
                toDelete: savedSet.to_delete,
            });
        }

        return response.data;
    } catch (error) {
        log(
            "ERROR",
            "API - DB",
            "Update",
            `Failed to patch ${saveItem.media_item.title} (${saveItem.media_item.tmdb_id} | ${saveItem.media_item.library_title}) in DB: ${
                error instanceof Error ? error.message : "Unknown error"
            }`,
            error
        );
        return ReturnErrorMessage<DBSavedItem>(error);
    }
};
