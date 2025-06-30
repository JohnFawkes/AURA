import localforage from "localforage";

// Library Sections storage
export const librarySectionsStorage = localforage.createInstance({
	name: "aura",
	storeName: "LibrarySections",
	version: 1.0,
	description: "Stores all of the Library Sections and their MediaItems",
});
