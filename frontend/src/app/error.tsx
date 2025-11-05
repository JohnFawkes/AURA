"use client";

import { useEffect } from "react";

import Link from "next/link";

import { Button } from "@/components/ui/button";

export default function GlobalError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
	useEffect(() => {
		// eslint-disable-next-line no-console
		console.error("App Error:", error);
	}, [error]);

	return (
		<div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-700 px-2 py-8">
			<div className="text-center max-w-full w-full" style={{ maxWidth: 420 }}>
				<h1 className="text-5xl sm:text-6xl md:text-7xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-red-500 to-orange-700 mb-3">
					Error
				</h1>
				<h2 className="text-xl sm:text-2xl md:text-3xl font-semibold text-white mb-2">Something went wrong</h2>
				<p className="text-gray-300 mb-3 text-base sm:text-lg break-words">
					{error.message || "An unexpected error occurred. Please try again later."}
				</p>
				<p className="text-gray-400 text-md mb-6">
					More technical details are available in your browser console.
				</p>
				<div className="flex flex-col sm:flex-row gap-2 justify-center mb-6">
					<Button
						variant="outline"
						onClick={() => reset()}
						className="shadow transition text-white hover:text-green-500"
					>
						Try Again
					</Button>
					<Button variant="outline" className="shadow transition text-white hover:text-primary">
						<Link href="/">Go to Home</Link>
					</Button>
				</div>
				<div className="text-gray-500 text-md">If the problem persists, please contact report the issue.</div>
			</div>
		</div>
	);
}
