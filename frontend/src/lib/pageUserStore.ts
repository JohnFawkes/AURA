import { create } from "zustand";
import { persist } from "zustand/middleware";

import { SortOptionsStore } from "@/lib/sortOptionsStore";

interface UserPageStore extends SortOptionsStore {
	clear: () => void;
}

export const useUserPageStore = create<UserPageStore>()(
	persist(
		(set) => ({
			sortOption: "date",
			sortOrder: "desc",
			setSortOption: (option) => set({ sortOption: option }),
			setSortOrder: (order) => set({ sortOrder: order }),

			clear: () => set({ sortOption: "date", sortOrder: "desc" }),
		}),
		{
			name: "user-page-storage", // key in localStorage
		}
	)
);
