import { create } from "zustand";
import { persist } from "zustand/middleware";

interface PaginationStore {
	currentPage: number;
	setCurrentPage: (page: number) => void;
	itemsPerPage: number;
	setItemsPerPage: (itemsPerPage: number) => void;
}

export const usePaginationStore = create<PaginationStore>()(
	persist(
		(set) => ({
			currentPage: 1,
			setCurrentPage: (page) => set({ currentPage: page }),
			itemsPerPage: 20, // Default items per page
			setItemsPerPage: (itemsPerPage) => set({ itemsPerPage }),
		}),
		{
			name: "pagination-storage", // key in localStorage
		}
	)
);
