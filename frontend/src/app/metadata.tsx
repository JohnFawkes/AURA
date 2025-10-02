import type { Metadata } from "next";

export const metadata: Metadata = {
	title: "aura",
	description: "An app to set images from MediUX to your Media Server",
	icons: {
		icon: "/favicon.ico", // main favicon
		shortcut: "/favicon.ico",
		apple: "/web-app-manifest-padded-192x192.png", // apple touch icon
		other: [
			{
				rel: "apple-touch-icon",
				sizes: "192x192",
				url: "/web-app-manifest-192x192.png",
			},
			{
				rel: "apple-touch-icon",
				sizes: "512x512",
				url: "/web-app-manifest-512x512.png",
			},
			{
				rel: "manifest",
				url: "/site.webmanifest",
			},
		],
	},
};
