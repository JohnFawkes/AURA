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

interface CustomPaginationProps {
	currentPage: number;
	totalPages: number;
	setCurrentPage: (page: number) => void;
	scrollToTop?: boolean;
}

export function CustomPagination({
	currentPage,
	totalPages,
	setCurrentPage,
	scrollToTop = true,
}: CustomPaginationProps) {
	const handlePageChange = (newPage: number) => {
		setCurrentPage(newPage);
		if (scrollToTop) {
			window.scrollTo({ top: 0, behavior: "smooth" });
		}
	};

	return (
		<div className="flex justify-center mt-8">
			{totalPages > 1 && (
				<Pagination>
					<PaginationContent>
						{/* Previous Page Button */}
						{totalPages > 1 && currentPage > 1 && (
							<PaginationItem>
								<PaginationPrevious onClick={() => handlePageChange(Math.max(currentPage - 1, 1))} />
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
	);
}
