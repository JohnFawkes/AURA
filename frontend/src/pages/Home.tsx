import React, { useEffect, useRef, useState } from "react";
import {
	Box,
	TextField,
	Typography,
	Card,
	CardContent,
	Grid,
	CircularProgress,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { fetchPlexSections } from "../services/api.plex";
import { LibrarySection, MediaItem } from "../types/mediaItem";
import InputAdornment from "@mui/material/InputAdornment";
import { MovieRounded } from "@mui/icons-material";

import { Skeleton } from "@mui/material";
import PlexPosterImage from "../components/PlexPosterImage";

const Home: React.FC = () => {
	const [librarySections, setLibrarySections] = useState<LibrarySection[]>(
		[]
	);
	const [filteredSections, setFilteredSections] = useState<MediaItem[]>([]);
	const [loading, setLoading] = useState<boolean>(true);
	const [searchTerm, setSearchTerm] = useState<string>("");
	const hasFetched = useRef(false);
	const navigate = useNavigate();
	const [errorLoading, setErrorLoading] = useState<boolean>(false);
	const [errorMessage, setErrorMessage] = useState<string>("");

	useEffect(() => {
		if (hasFetched.current) return;
		hasFetched.current = true;

		const getPlexItems = async () => {
			try {
				const resp = await fetchPlexSections();
				if (resp.status !== "success") {
					throw new Error(resp.message);
				}
				const sections = resp.data;
				if (!sections || sections === null || sections.length === 0) {
					throw new Error(
						"No sections found, please check the logs."
					);
				}
				setLibrarySections(sections);
				setFilteredSections(
					sections.flatMap((section) =>
						(section.MediaItems || []).map((item) => ({
							...item,
							LibraryTitle: section.Title, // Add library name to each item
						}))
					)
				);
			} catch (error) {
				setErrorLoading(true);
				if (error instanceof Error) {
					setErrorMessage(error.message);
				} else {
					setErrorMessage("An unknown error occurred");
				}
			} finally {
				setLoading(false);
			}
		};
		getPlexItems();
	}, []);

	const handleSearch = (searchValue: string) => {
		const searchValueTrimmed = searchValue.trim().toLowerCase();
		if (searchValueTrimmed === "") {
			setFilteredSections(
				librarySections.flatMap((section) =>
					(section.MediaItems || []).map((item) => ({
						...item,
						LibraryTitle: section.Title,
					}))
				)
			);
		} else {
			const searchWords = searchValueTrimmed.split(" ");
			const filtered = librarySections.flatMap((section) =>
				(section.MediaItems || [])
					.filter((item) =>
						searchWords.every((word) =>
							item.Title.toLowerCase().includes(word)
						)
					)
					.map((item) => ({
						...item,
						LibraryTitle: section.Title,
					}))
			);
			setFilteredSections(filtered);
		}
	};

	const handleCardClick = (item: MediaItem) => {
		navigate("/plex", { state: { item } });
	};

	if (loading) {
		return (
			<Box
				sx={{
					display: "flex",
					flexDirection: "column", // Stack the text and spinner vertically
					justifyContent: "center",
					alignItems: "center",
					height: "100vh",
					backgroundColor: "background.default", // Use theme background color
					color: "text.primary", // Use theme text color
				}}
			>
				<Typography
					variant="h6"
					sx={{
						marginBottom: 2, // Add spacing between text and spinner
						fontWeight: 500,
					}}
				>
					Loading all content from Plex...
				</Typography>
				<CircularProgress size={50} thickness={4} />
			</Box>
		);
	}

	return (
		<Box
			sx={{
				width: "90%", // Takes up 80% of the viewport width
				margin: "0 auto", // Centers the container horizontally
				padding: 2, // Adds padding inside the container
				display: "flex",
				flexDirection: "column",
				alignItems: "center", // Centers the grid horizontally
				minHeight: "100vh", // Ensures the container spans the full height of the viewport
			}}
		>
			{errorLoading ? (
				<Typography variant="h6" color="error">
					{errorMessage}
				</Typography>
			) : filteredSections.length === 0 && searchTerm === "" ? (
				<Grid
					container
					spacing={2}
					sx={{
						justifyContent: "center", // Centers the grid items horizontally
					}}
				>
					<Typography variant="h6" color="error">
						No items found in your Plex library sections
					</Typography>
				</Grid>
			) : filteredSections.length === 0 && searchTerm !== "" ? (
				<>
					<TextField
						id="input-with-icon-textfield"
						label="Search Media Items"
						placeholder="Media Title"
						fullWidth
						slotProps={{
							input: {
								startAdornment: (
									<InputAdornment position="start">
										<MovieRounded />
									</InputAdornment>
								),
							},
						}}
						variant="outlined"
						onChange={(e) => {
							setSearchTerm(e.target.value);
							handleSearch(e.target.value);
						}}
					/>
					<Grid
						container
						spacing={2}
						sx={{
							justifyContent: "center", // Centers the grid items horizontally
						}}
					>
						<Typography variant="h6" color="error">
							No items found matching '{searchTerm}'
						</Typography>
						{[...Array(1)].map((_, index) => (
							<Grid size={{ xs: 12, sm: 12, md: 12 }} key={index}>
								<Skeleton
									variant="rectangular"
									width="100%"
									height={150}
								/>
							</Grid>
						))}
					</Grid>
				</>
			) : (
				<>
					<TextField
						id="input-with-icon-textfield"
						label="Search Media Items"
						placeholder="Media Title"
						fullWidth
						slotProps={{
							input: {
								startAdornment: (
									<InputAdornment position="start">
										<MovieRounded />
									</InputAdornment>
								),
							},
						}}
						variant="outlined"
						onChange={(e) => {
							setSearchTerm(e.target.value);
							handleSearch(e.target.value);
						}}
					/>
					<Grid
						container
						spacing={2}
						sx={{
							justifyContent: "center", // Centers the grid items horizontally
						}}
						paddingTop={2}
					>
						{filteredSections.map((item) => (
							<Grid
								key={item.RatingKey}
								sx={{
									display: "flex",
								}}
							>
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
									onClick={() => handleCardClick(item)}
								>
									<PlexPosterImage
										ratingKey={item.RatingKey}
										alt={item.Title}
									/>
									<CardContent
										sx={{
											textAlign: "left",
										}}
									>
										<Typography variant="h6">
											{item.Title}
										</Typography>
										<Typography
											variant="body2"
											color="text.secondary"
										>
											{item.Year}
										</Typography>
										<Typography
											variant="body2"
											color="text.secondary"
										>
											Library: {item.LibraryTitle}
										</Typography>
									</CardContent>
								</Card>
							</Grid>
						))}
					</Grid>
				</>
			)}
		</Box>
	);
};

export default Home;
