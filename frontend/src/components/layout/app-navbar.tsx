"use client";

import { getAuthToken } from "@/services/auth/api-auth";
import {
	ArrowLeftCircle,
	ArrowRightCircle,
	Bookmark as BookmarkIcon,
	FileCog as FileCogIcon,
	LayoutGrid,
	ListOrdered,
	LogOutIcon,
	Logs,
	Settings as SettingsIcon,
} from "lucide-react";

import { useEffect, useState } from "react";

import Image from "next/image";
import { usePathname, useRouter } from "next/navigation";

import { DynamicSearch } from "@/components/layout/app-search-bar";
import { ViewDensitySlider } from "@/components/shared/view-density-context";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { useCollectionStore } from "@/lib/stores/global-store-collection-store";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { useOnboardingStore } from "@/lib/stores/global-store-onboarding";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";
import { useCollectionsPageStore } from "@/lib/stores/page-store-collections";
import { useHomePageStore } from "@/lib/stores/page-store-home";

const placeholderTexts = {
	home: {
		desktop: "Search for Movies or Shows",
		mobile: "Search",
	},
	savedSets: {
		desktop: "Search Saved Sets",
		mobile: "Search",
	},
	user: {
		desktop: "Search Sets by",
		mobile: "Search",
	},
	collections: {
		desktop: "Search Collections",
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
	const isMediaPage = pathName.startsWith("/media-item") || pathName.startsWith("/media-item/");
	//const isSettingsPage = pathName === "/settings" || pathName === "/settings/";
	const isSavedSetsPage = pathName === "/saved-sets" || pathName === "/saved-sets/";
	const isUserPage = pathName.startsWith("/user/");
	const isOnboardingPage = pathName === "/onboarding" || pathName === "/onboarding/";
	const isLogsPage = pathName === "/logs" || pathName === "/logs/";
	const isChangeLogPage = pathName === "/change-log" || pathName === "/change-log/";
	const isCollectionsPage = pathName === "/collections" || pathName === "/collections/";
	const isCollectionItemPage = pathName.startsWith("/collection-item") || pathName.startsWith("/collection-item/");

	// Auth State
	const [isAuthed, setIsAuthed] = useState(false);

	// Logo state
	const [logoSrc, setLogoSrc] = useState("/aura_word_logo.svg");

	// Search States
	const { setSearchQuery } = useSearchQueryStore(); // Global store for search query
	const [placeholderText, setPlaceholderText] = useState(""); // Placeholder text based on page

	// Home Page Store
	const { setCurrentPage, setFilteredLibraries, setFilterInDB } = useHomePageStore();
	const nextMediaItem = useHomePageStore((state) => state.nextMediaItem);
	const previousMediaItem = useHomePageStore((state) => state.previousMediaItem);

	// Collection Item Page Store
	const nextCollectionItem = useCollectionsPageStore((state) => state.nextCollectionItem);
	const previousCollectionItem = useCollectionsPageStore((state) => state.previousCollectionItem);

	// Onboarding Store
	const { fetchStatus } = useOnboardingStore();
	const status = useOnboardingStore((state) => state.status);
	const hasHydrated = useOnboardingStore((state) => state.hasHydrated);

	// Check if the screen is mobile
	const [isWideScreen, setIsWideScreen] = useState(false);

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
				if (!isOnboardingPage && !isLogsPage && !isChangeLogPage) {
					router.replace("/onboarding");
				}
			} else if (!status.needsSetup) {
				// If does not need setup and on onboarding page, redirect to home
				if (isOnboardingPage) {
					router.replace("/");
				}
			}
		}
	}, [status, pathName, router, hasHydrated, isOnboardingPage, isLogsPage, isChangeLogPage]);

	// Change isWideScreen on window resize
	useEffect(() => {
		const handleResize = () => {
			setIsWideScreen(window.innerWidth >= 950);
		};
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, []);

	useEffect(() => {
		// Update the Logo based on the screen size
		setLogoSrc(isWideScreen ? "/aura_word_logo.svg" : "/aura_logo.svg");

		let username = "";
		// Update the placeholder text based on the page and screen size
		if (isUserPage) {
			const parts = pathName.split("/");
			username = parts[parts.length - 1] || parts[parts.length - 2] || "";
		}
		if (isSavedSetsPage) {
			setPlaceholderText(isWideScreen ? placeholderTexts.savedSets.desktop : placeholderTexts.savedSets.mobile);
		} else if (isCollectionsPage) {
			setPlaceholderText(
				isWideScreen ? placeholderTexts.collections.desktop : placeholderTexts.collections.mobile
			);
		} else if (isUserPage) {
			setPlaceholderText(
				isWideScreen
					? `${placeholderTexts.user.desktop} ${username}`
					: `${placeholderTexts.user.mobile} ${username}`
			);
		} else {
			setPlaceholderText(isWideScreen ? placeholderTexts.home.desktop : placeholderTexts.home.mobile);
		}
	}, [isCollectionsPage, isHomePage, isSavedSetsPage, isUserPage, isWideScreen, pathName]);

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
			setCurrentPage(1);
			setFilteredLibraries([]);
			setFilterInDB("all");
		}
		router.push("/");
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
					<DynamicSearch placeholder={placeholderText} />
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
				{isCollectionItemPage && (
					<>
						<ArrowLeftCircle
							className={`h-8 w-8 hover:scale-105 active:scale-95 transition-colors cursor-pointer ${!previousCollectionItem ? "opacity-30 pointer-events-none" : "text-primary hover:text-primary/80"}`}
							onClick={() => {
								if (previousCollectionItem)
									useCollectionStore.setState({ collectionItem: previousCollectionItem });
							}}
						/>
						<ArrowRightCircle
							className={`h-8 w-8 hover:scale-105 active:scale-95 transition-colors cursor-pointer ${!nextCollectionItem ? "opacity-30 pointer-events-none" : "text-primary hover:text-primary/80"}`}
							onClick={() => {
								if (nextCollectionItem)
									useCollectionStore.setState({ collectionItem: nextCollectionItem });
							}}
						/>
					</>
				)}
				{(!isMediaPage || isWideScreen) && (
					<DropdownMenu>
						<DropdownMenuTrigger
							asChild
							className="cursor-pointer hover:brightness-120 active:scale-95 transition text-muted-foreground"
						>
							<SettingsIcon className="w-8 h-8 ml-2" />
						</DropdownMenuTrigger>
						<DropdownMenuContent className="w-56 md:w-64" side="bottom" align="end">
							{status && !status.needsSetup && (
								<>
									<DropdownMenuItem
										className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
										onClick={() => router.push("/saved-sets")}
									>
										<BookmarkIcon className="w-6 h-6 mr-2" />
										Saved Sets
									</DropdownMenuItem>
									<DropdownMenuItem
										className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
										onClick={() => router.push("/collections")}
									>
										<LayoutGrid className="w-6 h-6 mr-2" />
										Collections
									</DropdownMenuItem>
									<DropdownMenuItem
										className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
										onClick={() => router.push("/download-queue")}
									>
										<ListOrdered className="w-6 h-6 mr-2" />
										Download Queue
									</DropdownMenuItem>
									<DropdownMenuSeparator />
									<DropdownMenuItem
										className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
										onClick={() => router.push("/settings")}
									>
										<FileCogIcon className="w-6 h-6 mr-2" />
										Settings
									</DropdownMenuItem>
									{isWideScreen && (
										<DropdownMenuItem className="cursor-pointer flex items-center active:scale-95 hover:brightness-120">
											<ViewDensitySlider />
										</DropdownMenuItem>
									)}
								</>
							)}
							<DropdownMenuItem
								className="cursor-pointer flex items-center active:scale-95 hover:brightness-120"
								onClick={() => router.push("/logs")}
							>
								<Logs className="w-6 h-6 mr-2" />
								Logs
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
