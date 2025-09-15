export interface ViewOptionsStore {
	viewOption: "card" | "table";
	setViewOption: (option: "card" | "table") => void;
}
