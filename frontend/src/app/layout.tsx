"use client";

import { ThemeProvider } from "@/components/theme-provider";
import AppFooter from "@/components/ui/app-footer";
import { JumpToTop } from "@/components/ui/jump-to-top";
import Navbar from "@/components/ui/navbar";
import { Toaster } from "sonner";
import "./globals.css";
import { gabarito } from "../../public/fonts/Gabarito";

export default function RootLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={`${gabarito.className}`}>
				<ThemeProvider
					attribute="class"
					defaultTheme="dark"
					disableTransitionOnChange
				>
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
