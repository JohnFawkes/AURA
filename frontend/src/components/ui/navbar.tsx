"use client";

import { useState, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import Image from "next/image";
import { Button } from "@/components/ui/button";
import {
	Bookmark as BookmarkIcon,
	Settings as SettingsIcon,
	FileCog as FileCogIcon,
	Search as SearchIcon,
} from "lucide-react";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { useHomeSearchStore } from "@/lib/homeSearchStore";
import { openDB } from "idb";
import { CACHE_DB_NAME, CACHE_STORE_NAME } from "@/constants/cache";
import { LibrarySection, MediaItem } from "@/types/mediaItem";
import { useMediaStore } from "@/lib/mediaStore";
import { searchMediaItems } from "@/hooks/searchMediaItems";

export default function Navbar() {
	const {
		setSearchQuery,
		setCurrentPage,
		setFilteredLibraries,
		setFilterOutInDB,
	} = useHomeSearchStore();
	const router = useRouter();
	const pathName = usePathname();
	const isHomePage = pathName === "/";
	const isSavedSetsPage =
		pathName === "/saved-sets" || pathName === "/saved-sets/";

	// Local state for search input
	const [localSearch, setLocalSearch] = useState("");
	// State for placeholder (optional)
	const [placeholderText, setPlaceholderText] = useState(
		"Search for movies or shows"
	);
	// State for search dropdown results (for non-homepage)
	// This will be populated with results from the IDB
	// when the user types in the search input
	const [searchResults, setSearchResults] = useState<MediaItem[]>([]);
	const [showDropdown, setShowDropdown] = useState(false);
	const { setMediaItem } = useMediaStore();
	const [logoSrc, setLogoSrc] = useState("/aura_word_logo.svg");

	// Use matchMedia to update placeholder based on screen width
	useEffect(() => {
		const mediaQuery = window.matchMedia("(max-width: 768px)");
		setPlaceholderText(
			mediaQuery.matches ? "Search Media" : "Search for movies or shows"
		);
		setLogoSrc(
			mediaQuery.matches ? "/aura_logo.svg" : "/aura_word_logo.svg"
		);
		const handleMediaQueryChange = (event: MediaQueryListEvent) => {
			setPlaceholderText(
				event.matches ? "Search Media" : "Search for movies or shows"
			);
			setLogoSrc(
				event.matches ? "/aura_logo.svg" : "/aura_word_logo.svg"
			);
		};
		mediaQuery.addEventListener("change", handleMediaQueryChange);
		return () => {
			mediaQuery.removeEventListener("change", handleMediaQueryChange);
		};
	}, []);

	// Debounce updating the zustand store when on the homepage
	useEffect(() => {
		if (isHomePage || isSavedSetsPage) {
			const handler = setTimeout(() => {
				setSearchQuery(localSearch);
			}, 300);
			return () => clearTimeout(handler);
		}
	}, [localSearch, setSearchQuery, isHomePage, isSavedSetsPage]);

	// When not on homepage, search the IDB (cache) for matching MediaItems
	useEffect(() => {
		if (!isHomePage && localSearch.trim() !== "") {
			const handler = setTimeout(async () => {
				try {
					const db = await openDB(CACHE_DB_NAME, 1);
					// getAll cached sections from idb
					const cachedSections = await db.getAll(CACHE_STORE_NAME);
					if (cachedSections.length === 0) {
						setSearchResults([]);
						return;
					}
					let allMediaItems: MediaItem[] = [];
					let sections: LibrarySection[] = [];
					sections = cachedSections.map((s) => s.data);
					sections.forEach((section: LibrarySection) => {
						if (section.MediaItems) {
							allMediaItems = allMediaItems.concat(
								section.MediaItems
							);
						}
					});
					// Filter items based on the localSearch (case-insensitive)
					const query = localSearch.trim().toLowerCase();
					const results = searchMediaItems(allMediaItems, query, 10);
					setSearchResults(results);
				} catch (error) {
					console.error("Error fetching cached sections", error);
					setSearchResults([]);
				}
			}, 300);
			return () => clearTimeout(handler);
		} else {
			setSearchResults([]);
		}
	}, [localSearch, isHomePage]);

	const handleHomeClick = () => {
		if (isHomePage) {
			setSearchQuery("");
			setCurrentPage(1);
			setFilteredLibraries([]);
			setFilterOutInDB(false);
		}
		router.push("/");
	};

	// When clicking on a dropdown result (non-homepage), set the mediaStore and navigate
	const handleResultClick = (result: MediaItem) => {
		setMediaItem(result);
		// Format title for URL (replace spaces with underscores, remove special characters)
		const formattedTitle = result.Title.replace(/\s+/g, "_");
		const sanitizedTitle = formattedTitle.replace(/[^a-zA-Z0-9_]/g, "");
		router.push(`/media/${result.RatingKey}/${sanitizedTitle}`);
	};

	return (
		<nav
			suppressHydrationWarning
			className="sticky top-0 z-50 flex items-center px-6 py-4 justify-between shadow-md bg-background dark:bg-background-dark border-b border-border dark:border-border-dark"
		>
			{/* Logo */}
			<div className="relative">
				<div
					className={`relative h-[35px] cursor-pointer w-[${
						logoSrc === "/aura_logo.svg" ? "50px" : "150px"
					}]`}
				>
					<Image
						src={logoSrc}
						alt="Logo"
						fill
						className="object-contain filter dark:invert-0 invert"
						onClick={() => handleHomeClick()}
					/>
				</div>
			</div>

			{/* Search Section */}
			<div className="relative w-full max-w-2xl ml-1 mr-1">
				<SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					type="search"
					placeholder={placeholderText}
					className="pl-10 pr-10 bg-transparent text-foreground rounded-full border-muted"
					onChange={(e) => setLocalSearch(e.target.value)}
					onFocus={() => setShowDropdown(true)}
					onBlur={() => setShowDropdown(false)}
				/>
				{/* If not on homepage, display dropdown results */}
				{!isHomePage &&
					!isSavedSetsPage &&
					showDropdown &&
					searchResults.length > 0 && (
						<div className="absolute top-full mt-5 md:mt-4 w-[80vw] md:w-full max-w-md bg-background border border-border rounded shadow-lg z-50 left-1/2 -translate-x-1/2 md:transform-none">
							{searchResults.map((result) => (
								<div
									key={result.RatingKey}
									onMouseDown={() =>
										handleResultClick(result)
									}
									className="p-2 cursor-pointer hover:bg-muted flex items-center gap-2"
								>
									<div className="relative w-[24px] h-[35px] rounded overflow-hidden">
										<Image
											src={`/api/mediaserver/image/${result.RatingKey}/poster`}
											alt={result.Title}
											fill
											className="object-cover"
											loading="lazy"
											unoptimized
										/>
									</div>
									<div>
										<p className="font-medium text-sm md:text-base">
											{result.Title}
										</p>
										<p className="text-xs text-muted-foreground">
											{result.LibraryTitle} ·{" "}
											{result.Year}
										</p>
									</div>
								</div>
							))}
						</div>
					)}
			</div>
			{/* Settings */}
			<DropdownMenu>
				<DropdownMenuTrigger asChild>
					<Button>
						<SettingsIcon className="w-5 h-5" />
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent className="w-56">
					<DropdownMenuItem
						onClick={() => router.push("/saved-sets")}
					>
						<BookmarkIcon className="w-4 h-4 mr-2" />
						Saved Sets
					</DropdownMenuItem>
					<DropdownMenuSeparator />
					<DropdownMenuItem onClick={() => router.push("/settings")}>
						<FileCogIcon className="w-4 h-4 mr-2" />
						Settings
					</DropdownMenuItem>
				</DropdownMenuContent>
			</DropdownMenu>
		</nav>
	);
}
