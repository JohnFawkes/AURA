import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import {
	Accordion,
	AccordionDetails,
	AccordionSummary,
	Box,
	Button,
	Card,
	CardMedia,
	Stack,
	Typography,
} from "@mui/material";
import React, { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import CarouselWithCards from "../components/PosterSetCarousel";
import { fetchMediuxSets } from "../../../frontend/src/services/api.mediux";
import { fetchMediaServerItemContent } from "../../../frontend/src/services/api.mediaserver";
import { Guid, MediaItem } from "../../../frontend/src/types/mediaItem";
import { PosterSets } from "../../../frontend/src/types/posterSets";
import ErrorMessage from "../components/ErrorMessage";

const MediaItemPage: React.FC = () => {
	const location = useLocation();
	const navigate = useNavigate();

	const { pathname } = location;
	const pathParts = pathname.split("/");
	const ratingKey = pathParts[pathParts.length - 2];

	const [mediaItem, setMediaItem] = React.useState<MediaItem | null>(null);
	const [posterSrc, setPosterSrc] = React.useState<string | null>(null);
	const [backdropSrc, setBackdropSrc] = React.useState<string | null>(null);
	const [imdbLink, setImdbLink] = React.useState<string>("");
	const [posterSets, setPosterSets] = React.useState<PosterSets>({
		Sets: [],
	});

	// Error Handling
	const [hasError, setHasError] = React.useState(false);
	const [errorMessage, setErrorMessage] = React.useState("");

	useEffect(() => {
		const fetchIMDBLink = (guids: Guid[]) => {
			if (!guids || guids.length === 0) {
				console.error("No GUIDs found");
				return;
			}
			const imdbGuid = guids.find((guid) => guid.Provider === "imdb");
			if (imdbGuid) {
				setImdbLink(imdbGuid.ID);
			}
		};
		const fetchPosterImage = async (ratingKey: string) => {
			const posterUrl = `/api/mediaserver/image/${ratingKey}/poster`;
			setPosterSrc(posterUrl); // Directly set the URL
		};
		const fetchBackdropImage = async (ratingKey: string) => {
			const backdropUrl = `/api/mediaserver/image/${ratingKey}/backdrop`;
			setBackdropSrc(backdropUrl); // Directly set the URL
		};
		const fetchPosterSets = async (responseItem: MediaItem) => {
			// Check if the item has GUIDs
			try {
				if (!responseItem.Guids || responseItem.Guids.length === 0) {
					return;
				}
				const tmdbID = responseItem.Guids.find(
					(guid) => guid.Provider === "tmdb"
				)?.ID;
				if (tmdbID) {
					const resp = await fetchMediuxSets(
						tmdbID,
						responseItem.Type
					);
					if (!resp) {
						throw new Error("No response from Mediux API");
					} else if (resp.status !== "success") {
						throw new Error(resp.message);
					}
					const sets = resp.data;
					if (sets) {
						setPosterSets(sets);
					} else {
						setPosterSets({ Sets: [] });
					}
				} else {
					console.log(
						"No TMDB ID found in GUIDs, searching by external IDs"
					);
					// TODO: ADD THIS
				}
			} catch (error) {
				console.error("Error fetching poster sets:", error);
			}
		};

		const fetchAllInfo = async () => {
			try {
				const resp = await fetchMediaServerItemContent(ratingKey);
				if (!resp) {
					throw new Error("No response from Plex API");
				}
				if (resp.status !== "success") {
					throw new Error(resp.message);
				}
				const responseItem = resp.data;
				if (!responseItem) {
					throw new Error("No item found in response");
				}
				setMediaItem(responseItem);
				fetchIMDBLink(responseItem.Guids);
				fetchPosterImage(responseItem.RatingKey);
				fetchBackdropImage(responseItem.RatingKey);
				fetchPosterSets(responseItem);
			} catch (error) {
				setHasError(true);
				if (error instanceof Error) {
					console.error("Error fetching Plex item:", error.message);
					setErrorMessage(error.message);
				} else {
					console.error("Error fetching Plex item:", error);
					setErrorMessage("An unknown error occurred, check logs.");
				}
			}
		};

		fetchAllInfo();
	}, [ratingKey, navigate]);

	if (hasError) {
		document.title = "Poster-Setter - Error";
		return (
			<Box
				sx={{
					position: "relative",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					padding: 3,
					gap: 3,
					overflow: "hidden",
				}}
			>
				<ErrorMessage message={errorMessage} />
				<Button
					variant="contained"
					color="primary"
					onClick={() => navigate("/")}
					sx={{ marginTop: 2 }}
				>
					Go to Home
				</Button>

				<Button
					variant="contained"
					color="secondary"
					onClick={() => {
						navigate("/logs");
					}}
					sx={{ marginTop: 2 }}
				>
					Go To Logs
				</Button>
			</Box>
		);
	} else {
		document.title = `Poster-Setter - ${mediaItem?.Title}`;
	}

	return (
		<>
			<Box
				sx={{
					position: "relative",
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					padding: 3,
					gap: 3,
					overflow: "hidden",
				}}
			>
				{/* Backdrop Image */}
				<Box
					sx={{
						position: "fixed", // Fix the backdrop to the viewport
						top: 0,
						left: 0,
						width: "100vw", // Full viewport width
						height: "100vh", // Full viewport height
						backgroundImage: `linear-gradient(to bottom, rgba(0, 0, 0, 0.8), rgba(0, 0, 0, 1)), url(${
							backdropSrc || ""
						})`,
						backgroundSize: "cover",
						backgroundPosition: "center",
						zIndex: -1, // Place it behind the content
					}}
				/>

				{/* Main Content */}
				<Card>
					<CardMedia
						component="img"
						image={posterSrc || "logo.png"} // Replace with actual image URL
						alt={mediaItem?.Title || "Poster"}
						sx={{
							height: { xs: 400, sm: 400 },
							width: "auto",
						}}
					/>
				</Card>

				<Box
					sx={{
						width: "100%",
						maxWidth: "800px",
						textAlign: "center",
						padding: 2,
						borderRadius: 2,
					}}
				>
					<Typography
						variant="h4"
						fontWeight="bold"
						gutterBottom
						color="white"
					>
						{mediaItem?.Title || "No title available."}
					</Typography>
					<Typography variant="body1" gutterBottom color="grey">
						{mediaItem?.Year || "N/A"} •{" "}
						{mediaItem?.ContentRating || "N/A"}
					</Typography>
					{mediaItem?.Type === "show" && (
						<Stack
							direction="row"
							justifyContent="center"
							spacing={2}
							mt={1}
							color="grey"
						>
							<Typography variant="body2">
								Seasons:{" "}
								{mediaItem?.Series?.SeasonCount || "N/A"}
							</Typography>
							<Typography variant="body2">
								Episodes:{" "}
								{mediaItem?.Series?.EpisodeCount || "N/A"}
							</Typography>
						</Stack>
					)}
					<Box
						display="flex"
						alignItems="center"
						justifyContent="center"
						mt={2}
					>
						<a
							href={
								imdbLink
									? `https://www.imdb.com/title/${imdbLink}`
									: "#"
							}
							target="_blank"
							rel="noopener noreferrer"
						>
							<img
								src="/imdb_logo.png"
								alt="IMDB"
								style={{ height: 20, marginRight: 8 }}
							/>
						</a>
						<Typography variant="body2" color="grey">
							{mediaItem?.AudienceRating || "N/A"}
						</Typography>
					</Box>
				</Box>

				<Accordion
					sx={{
						width: "100%",
						maxWidth: "800px",
						boxShadow: 2,
						borderRadius: 2,
						"&:before": { display: "none" },
						backgroundColor: "rgba(255, 255, 255, 0.1)",
						color: "white",
					}}
				>
					<AccordionSummary
						expandIcon={<ExpandMoreIcon />}
						aria-controls="panel1-content"
						id="summary-header"
						sx={{
							"&:hover": {
								backgroundColor: "rgba(0, 0, 0, .7)",
							},
						}}
					>
						<Typography variant="h6" fontWeight="bold">
							{mediaItem?.Type.replace(
								/\w\S*/g,
								(word) =>
									word.charAt(0).toUpperCase() +
									word.slice(1).toLowerCase()
							)}{" "}
							Details
						</Typography>
					</AccordionSummary>
					<AccordionDetails>
						<Typography variant="body1" gutterBottom>
							Summary
						</Typography>
						<Typography variant="body2" paddingBottom={2}>
							{mediaItem?.Summary || "No summary available."}
						</Typography>
						{mediaItem?.Type === "movie" && (
							<>
								<Typography variant="body1" gutterBottom>
									Path
								</Typography>
								<Typography variant="body2" paddingBottom={2}>
									{mediaItem?.Movie?.File?.Path ||
										"Path not available."}
								</Typography>
								<Typography variant="body1" gutterBottom>
									Duration
								</Typography>
								<Typography variant="body2" paddingBottom={2}>
									{mediaItem?.Movie?.File?.Duration
										? mediaItem.Movie.File.Duration <
										  3600000
											? `${Math.floor(
													mediaItem.Movie.File
														.Duration / 60000
											  )} minutes`
											: `${Math.floor(
													mediaItem.Movie.File
														.Duration / 3600000
											  )}hr ${Math.floor(
													(mediaItem.Movie.File
														.Duration %
														3600000) /
														60000
											  )}min`
										: "Duration not available."}
								</Typography>
								<Typography variant="body1" gutterBottom>
									Size
								</Typography>
								<Typography variant="body2" paddingBottom={2}>
									{mediaItem?.Movie?.File?.Size
										? mediaItem.Movie.File.Size >=
										  1024 * 1024 * 1024
											? `${(
													mediaItem.Movie.File.Size /
													(1024 * 1024 * 1024)
											  ).toFixed(2)} GB`
											: `${(
													mediaItem.Movie.File.Size /
													(1024 * 1024)
											  ).toFixed(2)} MB`
										: "Size not available."}
								</Typography>
							</>
						)}
					</AccordionDetails>
				</Accordion>
			</Box>
			{mediaItem && (
				<CarouselWithCards
					posterSets={posterSets}
					mediaItem={mediaItem}
				></CarouselWithCards>
			)}
			<Box
				sx={{
					display: "flex",
					flexDirection: "column",
					alignItems: "center",
					gap: 2, // Add spacing between the TextField and Button
					marginTop: 3, // Optional: Add some spacing from the top
				}}
			></Box>
		</>
	);
};

export default MediaItemPage;
