import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";

import {
	Carousel,
	CarouselItem,
	CarouselContent,
	CarouselNext,
	CarouselPrevious,
} from "@/components/ui/carousel";
import { cn } from "@/lib/utils";
import Image from "next/image";

const PosterSetCard: React.FC<{
	posterSet: PosterSet;
	mediaItem: MediaItem;
}> = ({ posterSet, mediaItem }) => {
	// Helper function to get the count and label for each type
	const getFileTypeCount = (type: string, label: string) => {
		const count = posterSet.Files.filter(
			(file) => file.Type === type
		).length;
		return count > 0 ? `${count} ${label}${count > 1 ? "s" : ""}` : null;
	};

	// Get counts for each type
	const posterCount = getFileTypeCount("poster", "Poster");
	const backdropCount = getFileTypeCount("backdrop", "Backdrop");
	const seasonPosterCount = getFileTypeCount("seasonPoster", "Season Poster");
	const titlecardCount = getFileTypeCount("titlecard", "Titlecard");

	// Combine counts into a single string
	const fileCounts = [
		posterCount,
		backdropCount,
		seasonPosterCount,
		titlecardCount,
	]
		.filter(Boolean) // Remove null values
		.join(" • ");

	return (
		<>
			{/* <div
				style={{
					display: "flex",
					justifyContent: "space-between",
					alignItems: "center",
					width: "100%",
					padding: "20px",
					boxSizing: "border-box",
				}}
			>
				<div>
					<h1>Set by: {posterSet.User.Name}</h1>
					<p className="text-sm text-gray-500">{fileCounts}</p>
				</div>
				<div>
					<PosterSetModal
						posterSet={posterSet}
						mediaItem={mediaItem}
					/>
				</div>
			</div>
			<div>
				 <PosterSetCarousel posterFiles={posterSet.Files} /> 
		</div > */}

			<Carousel
				opts={{
					align: "start",
					dragFree: true,
					slidesToScroll: "auto",
				}}
				className="w-full"
			>
				<div className="flex flex-col">
					<div className="flex items-center gap-1 justify-between mb-2">
						<span className="text-primary-dynamic">
							Set by: {posterSet.User.Name}
							<p className="text-sm text-gray-500">
								{fileCounts}
							</p>
						</span>
					</div>
				</div>

				<CarouselContent>
					{posterSet.Files.map((file) => (
						<CarouselItem key={`${posterSet.ID}-${file.ID}`}>
							<div
								className={cn(
									"relative overflow-hidden rounded-md border border-primary-dynamic/40 hover:border-primary-dynamic transition-all duration-300 group"
								)}
								style={{ height: "300px", width: "100%" }} // Set height and width for the parent container
							>
								<Image
									src={`/api/mediux/image/${file.ID}?modifiedDate=${file.Modified}`}
									alt={file.ID}
									fill
									quality={70}
									sizes={"300px"}
									className={cn(
										"transition-opacity duration-500"
									)}
									priority={false}
									loading="lazy"
									unoptimized
								/>
							</div>
						</CarouselItem>
					))}
				</CarouselContent>
				<CarouselNext className="right-2 bottom-0" />
				<CarouselPrevious className="right-8 bottom-0" />
			</Carousel>
		</>
	);
};

export default PosterSetCard;
