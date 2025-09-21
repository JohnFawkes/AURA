import { Lead } from "@/components/ui/typography";

import { cn } from "@/lib/cn";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

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
			primary = "Poster & Backdrop";
		} else if (set.Poster) {
			primary = "Poster Only";
		} else if (set.Backdrop) {
			primary = "Backdrop Only";
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
		const primary: React.ReactNode[] = [];
		let secondary = "";

		if (set.Poster) primary.push(<span key="poster">1 Poster</span>);
		if (set.Backdrop) {
			if (primary.length) primary.push(<span key="sep1"> • </span>);
			primary.push(<span key="backdrop">1 Backdrop</span>);
		}

		const seasonPosterCount =
			set.SeasonPosters?.filter((file) => file.Type === "seasonPoster" || file.Type === "specialSeasonPoster")
				.length ?? 0;
		const mediaSeasonCount = mediaItem.Series?.SeasonCount ?? 0;

		if (seasonPosterCount > 0 && mediaSeasonCount > 0) {
			if (primary.length) primary.push(<span key="sep2"> • </span>);
			if (seasonPosterCount < mediaSeasonCount) {
				primary.push(
					<span key="season-count" className="text-yellow-500">
						{seasonPosterCount}/{mediaSeasonCount} Season Poster{mediaSeasonCount > 1 ? "s" : ""}
					</span>
				);
			} else {
				primary.push(
					<span key="season-label">
						{seasonPosterCount} Season Poster{seasonPosterCount > 1 ? "s" : ""}
					</span>
				);
			}
		} else if (seasonPosterCount > 0 && mediaSeasonCount === 0) {
			if (primary.length) primary.push(<span key="sep3"> • </span>);
			primary.push(
				<span key="season-label">
					{seasonPosterCount} Season Poster{seasonPosterCount > 1 ? "s" : ""}
				</span>
			);
		}

		// Secondary line logic for titlecards (unchanged)
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

	const { primary, secondary } = mediaItem.Type === "movie" ? getMovieFileCounts() : getShowFileCounts();

	const secondaryLineClass = cn("text-sm ml-1", {
		"text-yellow-500": secondary.toLowerCase().includes("server is missing"),
		"text-orange-500": secondary.toLowerCase().includes("set is missing"),
		"text-muted-foreground": !secondary.toLowerCase().includes("missing"),
	});

	return (
		<div>
			{primary && <Lead className="text-sm text-muted-foreground ml-1">{primary}</Lead>}
			{secondary && <Lead className={secondaryLineClass}>{secondary}</Lead>}
		</div>
	);
}
