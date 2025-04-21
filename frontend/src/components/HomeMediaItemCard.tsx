import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import React from "react";
import { useNavigate } from "react-router-dom";
import MediaItemPosterImage from "./MediaItemPosterImage";

const HomeMediaItemCard: React.FC<{
	ratingKey: string;
	title: string;
	year: number;
	libraryTitle: string;
}> = ({ ratingKey, title, year, libraryTitle }) => {
	const navigate = useNavigate();

	const handleCardClick = (ratingKey: string, title: string) => {
		// Replace space with underscore for URL compatibility
		const formattedTitle = title.replace(/\s+/g, "_");
		navigate(`/media/${ratingKey}/${formattedTitle}`);
	};

	return (
		<Card
			sx={{
				display: "flex",
				flexDirection: "row",
				alignItems: "center",
				padding: 1,
				cursor: "pointer",
				height: 150, // Uniform card height
				width: {
					xs: 300,
					sm: 350,
				}, // Responsive width
			}}
			onClick={() => handleCardClick(ratingKey, title)}
		>
			<MediaItemPosterImage ratingKey={ratingKey} alt={title} />
			<CardContent
				sx={{
					textAlign: "left",
				}}
			>
				<Typography variant="h6">
					{title.length > 45 ? `${title.slice(0, 45)}...` : title}
				</Typography>
				<Typography variant="body2" color="text.secondary">
					{year}
				</Typography>
				<Typography variant="body2" color="text.secondary">
					Library: {libraryTitle}
				</Typography>
			</CardContent>
		</Card>
	);
};

export default HomeMediaItemCard;
