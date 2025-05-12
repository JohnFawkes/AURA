"use client";

import { ThemeProvider } from "@/components/theme-provider";
import AppFooter from "@/components/ui/app-footer";
import { JumpToTop } from "@/components/ui/jump-to-top";
import Navbar from "@/components/ui/navbar";
import { Gabarito } from "next/font/google";
import { createContext, useState } from "react";
import { Toaster } from "sonner";
import "./globals.css";
const gabarito = Gabarito({ subsets: ["latin"] });

export const SearchContext = createContext<{
	searchQuery: string;
	setSearchQuery: (query: string) => void;
}>({
	searchQuery: "",
	setSearchQuery: () => {},
});

export default function RootLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const [searchQuery, setSearchQuery] = useState<string>("");
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={`${gabarito.className}`}>
				<ThemeProvider
					attribute="class"
					defaultTheme="dark"
					disableTransitionOnChange
				>
					<SearchContext.Provider
						value={{
							searchQuery,
							setSearchQuery,
						}}
					>
						{/* Navbar */}
						<Navbar />

						{/* Main Content */}
						<main className="min-h-screen"> {children}</main>

						{/* Footer */}
						<AppFooter
							version={process.env.NEXT_PUBLIC_APP_VERSION}
						/>
					</SearchContext.Provider>
					<Toaster richColors position="top-center" />
					<JumpToTop />
				</ThemeProvider>
			</body>
		</html>
	);
}
