import axios from "axios";

const apiClient = axios.create({
	baseURL: "/api",
	timeout: 3000000,
	headers: {
		"Content-Type": "application/json",
	},
});

export default apiClient;
