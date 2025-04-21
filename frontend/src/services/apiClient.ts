import axios from "axios";

const apiClient = axios.create({
	baseURL: "/api",
	timeout: 1200000, // 120 seconds
	headers: {
		"Content-Type": "application/json",
	},
});

export default apiClient;
