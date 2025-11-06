"use client";

import { ArrowDown01, ArrowDown10, ArrowDownAZ, ArrowDownZA, Filter, SortDescIcon } from "lucide-react";

import { useMemo, useState } from "react";

import { SelectItemsPerPage } from "@/components/shared/select-items-per-page";
import { SortControl } from "@/components/shared/select-sort";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { cn } from "@/lib/cn";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

import { extractInfoFromSearchQuery } from "@/hooks/search-query";

import { TYPE_ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

type CollectionsFilterProps = {
	// Filtering
	librarySections: string[];
	filteredLibraries: string[];
	setFilteredLibraries: (libs: string[]) => void;

	// Sorting
	sortOption: string;
	setSortOption: (option: string) => void;
	sortOrder: "asc" | "desc";
	setSortOrder: (order: "asc" | "desc") => void;

	// Items Per Page
	setCurrentPage: (page: number) => void;
	itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS;
	setItemsPerPage: (num: TYPE_ITEMS_PER_PAGE_OPTIONS) => void;
};

function FilterCollectionsContent({
	librarySections = [],
	filteredLibraries,
	setFilteredLibraries,

	sortOption,
	setSortOption,
	sortOrder,
	setSortOrder,

	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: CollectionsFilterProps) {
	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const { searchLibrary, searchYear, searchTitle } = extractInfoFromSearchQuery(searchQuery);

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
						value: "title",
						label: "Title",
						ascIcon: <ArrowDownAZ />,
						descIcon: <ArrowDownZA />,
						type: "string",
					},
					{
						value: "numberOfItems",
						label: "Number of Items",
						ascIcon: <ArrowDown01 />,
						descIcon: <ArrowDown10 />,
						type: "number" as const,
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
			{(searchTitle || searchYear || searchLibrary) && (
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
								key={section}
								className="cursor-pointer text-sm active:scale-95 hover:brightness-120"
								variant={filteredLibraries.includes(section) ? "default" : "outline"}
								onClick={() => {
									if (filteredLibraries.includes(section)) {
										setFilteredLibraries(filteredLibraries.filter((lib) => lib !== section));
									} else {
										setFilteredLibraries([...filteredLibraries, section]);
									}
								}}
							>
								{section}
							</Badge>
						))}
					</ToggleGroup>
					<Separator className="my-4 w-full" />
				</>
			</div>
		</div>
	);
}

export function FilterCollections({
	librarySections,
	filteredLibraries,
	setFilteredLibraries,
	sortOption,
	setSortOption,
	sortOrder,
	setSortOrder,
	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: CollectionsFilterProps) {
	// State - Open/Close Modal
	const [modalOpen, setModalOpen] = useState(false);

	// Calculate number of active filters
	const numberOfActiveFilters = useMemo(() => {
		let count = 0;
		if (filteredLibraries.length > 0) count++;
		return count;
	}, [filteredLibraries]);

	return (
		<Dialog open={modalOpen} onOpenChange={setModalOpen}>
			<DialogTrigger asChild>
				<Button
					variant="outline"
					className={cn(numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1")}
				>
					<SortDescIcon className="h-5 w-5" />
					Sort & Filter {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
					<Filter className="h-5 w-5" />
				</Button>
			</DialogTrigger>
			<DialogContent
				className={cn("z-50", "max-h-[80vh] overflow-y-auto", "sm:max-w-[700px]", "border border-primary")}
			>
				<DialogHeader>
					<DialogTitle>Sort & Filter</DialogTitle>
					<DialogDescription>Use the options below to sort and filter your media items.</DialogDescription>
				</DialogHeader>
				<Separator className="my-1 w-full" />
				<FilterCollectionsContent
					librarySections={librarySections}
					filteredLibraries={filteredLibraries}
					setFilteredLibraries={setFilteredLibraries}
					sortOption={sortOption}
					setSortOption={setSortOption}
					sortOrder={sortOrder}
					setSortOrder={setSortOrder}
					setCurrentPage={setCurrentPage}
					itemsPerPage={itemsPerPage}
					setItemsPerPage={setItemsPerPage}
				/>
			</DialogContent>
		</Dialog>
	);
}
