import React from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export type SortOption = {
	value: string;
	label: string;
	ascIcon: React.ReactNode;
	descIcon: React.ReactNode;
};

type SortControlProps = {
	options: SortOption[];
	sortOption: string;
	sortOrder: "asc" | "desc";
	setSortOption: (value: string) => void;
	setSortOrder: (order: "asc" | "desc") => void;
	label?: string;
	showLabel?: boolean;
};

export const SortControl: React.FC<SortControlProps> = ({
	options,
	sortOption,
	sortOrder,
	setSortOption,
	setSortOrder,
	label = "Sort:",
	showLabel = true,
}) => {
	const selected = options.find((opt) => opt.value === sortOption);

	return (
		<div className="flex flex-row gap-2 items-center mb-2">
			{showLabel && <Label className="text-lg font-semibold mb-2 sm:mb-0 sm:mr-4">{label}</Label>}
			<Select
				onValueChange={(value) => {
					setSortOption(value);
					// Optionally set default order per option here if needed
				}}
				value={sortOption}
			>
				<SelectTrigger className="w-[140px] sm:w-[180px]">
					<SelectValue placeholder="Sort By" />
				</SelectTrigger>
				<SelectContent>
					{options.map((opt) => (
						<SelectItem className="cursor-pointer" key={opt.value} value={opt.value}>
							{opt.label}
						</SelectItem>
					))}
				</SelectContent>
			</Select>
			{sortOption && selected && (
				<Button
					variant="ghost"
					size="icon"
					className="p-2 cursor-pointer"
					onClick={() => setSortOrder(sortOrder === "asc" ? "desc" : "asc")}
				>
					{sortOrder === "asc" ? selected.ascIcon : selected.descIcon}
				</Button>
			)}
		</div>
	);
};
