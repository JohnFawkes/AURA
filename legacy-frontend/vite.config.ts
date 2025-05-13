import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";

const backendPort = process.env.VITE_APP_PORT || "8888"; // Default to 8888 if not set

// https://vite.dev/config/
export default defineConfig({
	plugins: [react()],
	server: {
		port: 3000,
		host: "0.0.0.0", // Allow access from external IPs
		proxy: {
			"/api": {
				target: `http://localhost:${backendPort}`,
				changeOrigin: true,
				secure: false,
			},
		},
	},
	build: {
		outDir: "dist",
		rollupOptions: {
			output: {
				manualChunks(id) {
					if (id.includes("node_modules")) {
						return id
							.toString()
							.split("node_modules/")[1]
							.split("/")[0]
							.toString();
					}
				},
			},
		},
	},
});
