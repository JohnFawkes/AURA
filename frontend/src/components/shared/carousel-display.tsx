"use client";

import React from "react";

import { AssetImage } from "@/components/shared/asset-image";
import { CarouselItem } from "@/components/ui/carousel";

import { PosterSet } from "@/types/media-and-posters/poster-sets";

export function CarouselDisplay({ sets }: { sets: PosterSet[] }) {
	return (
		<>
			{sets.map((set) => (
				<React.Fragment key={set.ID}>
					{/* Primary Poster and Backdrop */}
					{(set.Poster || set.Backdrop) && (
						<CarouselItem key={`${set.ID}-primary`}>
							<div className="space-y-2">
								{set.Poster && <AssetImage image={set.Poster} aspect="poster" className="w-full" />}
								{set.Backdrop && (
									<AssetImage image={set.Backdrop} aspect="backdrop" className="w-full" />
								)}
							</div>
						</CarouselItem>
					)}

					{/* All Other Posters and their matching Backdrops */}
					{set.OtherPosters?.map((poster) => {
						const matchingBackdrop = set.OtherBackdrops?.find(
							(backdrop) => backdrop.Movie?.ID === poster.Movie?.ID
						);

						return (
							<CarouselItem key={`${set.ID}-other-${poster.ID}`}>
								<div className="space-y-2">
									<AssetImage
										image={poster}
										aspect="poster"
										className={`w-full ${!poster.Movie?.MediaItem.RatingKey ? "opacity-35" : ""}`}
									/>
									{matchingBackdrop && (
										<AssetImage
											image={matchingBackdrop}
											aspect="backdrop"
											className={`w-full ${!matchingBackdrop.Movie?.MediaItem.RatingKey ? "opacity-35" : ""}`}
										/>
									)}
								</div>
							</CarouselItem>
						);
					})}

					{/* Season Posters with Latest Titlecards */}
					{set.SeasonPosters?.filter(
						(poster) => poster.Type === "seasonPoster" || poster.Type === "specialSeasonPoster"
					)
						.sort((a, b) => (b.Season?.Number ?? 0) - (a.Season?.Number ?? 0))
						.map((seasonPoster) => {
							const matchingTitlecards = set.TitleCards?.filter(
								(titleCard) =>
									titleCard.Type === "titlecard" &&
									titleCard.Episode?.SeasonNumber === seasonPoster.Season?.Number
							);

							const latestTitlecard = matchingTitlecards?.length
								? [...matchingTitlecards].sort(
										(a, b) => new Date(b.Modified).getTime() - new Date(a.Modified).getTime()
									)[0]
								: null;

							return (
								<CarouselItem key={`${set.ID}-season-${seasonPoster.ID}`}>
									<div className="space-y-2">
										<AssetImage
											image={seasonPoster}
											aspect="poster"
											className={`w-full ${!seasonPoster.Show?.MediaItem.RatingKey ? "opacity-35" : ""}`}
										/>
										{latestTitlecard && (
											<AssetImage
												image={latestTitlecard}
												aspect="titlecard"
												className={`w-full ${
													!latestTitlecard.Show?.MediaItem.RatingKey ? "opacity-35" : ""
												}`}
											/>
										)}
									</div>
								</CarouselItem>
							);
						})}

					{/* Standalone Titlecards (only if no posters/season posters exist) */}
					{!set.SeasonPosters?.some(
						(poster) => poster.Type === "seasonPoster" || poster.Type === "specialSeasonPoster"
					) &&
						!set.Poster &&
						set.TitleCards && (
							<>
								{Object.entries(
									set.TitleCards.filter(
										(card) => card.Type === "titlecard" && card.Episode?.SeasonNumber != null
									).reduce(
										(acc, card) => {
											const season = card.Episode!.SeasonNumber;
											if (!acc[season]) acc[season] = [];
											acc[season].push(card);
											return acc;
										},
										{} as Record<number, typeof set.TitleCards>
									)
								)
									.sort(([a], [b]) => Number(b) - Number(a))
									.map(([season, cards]) =>
										cards
											.sort(
												(a, b) =>
													new Date(b.Modified).getTime() - new Date(a.Modified).getTime()
											)
											.slice(0, 3)
											.map((card) => (
												<CarouselItem key={`${set.ID}-titlecard-${season}-${card.ID}`}>
													<div className="space-y-2">
														<AssetImage
															image={card}
															aspect="titlecard"
															className={`w-full ${
																!card.Show?.MediaItem.RatingKey ? "opacity-35" : ""
															}`}
														/>
													</div>
												</CarouselItem>
											))
									)}
							</>
						)}
				</React.Fragment>
			))}
		</>
	);
}
