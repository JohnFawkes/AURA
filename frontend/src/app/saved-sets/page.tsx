"use client";

import { ReturnErrorMessage } from "@/services/api-error-return";
import { fetchAllItemFromDBWithFilters } from "@/services/database/api-db-get-all";
import { AutodownloadResult, postForceRecheckDBItemForAutoDownload } from "@/services/database/api-db-items-recheck";
import {
	ArrowDownAZ,
	ArrowDownZA,
	CalendarArrowDown,
	Check,
	ChevronDown,
	ClockArrowDown,
	ClockArrowUp,
	RefreshCcw as RefreshIcon,
	XCircle,
} from "lucide-react";
import { toast } from "sonner";

import React, { useCallback, useEffect, useRef, useState } from "react";

import { CustomPagination } from "@/components/shared/custom-pagination";
import { ErrorMessage } from "@/components/shared/error-message";
import Loader from "@/components/shared/loader";
import { RefreshButton } from "@/components/shared/refresh-button";
import SavedSetsCard from "@/components/shared/saved-sets-cards";
import SavedSetsTableRow from "@/components/shared/saved-sets-table";
import { SelectItemsPerPage } from "@/components/shared/select_items_per_page";
import { SortControl } from "@/components/shared/select_sort";
import { ViewControl } from "@/components/shared/select_view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { cn } from "@/lib/cn";
import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useSavedSetsPageStore } from "@/lib/stores/page-store-saved-sets";

import { extractYearAndMediaItemID } from "@/hooks/search-query";

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
		filterAutoDownloadOnly,
		setFilterAutoDownloadOnly,
		filteredLibraries,
		setFilteredLibraries,
	} = useSavedSetsPageStore();

	// Get the library options from Global Library Store
	const { getSectionSummaries } = useLibrarySectionsStore();

	const [filterUserOptions, setFilterUserOptions] = useState<string[]>([]);
	const [filteredUsers, setFilteredUsers] = useState<string[]>([]);
	const [userFilterSearch, setUserFilterSearch] = useState("");
	const [totalItems, setTotalItems] = useState(0);
	const [isWideScreen, setIsWideScreen] = useState(typeof window !== "undefined" ? window.innerWidth >= 1300 : false);
	// Set sortOption to "dateUpdated" if its not title, dateUpdated, year, or library
	useEffect(() => {
		if (
			sortOption !== "title" &&
			sortOption !== "dateUpdated" &&
			sortOption !== "year" &&
			sortOption !== "library"
		) {
			setSortOption("dateUpdated");
		}
	}, [sortOption, setSortOption]);

	// Fetch saved sets with filters from store
	const fetchSavedSets = useCallback(async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;
		try {
			setLoading(true);
			// From Search Query
			// Get the mediaItemID (if any)
			// Get the mediaItemYear (if any)
			const {
				cleanedQuery,
				year: searchMediaItemYear,
				mediaItemID: searchMediaItemID,
			} = extractYearAndMediaItemID(searchQuery);

			const response = await fetchAllItemFromDBWithFilters(
				searchMediaItemID,
				cleanedQuery,
				filteredLibraries,
				searchMediaItemYear,
				filterAutoDownloadOnly,
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
		searchQuery,
		filteredLibraries,
		filterAutoDownloadOnly,
		filteredUsers,
		itemsPerPage,
		currentPage,
		sortOption,
		sortOrder,
	]);

	// Load values from store first, then fetch data
	useEffect(() => {
		if (typeof window !== "undefined") {
			document.title = "aura | Saved Sets";
		}
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

	useEffect(() => {
		// If any of the following is changed, reset to page 1
		// - searchQuery
		// - filteredLibraries
		// - filterAutoDownloadOnly
		// - filteredUsers
		// - sortOption
		// - sortOrder
		// - itemsPerPage
		setCurrentPage(1);
	}, [
		searchQuery,
		filteredLibraries,
		filterAutoDownloadOnly,
		filteredUsers,
		sortOption,
		sortOrder,
		itemsPerPage,
		setCurrentPage,
	]);

	const paginatedSets = savedSets;

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
				toast.error(response.error?.Message || "Failed to recheck item");
				return;
			}

			setRecheckStatus((prev) => ({
				...prev,
				[title]: response.data as AutodownloadResult,
			}));
		} catch (error) {
			const errorResponse = ReturnErrorMessage<unknown>(error);
			toast.error(errorResponse.error?.Message || "An unexpected error occurred");
		}
	};

	const forceRecheckAll = async () => {
		if (isFetchingRef.current) return;
		isFetchingRef.current = true;

		setRecheckStatus({}); // Reset recheck status

		// Get all saved sets that have AutoDownload enabled
		const setsToRecheck = savedSets.filter((set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload));

		log("Sets to recheck:", setsToRecheck);

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
	};

	return (
		<div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
			<div className="w-full flex items-center justify-between mb-2">
				<div className="w-full">
					<div className="w-full flex flex-col gap-4">
						<div className="flex items-center justify-between mb-2">
							<Label htmlFor="filters" className="text-lg font-semibold">
								Filters:
							</Label>
							{savedSets &&
								savedSets.some(
									(set) => set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)
								) && (
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
													(set) =>
														set.PosterSets && set.PosterSets.some((ps) => ps.AutoDownload)
												).length
											}
											)
										</span>
										<RefreshIcon className="h-3 w-3" />
									</Button>
								)}
						</div>
					</div>
					<div className="flex flex-col gap-4">
						{/* Library Filter */}
						<div className="flex flex-row" id="library-filter">
							<Label htmlFor="library-filter" className="text-md font-semibold mb-1 block">
								Library
							</Label>
							<ToggleGroup
								type="multiple"
								className="flex flex-wrap gap-2 ml-2"
								value={filteredLibraries}
								onValueChange={setFilteredLibraries}
							>
								{getSectionSummaries()
									.map((section) => section.title || "Unknown Library")
									.filter((value, index, self) => self.indexOf(value) === index)
									.map((section) => (
										<Badge
											key={section}
											className="cursor-pointer text-sm"
											variant={filteredLibraries.includes(section) ? "default" : "outline"}
											onClick={() => {
												if (filteredLibraries.includes(section)) {
													setFilteredLibraries(
														filteredLibraries.filter((lib: string) => lib !== section)
													);
													setCurrentPage(1);
												} else {
													setFilteredLibraries([...filteredLibraries, section]);
													setCurrentPage(1);
												}
											}}
										>
											{section}
										</Badge>
									))}
							</ToggleGroup>
						</div>

						{/* AutoDownload Filter */}
						<div className="flex flex-row" id="auto-download-filter">
							<Label className="text-md font-semibold mb-1 block">AutoDownload</Label>
							<Badge
								key={"filter-auto-download-only"}
								className="cursor-pointer text-sm ml-2"
								variant={filterAutoDownloadOnly ? "default" : "outline"}
								onClick={() => {
									if (setFilterAutoDownloadOnly) {
										setFilterAutoDownloadOnly(!filterAutoDownloadOnly);
									}
								}}
							>
								{filterAutoDownloadOnly ? "AutoDownload Only" : "All Items"}
							</Badge>
						</div>

						{/* User Filter */}
						{filterUserOptions.length > 0 && (
							<div className="flex flex-row" id="user-filter">
								<Label className="text-md font-semibold mb-1 block">User</Label>
								<Popover>
									<PopoverTrigger asChild>
										<Button
											variant="outline"
											className="min-w-[180px] max-w-[220px] justify-between border-muted-foreground/30 ml-2"
										>
											<span className="truncate">
												{filteredUsers.length === 0
													? "All Users"
													: filteredUsers.length === 1
														? filteredUsers[0]
														: `${filteredUsers.length} users selected`}
											</span>
											<ChevronDown className="ml-2 h-4 w-4 shrink-0" />
										</Button>
									</PopoverTrigger>
									<PopoverContent className="w-72 p-2 shadow-lg border-muted-foreground/20">
										<Input
											type="search"
											placeholder="Search users..."
											className="mb-2"
											value={userFilterSearch || ""}
											onChange={(e) => setUserFilterSearch(e.target.value)}
											tabIndex={-1} // Prevents auto-focus when popover opens
											autoFocus={false} // Auto-focus when popover opens
										/>
										<div className="flex flex-col gap-1 max-h-64 overflow-y-auto">
											<div
												className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
													filteredUsers.length === 0 ? "bg-muted" : "hover:bg-muted/60"
												}`}
												onClick={() => {
													setFilteredUsers([]);
													setCurrentPage(1);
												}}
											>
												<Checkbox checked={filteredUsers.length === 0} id={`users-all`} />
												<Label
													htmlFor={`users-all`}
													className="text-sm flex-1 cursor-pointer truncate"
													onClick={(e) => e.stopPropagation()}
												>
													All User
												</Label>
												{filteredUsers.length === 0 && (
													<Check className="h-4 w-4 text-primary" />
												)}
											</div>
											<div className="border-b my-1" />
											{filterUserOptions
												.filter(
													(user) =>
														!userFilterSearch ||
														user.toLowerCase().includes(userFilterSearch.toLowerCase())
												)
												.sort((a, b) => a.localeCompare(b))
												.map((user) => (
													<div
														key={user}
														className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
															filteredUsers.includes(user)
																? "bg-muted"
																: "hover:bg-muted/60"
														}`}
														onClick={() => {
															let newUsers: string[];
															if (filteredUsers.includes(user)) {
																newUsers = filteredUsers.filter((u) => u !== user);
															} else {
																newUsers = [...filteredUsers, user];
															}
															setFilteredUsers(newUsers);
															setCurrentPage(1);
														}}
													>
														<Checkbox
															checked={filteredUsers.includes(user)}
															id={`user-${user}`}
														/>
														<Label
															htmlFor={`user-${user}`}
															className="text-sm flex-1 cursor-pointer truncate"
															onClick={(e) => e.stopPropagation()}
														>
															{user}
														</Label>
														{filteredUsers.includes(user) && (
															<Check className="h-4 w-4 text-primary" />
														)}
													</div>
												))}
										</div>
									</PopoverContent>
								</Popover>
							</div>
						)}
					</div>
				</div>
			</div>

			<Separator className="my-4 w-full" />

			{/* Sorting controls */}
			<div className="w-full flex items-center justify-between mb-4">
				{viewOption === "card" && (
					<SortControl
						options={[
							{
								value: "dateUpdated",
								label: "Date Updated",
								ascIcon: <ClockArrowUp />,
								descIcon: <ClockArrowDown />,
							},
							{ value: "title", label: "Title", ascIcon: <ArrowDownAZ />, descIcon: <ArrowDownZA /> },
							{
								value: "year",
								label: "Year",
								ascIcon: <CalendarArrowDown />,
								descIcon: <CalendarArrowDown />,
							},
							{
								value: "library",
								label: "Library",
								ascIcon: <ArrowDownAZ />,
								descIcon: <ArrowDownZA />,
							},
						]}
						sortOption={sortOption}
						sortOrder={sortOrder}
						setSortOption={(value) => {
							setSortOption(value as "title" | "dateUpdated" | "year" | "library" | "");
							if (value === "title") setSortOrder("asc");
							else if (value === "dateUpdated") setSortOrder("desc");
							else if (value === "year") setSortOrder("desc");
							else if (value === "library") setSortOrder("asc");
						}}
						setSortOrder={setSortOrder}
					/>
				)}

				{/* View Control - only show this if not on mobile */}
				<div className="hidden sm:flex w-full items-center justify-end">
					<ViewControl
						options={[
							{ value: "card", label: "Card View" },
							{ value: "table", label: "Table View" },
						]}
						viewOption={viewOption}
						setViewOption={(value) => setViewOption(value as "card" | "table")}
						label="View:"
						showLabel={true}
					/>
				</div>
			</div>

			{/* Items Per Page Selection */}
			{
				// Only show the items per page selection if there are more than 10 sets
				totalItems > 10 && (
					<div className="w-full flex items-center mb-2">
						<SelectItemsPerPage
							setCurrentPage={setCurrentPage}
							itemsPerPage={itemsPerPage}
							setItemsPerPage={setItemsPerPage}
						/>
					</div>
				)
			}

			{Object.keys(recheckStatus).length > 0 && (
				<div className="w-full mb-4">
					<h3 className="text-lg font-semibold mb-2">Recheck Status:</h3>
					<div className="rounded-md border">
						<table className="w-full divide-y divide-border">
							<thead className="bg-muted/50">
								<tr>
									<th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground w-[200px]">
										Title
									</th>
									<th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground w-[80px]">
										Status
									</th>
									<th className="px-3 py-2 text-left text-sm font-medium text-muted-foreground">
										Details
									</th>
									<th className="px-3 py-2 text-right">
										<Button
											variant="ghost"
											size="icon"
											onClick={() => setRecheckStatus({})}
											className="h-8 w-8 cursor-pointer text-muted-foreground hover:text-red-500"
										>
											<XCircle className="h-4 w-4" />
										</Button>
									</th>
								</tr>
							</thead>
							<tbody className="divide-y divide-border">
								{Object.entries(recheckStatus)
									// Sort entries by MediaItemTitle
									.sort(([, a], [, b]) => a.MediaItemTitle.localeCompare(b.MediaItemTitle))
									.map(([title, result]) => (
										<tr key={title}>
											<td className="px-3 py-2 text-sm">{result.MediaItemTitle}</td>
											<td className="px-3 py-2">
												<Badge
													className={cn(
														"inline-flex items-center rounded-full px-2 py-0.5 text-sm font-medium",
														{
															"bg-green-100 text-green-700":
																result.OverAllResult === "Success",
															"bg-yellow-100 text-yellow-700":
																result.OverAllResult === "Warn",
															"bg-red-100 text-red-700": result.OverAllResult === "Error",
															"bg-gray-100 text-gray-700":
																result.OverAllResult === "Skipped",
														}
													)}
												>
													{result.OverAllResult}
												</Badge>
											</td>
											<td className="px-3 py-2">
												<div className="space-y-1">
													<p className="text-md text-muted-foreground">
														{result.OverAllResultMessage}
													</p>
													{result.Sets.map((set, index) => (
														<div
															key={`${set.PosterSetID}-${index}`}
															className="text-xs text-muted-foreground pl-4"
														>
															• Set {set.PosterSetID}: {set.Result} - {set.Reason}
														</div>
													))}
												</div>
											</td>
											<td className="px-3 py-2">
												<RefreshIcon
													className="h-4 w-4 cursor-pointer text-primary-dynamic hover:text-primary"
													onClick={async () => {
														const item = savedSets.find(
															(set) => set.MediaItem.Title === title
														);
														if (!item) return;
														// Remove the item from the recheck status list
														setRecheckStatus((prev) => {
															const newStatus = {
																...prev,
															};
															delete newStatus[title];
															return newStatus;
														});
														await handleRecheckItem(title, item);
													}}
												/>
											</td>
										</tr>
									))}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{/* If there are no saved sets, show a message */}
			{(!savedSets || savedSets.length === 0) && !loading && !error && !Object.keys(recheckStatus).length && (
				<div className="w-full">
					<ErrorMessage
						error={ReturnErrorMessage<string>(`No items found ${searchQuery ? `matching "${searchQuery}"` : ""} in 
                ${filteredLibraries.length > 0 ? filteredLibraries.join(", ") : "any library"} 
                ${filterAutoDownloadOnly ? "that are set to AutoDownload" : ""}
				${filteredUsers.length > 0 ? ` for user${filteredUsers.length > 1 ? "s" : ""} ${filteredUsers.join(", ")}` : ""}.`)}
					/>
					<div className="text-center text-muted-foreground mt-4">
						<Button
							variant="outline"
							size="sm"
							onClick={() => {
								setSearchQuery("");
								setFilteredLibraries([]);
								if (setFilterAutoDownloadOnly) setFilterAutoDownloadOnly(false);
								setFilteredUsers([]);
								setCurrentPage(1);
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
			{viewOption === "table" && paginatedSets && paginatedSets.length > 0 && (
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
									if (sortOption === "dateUpdated") {
										setSortOrder(sortOrder === "asc" ? "desc" : "asc");
									} else {
										setSortOption("dateUpdated");
										setSortOrder("desc");
									}
								}}
							>
								<span className="inline-flex items-center gap-1">
									Last Updated
									<span className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 flex items-center">
										{sortOption === "dateUpdated" ? (
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
						{paginatedSets.length > 0 &&
							paginatedSets.map((savedSet) => (
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
					{paginatedSets.length > 0 &&
						paginatedSets.map((savedSet) => (
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
