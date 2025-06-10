import localforage from "localforage";

// Initialize localforage
localforage.config({
	name: "aura",
	storeName: "LibrarySections",
	version: 1.0,
	description: "Library sections cache for Aura",
});

export const storage = localforage;
