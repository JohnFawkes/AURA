import { Download } from "@mui/icons-material";
import {
	Box,
	Button,
	Card,
	CardContent,
	CardMedia,
	Checkbox,
	Dialog,
	DialogActions,
	DialogContent,
	DialogTitle,
	FormControlLabel,
	FormGroup,
	FormHelperText,
	IconButton,
	Skeleton,
	Typography,
	useMediaQuery,
	useTheme,
} from "@mui/material";
import LinearProgress from "@mui/material/LinearProgress";
import React, { useEffect, useState } from "react";
import { useInView } from "react-intersection-observer";
import Slider, { Settings } from "react-slick"; // Import the Settings type
import "slick-carousel/slick/slick-theme.css";
import "slick-carousel/slick/slick.css";
import { fetchMediuxImageData } from "../services/api.mediux";
import { postSendSetToAPI } from "../services/api.mediaserver";
import { MediaItem } from "../types/mediaItem";
import { PosterSet, PosterSets } from "../types/posterSets";

const imageCache = new Map<string, string>(); // Cache to store loaded image URLs

const PosterSetCarousel: React.FC<{
	posterSets: PosterSets;
	mediaItem: MediaItem;
}> = ({ posterSets, mediaItem }) => {
	const theme = useTheme();
	const isSmallScreen = useMediaQuery(theme.breakpoints.down("sm"));
	const isMediumScreen = useMediaQuery(theme.breakpoints.between("sm", "md"));
	const isLargeScreen = useMediaQuery(theme.breakpoints.between("md", "lg"));
	const isExtraLargeScreen = useMediaQuery(theme.breakpoints.up("lg"));

	// Adjust slidesToShow based on screen size
	const slidesToShow = isSmallScreen
		? 1
		: isMediumScreen
		? 3
		: isLargeScreen
		? 4
		: isExtraLargeScreen
		? 5
		: 2;

	const settings = {
		arrows: true,
		autoplay: false,
		autoplaySpeed: 2000,
		dots: false,
		draggable: true,
		initialSlide: 0,
		infinite: false,
		lazyLoad: "progressive" as const,
		speed: 500,
		slidesToShow,
		slidesToScroll: 1,
	};

	// State for modal
	const [openModal, setOpenModal] = useState(false);
	const [selectedSet, setSelectedSet] = useState<PosterSet | null>(null);
	const [modalErrorMessage, setModalErrorMessage] = useState("");

	// Tracking selected checkboxes for what to download
	const [selectedCheckboxes, setSelectedCheckboxes] = useState<string[]>([]);
	const handleCheckboxChange = (type: string) => {
		setSelectedCheckboxes((prev) =>
			prev.includes(type)
				? prev.filter((item) => item !== type)
				: [...prev, type]
		);
	};
	const [autoDownload, setAutoDownload] = useState(false);
	const handleAutoDownloadChange = () => {
		setAutoDownload((prev) => !prev);
	};

	// Download Progress
	const [progressValue, setProgressValue] = useState(0);
	const [progressColor, setProgressColor] = useState<
		| "primary"
		| "inherit"
		| "success"
		| "error"
		| "secondary"
		| "info"
		| "warning"
	>("primary");
	const [progressText, setProgressText] = useState("");
	const [progressNextStep, setProgressNextStep] = useState("");

	const handleSaveIconClick = (set: PosterSet) => {
		// Open the modal and set the selected set
		console.log("Save clicked for set:", set);
		setSelectedSet(set);
		setOpenModal(true);
		setModalErrorMessage("");

		// Set all file types as selected
		const allFileTypes = Array.from(
			new Set(set.Files.map((file) => file.Type))
		);
		setSelectedCheckboxes(allFileTypes);
	};

	const handleModalClose = () => {
		setOpenModal(false);
		setSelectedSet(null);
		setModalErrorMessage("");
		setSelectedCheckboxes([]);
		setProgressValue(0);
		setProgressColor("primary");
		setProgressText("");
		setProgressNextStep("");
		console.log("Modal closed");
	};

	const handleDownloadButtonClick = async (set: PosterSet) => {
		if (selectedCheckboxes.length === 0) {
			setModalErrorMessage("Please select at least one option.");
			return;
		} else {
			setModalErrorMessage("");
		}

		try {
			await postSendSetToAPI({
				Set: set,
				SelectedTypes: selectedCheckboxes,
				Plex: mediaItem,
				AutoDownload: autoDownload,
			}).then((response) => {
				// If the response was not successful, show an error message
				if (!response || response.status !== "success") {
					setProgressColor("error");
					setProgressText(
						"Failed to start the task. Please try again."
					);
					return;
				}

				// Create a new SSE connection to the backend server
				const eventSource = new EventSource(
					`/api/plex/update/set/${mediaItem.RatingKey}`
				);

				eventSource.onmessage = (event) => {
					// Parse the incoming data
					const data = JSON.parse(event.data);

					// Update progress bar and text
					if (data.response.status === "success") {
						setProgressValue(data.progress.value || 0);
						setProgressText(data.progress.text || "");
						setProgressNextStep(data.progress.nextStep || "");
					}

					// Close the connection when the task is complete
					if (data.response.status === "complete") {
						setProgressValue(100);
						setProgressText("Task complete");
						setProgressNextStep("");
						eventSource.close();
					}
				};

				eventSource.onerror = () => {
					setProgressColor("error");
					setProgressText(
						"An error occurred while processing the task."
					);
					eventSource.close();
				};
			});
		} catch {
			setProgressColor("error");
			setProgressText("Failed to start the task. Please try again.");
		}
	};

	return (
		<Box padding={1}>
			{(posterSets.Sets ?? []).length > 0 ? (
				posterSets.Sets?.map((set) => (
					<Card
						key={set.ID}
						sx={{
							marginBottom: 4,
							backgroundColor: "rgba(255, 255, 255, 0.05)", // 90% transparent
							boxShadow: "none",
						}}
					>
						<CardContent>
							<Box
								display="flex"
								justifyContent="space-between"
								alignItems="center"
							>
								<Typography
									variant="h6"
									gutterBottom
									color="white"
								>
									<Box
										component="span"
										sx={{
											fontSize: "0.875rem",
											color: "grey",
										}}
									>
										Set By: {set.User.Name}
									</Box>
								</Typography>
								<Typography
									variant="body2"
									gutterBottom
									color="white"
								>
									{set.Files && (
										<Box component="span">
											{set.Files.filter(
												(file) =>
													file.Type === "poster" ||
													file.Type === "seasonPoster"
											).length > 0 &&
												`${
													set.Files.filter(
														(file) =>
															file.Type ===
																"poster" ||
															file.Type ===
																"seasonPoster"
													).length
												} Posters`}
											{set.Files.filter(
												(file) =>
													file.Type === "backdrop"
											).length > 0 &&
												` • ${
													set.Files.filter(
														(file) =>
															file.Type ===
															"backdrop"
													).length
												} Backdrops`}
											{set.Files.filter(
												(file) =>
													file.Type === "titlecard"
											).length > 0 &&
												` • ${
													set.Files.filter(
														(file) =>
															file.Type ===
															"titlecard"
													).length
												} Titlecards`}
										</Box>
									)}
								</Typography>
								<IconButton
									aria-label="download"
									size="large"
									sx={{ color: "white" }}
									onClick={() => handleSaveIconClick(set)}
								>
									<Download />
								</IconButton>
							</Box>
							<Box>
								<LazyCarousel
									files={set.Files}
									types={["poster", "seasonPoster"]}
									settings={settings}
								/>

								<LazyCarousel
									files={set.Files}
									types={["backdrop", "titlecard"]}
									settings={settings}
								/>
							</Box>
						</CardContent>
					</Card>
				))
			) : (
				<Typography variant="h6" color="white" textAlign="center">
					No poster sets or collections found for {mediaItem.Title}
				</Typography>
			)}

			{/* Modal */}
			<Dialog
				open={openModal}
				onClose={handleModalClose}
				maxWidth="sm"
				fullWidth
			>
				<DialogTitle>Download Poster Set</DialogTitle>
				<DialogContent>
					{/* Display Item Name, Set ID, and Set Author */}
					<Box mb={2}>
						<Typography variant="body1" gutterBottom>
							<strong>Name:</strong> {posterSets.Item?.Title}
						</Typography>
						<Typography variant="body1" gutterBottom>
							<strong>Set ID:</strong>{" "}
							<a
								href={`https://mediux.pro/sets/${selectedSet?.ID}`}
								target="_blank"
								rel="noopener noreferrer"
							>
								{selectedSet?.ID}
							</a>{" "}
						</Typography>
						<Typography variant="body1" gutterBottom>
							<strong>Author:</strong> {selectedSet?.User.Name}
						</Typography>
					</Box>

					{/* Checkboxes for selecting download options */}
					<Typography variant="body2" gutterBottom>
						Select what you would like to download for the set:
					</Typography>
					<FormGroup>
						{selectedSet?.Files.some(
							(file) => file.Type === "poster"
						) && (
							<FormControlLabel
								control={
									<Checkbox
										checked={selectedCheckboxes.includes(
											"poster"
										)}
										onChange={() =>
											handleCheckboxChange("poster")
										}
									/>
								}
								disabled={
									!selectedSet?.Files.some(
										(file) => file.Type === "poster"
									)
								}
								label="Posters"
							/>
						)}
						{selectedSet?.Files.some(
							(file) => file.Type === "seasonPoster"
						) && (
							<FormControlLabel
								control={
									<Checkbox
										checked={selectedCheckboxes.includes(
											"seasonPoster"
										)}
										onChange={() =>
											handleCheckboxChange("seasonPoster")
										}
									/>
								}
								disabled={
									!selectedSet?.Files.some(
										(file) => file.Type === "seasonPoster"
									)
								}
								label="Season Posters"
							/>
						)}
						{selectedSet?.Files.some(
							(file) => file.Type === "backdrop"
						) && (
							<FormControlLabel
								control={
									<Checkbox
										checked={selectedCheckboxes.includes(
											"backdrop"
										)}
										onChange={() =>
											handleCheckboxChange("backdrop")
										}
									/>
								}
								disabled={
									!selectedSet?.Files.some(
										(file) => file.Type === "backdrop"
									)
								}
								label="Backdrops"
							/>
						)}
						{selectedSet?.Files.some(
							(file) => file.Type === "titlecard"
						) && (
							<FormControlLabel
								control={
									<Checkbox
										checked={selectedCheckboxes.includes(
											"titlecard"
										)}
										onChange={() =>
											handleCheckboxChange("titlecard")
										}
									/>
								}
								disabled={
									!selectedSet?.Files.some(
										(file) => file.Type === "titlecard"
									)
								}
								label="Titlecards"
							/>
						)}
						{modalErrorMessage && (
							<FormHelperText error>
								{modalErrorMessage}
							</FormHelperText>
						)}
					</FormGroup>

					<FormGroup>
						<FormControlLabel
							control={
								<Checkbox
									checked={autoDownload}
									onChange={() => handleAutoDownloadChange()}
								/>
							}
							label="Automatically download updates"
						/>
						<FormHelperText>
							This will automatically download any new files added
							to this set.
						</FormHelperText>
					</FormGroup>

					{progressValue > 0 && (
						<Box mb={2}>
							<Box sx={{ display: "flex", alignItems: "center" }}>
								<Box sx={{ width: "100%", mr: 1 }}>
									{/* Pass the value explicitly to LinearProgress */}
									<LinearProgress
										color={progressColor}
										variant="determinate"
										value={progressValue}
									/>
								</Box>
								<Box sx={{ minWidth: 35 }}>
									<Typography
										variant="body2"
										color="text.secondary"
									>
										{`${Math.round(progressValue)}%`}
									</Typography>
								</Box>
							</Box>
							<Typography variant="body2" color="text.secondary">
								{progressText}
							</Typography>
							<Typography variant="body2" color="text.secondary">
								{progressNextStep}
							</Typography>
						</Box>
					)}
				</DialogContent>
				<DialogActions>
					<Button
						onClick={handleModalClose}
						color="secondary"
						variant="outlined"
					>
						Cancel
					</Button>
					<Button
						onClick={() =>
							selectedSet &&
							handleDownloadButtonClick(selectedSet)
						}
						color="primary"
						variant="contained"
					>
						Download
					</Button>
				</DialogActions>
			</Dialog>
		</Box>
	);
};

interface LazyCarouselProps {
	files: { ID: string; Type: string; Modified: string }[];
	types: string[]; // Types to filter (e.g., "poster", "backdrop", etc.)
	settings: Settings; // Slider settings
}

const LazyCarousel: React.FC<LazyCarouselProps> = ({
	files = [],
	types,
	settings,
}) => {
	const [imageSources, setImageSources] = useState<Map<string, string>>(
		new Map()
	);
	const { ref, inView } = useInView({ triggerOnce: true });

	useEffect(() => {
		if (inView) {
			const loadImages = async () => {
				const filteredFiles = files.filter((file) =>
					types.includes(file.Type)
				);
				const newImageSources = new Map<string, string>();

				await Promise.all(
					filteredFiles.map(async (file) => {
						if (!imageCache.has(file.ID)) {
							try {
								const imageUrl = await fetchMediuxImageData(
									file.ID,
									file.Modified
								);
								imageCache.set(file.ID, imageUrl);
								newImageSources.set(file.ID, imageUrl);
							} catch (error) {
								console.error(
									`Error fetching image for ${file.ID}:`,
									error
								);
							}
						} else {
							newImageSources.set(
								file.ID,
								imageCache.get(file.ID)!
							);
						}
					})
				);

				setImageSources(
					(prev) => new Map([...prev, ...newImageSources])
				);
			};

			loadImages();
		}
	}, [inView, files, types]);

	const filteredFiles = files.filter((file) => types.includes(file.Type));

	if (filteredFiles.length === 0) return null;

	return (
		<Box ref={ref}>
			<Slider {...settings}>
				{filteredFiles.map((file) => (
					<Box key={file.ID}>
						{imageSources.has(file.ID) ? (
							<CardMedia
								component="img"
								image={imageSources.get(file.ID) || ""}
								alt={file.ID}
								sx={{
									height: "auto",
									width: {
										xs: 300,
										sm: 300,
										md: 300,
									},
								}}
							/>
						) : (
							<Skeleton
								sx={{ bgcolor: "white" }}
								animation="wave"
								variant="rectangular"
								width={300}
								height={"auto"}
							/>
						)}
					</Box>
				))}
			</Slider>
		</Box>
	);
};

export default PosterSetCarousel;
