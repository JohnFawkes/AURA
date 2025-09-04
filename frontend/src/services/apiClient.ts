import axios from "axios";

const apiClient = axios.create({
	baseURL: "/api",
	timeout: 3000000,
	headers: {
		"Content-Type": "application/json",
	},
});

apiClient.interceptors.request.use((config) => {
	if (typeof window !== "undefined") {
		const token = localStorage.getItem("aura-auth-token");
		if (token) {
			config.headers = config.headers || {};
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(config.headers as any).Authorization = `Bearer ${token}`;
		}
	}
	return config;
});

// Optional: auto redirect on 401
apiClient.interceptors.response.use(
	(res) => res,
	(err) => {
		if (err?.response?.status === 401 && typeof window !== "undefined") {
			// Clear token and go to login
			localStorage.removeItem("aura-auth-token");
			if (!window.location.pathname.startsWith("/login")) {
				window.location.href = "/login";
			}
		}
		return Promise.reject(err);
	}
);

export default apiClient;
