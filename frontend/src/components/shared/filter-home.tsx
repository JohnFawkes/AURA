"use client";

import { ArrowDownAZ, ArrowDownZA, ClockArrowDown, ClockArrowUp, Filter, SortDescIcon } from "lucide-react";

import { useEffect, useMemo, useState } from "react";

import { SelectItemsPerPage } from "@/components/shared/select-items-per-page";
import { SortControl } from "@/components/shared/select-sort";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Drawer,
	DrawerContent,
	DrawerDescription,
	DrawerHeader,
	DrawerTitle,
	DrawerTrigger,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { cn } from "@/lib/cn";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

import { extractInfoFromSearchQuery } from "@/hooks/search-query";

import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";
import { FILTER_IN_DB_OPTIONS, TYPE_FILTER_IN_DB_OPTIONS, TYPE_ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

type HomeFilterProps = {
	// Filtering
	librarySections: LibrarySection[];
	filteredLibraries: string[];
	setFilteredLibraries: (libs: string[]) => void;
	filterInDB: string;
	setFilterInDB: (filter: TYPE_FILTER_IN_DB_OPTIONS) => void;

	// Sorting
	hasUpdatedAt: boolean;
	sortOption: string;
	setSortOption: (option: string) => void;
	sortOrder: "asc" | "desc";
	setSortOrder: (order: "asc" | "desc") => void;

	// Items Per Page
	setCurrentPage: (page: number) => void;
	itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS;
	setItemsPerPage: (num: TYPE_ITEMS_PER_PAGE_OPTIONS) => void;
};

function FilterHomeContent({
	librarySections,
	filteredLibraries,
	setFilteredLibraries,
	filterInDB,
	setFilterInDB,

	hasUpdatedAt,
	sortOption,
	setSortOption,
	sortOrder,
	setSortOrder,

	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: HomeFilterProps) {
	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const { searchTMDBID, searchLibrary, searchYear, searchTitle } = extractInfoFromSearchQuery(searchQuery);

	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4">
			{/* Sort Header */}
			<div className="flex items-center justify-center mb-0 mt-0">
				<Label className="text-lg font-semibold">Sort</Label>
			</div>
			<Separator className="my-2 w-full" />
			{/* Sort Control */}
			<SortControl
				options={[
					{
						value: "dateAdded",
						label: "Date Added",
						ascIcon: <ClockArrowUp />,
						descIcon: <ClockArrowDown />,
						type: "date",
					},
					...(hasUpdatedAt
						? [
								{
									value: "dateUpdated",
									label: "Date Updated",
									ascIcon: <ClockArrowUp />,
									descIcon: <ClockArrowDown />,
									type: "date" as const,
								},
							]
						: []),
					{
						value: "dateReleased",
						label: "Date Released",
						ascIcon: <ClockArrowUp />,
						descIcon: <ClockArrowDown />,
						type: "date",
					},
					{
						value: "title",
						label: "Title",
						ascIcon: <ArrowDownAZ />,
						descIcon: <ArrowDownZA />,
						type: "string",
					},
				]}
				sortOption={sortOption}
				sortOrder={sortOrder}
				setSortOption={(value) => {
					setSortOption(value as "title" | "dateUpdated" | "dateAdded" | "dateReleased");
					if (value === "title") setSortOrder("asc");
					else if (value === "dateUpdated") setSortOrder("desc");
					else if (value === "dateAdded") setSortOrder("desc");
					else if (value === "dateReleased") setSortOrder("desc");
				}}
				setSortOrder={setSortOrder}
			/>
			{/* Items Per Page Selection */}
			<div className="flex items-center mb-4">
				<SelectItemsPerPage
					setCurrentPage={setCurrentPage}
					itemsPerPage={itemsPerPage}
					setItemsPerPage={setItemsPerPage}
				/>
			</div>

			<Separator className="my-2 w-full" />
			<Separator className="my-1 w-full" />

			{/* Filters Header */}
			<div className="flex items-center justify-center mb-0">
				<Label className="text-lg font-semibold">Filters</Label>
			</div>
			<Separator className="my-2 w-full" />

			{/* Search Info */}
			{(searchTitle || searchYear || searchTMDBID || searchLibrary) && (
				<div className="p-2 bg-secondary rounded-md">
					<Label className="text-md font-semibold mb-1 block">Current Search</Label>
					<div className="flex flex-col gap-1">
						{searchTitle && (
							<div className="text-sm">
								<span className="font-semibold">Search:</span> {searchTitle}
							</div>
						)}
						{typeof searchYear === "number" && searchYear > 0 && (
							<div className="text-sm">
								<span className="font-semibold">Year:</span> {searchYear}
							</div>
						)}
						{searchTMDBID && (
							<div className="text-sm">
								<span className="font-semibold">ID:</span> {searchTMDBID}
							</div>
						)}
						{searchLibrary && (
							<div className="text-sm">
								<span className="font-semibold">Library:</span> {searchLibrary}
							</div>
						)}
					</div>
					<Button
						variant={"destructive"}
						className="mt-2"
						onClick={() => {
							setSearchQuery("");
						}}
					>
						Clear Search
					</Button>
				</div>
			)}
			{/* Library Sections Filter */}
			<div className="flex flex-col">
				<>
					<Label className="text-md font-semibold mb-1">Library Sections</Label>
					<ToggleGroup
						type="multiple"
						className="flex flex-wrap gap-2 ml-2"
						value={filteredLibraries}
						onValueChange={setFilteredLibraries}
					>
						{librarySections.map((section) => (
							<Badge
								key={section.Title}
								className="cursor-pointer text-sm active:scale-95 hover:brightness-120"
								variant={filteredLibraries.includes(section.Title) ? "default" : "outline"}
								onClick={() => {
									if (filteredLibraries.includes(section.Title)) {
										setFilteredLibraries(filteredLibraries.filter((lib) => lib !== section.Title));
									} else {
										setFilteredLibraries([...filteredLibraries, section.Title]);
									}
								}}
							>
								{section.Title}
							</Badge>
						))}
					</ToggleGroup>
					<Separator className="my-4 w-full" />
				</>
				{/* In-Database Filter */}
				<div className="flex flex-col">
					<Label className="text-md font-semibold mb-1">In-Database Filter</Label>
					<ToggleGroup
						type="single"
						className="flex flex-wrap gap-2 ml-2"
						value={filterInDB}
						onValueChange={(value) => {
							if (value) {
								setFilterInDB(value as TYPE_FILTER_IN_DB_OPTIONS);
							}
						}}
					>
						{FILTER_IN_DB_OPTIONS.map((option) => (
							<Badge
								key={option}
								className={cn(
									"cursor-pointer text-sm active:scale-95 hover:brightness-120",
									filterInDB === option && option === "all" && "bg-primary text-primary-foreground",
									filterInDB === option &&
										option === "inDB" &&
										"bg-green-500 text-primary-foreground",
									filterInDB === option &&
										option === "notInDB" &&
										"bg-red-500 text-primary-foreground"
								)}
								variant={filterInDB === option ? "default" : "outline"}
								onClick={() => {
									if (filterInDB === option) {
										setFilterInDB("all");
									} else {
										setFilterInDB(option as TYPE_FILTER_IN_DB_OPTIONS);
									}
								}}
							>
								{option === "all" ? "All Items" : option === "inDB" ? "In Database" : "Not In Database"}
							</Badge>
						))}
					</ToggleGroup>
				</div>
			</div>
		</div>
	);
}

export function FilterHome({
	librarySections,
	filteredLibraries,
	setFilteredLibraries,
	filterInDB,
	setFilterInDB,

	hasUpdatedAt,
	sortOption,
	setSortOption,
	sortOrder,
	setSortOrder,

	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: HomeFilterProps) {
	const [isWideScreen, setIsWideScreen] = useState(false);

	// Change isWideScreen on window resize
	useEffect(() => {
		const handleResize = () => {
			setIsWideScreen(window.innerWidth >= 1300);
		};
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, []);

	// Calculate number of active filters
	const numberOfActiveFilters = useMemo(() => {
		let count = 0;
		if (filteredLibraries.length > 0) count++;
		if (filterInDB !== "all") count++;
		return count;
	}, [filteredLibraries, filterInDB]);

	return (
		<>
			{isWideScreen ? (
				<Popover>
					<PopoverTrigger asChild>
						<Button
							variant="outline"
							className={cn(numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1")}
						>
							<SortDescIcon className="h-5 w-5" />
							Sort & Filter {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
							<Filter className="h-5 w-5" />
						</Button>
					</PopoverTrigger>
					<PopoverContent
						side="right"
						align="start"
						className="w-[350px] p-2 bg-background border border-primary"
					>
						<FilterHomeContent
							librarySections={librarySections}
							filteredLibraries={filteredLibraries}
							setFilteredLibraries={setFilteredLibraries}
							filterInDB={filterInDB}
							setFilterInDB={setFilterInDB}
							hasUpdatedAt={hasUpdatedAt}
							sortOption={sortOption}
							setSortOption={setSortOption}
							sortOrder={sortOrder}
							setSortOrder={setSortOrder}
							setCurrentPage={setCurrentPage}
							itemsPerPage={itemsPerPage}
							setItemsPerPage={setItemsPerPage}
						/>
					</PopoverContent>
				</Popover>
			) : (
				<Drawer direction="left">
					<DrawerTrigger asChild>
						<Button
							variant="outline"
							className={cn(numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1")}
						>
							<SortDescIcon className="h-5 w-5" />
							Sort & Filter {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
							<Filter className="h-5 w-5" />
						</Button>
					</DrawerTrigger>
					<DrawerContent>
						<DrawerHeader className="my-0">
							<DrawerTitle className="mb-0">Sort & Filter</DrawerTitle>
							<DrawerDescription className="mb-0">
								Use the options below to sort and filter your media items.
							</DrawerDescription>
						</DrawerHeader>
						<Separator className="my-1 w-full" />
						<FilterHomeContent
							librarySections={librarySections}
							filteredLibraries={filteredLibraries}
							setFilteredLibraries={setFilteredLibraries}
							filterInDB={filterInDB}
							setFilterInDB={setFilterInDB}
							hasUpdatedAt={hasUpdatedAt}
							sortOption={sortOption}
							setSortOption={setSortOption}
							sortOrder={sortOrder}
							setSortOrder={setSortOrder}
							setCurrentPage={setCurrentPage}
							itemsPerPage={itemsPerPage}
							setItemsPerPage={setItemsPerPage}
						/>
					</DrawerContent>
				</Drawer>
			)}
		</>
	);
}
