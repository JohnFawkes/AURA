import React from "react";
import { useNavigate } from "react-router-dom";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Typography from "@mui/material/Typography";
import { Chip, Link, Box, Grid, Divider } from "@mui/material";
import { Theme } from "@mui/material/styles";
import MediaItemPosterImage from "./MediaItemPosterImage";
import { ClientMessage } from "../types/clientMessage";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import CancelIcon from "@mui/icons-material/Cancel";

const formatDate = (dateString: string) => {
	try {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat("en-US", {
			year: "numeric",
			month: "long",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		}).format(date);
	} catch {
		return "Invalid Date";
	}
};

const SavedSetsCard: React.FC<{
	id: string;
	savedSet: ClientMessage;
}> = ({ id, savedSet }) => {
	const navigate = useNavigate();

	const handleCardClick = (id: string) => {
		console.log("Card clicked:", id);
		navigate(`/savedset/${id}`);
	};

	const chipStyles = {
		margin: "2px",
		backgroundColor: (theme: Theme) => theme.palette.background.paper,
		color: (theme: Theme) => theme.palette.text.primary,
	};

	const renderChips = () =>
		savedSet.SelectedTypes.map((type) => (
			<Chip
				key={type}
				label={
					type === "poster"
						? "Poster"
						: type === "backdrop"
						? "Backdrop"
						: type === "seasonPoster"
						? "Season Posters"
						: type === "titlecard"
						? "Title Card"
						: type
				}
				variant="outlined"
				sx={chipStyles}
			/>
		));

	return (
		<Grid container spacing={1} sx={{ marginBottom: 2 }}>
			<Grid size={{ xs: 12, sm: 12, md: 12 }}>
				<Card
					sx={{
						display: "flex",
						flexDirection: "column",
						padding: 2,
						boxShadow: 3,
						position: "relative",
						transition: "transform 0.2s",
						"&:hover": {
							transform: "scale(1.02)",
						},
						marginBottom: 2, // Add margin between cards
						width: {
							xs: 300,
							sm: 450,
						}, // Responsive width
					}}
					onClick={() => handleCardClick(id)}
				>
					{/* Auto Download Icon */}
					<Box sx={{ position: "absolute", top: 8, right: 8 }}>
						{savedSet.AutoDownload ? (
							<>
								<CheckCircleIcon color="success" />
							</>
						) : (
							<>
								Not set to autodownload
								<CancelIcon color="error" />
							</>
						)}
					</Box>

					{/* Rest of the card content */}
					<Grid container spacing={2}>
						{/* Poster Image */}
						<Grid size={{ xs: 12, sm: 12, md: 12 }}>
							<MediaItemPosterImage
								ratingKey={id}
								alt={savedSet.MediaItem.Title}
							/>
						</Grid>

						{/* Card Content */}
						<Grid size={{ xs: 12, sm: 12, md: 12 }}>
							<CardContent sx={{ padding: 0 }}>
								{/* Title */}
								<Typography variant="h6" gutterBottom>
									{savedSet.MediaItem.Title.length > 45
										? `${savedSet.MediaItem.Title.slice(
												0,
												45
										  )}...`
										: savedSet.MediaItem.Title}
								</Typography>

								{/* Year */}
								<Typography
									variant="body2"
									color="text.secondary"
								>
									{savedSet.MediaItem.Year}
								</Typography>

								{/* Last Updated */}
								<Typography
									variant="body2"
									color="text.secondary"
								>
									Last Updated:{" "}
									{formatDate(savedSet.LastUpdate || "")}
								</Typography>

								{/* Divider */}
								<Divider sx={{ marginY: 2 }} />

								{/* Selected Types */}
								<Box sx={{ display: "flex", flexWrap: "wrap" }}>
									{renderChips()}
								</Box>

								{/* Set Link */}
								<Typography sx={{ marginTop: 2 }}>
									<Link
										href={`https://mediux.pro/sets/${savedSet.Set.ID}`}
										target="_blank"
										rel="noopener noreferrer"
										sx={{
											color: (theme: Theme) =>
												theme.palette.primary.main,
											fontWeight: "bold",
										}}
									>
										View Set: {savedSet.Set.ID}
									</Link>
								</Typography>
							</CardContent>
						</Grid>
					</Grid>
				</Card>
			</Grid>
		</Grid>
	);
};

export default SavedSetsCard;
