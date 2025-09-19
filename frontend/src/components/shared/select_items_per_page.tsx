"use client";

import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

import { ITEMS_PER_PAGE_OPTIONS, TYPE_ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

export function SelectItemsPerPage({
	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: {
	setCurrentPage: (page: number) => void;
	itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS;
	setItemsPerPage: (itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS) => void;
}) {
	return (
		<>
			<Label htmlFor="items-per-page-trigger" className="text-lg font-semibold mb-2 sm:mb-0 mr-2">
				Items per page:
			</Label>
			<Select
				value={itemsPerPage.toString()}
				onValueChange={(value) => {
					const newItemsPerPage = Number(value) as TYPE_ITEMS_PER_PAGE_OPTIONS;
					if (ITEMS_PER_PAGE_OPTIONS.includes(newItemsPerPage)) {
						setItemsPerPage(newItemsPerPage);
						setCurrentPage(1);
					}
				}}
			>
				<SelectTrigger id="items-per-page-trigger">
					<SelectValue placeholder="Select" />
				</SelectTrigger>
				<SelectContent>
					{ITEMS_PER_PAGE_OPTIONS.map((option) => (
						<SelectItem className="cursor-pointer" key={option} value={option.toString()}>
							{option}
						</SelectItem>
					))}
					<SelectScrollUpButton />
					<SelectScrollDownButton />
				</SelectContent>
			</Select>
		</>
	);
}
