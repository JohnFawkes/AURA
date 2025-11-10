import { useEffect } from "react";

import { Input } from "@/components/ui/input";
import {
	Pagination,
	PaginationContent,
	PaginationEllipsis,
	PaginationItem,
	PaginationLink,
	PaginationNext,
	PaginationPrevious,
} from "@/components/ui/pagination";

import { TYPE_ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

interface CustomPaginationProps {
	currentPage: number;
	totalPages: number;
	setCurrentPage: (page: number) => void;
	scrollToTop?: boolean;
	filterItemsLength?: number; // Optional prop to filter items length
	itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS;
}

export function CustomPagination({
	currentPage,
	totalPages,
	setCurrentPage,
	scrollToTop = true,
	filterItemsLength,
	itemsPerPage,
}: CustomPaginationProps) {
	const handlePageChange = (newPage: number) => {
		setCurrentPage(newPage);
		if (scrollToTop) {
			window.scrollTo({ top: 0, behavior: "smooth" });
		}
	};

	useEffect(() => {
		// Reset to first page if filterItemsLength changes and currentPage is out of bounds
		if (filterItemsLength !== undefined) {
			const maxPage = Math.ceil(filterItemsLength / itemsPerPage);
			if (currentPage > maxPage) {
				setCurrentPage(1);
			}
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [filterItemsLength]);

	return (
		<>
			<div className="w-full flex justify-center items-center text-sm mt-4 text-muted-foreground">
				<span>
					{(() => {
						if (filterItemsLength === 0) {
							return "No Results";
						}
						const start = currentPage * itemsPerPage - itemsPerPage + 1;
						const end = filterItemsLength
							? Math.min(currentPage * itemsPerPage, filterItemsLength)
							: currentPage * itemsPerPage;
						if (filterItemsLength && end === filterItemsLength) {
							return `Showing ${start} - ${end}`;
						}
						if (filterItemsLength) {
							return `Showing ${start} - ${end} of ${filterItemsLength}`;
						}
						return `Showing ${start} - ${end}`;
					})()}
				</span>
			</div>
			<div className="flex justify-center mt-8">
				{/* Showing XX of XXX */}

				{totalPages > 1 && (
					<Pagination>
						<PaginationContent>
							{/* Previous Page Button */}
							{totalPages > 1 && currentPage > 1 && (
								<PaginationItem>
									<PaginationPrevious
										onClick={() => handlePageChange(Math.max(currentPage - 1, 1))}
									/>
								</PaginationItem>
							)}

							{/* Page Input */}
							<PaginationItem>
								<div className="flex items-center gap-2">
									<Input
										type="number"
										min={1}
										max={totalPages}
										value={currentPage}
										onChange={(e) => {
											const value = parseInt(e.target.value);
											if (!isNaN(value) && value >= 1 && value <= totalPages) {
												handlePageChange(value);
											}
										}}
										className="w-16 h-8 text-center [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
										onBlur={(e) => {
											const value = parseInt(e.target.value);
											if (isNaN(value) || value < 1) {
												handlePageChange(1);
											} else if (value > totalPages) {
												handlePageChange(totalPages);
											}
										}}
									/>
								</div>
							</PaginationItem>

							{/* Next Page Button */}
							{totalPages > 1 && currentPage < totalPages && (
								<PaginationItem>
									<PaginationNext
										className="cursor-pointer"
										onClick={() => handlePageChange(Math.min(currentPage + 1, totalPages))}
									/>
								</PaginationItem>
							)}

							{/* Ellipsis and End Page */}
							{totalPages > 3 && currentPage < totalPages - 1 && (
								<>
									<PaginationItem>
										<PaginationEllipsis />
									</PaginationItem>
									<PaginationItem>
										<PaginationLink
											className="cursor-pointer"
											onClick={() => handlePageChange(totalPages)}
										>
											{totalPages}
										</PaginationLink>
									</PaginationItem>
								</>
							)}
						</PaginationContent>
					</Pagination>
				)}
			</div>
		</>
	);
}
