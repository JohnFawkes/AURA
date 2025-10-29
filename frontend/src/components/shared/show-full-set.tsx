import { ReturnErrorMessage } from "@/services/api-error-return";

import { useEffect, useState } from "react";

import { AssetImage } from "@/components/shared/asset-image";
import { DimmedBackground } from "@/components/shared/dimmed_backdrop";
import DownloadModal, { DownloadModalProps } from "@/components/shared/download-modal";
import { ErrorMessage } from "@/components/shared/error-message";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { Carousel, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious } from "@/components/ui/carousel";
import { H1, Lead } from "@/components/ui/typography";

import { log } from "@/lib/logger";

import { PosterFile, PosterSet } from "@/types/media-and-posters/poster-sets";

export const ShowFullSetsDisplay: React.FC<DownloadModalProps> = ({
	setType,
	setTitle,
	setAuthor,
	setID,
	posterSets,
}) => {
	const allPosters: PosterFile[] = [];
	const allBackdrops: PosterFile[] = [];
	const seasonPostersByShow: Record<string, PosterFile[]> = {};
	const titleCardsByShowAndSeason: Record<string, Record<number, PosterFile[]>> = {};

	if (!posterSets || posterSets.length === 0) {
		return <ErrorMessage error={ReturnErrorMessage<string>("No Poster Sets found")} />;
	}

	posterSets.forEach((posterSet: PosterSet) => {
		// All Posters are the poster and other posters
		if (posterSet.Poster) {
			allPosters.push(posterSet.Poster as PosterFile);
		}
		if (posterSet.OtherPosters) {
			allPosters.push(...posterSet.OtherPosters);
		}

		// All Backdrops are the backdrop and other backdrops
		if (posterSet.Backdrop) {
			allBackdrops.push(posterSet.Backdrop as PosterFile);
		}
		if (posterSet.OtherBackdrops) {
			allBackdrops.push(...posterSet.OtherBackdrops);
		}

		// Season Posters by Show
		if (posterSet.SeasonPosters) {
			posterSet.SeasonPosters.forEach((poster) => {
				if (poster.Type === "seasonPoster" || poster.Type === "specialSeasonPoster") {
					const showTitle = poster.Show?.Title || "";
					if (!seasonPostersByShow[showTitle]) {
						seasonPostersByShow[showTitle] = [];
					}
					seasonPostersByShow[showTitle].push(poster as PosterFile);
				}
			});
		}
		// Title Cards by Show and Season
		if (posterSet.TitleCards) {
			posterSet.TitleCards.forEach((card) => {
				const showTitle = card.Show?.Title || "";
				const seasonNumber = card.Episode?.SeasonNumber ?? "Unknown Season";

				if (!titleCardsByShowAndSeason[showTitle]) {
					titleCardsByShowAndSeason[showTitle] = {};
				}
				if (!titleCardsByShowAndSeason[showTitle][seasonNumber as number]) {
					titleCardsByShowAndSeason[showTitle][seasonNumber as number] = [];
				}
				titleCardsByShowAndSeason[showTitle][seasonNumber as number].push(card as PosterFile);
			});
		}
	});

	return (
		<div>
			<div className="p-2 lg:p-3">
				<div className="pb-4">
					{setType !== "boxset" && (
						<ShowFullSetDetails
							setType={setType}
							setTitle={setTitle}
							setAuthor={setAuthor}
							setID={setID}
							posterSets={posterSets}
						/>
					)}

					<Accordion
						type="multiple"
						className="w-full"
						defaultValue={["posters", "backdrops", "season-posters", "title-cards"]}
					>
						{/* All Posters */}
						{(allPosters.length > 0 || allBackdrops.length > 0) && (
							<AccordionItem value="posters">
								<AccordionTrigger>
									{(() => {
										const posterText = allPosters.length === 1 ? "Poster" : "Posters";
										const backdropText = allBackdrops.length === 1 ? "Backdrop" : "Backdrops";

										if (allBackdrops.length === 0) return posterText;
										return `${posterText} & ${backdropText}`;
									})()}
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
											{allPosters.map((poster) => {
												// Determine the type from the first poster
												const isShowType = poster.Show?.ID !== undefined;

												const matchingBackdrop =
													allBackdrops.length > 0 &&
													allBackdrops.find((backdrop) => {
														if (isShowType) {
															return backdrop.Show?.ID === poster.Show?.ID;
														} else {
															return backdrop.Movie?.ID === poster.Movie?.ID;
														}
													});

												return (
													<CarouselItem key={`poster-${poster.ID}`}>
														<div className="space-y-2">
															<AssetImage
																image={poster as unknown as PosterFile}
																aspect="poster"
																className="w-full h-auto"
															/>
															{matchingBackdrop && (
																<AssetImage
																	image={matchingBackdrop as unknown as PosterFile}
																	aspect="backdrop"
																	className="w-full h-auto"
																/>
															)}
														</div>
													</CarouselItem>
												);
											})}
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
															<CarouselItem key={`season-poster-${poster.ID}`}>
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
													{Object.entries(seasons).map(([seasonNumber, cards]) => (
														<AccordionItem
															key={`${showTitle}-season-${seasonNumber}`}
															value={`${showTitle}-season-${seasonNumber}`}
														>
															<AccordionTrigger>Season {seasonNumber}</AccordionTrigger>
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
																			<CarouselItem key={`title-card-${card.ID}`}>
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
													))}
												</Accordion>
											</div>
										))}
								</AccordionContent>
							</AccordionItem>
						)}
					</Accordion>
				</div>
			</div>
		</div>
	);
};

const ShowFullSetDetails: React.FC<{
	setType: "show" | "movie" | "collection" | "boxset" | "set";
	setTitle: string;
	setAuthor: string;
	setID: string;
	posterSets: PosterSet[];
}> = ({ setType, setTitle, setAuthor, setID, posterSets }) => {
	const [backdropURL, setBackdropURL] = useState("");
	// Construct the backdrop URL
	// If the posterSet has a backdrop, use that
	// Otherwise, use the mediaItem's backdrop
	useEffect(() => {
		if (typeof window !== "undefined") {
			document.title = "aura | Poster Set";
		}

		if (posterSets.length > 0 && posterSets.some((set) => set.Backdrop)) {
			const backdropSet = posterSets.find((set) => set.Backdrop);
			if (backdropSet && backdropSet.Backdrop) {
				setBackdropURL(
					`/api/mediux/image/?assetID=${backdropSet.Backdrop.ID}&modifiedDate=${backdropSet.Backdrop.Modified}&quality=optimized`
				);
			}
		} else {
			// Get all unique RatingKeys from all assets
			const uniqueRatingKeys = new Set<string>();

			posterSets.forEach((set) => {
				// Check Poster
				if (set.Poster?.Movie?.MediaItem.RatingKey) {
					uniqueRatingKeys.add(set.Poster.Movie.MediaItem.RatingKey);
				}

				// Check Backdrop
				if (set.Backdrop?.Movie?.MediaItem.RatingKey) {
					uniqueRatingKeys.add(set.Backdrop.Movie.MediaItem.RatingKey);
				}

				// Check OtherPosters
				set.OtherPosters?.forEach((poster) => {
					if (poster.Movie?.MediaItem.RatingKey) {
						uniqueRatingKeys.add(poster.Movie.MediaItem.RatingKey);
					}
				});

				// Check OtherBackdrops
				set.OtherBackdrops?.forEach((backdrop) => {
					if (backdrop.Movie?.MediaItem.RatingKey) {
						uniqueRatingKeys.add(backdrop.Movie.MediaItem.RatingKey);
					}
				});
			});

			// Convert Set to Array and get random RatingKey
			const ratingKeysArray = Array.from(uniqueRatingKeys);
			log(
				"INFO",
				"ShowFullSetDetails",
				"Set to Array",
				"Unique RatingKeys for Backdrop Selection",
				ratingKeysArray
			);
			if (ratingKeysArray.length > 0) {
				const randomRatingKey = ratingKeysArray[Math.floor(Math.random() * ratingKeysArray.length)];
				setBackdropURL(`/api/mediaserver/image?ratingKey=${randomRatingKey}&imageType=backdrop`);
			}
		}
	}, [posterSets]);

	return (
		<>
			{/* Backdrop Background */}
			{backdropURL && <DimmedBackground backdropURL={backdropURL} />}

			{/* Title */}
			<div className="flex flex-col pt-40 justify-end items-center text-center lg:items-start lg:text-left">
				<H1
					className="mb-1"
					onClick={() => {
						log("INFO", "ShowFullSetDetails", "Return to Set Details", "Set Details Clicked", {
							setID,
							setType,
							setTitle,
							setAuthor,
						});
					}}
				>
					{setTitle}
				</H1>
			</div>

			{/* Set Author */}
			<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center text-white gap-4 tracking-wide mt-4">
				<div className="flex items-center gap-2">
					<Badge
						className="flex items-center text-sm hover:text-white transition-colors hover:brightness-120 cursor-pointer active:scale-95"
						onClick={(e) => {
							e.stopPropagation();
							window.location.href = `/user/${setAuthor}`;
						}}
					>
						Set Author: {setAuthor}
					</Badge>
					<Badge
						className="flex items-center text-sm hover:text-white transition-colors hover:brightness-120 cursor-pointer active:scale-95"
						onClick={(e) => {
							e.stopPropagation();
							window.open(`https://mediux.io/${setType}-set/${setID}`, "_blank");
						}}
					>
						View on Mediux
					</Badge>
				</div>

				{/* Season/Episode Information */}
				{setType === "show" &&
					posterSets.flatMap((set) => set.SeasonPosters || [] || set.TitleCards || []).length > 0 && (
						<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide">
							<Lead className="flex items-center text-md text-primary-dynamic">
								{posterSets.flatMap((set) => set.SeasonPosters || [] || set.TitleCards || []).length}{" "}
								{posterSets.flatMap((set) => set.SeasonPosters || [] || set.TitleCards || []).length ===
								1
									? "Season"
									: "Seasons"}
								{posterSets.flatMap((set) => set.TitleCards || []).length > 0 && (
									<>
										{" with "}
										{posterSets.flatMap((set) => set.TitleCards || []).length}{" "}
										{posterSets.flatMap((set) => set.TitleCards || []).length === 1
											? "Title Card"
											: "Title Cards"}
									</>
								)}
							</Lead>
						</div>
					)}

				{/* Movies Information 
				Get a count of total number of posters and backdrops for movies in the set
				Only display if posters or backdrops exist
				*/}
				{setType === "movie" &&
					(() => {
						// Collect all posters and backdrops (including "other" ones)
						const allPosters = posterSets.flatMap((set) => [
							...(set.Poster ? [set.Poster] : []),
							...(set.OtherPosters || []),
						]);
						const allBackdrops = posterSets.flatMap((set) => [
							...(set.Backdrop ? [set.Backdrop] : []),
							...(set.OtherBackdrops || []),
						]);
						// Use a Set to ensure uniqueness by ID
						const uniquePosters = Array.from(new Map(allPosters.map((p) => [p.ID, p])).values());
						const uniqueBackdrops = Array.from(new Map(allBackdrops.map((b) => [b.ID, b])).values());

						const posterCount = uniquePosters.length;
						const backdropCount = uniqueBackdrops.length;

						if (posterCount === 0 && backdropCount === 0) return null;

						let text = "";
						if (posterCount > 0 && backdropCount > 0) {
							text = `${posterCount} ${posterCount === 1 ? "Poster" : "Posters"} and ${backdropCount} ${backdropCount === 1 ? "Backdrop" : "Backdrops"}`;
						} else if (posterCount > 0) {
							text = `${posterCount} ${posterCount === 1 ? "Poster" : "Posters"}`;
						} else if (backdropCount > 0) {
							text = `${backdropCount} ${backdropCount === 1 ? "Backdrop" : "Backdrops"}`;
						}

						return <Lead className="flex items-center text-md text-primary-dynamic">{text}</Lead>;
					})()}

				{/* Download Button */}
				<div className="ml-auto">
					<DownloadModal
						setType={setType}
						setTitle={setTitle}
						setAuthor={setAuthor}
						setID={setID}
						posterSets={posterSets}
					/>
				</div>
			</div>
		</>
	);
};
