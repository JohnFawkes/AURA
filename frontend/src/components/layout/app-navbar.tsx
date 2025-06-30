"use client";

import { formatMediaItemUrl } from "@/helper/formatMediaItemURL";
import {
	Bookmark as BookmarkIcon,
	FileCog as FileCogIcon,
	Search as SearchIcon,
	Settings as SettingsIcon,
} from "lucide-react";

import { useEffect, useState } from "react";

import Image from "next/image";
import { usePathname, useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";

import { useMediaStore } from "@/lib/mediaStore";
import { useHomePageStore } from "@/lib/pageHomeStore";
import { usePaginationStore } from "@/lib/paginationStore";
import { useSearchQueryStore } from "@/lib/searchQueryStore";
import { librarySectionsStorage } from "@/lib/storage";

import { searchMediaItems } from "@/hooks/searchMediaItems";

import { LibrarySection, MediaItem } from "@/types/mediaItem";

const placeholderTexts = {
	home: {
		desktop: "Search for Movies or Shows",
		mobile: "Search Media",
	},
	savedSets: {
		desktop: "Search Saved Sets",
		mobile: "Search Sets",
	},
	user: {
		desktop: "Search Sets by",
		mobile: "Search",
	},
};

export default function Navbar() {
	const { searchQuery, setSearchQuery } = useSearchQueryStore();
	const { setCurrentPage } = usePaginationStore();
	const { setFilteredLibraries, setFilterOutInDB } = useHomePageStore();
	const router = useRouter();
	const pathName = usePathname();
	const isHomePage = pathName === "/";
	const isSavedSetsPage = pathName === "/saved-sets" || pathName === "/saved-sets/";
	const isUserPage = pathName.startsWith("/user/");

	const [placeholderText, setPlaceholderText] = useState("");

	// State for search dropdown results (for non-homepage)
	// This will be populated with results from the IDB
	// when the user types in the search input
	const [searchResults, setSearchResults] = useState<MediaItem[]>([]);
	const [showDropdown, setShowDropdown] = useState(false);
	const { setMediaItem } = useMediaStore();
	const [logoSrc, setLogoSrc] = useState("/aura_word_logo.svg");

	// Set the placeholder text based on the current page
	useEffect(() => {
		const mediaQuery = window.matchMedia("(max-width: 768px)");
		let username = "";
		if (isUserPage) {
			const parts = pathName.split("/");
			username = parts[parts.length - 1] || parts[parts.length - 2] || "";
		}
		if (isSavedSetsPage) {
			setPlaceholderText(
				mediaQuery.matches ? placeholderTexts.savedSets.mobile : placeholderTexts.savedSets.desktop
			);
		} else if (isUserPage) {
			setPlaceholderText(
				mediaQuery.matches
					? `${placeholderTexts.user.mobile} ${username}`
					: `${placeholderTexts.user.desktop} ${username}`
			);
		} else {
			setPlaceholderText(mediaQuery.matches ? placeholderTexts.home.mobile : placeholderTexts.home.desktop);
		}
	}, [isHomePage, isSavedSetsPage, isUserPage, pathName]);

	// Use matchMedia to update placeholder based on screen width
	useEffect(() => {
		const mediaQuery = window.matchMedia("(max-width: 768px)");
		setLogoSrc(mediaQuery.matches ? "/aura_logo.svg" : "/aura_word_logo.svg");
		const handleMediaQueryChange = (event: MediaQueryListEvent) => {
			setLogoSrc(event.matches ? "/aura_logo.svg" : "/aura_word_logo.svg");
		};
		mediaQuery.addEventListener("change", handleMediaQueryChange);
		return () => {
			mediaQuery.removeEventListener("change", handleMediaQueryChange);
		};
	}, []);

	// Debounce updating the zustand store when on the homepage
	useEffect(() => {
		if (isHomePage || isSavedSetsPage || isUserPage) {
			const handler = setTimeout(() => {
				setSearchQuery(searchQuery);
			}, 300);
			return () => clearTimeout(handler);
		}
	}, [searchQuery, setSearchQuery, isHomePage, isSavedSetsPage, isUserPage]);

	// When not on homepage, search the IDB (cache) for matching MediaItems
	useEffect(() => {
		if (!isHomePage && searchQuery.trim() !== "") {
			const handler = setTimeout(async () => {
				try {
					// Get all cached sections from librarySectionsStorage
					const keys = await librarySectionsStorage.keys();
					const cachedSectionsPromises = keys.map((key) =>
						librarySectionsStorage.getItem<{
							data: LibrarySection;
							timestamp: number;
						}>(key)
					);
					const cachedSections = (await Promise.all(cachedSectionsPromises)).filter(
						(
							section
						): section is {
							data: LibrarySection;
							timestamp: number;
						} => section !== null
					);

					if (cachedSections.length === 0) {
						setSearchResults([]);
						return;
					}

					let allMediaItems: MediaItem[] = [];
					const sections = cachedSections.map((s) => s.data);

					sections.forEach((section: LibrarySection) => {
						if (section.MediaItems) {
							allMediaItems = allMediaItems.concat(section.MediaItems);
						}
					});

					// Filter items based on the searchQuery (case-insensitive)
					const query = searchQuery.trim().toLowerCase();
					const results = searchMediaItems(allMediaItems, query, 10);
					setSearchResults(results);
				} catch {
					setSearchResults([]);
				}
			}, 300);
			return () => clearTimeout(handler);
		} else {
			setSearchResults([]);
		}
	}, [searchQuery, isHomePage]);

	const handleHomeClick = () => {
		if (isHomePage) {
			setSearchQuery("");
			setCurrentPage(1);
			setFilteredLibraries([]);
			setFilterOutInDB(false);
			setSearchResults([]);
		}
		router.push("/");
	};

	// When clicking on a dropdown result (non-homepage), set the mediaStore and navigate
	const handleResultClick = (result: MediaItem) => {
		setMediaItem(result);
		router.push(formatMediaItemUrl(result));
	};

	return (
		<nav
			suppressHydrationWarning
			className="sticky top-0 z-50 flex items-center px-6 py-4 justify-between shadow-md bg-background dark:bg-background-dark border-b border-border dark:border-border-dark"
		>
			{/* Logo */}
			<div className="relative">
				<div
					onClick={() => handleHomeClick()}
					className="relative cursor-pointer"
					style={{
						width: logoSrc === "/aura_logo.svg" ? "50px" : "150px",
						height: logoSrc === "/aura_logo.svg" ? "30px" : "35px",
					}}
				>
					<Image src={logoSrc} alt="Logo" fill className="object-contain filter" priority />
				</div>
			</div>

			{/* Search Section */}
			<div className="relative w-full max-w-2xl ml-1 mr-3">
				<SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					type="search"
					placeholder={placeholderText}
					className="pl-10 pr-10 bg-transparent text-foreground rounded-full border-muted"
					onChange={(e) => setSearchQuery(e.target.value)}
					value={searchQuery}
					onFocus={() => setShowDropdown(true)}
					onBlur={() => setShowDropdown(false)}
				/>
				{/* If not on homepage, display dropdown results */}
				{!isHomePage && !isSavedSetsPage && !isUserPage && showDropdown && searchResults.length > 0 && (
					<div className="absolute top-full mt-5 md:mt-4 w-[80vw] md:w-full max-w-md bg-background border border-border rounded shadow-lg z-50 left-1/2 -translate-x-1/2 md:transform-none">
						{searchResults.map((result) => (
							<div
								key={result.RatingKey}
								onMouseDown={() => handleResultClick(result)}
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
									<p className="font-medium text-sm md:text-base">{result.Title}</p>
									<p className="text-xs text-muted-foreground">
										{result.LibraryTitle} · {result.Year}
									</p>
								</div>
							</div>
						))}
					</div>
				)}
			</div>
			{/* Settings */}
			<DropdownMenu>
				<DropdownMenuTrigger asChild className="cursor-pointer">
					<Button>
						<SettingsIcon className="w-5 h-5" />
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent className="w-56">
					<DropdownMenuItem
						className="cursor-pointer flex items-center hover:bg-primary/10"
						onClick={() => router.push("/saved-sets")}
					>
						<BookmarkIcon className="w-4 h-4 mr-2" />
						Saved Sets
					</DropdownMenuItem>
					<DropdownMenuSeparator />
					<DropdownMenuItem
						className="cursor-pointer flex items-center hover:bg-primary/10"
						onClick={() => router.push("/settings")}
					>
						<FileCogIcon className="w-4 h-4 mr-2" />
						Settings
					</DropdownMenuItem>
				</DropdownMenuContent>
			</DropdownMenu>
		</nav>
	);
}
