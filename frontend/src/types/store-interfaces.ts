export interface SortStore<TSortOption = string, TSortOrder = string> {
	sortOption: TSortOption;
	setSortOption: (option: TSortOption) => void;
	sortOrder: TSortOrder;
	setSortOrder: (order: TSortOrder) => void;
}

export interface PaginationStore<TCurrentPage = number, TItemsPerPage = number> {
	currentPage: TCurrentPage;
	setCurrentPage: (page: TCurrentPage) => void;
	itemsPerPage: TItemsPerPage;
	setItemsPerPage: (itemsPerPage: TItemsPerPage) => void;
}
