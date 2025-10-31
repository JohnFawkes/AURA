import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

import { extractInfoFromSearchQuery } from "@/hooks/search-query";

import { TYPE_FILTER_IN_DB_OPTIONS } from "@/types/ui-options";

type HomeFilterProps = {
	librarySections: { label: string; value: string }[];
	filteredLibraries: string[];
	setFilteredLibraries: (libs: string[]) => void;
	inDBOptions: { label: string; value: string }[];
	filterInDB: string;
	setFilterInDB: (filter: TYPE_FILTER_IN_DB_OPTIONS) => void;
};

export function FilterHome({
	librarySections,
	filteredLibraries,
	setFilteredLibraries,
	inDBOptions,
	filterInDB,
	setFilterInDB,
}: HomeFilterProps) {
	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const { searchTMDBID, searchLibrary, searchYear, searchTitle } = extractInfoFromSearchQuery(searchQuery);

	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
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
								key={section.label}
								className="cursor-pointer text-sm active:scale-95 hover:brightness-120"
								variant={filteredLibraries.includes(section.value) ? "default" : "outline"}
								onClick={() => {
									if (filteredLibraries.includes(section.value)) {
										setFilteredLibraries(filteredLibraries.filter((lib) => lib !== section.value));
									} else {
										setFilteredLibraries([...filteredLibraries, section.value]);
									}
								}}
							>
								{section.label}
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
						{inDBOptions.map((option) => (
							<Badge
								key={option.value}
								className="cursor-pointer text-sm active:scale-95 hover:brightness-120"
								variant={filterInDB === option.value ? "default" : "outline"}
								onClick={() => {
									if (filterInDB === option.value) {
										setFilterInDB("all");
									} else {
										setFilterInDB(option.value as TYPE_FILTER_IN_DB_OPTIONS);
									}
								}}
							>
								{option.label}
							</Badge>
						))}
					</ToggleGroup>
				</div>
			</div>
		</div>
	);
}
