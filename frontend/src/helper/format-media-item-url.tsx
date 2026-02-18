import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export function formatMediaItemUrl(mediaItem: MediaItem): string {
  const formattedTitle = mediaItem.title.replace(/\s+/g, "_");
  const sanitizedTitle = formattedTitle.replace(/[^a-zA-Z0-9_]/g, "");
  return `/media/${mediaItem.rating_key}/${sanitizedTitle}`;
}
