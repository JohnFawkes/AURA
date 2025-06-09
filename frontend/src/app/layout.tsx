"use client";

import { ThemeProvider } from "@/providers/theme-provider";
import { Toaster } from "sonner";

import AppFooter from "@/components/layout/app-footer";
import Navbar from "@/components/layout/navbar";
import { JumpToTop } from "@/components/shared/buttons/jump-to-top";

import { gabarito } from "../../public/fonts/Gabarito";
import "./globals.css";

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={`${gabarito.className}`}>
				<ThemeProvider attribute="class" defaultTheme="dark" disableTransitionOnChange>
					{/* Navbar */}
					<Navbar />

					{/* Main Content */}
					<main className="min-h-screen"> {children}</main>

					{/* Footer */}
					<AppFooter version={process.env.NEXT_PUBLIC_APP_VERSION} />
					<Toaster richColors position="top-center" />
					<JumpToTop />
				</ThemeProvider>
			</body>
		</html>
	);
}
