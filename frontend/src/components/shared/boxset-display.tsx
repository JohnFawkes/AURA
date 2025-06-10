"use client";

import { AssetImage } from "@/components/shared/asset-image";
import DownloadModalBoxset from "@/components/shared/download-modal-boxset";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";

import {
	MediuxUserBoxset,
	MediuxUserImage,
	MediuxUserSeasonPoster,
	MediuxUserTitlecard,
} from "@/types/mediuxUserAllSets";
import { PosterFile } from "@/types/posterSets";

import { Badge } from "../ui/badge";
import {
	Carousel,
	CarouselContent,
	CarouselItem,
	CarouselNext,
	CarouselPrevious,
} from "../ui/carousel";
import { Lead } from "../ui/typography";

export function BoxsetDisplay({
	boxset,
	libraryType,
}: {
	boxset: MediuxUserBoxset;
	libraryType: string;
}) {
	const allPosters: PosterFile[] = [];
	const allBackdrops: PosterFile[] = [];
	const seasonPostersByShow: Record<string, PosterFile[]> = {};
	const titleCardsByShowAndSeason: Record<string, Record<number, PosterFile[]>> = {};

	const showSets = boxset.show_sets;
	const movieSets = boxset.movie_sets;
	const collectionSets = boxset.collection_sets;

	// Show Sets
	showSets.forEach((showSet) => {
		// Posters
		showSet.show_poster.forEach((poster: MediuxUserImage) => {
			allPosters.push({
				ID: poster.id,
				Type: "poster",
				Modified: poster.modified_on,
				FileSize: Number(poster.filesize),
			});
		});
		// Backdrops
		showSet.show_backdrop.forEach((backdrop: MediuxUserImage) => {
			allBackdrops.push({
				ID: backdrop.id,
				Type: "backdrop",
				Modified: backdrop.modified_on,
				FileSize: Number(backdrop.filesize),
			});
		});
		// Season Posters (grouped by show)
		const showTitle = showSet.MediaItem?.Title || showSet.set_title || "Unknown Show";
		showSet.season_posters.forEach((seasonPoster: MediuxUserSeasonPoster) => {
			if (!seasonPostersByShow[showTitle]) {
				seasonPostersByShow[showTitle] = [];
			}
			seasonPostersByShow[showTitle].push({
				ID: seasonPoster.id,
				Type: "season_poster",
				Modified: seasonPoster.modified_on,
				FileSize: Number(seasonPoster.filesize),
				Season: { Number: seasonPoster.season.season_number },
			});
		});
		// Title Cards (grouped by show and season)
		showSet.titlecards.forEach((titleCard: MediuxUserTitlecard) => {
			const seasonNumber = titleCard.episode?.season_id?.season_number ?? 0;
			if (!titleCardsByShowAndSeason[showTitle]) {
				titleCardsByShowAndSeason[showTitle] = {};
			}
			if (!titleCardsByShowAndSeason[showTitle][seasonNumber]) {
				titleCardsByShowAndSeason[showTitle][seasonNumber] = [];
			}
			titleCardsByShowAndSeason[showTitle][seasonNumber].push({
				ID: titleCard.id,
				Type: "title_card",
				Modified: titleCard.modified_on,
				FileSize: Number(titleCard.filesize),
			});
		});
	});

	// Movie Sets
	movieSets.forEach((movieSet) => {
		movieSet.movie_poster.forEach((poster: MediuxUserImage) => {
			allPosters.push({
				ID: poster.id,
				Type: "poster",
				Modified: poster.modified_on,
				FileSize: Number(poster.filesize),
			});
		});
		movieSet.movie_backdrop.forEach((backdrop: MediuxUserImage) => {
			allBackdrops.push({
				ID: backdrop.id,
				Type: "backdrop",
				Modified: backdrop.modified_on,
				FileSize: Number(backdrop.filesize),
			});
		});
	});

	// Collection Sets
	collectionSets.forEach((collectionSet) => {
		collectionSet.movie_posters.forEach((poster: MediuxUserImage) => {
			allPosters.push({
				ID: poster.id,
				Type: "poster",
				Modified: poster.modified_on,
				FileSize: Number(poster.filesize),
			});
		});
		collectionSet.movie_backdrops.forEach((backdrop: MediuxUserImage) => {
			allBackdrops.push({
				ID: backdrop.id,
				Type: "backdrop",
				Modified: backdrop.modified_on,
				FileSize: Number(backdrop.filesize),
			});
		});
	});

	const getUniqueMovieCount = (boxset: MediuxUserBoxset) => {
		const uniqueMovies = new Set();

		// Add movies from movie_sets
		boxset.movie_sets.forEach((movieSet) => {
			uniqueMovies.add(movieSet.MediaItem.RatingKey);
		});

		// Add movies from collection_sets
		boxset.collection_sets.forEach((collection) => {
			collection.movie_posters.forEach((poster) => {
				uniqueMovies.add(poster.movie.MediaItem.RatingKey);
			});
		});

		return uniqueMovies.size;
	};

	return (
		<div className="flex flex-col gap-2 mt-4">
			{/* Summary counts */}
			<div className="mb-4 flex flex-wrap gap-6 text-sm text-muted-foreground">
				{boxset.show_sets.length > 0 && (
					<div>
						<Badge>Shows: {boxset.show_sets.length}</Badge>
					</div>
				)}
				{(boxset.movie_sets.length > 0 || boxset.collection_sets.length > 0) && (
					<div>
						<Badge>Movies: {getUniqueMovieCount(boxset)}</Badge>
					</div>
				)}
				{allPosters.length > 0 && (
					<div>
						<Badge>Posters: {allPosters.length}</Badge>
					</div>
				)}
				{allBackdrops.length > 0 && (
					<div>
						<Badge>Backdrops: {allBackdrops.length}</Badge>
					</div>
				)}
				{Object.values(seasonPostersByShow).reduce(
					(acc, posters) => acc + posters.length,
					0
				) > 0 && (
					<div>
						<Badge>
							Season Posters:{" "}
							{Object.values(seasonPostersByShow).reduce(
								(acc, posters) => acc + posters.length,
								0
							)}
						</Badge>
					</div>
				)}
				{Object.values(titleCardsByShowAndSeason).reduce(
					(acc, seasons) =>
						acc +
						Object.values(seasons).reduce((acc2, cards) => acc2 + cards.length, 0),
					0
				) > 0 && (
					<div>
						<Badge>
							Title Cards:{" "}
							{Object.values(titleCardsByShowAndSeason).reduce(
								(acc, seasons) =>
									acc +
									Object.values(seasons).reduce(
										(acc2, cards) => acc2 + cards.length,
										0
									),
								0
							)}
						</Badge>
					</div>
				)}
			</div>

			{/* Download Button */}
			<div className="mb-4">
				<DownloadModalBoxset boxset={boxset} libraryType={libraryType} />
			</div>

			<Accordion type="multiple" className="w-full">
				{/* All Posters */}
				{allPosters.length > 0 && (
					<AccordionItem value="posters">
						<AccordionTrigger>All Posters</AccordionTrigger>
						<AccordionContent>
							<Carousel
								opts={{
									align: "start",
									dragFree: true,
									slidesToScroll: "auto",
								}}
								className="w-full"
							>
								<CarouselContent>
									{allPosters.map((poster) => (
										<CarouselItem key={`${boxset.id}-poster-${poster.ID}`}>
											<div className="space-y-2">
												<AssetImage
													image={poster as unknown as PosterFile}
													aspect="poster"
													className="w-full h-auto"
												/>
											</div>
										</CarouselItem>
									))}
								</CarouselContent>
								<CarouselNext className="right-2 bottom-0" />
								<CarouselPrevious className="right-8 bottom-0" />
							</Carousel>
						</AccordionContent>
					</AccordionItem>
				)}

				{/* All Backdrops */}
				{allBackdrops.length > 0 && (
					<AccordionItem value="backdrops">
						<AccordionTrigger>All Backdrops</AccordionTrigger>
						<AccordionContent>
							<Carousel
								opts={{
									align: "start",
									dragFree: true,
									slidesToScroll: "auto",
								}}
								className="w-full"
							>
								<CarouselContent>
									{allBackdrops.map((backdrop) => (
										<CarouselItem key={`${boxset.id}-backdrop-${backdrop.ID}`}>
											<div className="space-y-2">
												<AssetImage
													image={backdrop as unknown as PosterFile}
													aspect="backdrop"
													className="w-full h-auto"
												/>
											</div>
										</CarouselItem>
									))}
								</CarouselContent>
								<CarouselNext className="right-2 bottom-0" />
								<CarouselPrevious className="right-8 bottom-0" />
							</Carousel>
						</AccordionContent>
					</AccordionItem>
				)}

				{/* Season Posters by Show */}
				{Object.values(seasonPostersByShow).some((posters) => posters.length > 0) && (
					<AccordionItem value="season-posters">
						<AccordionTrigger>Season Posters</AccordionTrigger>
						<AccordionContent>
							{Object.entries(seasonPostersByShow)
								.filter(([, posters]) => posters.length > 0)
								.map(([showTitle, posters]) => (
									<div key={showTitle} className="mb-8">
										<Lead className="mb-4">{showTitle}</Lead>
										<Carousel
											opts={{
												align: "start",
												dragFree: true,
												slidesToScroll: "auto",
											}}
											className="w-full"
										>
											<CarouselContent>
												{posters.map((poster) => (
													<CarouselItem
														key={`season-poster-${poster.ID}`}
													>
														<div className="space-y-2">
															<AssetImage
																key={poster.ID}
																image={poster}
																aspect="poster"
																className="w-full h-auto"
															/>
														</div>
													</CarouselItem>
												))}
											</CarouselContent>
											<CarouselNext className="right-2 bottom-0" />
											<CarouselPrevious className="right-8 bottom-0" />
										</Carousel>
									</div>
								))}
						</AccordionContent>
					</AccordionItem>
				)}

				{/* Title Cards by Show and Season */}
				{Object.values(titleCardsByShowAndSeason).some((seasons) =>
					Object.values(seasons).some((cards) => cards.length > 0)
				) && (
					<AccordionItem value="title-cards">
						<AccordionTrigger>Title Cards</AccordionTrigger>
						<AccordionContent>
							{Object.entries(titleCardsByShowAndSeason)
								.filter(([, seasons]) =>
									Object.values(seasons).some((cards) => cards.length > 0)
								)
								.map(([showTitle, seasons]) => (
									<div key={showTitle} className="mb-8">
										<Lead className="mb-4">{showTitle}</Lead>
										<Accordion type="multiple" className="w-full">
											{Object.entries(seasons).map(
												([seasonNumber, cards]) => (
													<AccordionItem
														key={`${showTitle}-season-${seasonNumber}`}
														value={`${showTitle}-season-${seasonNumber}`}
													>
														<AccordionTrigger>
															Season {seasonNumber}
														</AccordionTrigger>
														<AccordionContent>
															<Carousel
																opts={{
																	align: "start",
																	dragFree: true,
																	slidesToScroll: "auto",
																}}
																className="w-full"
															>
																<CarouselContent>
																	{cards.map((card) => (
																		<CarouselItem
																			key={`title-card-${card.ID}`}
																		>
																			<div className="space-y-2">
																				<AssetImage
																					image={
																						card as unknown as PosterFile
																					}
																					aspect="backdrop"
																					className="w-full h-auto"
																				/>
																			</div>
																		</CarouselItem>
																	))}
																</CarouselContent>
																<CarouselNext className="right-2 bottom-0" />
																<CarouselPrevious className="right-8 bottom-0" />
															</Carousel>
														</AccordionContent>
													</AccordionItem>
												)
											)}
										</Accordion>
									</div>
								))}
						</AccordionContent>
					</AccordionItem>
				)}
			</Accordion>
		</div>
	);
}
