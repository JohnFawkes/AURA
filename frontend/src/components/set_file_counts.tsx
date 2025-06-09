import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";

import { Lead } from "./ui/typography";

type SetFileCountsProps = {
	mediaItem: MediaItem;
	set: PosterSet;
};

export function SetFileCounts({ mediaItem, set }: SetFileCountsProps) {
	if (!set) {
		return null;
	}

	let primaryLine = "";
	let secondaryLine = "";
	if (mediaItem.Type === "movie") {
		// First line: if both primary files are present, or only one of them.
		primaryLine = "For This Movie: ";
		if (set.Poster && set.Backdrop) {
			primaryLine += "Poster & Backdrop";
		} else if (set.Poster) {
			primaryLine += "Poster Only";
		} else if (set.Backdrop) {
			primaryLine += "Backdrop Only";
		} else {
			primaryLine = "";
		}

		// Second line: count files in set.OtherPosters & set.OtherBackdrops.
		const otherPosterCount = set.OtherPosters ? set.OtherPosters.length : 0;
		const otherBackdropCount = set.OtherBackdrops ? set.OtherBackdrops.length : 0;
		if (otherPosterCount || otherBackdropCount) {
			secondaryLine = "Other Movies: ";
			if (otherPosterCount > 0) {
				secondaryLine += `${otherPosterCount} Posters`;
			}
			if (otherBackdropCount > 0) {
				if (secondaryLine) {
					secondaryLine += " & ";
				}
				secondaryLine += `${otherBackdropCount} Backdrops`;
			}
			if (secondaryLine) {
				secondaryLine += " in Set";
			}
		}
	} else if (mediaItem.Type === "show") {
		// First Line: # Posters • # Backdrops • # Season Posters
		const posterCount = set.Poster ? 1 : 0;
		const backdropCount = set.Backdrop ? 1 : 0;
		const seasonPosterCount =
			set.SeasonPosters?.filter(
				(file) => file.Type === "seasonPoster" || file.Type === "specialSeasonPoster"
			).length ?? 0;
		if (posterCount > 0) {
			primaryLine += `${posterCount} Poster${posterCount > 1 ? "s" : ""}`;
		}
		if (backdropCount > 0) {
			if (primaryLine) {
				primaryLine += " • ";
			}
			primaryLine += `${backdropCount} Backdrop${backdropCount > 1 ? "s" : ""}`;
		}
		if (seasonPosterCount > 0) {
			if (primaryLine) {
				primaryLine += " • ";
			}
			primaryLine += `${seasonPosterCount} Season Poster${seasonPosterCount > 1 ? "s" : ""}`;
		}

		// Second Line:
		// # Titlecards (set is missing # episodes)
		// or
		// # Titlecards (server is missing # episodes from set)
		// or
		// # Titlecards
		const titlecardCount = set.TitleCards?.length ?? 0;
		const episodeCount = mediaItem.Series?.EpisodeCount ?? 0;

		if (titlecardCount > 0 && episodeCount > 0) {
			if (titlecardCount === episodeCount) {
				secondaryLine = `${titlecardCount} Titlecards`;
			} else if (titlecardCount > episodeCount) {
				const missingEpisodes = titlecardCount - episodeCount;
				secondaryLine = `${titlecardCount} Titlecards (Server is Missing ${missingEpisodes} Episode${
					missingEpisodes > 1 ? "s" : ""
				})`;
			} else {
				const missingEpisodes = episodeCount - titlecardCount;
				secondaryLine = `${titlecardCount} Titlecards (Set is Missing ${missingEpisodes} Episode${
					missingEpisodes > 1 ? "s" : ""
				})`;
			}
		}
	}

	// Set text color based on the content of secondaryLine.
	const lowerSecondaryLine = secondaryLine.toLowerCase();
	let secondaryLineClass = "text-sm ";
	if (lowerSecondaryLine.includes("server is missing")) {
		secondaryLineClass += "text-yellow-500";
	} else if (lowerSecondaryLine.includes("set is missing")) {
		secondaryLineClass += "text-orange-500";
	} else {
		secondaryLineClass += "text-muted-foreground";
	}

	return (
		<div>
			{primaryLine && <Lead className="text-sm text-muted-foreground">{primaryLine}</Lead>}
			{secondaryLine && <Lead className={secondaryLineClass}>{secondaryLine}</Lead>}
		</div>
	);
}
