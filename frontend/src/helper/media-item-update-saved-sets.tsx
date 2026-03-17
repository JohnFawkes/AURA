import type { DBSavedSet, MediaItem, SelectedTypes } from "@/types/media-and-posters/media-item-and-library";

export function upsertSavedSets(
  mediaItem: MediaItem,
  setID: string,
  setUser: string,
  selectedTypes: SelectedTypes
): MediaItem {
  const currentSets: DBSavedSet[] = mediaItem.db_saved_sets ?? [];
  const nextSets: DBSavedSet[] = [
    ...currentSets.filter((s) => String(s.id) !== String(setID)),
    { id: setID, user_created: setUser, selected_types: selectedTypes },
  ];

  return { ...mediaItem, db_saved_sets: nextSets };
}

export function removeSavedSet(mediaItem: MediaItem, setID: string): MediaItem {
  const currentSets: DBSavedSet[] = mediaItem.db_saved_sets ?? [];
  const nextSets: DBSavedSet[] = currentSets.filter((s) => String(s.id) !== String(setID));
  return { ...mediaItem, db_saved_sets: nextSets };
}
