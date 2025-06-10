import { Lead } from "@/components/ui/typography";

import { cn } from "@/lib/utils";

import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";

interface SetFileCountsProps {
	mediaItem: MediaItem;
	set: PosterSet;
}

export function SetFileCounts({ mediaItem, set }: SetFileCountsProps) {
	if (!set) return null;

	const getMovieFileCounts = () => {
		let primary = "";
		let secondary = "";

		// Primary line logic
		if (set.Poster && set.Backdrop) {
			primary = "For This Movie: Poster & Backdrop";
		} else if (set.Poster) {
			primary = "For This Movie: Poster Only";
		} else if (set.Backdrop) {
			primary = "For This Movie: Backdrop Only";
		}

		// Secondary line logic
		const otherPosterCount = set.OtherPosters?.length ?? 0;
		const otherBackdropCount = set.OtherBackdrops?.length ?? 0;

		if (otherPosterCount || otherBackdropCount) {
			const parts = [];
			if (otherPosterCount > 0) parts.push(`${otherPosterCount} Posters`);
			if (otherBackdropCount > 0) parts.push(`${otherBackdropCount} Backdrops`);
			secondary = `Other Movies: ${parts.join(" & ")} in Set`;
		}

		return { primary, secondary };
	};

	const getShowFileCounts = () => {
		let primary = "";
		let secondary = "";

		// Primary line logic
		const parts = [];
		if (set.Poster) parts.push("1 Poster");
		if (set.Backdrop) parts.push("1 Backdrop");

		const seasonPosterCount =
			set.SeasonPosters?.filter(
				(file) => file.Type === "seasonPoster" || file.Type === "specialSeasonPoster"
			).length ?? 0;
		if (seasonPosterCount > 0) {
			parts.push(`${seasonPosterCount} Season Poster${seasonPosterCount > 1 ? "s" : ""}`);
		}
		primary = parts.join(" • ");

		// Secondary line logic for titlecards
		const titlecardCount = set.TitleCards?.length ?? 0;
		const episodeCount = mediaItem.Series?.EpisodeCount ?? 0;

		if (titlecardCount > 0 && episodeCount > 0) {
			if (titlecardCount === episodeCount) {
				secondary = `${titlecardCount} Titlecards`;
			} else if (titlecardCount > episodeCount) {
				const missing = titlecardCount - episodeCount;
				secondary = `${titlecardCount} Titlecards (Server is Missing ${missing} Episode${missing > 1 ? "s" : ""})`;
			} else {
				const missing = episodeCount - titlecardCount;
				secondary = `${titlecardCount} Titlecards (Set is Missing ${missing} Episode${missing > 1 ? "s" : ""})`;
			}
		}

		return { primary, secondary };
	};

	const { primary, secondary } =
		mediaItem.Type === "movie" ? getMovieFileCounts() : getShowFileCounts();

	const secondaryLineClass = cn("text-sm", {
		"text-yellow-500": secondary.toLowerCase().includes("server is missing"),
		"text-orange-500": secondary.toLowerCase().includes("set is missing"),
		"text-muted-foreground": !secondary.toLowerCase().includes("missing"),
	});

	return (
		<div>
			{primary && <Lead className="text-sm text-muted-foreground">{primary}</Lead>}
			{secondary && <Lead className={secondaryLineClass}>{secondary}</Lead>}
		</div>
	);
}
