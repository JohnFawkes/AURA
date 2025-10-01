"use client";

import { getAuthToken } from "@/services/auth/api-auth";
import {
	ArrowLeftCircle,
	ArrowRightCircle,
	Bookmark as BookmarkIcon,
	FileCog as FileCogIcon,
	LogOutIcon,
	Search as SearchIcon,
	Settings as SettingsIcon,
} from "lucide-react";

import { useEffect, useState } from "react";

import Image from "next/image";
import { usePathname, useRouter } from "next/navigation";

import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";

import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useOnboardingStore } from "@/lib/stores/global-store-onboarding";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useHomePageStore } from "@/lib/stores/page-store-home";

import { searchMediaItems } from "@/hooks/search-query";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

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
	// Router
	const router = useRouter();

	// Pathname
	const pathName = usePathname();
	// Page Logic
	const isHomePage = pathName === "/";
	const isMediaPage = pathName.startsWith("/media/");
	//const isSettingsPage = pathName === "/settings" || pathName === "/settings/";
	const isSavedSetsPage = pathName === "/saved-sets" || pathName === "/saved-sets/";
	const isUserPage = pathName.startsWith("/user/");
	const isOnboardingPage = pathName === "/onboarding" || pathName === "/onboarding/";

	// Auth State
	const [isAuthed, setIsAuthed] = useState(false);

	// Logo state
	const [logoSrc, setLogoSrc] = useState("/aura_word_logo.svg");

	// Search States
	const { searchQuery, setSearchQuery } = useSearchQueryStore(); // Global store for search query
	const [searchInput, setSearchInput] = useState(searchQuery); // Local state for input field
	const [placeholderText, setPlaceholderText] = useState(""); // Placeholder text based on page

	// Home Page Store
	const { setCurrentPage, setFilteredLibraries, setFilterInDB } = useHomePageStore();
	const nextMediaItem = useHomePageStore((state) => state.nextMediaItem);
	const previousMediaItem = useHomePageStore((state) => state.previousMediaItem);

	// Onboarding Store
	const { fetchStatus } = useOnboardingStore();
	const status = useOnboardingStore((state) => state.status);
	const hasHydrated = useOnboardingStore((state) => state.hasHydrated);

	// Library Sections Store (for searching cached media items)
	const sectionsMap = useLibrarySectionsStore((s) => s.sections); // subscribe so effect reacts to cache changes

	// State for search dropdown results (for non-homepage)
	// This will be populated with results from the IDB
	// when the user types in the search input
	const [searchResults, setSearchResults] = useState<MediaItem[]>([]);
	const [showDropdown, setShowDropdown] = useState(false);

	// Media Store
	const { setMediaItem } = useMediaStore();

	// Check if the screen is mobile
	const [isMobile, setIsMobile] = useState(false);

	// Onboarding Status Check on mount and path change
	useEffect(() => {
		const checkOnboarding = async () => {
			await fetchStatus();
		};
		checkOnboarding();
	}, [pathName, fetchStatus]);

	// Onboarding Redirect Logic
	useEffect(() => {
		if (!hasHydrated) return;
		if (status) {
			// If needs setup and not on onboarding page, redirect to onboarding
			if (status.needsSetup) {
				if (!isOnboardingPage) {
					router.replace("/onboarding");
				}
			} else if (!status.needsSetup) {
				// If does not need setup and on onboarding page, redirect to home
				if (isOnboardingPage) {
					router.replace("/");
				}
			}
		}
	}, [status, pathName, router, hasHydrated, isOnboardingPage]);

	useEffect(() => {
		const handleResize = () => {
			const isNowMobile = window.innerWidth < 768;
			setIsMobile(isNowMobile);

			// Update the Logo based on the screen size
			setLogoSrc(isNowMobile ? "/aura_logo.svg" : "/aura_word_logo.svg");

			let username = "";
			// Update the placeholder text based on the page and screen size
			if (isUserPage) {
				const parts = pathName.split("/");
				username = parts[parts.length - 1] || parts[parts.length - 2] || "";
			}
			if (isSavedSetsPage) {
				setPlaceholderText(
					isNowMobile ? placeholderTexts.savedSets.mobile : placeholderTexts.savedSets.desktop
				);
			} else if (isUserPage) {
				setPlaceholderText(
					isNowMobile
						? `${placeholderTexts.user.mobile} ${username}`
						: `${placeholderTexts.user.desktop} ${username}`
				);
			} else {
				setPlaceholderText(isNowMobile ? placeholderTexts.home.mobile : placeholderTexts.home.desktop);
			}
		};

		handleResize(); // Set initial state

		window.addEventListener("resize", handleResize);
		return () => {
			window.removeEventListener("resize", handleResize);
		};
	}, [isHomePage, isSavedSetsPage, isUserPage, pathName]);

	// Debounce updating
	useEffect(() => {
		const handler = setTimeout(() => {
			setSearchQuery(searchInput);
		}, 300);
		return () => clearTimeout(handler);
	}, [searchInput, setSearchQuery]);

	// Sync local input if store value changes externally
	useEffect(() => {
		setSearchInput(searchQuery);
	}, [searchQuery]);

	// When not on homepage, search cached media items from zustand store
	useEffect(() => {
		if (isHomePage || searchQuery.trim() === "") {
			setSearchResults([]);
			return;
		}

		const handler = setTimeout(() => {
			try {
				const records = Object.values(sectionsMap);

				if (!records || records.length === 0) {
					setSearchResults([]);
					return;
				}

				const allMediaItems: MediaItem[] = records.flatMap((r) => r.MediaItems ?? []);
				if (allMediaItems.length === 0) {
					setSearchResults([]);
					return;
				}

				const query = searchQuery.trim().toLowerCase();
				const results = searchMediaItems(allMediaItems, query, 10);
				setSearchResults(results);
			} catch {
				setSearchResults([]);
			}
		}, 300);

		return () => clearTimeout(handler);
	}, [searchQuery, isHomePage, sectionsMap]); // sectionsMap included so results update when cache fills

	// On mount, check auth status
	useEffect(() => {
		if (status?.currentSetup.Auth.Enabled === false) {
			setIsAuthed(true);
			return;
		}

		// If auth is enabled, check for token
		const token = getAuthToken();
		setIsAuthed(!!token && token !== "null" && token !== "undefined");
	}, [pathName, status?.currentSetup.Auth.Enabled]);

	// When clicking on the logo, navigate to home
	// If already on homepage, reset home page states
	const handleHomeClick = () => {
		if (!isAuthed) {
			router.push("/login");
			return;
		}
		if (isHomePage) {
			setSearchQuery("");
			setSearchInput("");
			setCurrentPage(1);
			setFilteredLibraries([]);
			setFilterInDB("all");
			setSearchResults([]);
		}
		router.push("/");
	};

	// When clicking on a dropdown result (non-homepage), set the mediaStore and navigate
	const handleResultClick = (result: MediaItem) => {
		setMediaItem(result);
		//router.push(formatMediaItemUrl(result));
		router.push("/media/");
	};

	// Handle Logout
	const handleLogout = () => {
		localStorage.removeItem("aura-auth-token");
		setIsAuthed(false);
		// Redirect to login page
		router.replace("/login");
	};

	return (
		<nav
			suppressHydrationWarning
			className="sticky top-0 z-50 flex items-center px-6 py-4 justify-between shadow-md bg-background dark:bg-background-dark border-b border-border dark:border-border-dark"
		>
			{/* Left: Logo */}
			<div className="relative flex-shrink-0">
				<div
					onClick={handleHomeClick}
					className="relative cursor-pointer active:scale-95 transition-transform select-none"
					style={{
						width: logoSrc === "/aura_logo.svg" ? "50px" : "150px",
						height: logoSrc === "/aura_logo.svg" ? "30px" : "35px",
					}}
				>
					<Image src={logoSrc} alt="Logo" fill className="object-contain filter" priority />
				</div>
			</div>

			{/* Center: Search */}
			<div className="relative flex-1 flex justify-center mx-3">
				<div className="relative w-full max-w-2xl">
					<SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						type="search"
						placeholder={placeholderText}
						className="pl-10 pr-10 bg-transparent text-foreground rounded-full border-muted w-full focus:border-primary focus:ring-1 focus:ring-primary placeholder:text-muted-foreground transition hover:brightness-120"
						onChange={(e) => setSearchInput(e.target.value)}
						value={searchInput}
						onFocus={() => setShowDropdown(true)}
						onBlur={() => setShowDropdown(false)}
					/>
					{/* Dropdown results (unchanged) */}
					{!isHomePage && !isSavedSetsPage && !isUserPage && showDropdown && searchResults.length > 0 && (
						<div className="absolute top-full mt-5 md:mt-4 w-[80vw] md:w-full max-w-md bg-background border border-border rounded shadow-lg z-50 left-1/2 -translate-x-1/2 md:transform-none max-h-[400px] overflow-y-auto">
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
			</div>

			{/* Right: Arrows and/or Settings */}
			<div className="flex items-center gap-2 flex-shrink-0">
				{isMediaPage && (
					<>
						<ArrowLeftCircle
							className={`h-8 w-8 hover:scale-105 active:scale-95 transition-colors cursor-pointer ${!previousMediaItem ? "opacity-30 pointer-events-none" : "text-primary hover:text-primary/80"}`}
							onClick={() => {
								if (previousMediaItem) useMediaStore.setState({ mediaItem: previousMediaItem });
							}}
						/>
						<ArrowRightCircle
							className={`h-8 w-8 hover:scale-105 active:scale-95 transition-colors cursor-pointer ${!nextMediaItem ? "opacity-30 pointer-events-none" : "text-primary hover:text-primary/80"}`}
							onClick={() => {
								if (nextMediaItem) useMediaStore.setState({ mediaItem: nextMediaItem });
							}}
						/>
					</>
				)}
				{(!isMediaPage || !isMobile) && (
					<DropdownMenu>
						<DropdownMenuTrigger
							asChild
							className="cursor-pointer hover:brightness-120 active:scale-95 transition"
						>
							<SettingsIcon className="w-8 h-8 ml-2" />
						</DropdownMenuTrigger>
						<DropdownMenuContent className="w-56 md:w-64" side="bottom" align="end">
							<DropdownMenuItem
								className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
								onClick={() => router.push("/saved-sets")}
							>
								<BookmarkIcon className="w-6 h-6 mr-2" />
								Saved Sets
							</DropdownMenuItem>
							<DropdownMenuSeparator />
							<DropdownMenuItem
								className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
								onClick={() => router.push("/settings")}
							>
								<FileCogIcon className="w-6 h-6 mr-2" />
								Settings
							</DropdownMenuItem>
							{isAuthed && status?.currentSetup.Auth.Enabled && (
								<>
									<DropdownMenuSeparator />
									<DropdownMenuItem
										className="cursor-pointer flex items-center active:scale-95 hover:brightness-120 text-red-600 focus:text-red-700"
										onClick={handleLogout}
									>
										<LogOutIcon className="w-6 h-6 mr-2" />
										Logout
									</DropdownMenuItem>
								</>
							)}
						</DropdownMenuContent>
					</DropdownMenu>
				)}
			</div>
		</nav>
	);
}
