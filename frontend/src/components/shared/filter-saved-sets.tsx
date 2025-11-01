import {
	ArrowDownAZ,
	ArrowDownZA,
	CalendarArrowDown,
	CalendarArrowUp,
	Check,
	ClockArrowDown,
	ClockArrowUp,
	Filter,
	SortDescIcon,
} from "lucide-react";

import { useMemo, useState } from "react";

import { SelectItemsPerPage } from "@/components/shared/select-items-per-page";
import { SortControl } from "@/components/shared/select-sort";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { cn } from "@/lib/cn";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

import { TYPE_ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

type SavedSetsFilterProps = {
	getSectionSummaries: () => { title?: string; type?: string }[];
	librarySectionsLoaded: boolean;
	filteredLibraries: string[];
	setFilteredLibraries: (libs: string[]) => void;

	typeOptions: { label: string; value: string }[];
	filteredTypes: string[];
	setFilteredTypes: (types: string[]) => void;

	filterAutoDownload: "all" | "on" | "off";
	setFilterAutoDownload: (val: "all" | "on" | "off") => void;

	filterUserOptions: string[];
	filteredUsers: string[];
	setFilteredUsers: (users: string[]) => void;

	filterMultiSetOnly: boolean;
	setFilterMultiSetOnly: (val: boolean) => void;

	searchTMDBID?: string;
	searchLibrary?: string;
	searchYear?: number;
	searchTitle?: string;

	// Sorting
	sortOption: string;
	setSortOption: (option: string) => void;
	sortOrder: "asc" | "desc";
	setSortOrder: (order: "asc" | "desc") => void;

	// Items Per Page
	setCurrentPage: (page: number) => void;
	itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS;
	setItemsPerPage: (num: TYPE_ITEMS_PER_PAGE_OPTIONS) => void;

	setModalOpen?: (open: boolean) => void;
};

function SavedSetsFilterContent({
	getSectionSummaries,
	librarySectionsLoaded,
	filteredLibraries,
	setFilteredLibraries,
	typeOptions,
	filteredTypes,
	setFilteredTypes,
	filterAutoDownload,
	setFilterAutoDownload,
	filterUserOptions,
	filteredUsers,
	setFilteredUsers,
	filterMultiSetOnly,
	setFilterMultiSetOnly,

	searchTMDBID,
	searchLibrary,
	searchYear,
	searchTitle,

	sortOption,
	setSortOption,
	sortOrder,
	setSortOrder,

	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,

	setModalOpen,
}: SavedSetsFilterProps) {
	const { setSearchQuery } = useSearchQueryStore();

	const [userFilterSearch, setUserFilterSearch] = useState<string>("");
	const [pendingFilterLibraries, setPendingFilterLibraries] = useState<string[]>(filteredLibraries);
	const [pendingFilterAutoDownload, setPendingFilterAutoDownload] = useState<"all" | "on" | "off">(
		filterAutoDownload
	);
	const [pendingFilterTypes, setPendingFilterTypes] = useState<string[]>(filteredTypes);
	const [pendingFilterMultiSetOnly, setPendingFilterMultiSetOnly] = useState<boolean>(filterMultiSetOnly);
	const [pendingFilteredUsers, setPendingFilteredUsers] = useState<string[]>(filteredUsers);

	const handleResetFilters = () => {
		setSearchQuery("");
		setFilteredLibraries([]);
		setPendingFilterLibraries([]);
		setFilteredTypes([]);
		setPendingFilterTypes([]);
		setFilterAutoDownload("all");
		setPendingFilterAutoDownload("all");
		setFilteredUsers([]);
		setPendingFilteredUsers([]);
		setFilterMultiSetOnly(false);
		setPendingFilterMultiSetOnly(false);
		setCurrentPage(1);
		setUserFilterSearch("");
		if (setModalOpen) setModalOpen(false);
	};

	const handleApplyFilters = () => {
		setCurrentPage(1);
		setFilteredLibraries(pendingFilterLibraries);
		setFilteredTypes(pendingFilterTypes);
		setFilterAutoDownload(pendingFilterAutoDownload);
		setFilterMultiSetOnly(pendingFilterMultiSetOnly);
		setFilteredUsers(pendingFilteredUsers);
		if (setModalOpen) setModalOpen(false);
	};

	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
			{/* Sort Header */}
			<div className="flex items-center justify-center mb-0 mt-0">
				<Label className="text-lg font-semibold">Sort</Label>
			</div>
			<Separator className="my-2 w-full" />
			{/* Sort Control */}
			<SortControl
				options={[
					{
						value: "dateDownloaded",
						label: "Date Downloaded",
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
					{
						value: "year",
						label: "Year",
						ascIcon: <CalendarArrowUp />,
						descIcon: <CalendarArrowDown />,
						type: "number",
					},
					{
						value: "library",
						label: "Library",
						ascIcon: <ArrowDownAZ />,
						descIcon: <ArrowDownZA />,
						type: "string",
					},
				]}
				sortOption={sortOption}
				sortOrder={sortOrder}
				setSortOption={setSortOption}
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
			{/* Library Sections */}
			<div className="flex flex-col">
				{librarySectionsLoaded && getSectionSummaries().length > 0 && (
					<>
						<Label className="text-md font-semibold mb-2 block">Library</Label>
						<ToggleGroup
							type="multiple"
							className="flex flex-wrap gap-2 ml-2"
							value={filteredLibraries}
							onValueChange={setPendingFilterLibraries}
						>
							{getSectionSummaries()
								.map((section) => section.title || "Unknown Library")
								.filter((value, index, self) => self.indexOf(value) === index)
								.map((section) => (
									<Badge
										key={section}
										className="cursor-pointer text-sm active:scale-95 hover:brightness-120"
										variant={pendingFilterLibraries.includes(section) ? "default" : "outline"}
										onClick={() => {
											if (pendingFilterLibraries.includes(section)) {
												setPendingFilterLibraries(
													pendingFilterLibraries.filter((lib) => lib !== section)
												);
											} else {
												setPendingFilterLibraries([...pendingFilterLibraries, section]);
											}
										}}
									>
										{section}
									</Badge>
								))}
						</ToggleGroup>
						<Separator className="my-4 w-full" />
					</>
				)}
				{/* Selected Types */}
				<Label className="text-md font-semibold mb-2 block">Selected Types</Label>
				<ToggleGroup
					type="multiple"
					className="flex flex-wrap gap-2 ml-2"
					value={pendingFilterTypes}
					onValueChange={setPendingFilterTypes}
				>
					{typeOptions.map((type) => (
						<Badge
							key={type.value}
							className="cursor-pointer text-sm active:scale-95 hover:brightness-120s"
							variant={pendingFilterTypes.includes(type.value) ? "default" : "outline"}
							onClick={() => {
								if (pendingFilterTypes.includes(type.value)) {
									setPendingFilterTypes(pendingFilterTypes.filter((t) => t !== type.value));
								} else {
									setPendingFilterTypes([...pendingFilterTypes, type.value]);
								}
							}}
						>
							{type.label}
						</Badge>
					))}
				</ToggleGroup>
				<Separator className="my-4 w-full" />
				{/* AutoDownload Only */}
				<Label className="text-md font-semibold mb-1 block">AutoDownload</Label>
				<ToggleGroup
					type="single"
					className="flex flex-wrap gap-2 ml-2"
					value={pendingFilterAutoDownload}
					onValueChange={(val) => setPendingFilterAutoDownload(val as "all" | "on" | "off")}
				>
					<Badge
						variant={pendingFilterAutoDownload === "all" ? "default" : "outline"}
						className={cn(
							"cursor-pointer text-sm active:scale-95 hover:brightness-120",
							pendingFilterAutoDownload === "all" ? "bg-primary text-primary-foreground" : ""
						)}
						onClick={() => setPendingFilterAutoDownload("all")}
					>
						Any
					</Badge>
					<Badge
						variant={pendingFilterAutoDownload === "on" ? "default" : "outline"}
						className={cn(
							"cursor-pointer text-sm active:scale-95 hover:brightness-120",
							pendingFilterAutoDownload === "on" ? "bg-green-500 text-primary-foreground" : ""
						)}
						onClick={() => setPendingFilterAutoDownload("on")}
					>
						AutoDownload On
					</Badge>
					<Badge
						variant={pendingFilterAutoDownload === "off" ? "default" : "outline"}
						className={cn(
							"cursor-pointer text-sm active:scale-95 hover:brightness-120",
							pendingFilterAutoDownload === "off" ? "bg-red-500 text-primary-foreground" : ""
						)}
						onClick={() => setPendingFilterAutoDownload("off")}
					>
						AutoDownload Off
					</Badge>
				</ToggleGroup>
				<Separator className="my-4 w-full" />
				{/* MultiSet Only */}
				<Label className="text-md font-semibold mb-1 block">Multi Set Only</Label>
				<Badge
					key={"filter-multi-set-only"}
					className="cursor-pointer text-sm ml-2 active:scale-95 hover:brightness-120"
					variant={pendingFilterMultiSetOnly ? "default" : "outline"}
					onClick={() => {
						if (setPendingFilterMultiSetOnly) {
							setPendingFilterMultiSetOnly(!pendingFilterMultiSetOnly);
						}
					}}
				>
					{pendingFilterMultiSetOnly ? "Multi Set Only" : "All Items"}
				</Badge>
				<Separator className="my-4 w-full" />
				{/* Users */}
				<Label className="text-md font-semibold mb-2 block">Users</Label>
				<Input
					type="search"
					placeholder="Search users..."
					className="mb-2"
					value={userFilterSearch || ""}
					onChange={(e) => setUserFilterSearch(e.target.value)}
					tabIndex={-1}
					autoFocus={false}
				/>
				<div className="flex flex-col gap-1 max-h-48 overflow-y-auto border p-2 rounded-md">
					{/* All User Option */}
					<div
						className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
							pendingFilteredUsers.length === 0 ? "bg-muted" : "hover:bg-muted/60"
						}`}
						onClick={() => {
							setPendingFilteredUsers([]);
						}}
					>
						<Checkbox checked={pendingFilteredUsers.length === 0} id={`users-all`} />
						<Label
							htmlFor={`users-all`}
							className="text-sm flex-1 cursor-pointer truncate"
							onClick={(e) => e.stopPropagation()}
						>
							All User
						</Label>
						{pendingFilteredUsers.length === 0 && <Check className="h-4 w-4 text-primary" />}
					</div>

					{/* No User Option - only show if present in filterUserOptions */}
					{filterUserOptions.includes("|||no-user|||") && (
						<div
							className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
								pendingFilteredUsers.includes("|||no-user|||") ? "bg-muted" : "hover:bg-muted/60"
							}`}
							onClick={() => {
								if (pendingFilteredUsers.includes("|||no-user|||")) {
									setPendingFilteredUsers(pendingFilteredUsers.filter((u) => u !== "|||no-user|||"));
								} else {
									setPendingFilteredUsers([...pendingFilteredUsers, "|||no-user|||"]);
								}
							}}
						>
							<Checkbox checked={pendingFilteredUsers.includes("|||no-user|||")} id={`users-no-user`} />
							<Label
								htmlFor={`users-no-user`}
								className="text-sm flex-1 cursor-pointer truncate"
								onClick={(e) => e.stopPropagation()}
							>
								No User
							</Label>
							{pendingFilteredUsers.includes("|||no-user|||") && (
								<Check className="h-4 w-4 text-primary" />
							)}
						</div>
					)}

					<div className="border-b my-1" />

					{/* List of actual users, excluding "|||no-user|||" */}
					{filterUserOptions
						.filter(
							(user) =>
								user !== "|||no-user|||" &&
								(!userFilterSearch || user.toLowerCase().includes(userFilterSearch.toLowerCase()))
						)
						.sort((a, b) => a.localeCompare(b))
						.map((user) => (
							<div
								key={user}
								className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
									pendingFilteredUsers.includes(user) ? "bg-muted" : "hover:bg-muted/60"
								}`}
								onClick={() => {
									let newUsers;
									if (pendingFilteredUsers.includes(user)) {
										newUsers = pendingFilteredUsers.filter((u) => u !== user);
									} else {
										newUsers = [...pendingFilteredUsers, user];
									}
									setPendingFilteredUsers(newUsers);
								}}
							>
								<Checkbox checked={pendingFilteredUsers.includes(user)} id={`user-${user}`} />
								<Label
									htmlFor={`user-${user}`}
									className="text-sm flex-1 cursor-pointer truncate"
									onClick={(e) => e.stopPropagation()}
								>
									{user}
								</Label>
								{pendingFilteredUsers.includes(user) && <Check className="h-4 w-4 text-primary" />}
							</div>
						))}
				</div>
			</div>

			{/* Apply Filters Button */}
			<Button
				className="w-full mt-2 cursor-pointer hover:brightness-120 active:scale-95"
				onClick={handleApplyFilters}
			>
				Apply Filters
			</Button>

			{/* Reset Filters Button */}
			<Button
				variant={"destructive"}
				className="w-full cursor-pointer hover:brightness-120 active:scale-95"
				onClick={handleResetFilters}
			>
				Reset Filters
			</Button>
		</div>
	);
}

export function FilterSavedSets({
	getSectionSummaries,
	librarySectionsLoaded,
	filteredLibraries,
	setFilteredLibraries,
	typeOptions,
	filteredTypes,
	setFilteredTypes,
	filterAutoDownload,
	setFilterAutoDownload,

	filterUserOptions,
	filteredUsers,
	setFilteredUsers,
	filterMultiSetOnly,
	setFilterMultiSetOnly,

	searchTMDBID,
	searchLibrary,
	searchYear,
	searchTitle,

	sortOption,
	setSortOption,
	sortOrder,
	setSortOrder,

	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: SavedSetsFilterProps) {
	// State - Open/Close Modal
	const [modalOpen, setModalOpen] = useState(false);

	// Calculate number of active filters
	const numberOfActiveFilters = useMemo(() => {
		let count = 0;
		if (filteredLibraries.length > 0) count++;
		if (filterAutoDownload && filterAutoDownload !== "all") count++;
		if (filteredUsers.length > 0) count++;
		if (filteredTypes.length > 0) count++;
		if (filterMultiSetOnly) count++;
		if (searchTitle) count++;
		if (searchYear) count++;
		if (searchLibrary) count++;
		if (searchTMDBID) count++;
		return count;
	}, [
		filteredLibraries.length,
		filterAutoDownload,
		filteredUsers.length,
		filteredTypes.length,
		filterMultiSetOnly,
		searchTitle,
		searchYear,
		searchLibrary,
		searchTMDBID,
	]);

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
					<DialogDescription> Use the options below to sort and filter your saved sets.</DialogDescription>
				</DialogHeader>
				<Separator className="my-1 w-full" />
				<SavedSetsFilterContent
					getSectionSummaries={getSectionSummaries}
					librarySectionsLoaded={librarySectionsLoaded}
					typeOptions={typeOptions}
					filterUserOptions={filterUserOptions}
					filteredLibraries={filteredLibraries}
					setFilteredLibraries={setFilteredLibraries}
					filteredTypes={filteredTypes}
					setFilteredTypes={setFilteredTypes}
					filterAutoDownload={filterAutoDownload}
					setFilterAutoDownload={setFilterAutoDownload}
					filteredUsers={filteredUsers}
					setFilteredUsers={setFilteredUsers}
					filterMultiSetOnly={filterMultiSetOnly}
					setFilterMultiSetOnly={setFilterMultiSetOnly}
					searchTitle={searchTitle}
					searchYear={searchYear}
					searchTMDBID={searchTMDBID}
					searchLibrary={searchLibrary}
					sortOption={sortOption}
					setSortOption={setSortOption}
					sortOrder={sortOrder}
					setSortOrder={setSortOrder}
					setCurrentPage={setCurrentPage}
					itemsPerPage={itemsPerPage}
					setItemsPerPage={setItemsPerPage}
					setModalOpen={setModalOpen}
				/>
			</DialogContent>
		</Dialog>
	);
}
