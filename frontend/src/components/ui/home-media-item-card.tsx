import React from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import Image from "next/image";
import { Badge } from "@/components/ui/badge";

interface HomeMediaItemCardProps {
	ratingKey: string;
	title: string;
	year: number;
	libraryTitle: string;
}

const HomeMediaItemCard: React.FC<HomeMediaItemCardProps> = ({
	ratingKey,
	title,
	year,
	libraryTitle,
}) => {
	const router = useRouter();

	const handleCardClick = (ratingKey: string, title: string) => {
		// Replace space with underscore for URL compatibility
		const formattedTitle = title.replace(/\s+/g, "_");
		router.push(`/media/${ratingKey}/${formattedTitle}`);
	};

	return (
		<Card
			className="items-center cursor-pointer hover:shadow-xl transition-shadow"
			style={{
				backgroundColor: "var(--card)",
			}}
			onClick={() => handleCardClick(ratingKey, title)}
		>
			{/* Poster Image */}
			<div className="relative w-[150px] h-[220px] rounded-md overflow-hidden transform transition-transform duration-300 hover:scale-105">
				<Image
					src={`/api/mediaserver/image/${ratingKey}/poster`}
					alt={title}
					fill
					className="object-cover"
					loading="lazy"
					unoptimized
				/>
			</div>

			<CardHeader className="text-center text-xl font-semibold flex justify-center ">
				{title.length > 45 ? `${title.slice(0, 45)}...` : title}
			</CardHeader>

			{/* Badges */}
			<CardContent className="flex justify-center gap-2">
				<Badge variant="default" className=" text-xs">
					{year}
				</Badge>
				<Badge variant="default" className="text-xs">
					{libraryTitle}
				</Badge>
			</CardContent>
		</Card>
	);
};

export default HomeMediaItemCard;
