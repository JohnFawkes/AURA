"use client";

import { useHomeSearchStore } from "@/lib/homeSearchStore";

import { Label } from "./ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectScrollDownButton,
	SelectScrollUpButton,
	SelectTrigger,
	SelectValue,
} from "./ui/select";

export function SelectItemsPerPage({ setCurrentPage }: { setCurrentPage: (page: number) => void }) {
	const { itemsPerPage, setItemsPerPage } = useHomeSearchStore();
	const itemsPerPageOptions = [10, 20, 30, 50, 100];

	return (
		<>
			<Label
				htmlFor="items-per-page-trigger"
				className="text-lg font-semibold mb-2 sm:mb-0 mr-2"
			>
				Items per page:
			</Label>
			<Select
				value={itemsPerPage.toString()}
				onValueChange={(value) => {
					const newItemsPerPage = parseInt(value);
					if (!isNaN(newItemsPerPage)) {
						setItemsPerPage(newItemsPerPage);
						setCurrentPage(1);
					}
				}}
			>
				<SelectTrigger id="items-per-page-trigger">
					<SelectValue placeholder="Select" />
				</SelectTrigger>
				<SelectContent>
					{itemsPerPageOptions.map((option) => (
						<SelectItem key={option} value={option.toString()}>
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
