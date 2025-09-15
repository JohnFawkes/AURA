import { LayoutGrid, Table } from "lucide-react";

import React from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export type ViewControlOption = {
	value: string;
	label: string;
};

type ViewControlProps = {
	options: ViewControlOption[];
	viewOption: string;
	setViewOption: (option: string) => void;
	label?: string;
	showLabel?: boolean;
};

export const ViewControl: React.FC<ViewControlProps> = ({
	options,
	viewOption,
	setViewOption,
	label = "View:",
	showLabel = true,
}) => {
	const selected = options.find((opt) => opt.value === viewOption);
	return (
		<div className="flex flex-row gap-2 items-center mb-2">
			{showLabel && <Label className="text-md font-semibold mb-2 sm:mb-0">{label}</Label>}

			{selected?.value === "table" ? (
				<Button
					variant="ghost"
					size="icon"
					className="cursor-default hover:bg-transparent focus:bg-transparent p-0"
					onClick={() => {
						// Change to card view
						setViewOption("card");
					}}
				>
					<Table className="h-5 w-5 mr-1 text-muted-foreground" />
				</Button>
			) : selected?.value === "card" ? (
				<Button
					variant="ghost"
					size="icon"
					className="cursor-default hover:bg-transparent focus:bg-transparent p-0"
					onClick={() => {
						// Change to table view
						setViewOption("table");
					}}
				>
					<LayoutGrid className="h-5 w-5 mr-1 text-muted-foreground" />
				</Button>
			) : (
				<Select
					onValueChange={(value) => {
						setViewOption(value);
					}}
					value={viewOption}
				>
					<SelectTrigger className="w-[140px] sm:w-[180px]">
						<SelectValue placeholder="Select view" />
					</SelectTrigger>
					<SelectContent>
						{options.map((opt) => (
							<SelectItem className="cursor-pointer" key={opt.value} value={opt.value}>
								{opt.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			)}
		</div>
	);
};
