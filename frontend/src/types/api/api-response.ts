// Define the error structure
interface StandardError {
	Message: string; // Error message
	HelpText: string; // Helpful suggestion for the user
	Function: string; // Function where error occurred
	LineNumber: number; // Line number where error occurred
	Details?: string | Record<string, unknown>; // Optional additional details
}

// Define the main API response structure
export interface APIResponse<T> {
	status: string; // 'success' | 'error' | 'warn'
	elapsed: string; // Time taken to process request
	data?: T; // Optional response data
	error?: StandardError; // Optional error information
}

// Helper type guard to check if response contains an error
// export function hasError(response: APIResponse<unknown>): response is APIResponse<never> {
// 	return response.status === "error" && response.error !== undefined;
// }
