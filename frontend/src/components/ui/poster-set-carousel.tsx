import { PosterFile } from "@/types/posterSets";
import Image from "next/image";
import {
	Carousel,
	CarouselItem,
	CarouselContent,
	CarouselNext,
	CarouselPrevious,
} from "@/components/ui/carousel";

interface PosterSetCarouselProps {
	posterFiles: PosterFile[]; // Define the type for posterFiles
}

const PosterSetCarousel: React.FC<PosterSetCarouselProps> = ({
	posterFiles,
}) => {
	// Carousel One will contain Poster and SeasonPoster
	// Carousel Two will contain Backdrop and Titlecard
	const carouselOneFiles = posterFiles.filter(
		(file) => file.Type === "poster" || file.Type === "seasonPoster"
	);
	const carouselTwoFiles = posterFiles.filter(
		(file) => file.Type === "backdrop" || file.Type === "titlecard"
	);

	// Sort Carousel One files in the following order:
	// Poster then SeasonPoster
	carouselOneFiles.sort((a, b) => {
		const order = {
			poster: 1,
			seasonPoster: 2,
		};

		return (
			(order[a.Type as keyof typeof order] || 0) -
			(order[b.Type as keyof typeof order] || 0)
		);
	});

	// Sort Carousel Two files in the following order:
	// Backdrop then Titlecard
	carouselTwoFiles.sort((a, b) => {
		const order = {
			backdrop: 1,
			titlecard: 2,
		};

		return (
			(order[a.Type as keyof typeof order] || 0) -
			(order[b.Type as keyof typeof order] || 0)
		);
	});

	return (
		<div className="divide-y divide-primary-dynamic/20 space-y-6">
			{/* Carousel One */}
		</div>
	);
};

export default PosterSetCarousel;
