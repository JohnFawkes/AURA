"use client";

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
import Link from "next/link";
import { useHomeSearchStore } from "@/lib/homeSearchStore";
import { useEffect, useState } from "react";

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

	// Local state to hold the immediate search input
	const [localSearch, setLocalSearch] = useState("");

	// State to hold Placeholder text for the search input
	const [placeholderText, setPlaceholderText] = useState(
		"Search for movies or shows"
	);

	// Use matchMedia to determine if we're on mobile
	useEffect(() => {
		const mediaQuery = window.matchMedia("(max-width: 768px)");
		if (mediaQuery.matches) {
			setPlaceholderText("Search Media");
		} else {
			setPlaceholderText("Search for movies or shows");
		}
		const handleMediaQueryChange = (event: MediaQueryListEvent) => {
			if (event.matches) {
				setPlaceholderText("Search Media");
			} else {
				setPlaceholderText("Search for movies or shows");
			}
		};
		mediaQuery.addEventListener("change", handleMediaQueryChange);
		return () => {
			mediaQuery.removeEventListener("change", handleMediaQueryChange);
		};
	}, []);

	// Debounce updating the zustand store by 300ms
	useEffect(() => {
		const handler = setTimeout(() => {
			setSearchQuery(localSearch);
		}, 300);

		return () => clearTimeout(handler);
	}, [localSearch, setSearchQuery]);

	const handleHomeClick = () => {
		if (isHomePage) {
			setSearchQuery("");
			setCurrentPage(1);
			setFilteredLibraries([]);
			setFilterOutInDB(false);
		}
		router.push("/");
	};

	return (
		<nav
			suppressHydrationWarning
			className={
				"sticky top-0 z-50 flex items-center px-6 py-4 justify-between shadow-md bg-background dark:bg-background-dark border-b border-border dark:border-border-dark"
			}
		>
			{/* Logo */}
			<Link
				href="/"
				className="flex items-center gap-2 hover:text-primary transition-colors"
				onClick={() => handleHomeClick()}
			>
				<div className="relative w-[32px] h-[32px] rounded-t-md overflow-hidden">
					<Image
						src="/mediux.svg"
						alt="Logo"
						fill
						className="object-contain filter dark:invert-0 invert"
					/>
				</div>
				Poster-Setter
			</Link>
			{/* Search Section */}
			{isHomePage && (
				<div className="relative w-full max-w-2xl ml-1 mr-1">
					<SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
					{/* Desktop Search Input */}
					<Input
						type="search"
						placeholder={placeholderText}
						className="pl-10 pr-10 bg-transparent text-foreground rounded-full border-muted"
						onChange={(e) => {
							setLocalSearch(e.target.value);
						}}
					/>
				</div>
			)}
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
