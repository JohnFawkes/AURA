export interface APIResponse<T> {
	status: string;
	message: string;
	elapsed?: string;
	data?: T;
}
