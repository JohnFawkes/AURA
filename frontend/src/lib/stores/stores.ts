import localforage from "localforage";

export const PageStore = localforage.createInstance({
	name: "aura",
	storeName: "PageStores",
	version: 1.0,
	driver: localforage.INDEXEDDB,
});

export const GlobalStore = localforage.createInstance({
	name: "aura",
	storeName: "GlobalStores",
	version: 1.0,
	driver: localforage.INDEXEDDB,
});
