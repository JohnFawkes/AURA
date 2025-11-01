"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchAllItemFromDBWithFilters } from "@/services/database/api-db-get-all";
import { AutodownloadResult, postForceRecheckDBItemForAutoDownload } from "@/services/database/api-db-items-recheck";
import {
	ArrowDownAZ,
	ArrowDownZA,
	ClockArrowDown,
	ClockArrowUp,
	RefreshCcw as RefreshIcon,
	XCircle,
} from "lucide-react";
import { toast } from "sonner";

import React, { useCallback, useEffect, useRef, useState } from "react";

import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import { FilterSavedSets } from "@/components/shared/filter-saved-sets";
import Loader from "@/components/shared/loader";
import { RefreshButton } from "@/components/shared/refresh-button";
import SavedSetsCard from "@/components/shared/saved-sets-cards";
import SavedSetsTableRow from "@/components/shared/saved-sets-table";
import { ViewControl } from "@/components/shared/select-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useSavedSetsPageStore } from "@/lib/stores/page-store-saved-sets";

import { extractInfoFromSearchQuery } from "@/hooks/search-query";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

const SavedSetsPage: React.FC = () => {
	const [savedSets, setSavedSets] = useState<DBMediaItemWithPosterSets[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<APIResponse<unknown> | null>(null);
	const isFetchingRef = useRef(false);
	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const [recheckStatus, setRecheckStatus] = useState<Record<string, AutodownloadResult>>({});

	const {
		currentPage,
		setCurrentPage,
		itemsPerPage,
		setItemsPerPage,
		sortOption,
		setSortOption,
		sortOrder,
		setSortOrder,
		viewOption,
		setViewOption,
		filteredLibraries,
		setFilteredLibraries,
		filterAutoDownload,
		setFilterAutoDownload,
		filteredUsers,
		setFilteredUsers,
		filteredTypes,
		setFilteredTypes,
		filterMultiSetOnly,
		setFilterMultiSetOnly,
	} = useSavedSetsPageStore();

	// Get the library options from Global Library Store
	const { getSectionSummaries } = useLibrarySectionsStore();
	const librarySectionsLoaded = useLibrarySectionsStore((state) => state.hasHydrated);

	// User Filter Options: all unique users from the savedSets
	const [filterUserOptions, setFilterUserOptions] = useState<string[]>([]);

	// Selected Types Options
	const typeOptions = [
		{ label: "Poster", value: "poster" },
		{ label: "Backdrop", value: "backdrop" },
		{ label: "Season Posters", value: "seasonPoster" },
		{ label: "Title Cards", value: "titlecard" },
		{ label: "No Selected Types", value: "none" },
	]; // Total Items: Total items matching filters in DB (for pagination)
	const [totalItems, setTotalItems] = useState(0);
	// Is Wide Screen: for showing/hiding the ViewControl
	const [isWideScreen, setIsWideScreen] = useState(false);

	// Set the Document Title
	useEffect(() => {
		document.title = `aura | Saved Sets`;
	}, []);

	// Set sortOption to "dateDownloaded" if its not title, dateDownloaded, year, or library
	useEffect(() => {
		if (
			sortOption !== "title" &&
			sortOption !== "dateDownloaded" &&
			sortOption !== "year" &&
			sortOption !== "library"
		) {
			setSortOption("dateDownloaded");
			setSortOrder("desc");
		}
	}, [sortOption, setSortOption, setSortOrder]);

	const { searchTMDBID, searchLibrary, searchYear, searchTitle } = extractInfoFromSearchQuery(searchQuery);

	// Fetch saved sets with filters from store
	const fetchSavedSets = useCallback(async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;
		try {
			setLoading(true);

			const response = await fetchAllItemFromDBWithFilters(
				searchTMDBID,
				searchLibrary,
				searchYear,
				searchTitle,
				filteredLibraries,
				filteredTypes,
				filterAutoDownload,
				filterMultiSetOnly,
				filteredUsers,
				itemsPerPage,
				currentPage,
				sortOption,
				sortOrder
			);

			if (response.status === "error") {
				setError(response);
				setSavedSets([]);
				setTotalItems(0);
				setFilterUserOptions([]);
				return;
			}

			if (!response.data) {
				setError(ReturnErrorMessage<Error>("No saved sets found in the database"));
				setSavedSets([]);
				setTotalItems(0);
				setFilterUserOptions([]);
				return;
			}

			setSavedSets(response.data.items);
			setTotalItems(response.data.total_items || 0);
			setFilterUserOptions(response.data.unique_users || []);

			setError(null);
		} catch (error) {
			setError(ReturnErrorMessage<unknown>(error));
			setSavedSets([]);
			setTotalItems(0);
			setFilterUserOptions([]);
		} finally {
			setLoading(false);
			isFetchingRef.current = false;
		}
	}, [
		searchTMDBID,
		searchLibrary,
		searchYear,
		searchTitle,
		filteredLibraries,
		filteredTypes,
		filterAutoDownload,
		filterMultiSetOnly,
		filteredUsers,
		itemsPerPage,
		currentPage,
		sortOption,
		sortOrder,
	]);

	// Load values from store first, then fetch data
	useEffect(() => {
		// Fetch after store values are loaded
		fetchSavedSets();
	}, [fetchSavedSets]);

	// Change to Card View if on mobile
	useEffect(() => {
		const handleResize = () => {
			if (window.innerWidth < 1300) {
				setViewOption("card");
				setIsWideScreen(false);
			} else {
				setIsWideScreen(true);
			}
		};
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, [setViewOption]);

	// Calculate total pages
	const totalPages = Math.ceil(totalItems / itemsPerPage);

	if (loading) {
		return <Loader className="mt-10" message="Loading saved sets..." />;
	}

	if (error) {
		return (
			<div className="flex flex-col items-center p-6 gap-4">
				<ErrorMessage error={error} />
			</div>
		);
	}

	const handleRecheckItem = async (title: string, item: DBMediaItemWithPosterSets): Promise<void> => {
		try {
			const response = await postForceRecheckDBItemForAutoDownload(item);

			if (response.status === "error") {
				toast.error(response.error?.message || "Failed to recheck item");
				return;
			}

			setRecheckStatus((prev) => ({
				...prev,
				[title]: response.data as AutodownloadResult,
			}));
		} catch (error) {
			const errorResponse = ReturnErrorMessage<unknown>(error);
			toast.error(errorResponse.error?.message || "An unexpected error occurred");
		}
		fetchSavedSets();
	};

	const forceRecheckAll = async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;

		setRecheckStatus({}); // Reset recheck status

		// Get all saved sets that have AutoDownload enabled
		const setsToRecheck = savedSets.filter((set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload));

		log("INFO", "Saved Sets Page", "Force Recheck", `Forcing recheck for ${setsToRecheck.length} sets`, {
			setsToRecheck,
		});

		if (setsToRecheck.length === 0) {
			toast.warning("No sets with AutoDownload enabled found", {
				id: "force-recheck",
				duration: 2000,
			});
		}

		// Show loading toast
		toast.loading(`Rechecking ${setsToRecheck.length} sets...`, {
			id: "force-recheck",
			duration: 0, // Keep it open until we manually close it
		});

		for (const [index, set] of setsToRecheck.entries()) {
			toast.loading(`Rechecking ${index + 1} of ${setsToRecheck.length} - ${set.MediaItem.Title}`, {
				id: "force-recheck",
				duration: 0, // Keep it open until we manually close it
			});
			await handleRecheckItem(set.MediaItem.Title, set);
		}

		// Close loading toast
		toast.success("Recheck completed", {
			id: "force-recheck",
			duration: 2000,
		});

		isFetchingRef.current = false;
		fetchSavedSets();
	};

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			{/* Row 1: Filter Icon (left), Autodownload (right) */}
			<div className="w-full flex items-center justify-between mb-2">
				<div>
					<FilterSavedSets
						getSectionSummaries={getSectionSummaries}
						librarySectionsLoaded={librarySectionsLoaded}
						filteredLibraries={filteredLibraries}
						setFilteredLibraries={setFilteredLibraries}
						typeOptions={typeOptions}
						filteredTypes={filteredTypes}
						setFilteredTypes={setFilteredTypes}
						filterAutoDownload={filterAutoDownload}
						setFilterAutoDownload={setFilterAutoDownload}
						filterUserOptions={filterUserOptions}
						filteredUsers={filteredUsers}
						setFilteredUsers={setFilteredUsers}
						filterMultiSetOnly={filterMultiSetOnly}
						setFilterMultiSetOnly={setFilterMultiSetOnly}
						searchTitle={searchTitle}
						searchYear={searchYear}
						searchTMDBID={searchTMDBID}
						searchLibrary={searchLibrary}
						sortOption={sortOption}
						sortOrder={sortOrder}
						setSortOption={setSortOption}
						setSortOrder={setSortOrder}
						setCurrentPage={setCurrentPage}
						itemsPerPage={itemsPerPage}
						setItemsPerPage={setItemsPerPage}
					/>
				</div>

				<div>
					{savedSets &&
						savedSets.some((set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)) && (
							<Button
								variant="secondary"
								size="sm"
								onClick={() => forceRecheckAll()}
								className="flex items-center gap-1 text-xs sm:text-sm cursor-pointer"
							>
								<span className="hidden sm:inline">Force Autodownload Recheck</span>
								<span className="sm:hidden">Recheck</span>
								<span className="whitespace-nowrap">
									(
									{
										savedSets.filter(
											(set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)
										).length
									}
									)
								</span>
								<RefreshIcon className="h-3 w-3" />
							</Button>
						)}
				</div>
			</div>

			{/* Row 2: ViewControl (right) */}
			<div className="w-full flex items-center justify-end mb-0">
				{isWideScreen && (
					<ViewControl
						options={[
							{ value: "card", label: "Card View" },
							{ value: "table", label: "Table View" },
						]}
						viewOption={viewOption}
						setViewOption={setViewOption}
						label="View:"
						showLabel={true}
					/>
				)}
			</div>

			<Separator className="mt-0 my-4 w-full" />

			{Object.keys(recheckStatus).length > 0 && (
				<div className="w-full">
					<h3 className="text-lg font-semibold mb-2 ml-1">Recheck Status:</h3>
					<div className="grid grid-cols-1 gap-2">
						{Object.entries(recheckStatus)
							.sort(([, a], [, b]) => a.MediaItemTitle.localeCompare(b.MediaItemTitle))
							.map(([title, result]) => (
								<div key={title} className="rounded-md border p-3 flex flex-col gap-2 bg-background">
									<div className="flex justify-between items-center">
										<div className="flex items-center gap-2">
											<span className="font-semibold">{result.MediaItemTitle}</span>
										</div>
										<div className="flex items-center gap-2">
											<Badge
												className={cn(
													"inline-flex items-center rounded-full px-2 py-0.5 text-sm font-medium",
													{
														"bg-green-100 text-green-700":
															result.OverAllResult === "Success",
														"bg-yellow-100 text-yellow-700":
															result.OverAllResult === "Warn",
														"bg-red-100 text-red-700": result.OverAllResult === "Error",
														"bg-gray-100 text-gray-700": result.OverAllResult === "Skipped",
													}
												)}
											>
												{result.OverAllResult}
											</Badge>
										</div>
									</div>
									<p className="text-sm text-muted-foreground">{result.OverAllResultMessage}</p>
									{result.Sets &&
										result.Sets.map((set, index) => (
											<div
												key={`${set.PosterSetID}-${index}`}
												className="text-xs text-muted-foreground pl-2"
											>
												• Set {set.PosterSetID}: {set.Result} - {set.Reason}
											</div>
										))}
									<div className="flex justify-end gap-2 mt-2">
										<Button
											variant="ghost"
											size="icon"
											onClick={async () => {
												const item = savedSets.find((set) => set.MediaItem.Title === title);
												if (!item) return;
												setRecheckStatus((prev) => {
													const newStatus = { ...prev };
													delete newStatus[title];
													return newStatus;
												});
												await handleRecheckItem(title, item);
											}}
											className="h-6 w-6 text-muted-foreground hover:text-red-500"
										>
											<RefreshIcon className="h-4 w-4" />
										</Button>
										<Button
											variant="ghost"
											size="icon"
											onClick={() => {
												setRecheckStatus((prev) => {
													const newStatus = { ...prev };
													delete newStatus[title];
													return newStatus;
												});
											}}
											className="h-6 w-6 text-muted-foreground hover:text-red-500"
										>
											<XCircle className="h-4 w-4" />
										</Button>
									</div>
								</div>
							))}
					</div>
					<Separator className="my-4 w-full" />
				</div>
			)}

			{/* If there are no saved sets, show a message */}
			{(!savedSets || savedSets.length === 0) && !loading && !error && !Object.keys(recheckStatus).length && (
				<div className="w-full">
					<ErrorMessage
						error={ReturnErrorMessage<string>(
							[
								`No ${filterMultiSetOnly ? "Multi-Poster" : ""} Sets found`,
								filteredLibraries.length > 0
									? `in ${filteredLibraries.map((lib) => `"${lib}"`).join(", ")}`
									: null,
								filterAutoDownload === "on"
									? "that are set to Auto Download"
									: filterAutoDownload === "off"
										? "that are not set to Auto Download"
										: null,
								filteredTypes.length > 0
									? `with type${filteredTypes.length > 1 ? "s" : ""} ${filteredTypes
											.map((t) => `"${typeOptions.find((opt) => opt.value === t)?.label || t}"`)
											.join(", ")}`
									: null,
								filteredUsers.length > 0
									? `for user${filteredUsers.length > 1 ? "s" : ""} ${filteredUsers
											.map((u) => (u === "|||no-user|||" ? '"No User"' : `"${u}"`))
											.join(", ")}`
									: null,
								searchQuery ? `matching "${searchQuery}"` : null,
							]
								.filter(Boolean)
								.join("\n")
						)}
					/>
					<div className="text-center text-muted-foreground mt-4">
						<Button
							variant="outline"
							size="sm"
							onClick={() => {
								setSearchQuery("");
								setFilteredLibraries([]);
								if (setFilterAutoDownload) setFilterAutoDownload("all");
								setFilteredUsers([]);
								setFilteredTypes([]);
								setCurrentPage(1);
								if (filterMultiSetOnly) setFilterMultiSetOnly(false);
							}}
							className="text-sm"
						>
							<XCircle className="inline mr-1" />
							Clear All Filters & Search Query
						</Button>
					</div>
				</div>
			)}

			{/* Table View (only available for larger screens) */}
			{viewOption === "table" && savedSets && savedSets.length > 0 && (
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead className="w-[20px]"></TableHead>
							<TableHead
								className="w-[300px] group cursor-pointer select-none"
								onClick={() => {
									if (sortOption === "title") {
										// Toggle sort order
										setSortOrder(sortOrder === "asc" ? "desc" : "asc");
									} else {
										setSortOption("title");
										setSortOrder("asc");
									}
								}}
							>
								<span className="inline-flex items-center gap-1">
									Title
									<span className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center">
										{sortOption === "title" ? (
											sortOrder === "asc" ? (
												<ArrowDownAZ className="h-4 w-4 ml-1" />
											) : (
												<ArrowDownZA className="h-4 w-4 ml-1" />
											)
										) : (
											<ArrowDownAZ className="h-4 w-4 ml-1" />
										)}
									</span>
								</span>
							</TableHead>
							<TableHead
								className="w-[75px] group cursor-pointer select-none"
								onClick={() => {
									if (sortOption === "year") {
										// Toggle sort order
										setSortOrder(sortOrder === "asc" ? "desc" : "asc");
									} else {
										setSortOption("year");
										setSortOrder("desc");
									}
								}}
							>
								<span className="inline-flex items-center gap-1">
									Year
									<span className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center">
										{sortOption === "year" ? (
											sortOrder === "asc" ? (
												<ClockArrowUp className="h-4 w-4 ml-1" />
											) : (
												<ClockArrowDown className="h-4 w-4 ml-1" />
											)
										) : (
											<ClockArrowDown className="h-4 w-4 ml-1" />
										)}
									</span>
								</span>
							</TableHead>
							<TableHead
								className="group cursor-pointer select-none"
								onClick={() => {
									if (sortOption === "library") {
										setSortOrder(sortOrder === "asc" ? "desc" : "asc");
									} else {
										setSortOption("library");
										setSortOrder("asc");
									}
								}}
							>
								<span className="inline-flex items-center gap-1">
									Library
									<span className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center">
										{sortOption === "library" ? (
											sortOrder === "asc" ? (
												<ArrowDownAZ className="h-4 w-4 ml-1" />
											) : (
												<ArrowDownZA className="h-4 w-4 ml-1" />
											)
										) : (
											<ArrowDownAZ className="h-4 w-4 ml-1" />
										)}
									</span>
								</span>
							</TableHead>
							<TableHead
								className="w-[150px] group cursor-pointer select-none"
								onClick={() => {
									if (sortOption === "dateDownloaded") {
										setSortOrder(sortOrder === "asc" ? "desc" : "asc");
									} else {
										setSortOption("dateDownloaded");
										setSortOrder("desc");
									}
								}}
							>
								<span className="inline-flex items-center gap-1">
									Last Downloaded
									<span className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center">
										{sortOption === "dateDownloaded" ? (
											sortOrder === "asc" ? (
												<ClockArrowUp className="h-4 w-4 ml-1" />
											) : (
												<ClockArrowDown className="h-4 w-4 ml-1" />
											)
										) : (
											<ClockArrowDown className="h-4 w-4 ml-1" />
										)}
									</span>
								</span>
							</TableHead>
							<TableHead>Sets</TableHead>
							<TableHead>Types</TableHead>
							<TableHead className="text-right">Actions</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{savedSets &&
							savedSets.length > 0 &&
							savedSets.map((savedSet) => (
								<SavedSetsTableRow
									key={savedSet.MediaItem.RatingKey}
									savedSet={savedSet}
									onUpdate={fetchSavedSets}
									handleRecheckItem={handleRecheckItem}
								/>
							))}
					</TableBody>
				</Table>
			)}

			{/* Card View */}
			{viewOption === "card" && (
				<div className="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-2">
					{savedSets &&
						savedSets.length > 0 &&
						savedSets.map((savedSet) => (
							<SavedSetsCard
								key={savedSet.MediaItem.RatingKey}
								savedSet={savedSet}
								onUpdate={fetchSavedSets}
								handleRecheckItem={handleRecheckItem}
							/>
						))}
				</div>
			)}

			{/* Pagination */}
			{itemsPerPage && (
				<CustomPagination
					currentPage={currentPage}
					totalPages={totalPages}
					setCurrentPage={setCurrentPage}
					scrollToTop={true}
					filterItemsLength={totalItems}
					itemsPerPage={itemsPerPage}
				/>
			)}

			{/* Refresh Button */}
			<RefreshButton onClick={fetchSavedSets} />
		</div>
	);
};

export default SavedSetsPage;
