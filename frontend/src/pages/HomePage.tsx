import Refresh from "@mui/icons-material/Refresh";
import Box from "@mui/material/Box";
import Fab from "@mui/material/Fab";
import Grid from "@mui/material/Grid";
import Skeleton from "@mui/material/Skeleton";
import Typography from "@mui/material/Typography";
import React, { useEffect, useState } from "react";
import { fetchMediaServerLibraryItems } from "../services/api.mediaserver";
import { LibrarySection, MediaItem } from "../types/mediaItem";
import Pagination from "@mui/material/Pagination";
import debounce from "lodash.debounce"; // Install lodash if not already installed
import HomeMediaItemCard from "../components/HomeMediaItemCard";
import HomeTopSection from "../components/HomeTopSection";
import Loader from "../components/Loader";
import ErrorMessage from "../components/ErrorMessage";

const CACHE_KEY = "librarySectionsCache";
const CACHE_EXPIRY = 1000 * 60 * 60; // 1 hour in milliseconds

const Home: React.FC = () => {
	const [loading, setLoading] = useState<boolean>(true);
	const [errorLoading, setErrorLoading] = useState<boolean>(false);
	const [errorMessage, setErrorMessage] = useState<string>("");

	const [librarySections, setLibrarySections] = useState<LibrarySection[]>(
		[]
	);
	const [filteredLibraries, setFilteredLibraries] = useState<string[]>([]);
	const [filteredItems, setFilteredItems] = useState<MediaItem[]>([]);
	const [searchQuery, setSearchQuery] = useState<string>("");
	const [debouncedQuery, setDebouncedQuery] = useState<string>("");

	// state for pagination
	const [currentPage, setCurrentPage] = useState<number>(1);
	const itemsPerPage = 20; // Number of items per page

	// Calculate paginated items
	const paginatedItems = filteredItems.slice(
		(currentPage - 1) * itemsPerPage,
		currentPage * itemsPerPage
	);

	const totalPages = Math.ceil(filteredItems.length / itemsPerPage);

	// Fetch data from cache or API
	const getPlexItems = async (useCache: boolean) => {
		setLoading(true);
		setErrorLoading(false);
		try {
			if (useCache) {
				const cachedData = localStorage.getItem(CACHE_KEY);
				if (cachedData) {
					const parsedData = JSON.parse(cachedData);
					const now = new Date().getTime();
					if (now - parsedData.timestamp < CACHE_EXPIRY) {
						console.log(
							`Cache Size: ${new Blob([cachedData]).size} bytes`
						);
						setLibrarySections(parsedData.data);
						setLoading(false);
						return;
					}
				}
			}

			const resp = await fetchMediaServerLibraryItems();
			if (resp.status !== "success") {
				throw new Error(resp.message);
			}
			const sections = resp.data;
			if (!sections || sections.length === 0) {
				throw new Error("No sections found, please check the logs.");
			}

			console.log("Sections:", sections);
			console.log(
				`Sections Size: ${
					new Blob([JSON.stringify(sections)]).size
				} bytes`
			);

			// Extract only essential fields (Title and RatingKey) for each media item
			const minimalSections = sections.map((section: LibrarySection) => ({
				...section,
				MediaItems: section.MediaItems?.map((item: MediaItem) => ({
					Title: item.Title,
					RatingKey: item.RatingKey,
					Year: item.Year,
					LibraryTitle: section.Title,
					Type: item.Type,
					Guids: item.Guids,
				})),
			}));

			const dataToStore = JSON.stringify({
				data: minimalSections,
				timestamp: new Date().getTime(),
			});
			console.log(`Saved Size: ${new Blob([dataToStore]).size} bytes`);
			localStorage.setItem(CACHE_KEY, dataToStore);

			setLibrarySections(minimalSections);
		} catch (error) {
			setErrorLoading(true);
			console.log("Error fetching data:", error);
			setErrorMessage(
				error instanceof Error
					? error.message
					: "An unknown error occurred"
			);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		getPlexItems(true);
	}, []);

	const handleSearch = (searchValue: string) => {
		setSearchQuery(searchValue); // Update the immediate search query
		debounce((query: string) => {
			setDebouncedQuery(query); // Update the debounced query
			setCurrentPage(1); // Reset to the first page on new search
		}, 300)(searchValue); // Trigger the debounced function
	};

	// Update filteredItems whenever librarySections, filteredLibraries, or debouncedQuery changes
	useEffect(() => {
		let items = librarySections.flatMap((section) =>
			(section.MediaItems || []).map((item) => ({
				...item,
				LibraryTitle: section.Title, // Add library name to each item
			}))
		);

		// Filter by selected libraries
		if (filteredLibraries.length > 0) {
			items = items.filter((item) =>
				filteredLibraries.includes(item.LibraryTitle)
			);
		}

		// Filter by debounced search query
		if (debouncedQuery.trim() !== "") {
			const query = debouncedQuery.trim().toLowerCase();
			items = items.filter((item) =>
				item.Title.toLowerCase().includes(query)
			);
		}

		setFilteredItems(items);
	}, [librarySections, filteredLibraries, debouncedQuery]);

	if (loading) {
		return (
			<Box
				sx={{
					display: "flex",
					flexDirection: "column",
					justifyContent: "center",
					alignItems: "center",
					height: "100vh",
					backgroundColor: "background.default",
					color: "text.primary",
				}}
			>
				<Loader
					loadingText={`Loading all content from ${
						import.meta.env.VITE_MEDIA_SERVER_TYPE ||
						"your media server"
					}...`}
				/>
			</Box>
		);
	}

	const handlePageChange = (_: React.ChangeEvent<unknown>, value: number) => {
		setCurrentPage(value);
	};

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
				<ErrorMessage message={errorMessage} />
			) : !loading &&
			  librarySections.length === 0 &&
			  searchQuery === "" ? (
				<Grid
					container
					spacing={2}
					sx={{
						justifyContent: "center", // Centers the grid items horizontally
					}}
				>
					<ErrorMessage
						message={`No items found in ${
							import.meta.env.VITE_MEDIA_SERVER_TYPE ||
							"Media Server"
						}!`}
					/>
				</Grid>
			) : filteredItems.length === 0 && searchQuery !== "" ? (
				<>
					<HomeTopSection
						searchQuery={searchQuery}
						librarySections={librarySections}
						filteredLibraries={filteredLibraries}
						onSearchChange={handleSearch}
						onLibraryChange={setFilteredLibraries}
					/>
					<Grid
						container
						spacing={2}
						sx={{
							justifyContent: "center", // Centers the grid items horizontally
						}}
					>
						<Typography variant="h6" color="error">
							No items found matching '{searchQuery}' in{" "}
							{filteredLibraries.length > 0
								? filteredLibraries.join(", ")
								: "any library"}
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
					<HomeTopSection
						searchQuery={searchQuery}
						librarySections={librarySections}
						filteredLibraries={filteredLibraries}
						onSearchChange={handleSearch}
						onLibraryChange={setFilteredLibraries}
					/>
					<Grid
						container
						spacing={2}
						sx={{
							justifyContent: "center", // Centers the grid items horizontally
						}}
						paddingTop={2}
					>
						{paginatedItems.map((item) => (
							<Grid
								key={item.RatingKey}
								sx={{
									display: "flex",
								}}
							>
								<HomeMediaItemCard
									ratingKey={item.RatingKey}
									title={item.Title}
									year={item.Year}
									libraryTitle={item.LibraryTitle}
								/>
							</Grid>
						))}
					</Grid>

					<Box
						sx={{
							marginTop: 2,
							display: "flex",
							justifyContent: "center", // Center the pagination
						}}
					>
						<Pagination
							count={totalPages}
							page={currentPage}
							onChange={handlePageChange}
							variant="outlined"
							color="primary"
							sx={{
								"& .MuiPagination-ul": {
									flexWrap: "nowrap", // Prevent wrapping
								},
								"& .MuiPaginationItem-root": {
									minWidth: "28px", // Reduce button width
									height: "28px", // Reduce button height
									fontSize: "12px", // Reduce font size
									"@media (max-width: 600px)": {
										minWidth: "24px", // Further reduce for very small screens
										height: "24px",
										fontSize: "10px",
									},
								},
							}}
						/>
					</Box>

					{/* Floating Action Button */}
					<Fab
						color="primary"
						aria-label="refresh"
						sx={{
							position: "fixed",
							bottom: 8,
							right: 8,
							width: 56, // Default size
							height: 56, // Default size
							"@media (max-width: 600px)": {
								width: 40, // Smaller size for mobile
								height: 40, // Smaller size for mobile
							},
						}}
						onClick={() => getPlexItems(false)}
					>
						<Refresh />
					</Fab>
				</>
			)}
		</Box>
	);
};

export default Home;
