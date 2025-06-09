import { MediaItem } from "@/types/mediaItem";

export function formatMediaItemUrl(mediaItem: MediaItem): string {
	const formattedTitle = mediaItem.Title.replace(/\s+/g, "_");
	const sanitizedTitle = formattedTitle.replace(/[^a-zA-Z0-9_]/g, "");
	return `/media/${mediaItem.RatingKey}/${sanitizedTitle}`;
}
