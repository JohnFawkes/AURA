// FilterContent.tsx
import { Check } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

type FilterContentProps = {
	getSectionSummaries: () => { title?: string }[];
	librarySectionsLoaded: boolean;
	filteredLibraries: string[];
	setFilteredLibraries: (libs: string[]) => void;
	typeOptions: { label: string; value: string }[];
	filteredTypes: string[];
	setFilteredTypes: (types: string[]) => void;
	filterAutoDownloadOnly: boolean;
	setFilterAutoDownloadOnly: (val: boolean) => void;
	userFilterSearch: string;
	setUserFilterSearch: (val: string) => void;
	filterUserOptions: string[];
	filteredUsers: string[];
	setFilteredUsers: (users: string[]) => void;
	filterMultiSetOnly?: boolean;
	setFilterMultiSetOnly?: (val: boolean) => void;
	onApplyFilters: () => void;
	onResetFilters: () => void;
	searchString?: string;
	searchYear?: number;
	searchID?: string;
};

export function FilterContent({
	getSectionSummaries,
	librarySectionsLoaded,
	filteredLibraries,
	setFilteredLibraries,
	typeOptions,
	filteredTypes,
	setFilteredTypes,
	filterAutoDownloadOnly,
	setFilterAutoDownloadOnly,
	userFilterSearch,
	setUserFilterSearch,
	filterUserOptions,
	filteredUsers,
	setFilteredUsers,
	filterMultiSetOnly,
	setFilterMultiSetOnly,
	onApplyFilters,
	onResetFilters,
	searchString,
	searchYear,
	searchID,
}: FilterContentProps) {
	const { setSearchQuery } = useSearchQueryStore();

	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
			{/* Search Info */}
			{(searchString || searchYear || searchID) && (
				<div className="p-2 bg-secondary rounded-md">
					<Label className="text-md font-semibold mb-1 block">Current Search</Label>
					<div className="flex flex-col gap-1">
						{searchString && (
							<div className="text-sm">
								<span className="font-semibold">Search:</span> {searchString}
							</div>
						)}
						{typeof searchYear === "number" && searchYear > 0 && (
							<div className="text-sm">
								<span className="font-semibold">Year:</span> {searchYear}
							</div>
						)}
						{searchID && (
							<div className="text-sm">
								<span className="font-semibold">ID:</span> {searchID}
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
													filteredLibraries.filter((lib) => lib !== section)
												);
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
				)}
				{/* Selected Types */}
				<Label className="text-md font-semibold mb-2 block">Selected Types</Label>
				<ToggleGroup
					type="multiple"
					className="flex flex-wrap gap-2 ml-2"
					value={filteredTypes}
					onValueChange={setFilteredTypes}
				>
					{typeOptions.map((type) => (
						<Badge
							key={type.value}
							className="cursor-pointer text-sm"
							variant={filteredTypes.includes(type.value) ? "default" : "outline"}
							onClick={() => {
								if (filteredTypes.includes(type.value)) {
									setFilteredTypes(filteredTypes.filter((t) => t !== type.value));
								} else {
									setFilteredTypes([...filteredTypes, type.value]);
								}
							}}
						>
							{type.label}
						</Badge>
					))}
				</ToggleGroup>
				<Separator className="my-4 w-full" />
				{/* MultiSet Only */}
				<Label className="text-md font-semibold mb-1 block">Multi Set Only</Label>
				<Badge
					key={"filter-multi-set-only"}
					className="cursor-pointer text-sm ml-2"
					variant={filterMultiSetOnly ? "default" : "outline"}
					onClick={() => {
						if (setFilterMultiSetOnly) {
							setFilterMultiSetOnly(!filterMultiSetOnly);
						}
					}}
				>
					{filterMultiSetOnly ? "Multi Set Only" : "All Items"}
				</Badge>
				<Separator className="my-4 w-full" />
				{/* AutoDownload Only */}
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
					<div
						className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
							filteredUsers.length === 0 ? "bg-muted" : "hover:bg-muted/60"
						}`}
						onClick={() => {
							setFilteredUsers([]);
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
						{filteredUsers.length === 0 && <Check className="h-4 w-4 text-primary" />}
					</div>
					<div className="border-b my-1" />
					{filterUserOptions
						.filter(
							(user) => !userFilterSearch || user.toLowerCase().includes(userFilterSearch.toLowerCase())
						)
						.sort((a, b) => a.localeCompare(b))
						.map((user) => (
							<div
								key={user}
								className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
									filteredUsers.includes(user) ? "bg-muted" : "hover:bg-muted/60"
								}`}
								onClick={() => {
									let newUsers;
									if (filteredUsers.includes(user)) {
										newUsers = filteredUsers.filter((u) => u !== user);
									} else {
										newUsers = [...filteredUsers, user];
									}
									setFilteredUsers(newUsers);
								}}
							>
								<Checkbox checked={filteredUsers.includes(user)} id={`user-${user}`} />
								<Label
									htmlFor={`user-${user}`}
									className="text-sm flex-1 cursor-pointer truncate"
									onClick={(e) => e.stopPropagation()}
								>
									{user}
								</Label>
								{filteredUsers.includes(user) && <Check className="h-4 w-4 text-primary" />}
							</div>
						))}
				</div>
			</div>

			{/* Apply Filters Button */}
			<Button className="w-full mt-2" onClick={onApplyFilters}>
				Apply Filters
			</Button>

			{/* Reset Filters Button */}
			<Button variant={"destructive"} className="w-full" onClick={onResetFilters}>
				Reset Filters
			</Button>
		</div>
	);
}
