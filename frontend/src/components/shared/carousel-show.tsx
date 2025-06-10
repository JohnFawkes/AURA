"use client";

import { AssetImage } from "@/components/shared/asset-image";
import { CarouselItem } from "@/components/ui/carousel";

import { PosterFile, PosterSet } from "@/types/posterSets";

export function CarouselShow({ set }: { set: PosterSet }) {
	return (
		<>
			{/* Show Posters with Backdrop */}
			{(set.Poster || set.Backdrop) && (
				<CarouselItem key={`${set.ID}-${set.Poster?.ID || "no-poster"}`}>
					<div className="space-y-2">
						{set.Poster && (
							<AssetImage
								image={set.Poster as unknown as PosterFile}
								aspect="poster"
								className="w-full"
							/>
						)}

						{set.Backdrop && (
							<AssetImage
								image={set.Backdrop as unknown as PosterFile}
								aspect="backdrop"
								className="w-full"
							/>
						)}
					</div>
				</CarouselItem>
			)}

			{/* Season Posters with Titlecards */}
			{set.SeasonPosters?.filter(
				(seasonPoster) =>
					seasonPoster.Type === "seasonPoster" ||
					seasonPoster.Type === "specialSeasonPoster"
			)
				// Sort seasons descending by season number (high to low)
				?.sort((a, b) => (b.Season?.Number ?? 0) - (a.Season?.Number ?? 0))
				?.map((poster) => {
					// For the matching titlecard, select the latest one based on UpdatedAt
					const matchingTitlecards = set.TitleCards?.filter(
						(titleCard) =>
							titleCard.Type === "titlecard" &&
							titleCard.Episode?.SeasonNumber === poster.Season?.Number
					);
					const latestTitlecard = matchingTitlecards
						? matchingTitlecards.sort(
								(a, b) =>
									new Date(b.Modified).getTime() - new Date(a.Modified).getTime()
							)[0]
						: null;

					return (
						<CarouselItem key={`${set.ID}-${poster.ID}`}>
							<div className="space-y-2">
								<AssetImage
									image={poster as unknown as PosterFile}
									aspect="poster"
									className="w-full"
								/>

								{latestTitlecard && (
									<AssetImage
										image={latestTitlecard as unknown as PosterFile}
										aspect="titlecard"
										className="w-full"
									/>
								)}
							</div>
						</CarouselItem>
					);
				})}

			{/* Standalone Titlecards if no Posters or Season Posters */}
			{!set.SeasonPosters?.some(
				(poster) => poster.Type === "seasonPoster" || poster.Type === "specialSeasonPoster"
			) &&
				!set.Poster &&
				set.TitleCards && (
					<>
						{Object.entries(
							set.TitleCards.filter(
								(titleCard) =>
									titleCard.Type === "titlecard" &&
									titleCard.Episode?.SeasonNumber != null
							).reduce(
								(acc, titleCard) => {
									const season = titleCard.Episode!.SeasonNumber;
									if (!acc[season]) {
										acc[season] = [];
									}
									acc[season].push(titleCard);
									return acc;
								},
								{} as Record<number, typeof set.TitleCards>
							)
						)
							.sort((a, b) => Number(b[0]) - Number(a[0]))
							.map(([season, cards]) => {
								const sortedCards = cards.sort(
									(a, b) =>
										new Date(b.Modified).getTime() -
										new Date(a.Modified).getTime()
								);
								return sortedCards.slice(0, 3).map((titleCard) => (
									<CarouselItem
										key={`${set.ID}-season-${season}-${titleCard.ID}`}
									>
										<div className="space-y-2">
											<AssetImage
												image={titleCard as unknown as PosterFile}
												aspect="titlecard"
												className="w-full"
											/>
										</div>
									</CarouselItem>
								));
							})}
					</>
				)}
		</>
	);
}
