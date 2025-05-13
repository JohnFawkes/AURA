"use client";

import { CarouselItem } from "@/components/ui/carousel";
import { AssetImage } from "@/components/ui/asset-image";
import { PosterFile, PosterSet } from "@/types/posterSets";

interface ShowCarouselProps {
	set: PosterSet;
}

export function ShowCarousel({ set }: ShowCarouselProps) {
	return (
		<>
			{/* Show Posters with Backdrop */}

			{set.Files?.filter((file) => file.Type === "poster").map(
				(poster) => (
					<CarouselItem key={`${set.ID}-${poster.ID}`}>
						<div className="space-y-2">
							<AssetImage
								image={poster as unknown as PosterFile}
								displayUser={true}
								displayMediaType={true}
								aspect="poster"
								className="w-full"
							/>

							{set.Files?.filter(
								(file) => file.Type === "backdrop"
							)[0] && (
								<AssetImage
									image={
										set.Files.filter(
											(file) => file.Type === "backdrop"
										)[0] as unknown as PosterFile
									}
									displayUser={true}
									displayMediaType={true}
									aspect="backdrop"
									className="w-full"
								/>
							)}
						</div>
					</CarouselItem>
				)
			)}

			{/* Season Posters with Titlecards */}
			{set.Files?.filter((file) => file.Type === "seasonPoster").map(
				(poster) => {
					const matchingTitlecard = set.Files?.find(
						(file) =>
							file.Type === "titlecard" &&
							file.Episode?.SeasonNumber === poster.Season?.Number
					);

					return (
						<CarouselItem key={`${set.ID}-${poster.ID}`}>
							<div className="space-y-2">
								<AssetImage
									image={poster as unknown as PosterFile}
									displayUser={true}
									displayMediaType={true}
									aspect="poster"
									className="w-full"
								/>

								{matchingTitlecard && (
									<AssetImage
										image={
											matchingTitlecard as unknown as PosterFile
										}
										displayUser={true}
										displayMediaType={true}
										aspect="titlecard"
										className="w-full"
									/>
								)}
							</div>
						</CarouselItem>
					);
				}
			)}

			{/* Standaline Backdrops if no Posters or Season Posters */}
			{set.Files?.filter(
				(file) =>
					file.Type === "backdrop" &&
					!set.Files?.some(
						(poster) =>
							poster.Type === "poster" ||
							poster.Type === "seasonPoster"
					)
			).map((backdrop) => (
				<CarouselItem key={`${set.ID}-${backdrop.ID}`}>
					<div className="space-y-2">
						<AssetImage
							image={backdrop as unknown as PosterFile}
							displayUser={true}
							displayMediaType={true}
							aspect="backdrop"
							className="w-full"
						/>
					</div>
				</CarouselItem>
			))}

			{/* Standalone Titlecards if no Posters or Season Posters */}
			{set.Files?.filter(
				(file) =>
					file.Type === "titlecard" &&
					!set.Files?.some(
						(poster) =>
							poster.Type === "poster" ||
							poster.Type === "seasonPoster"
					)
			).map((titlecard) => (
				<CarouselItem key={`${set.ID}-${titlecard.ID}`}>
					<div className="space-y-2">
						<AssetImage
							image={titlecard as unknown as PosterFile}
							displayUser={true}
							displayMediaType={true}
							aspect="titlecard"
							className="w-full"
						/>
					</div>
				</CarouselItem>
			))}
		</>
	);
}
