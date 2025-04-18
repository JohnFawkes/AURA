// src/theme.ts
import { createTheme } from "@mui/material/styles";

const orange = {
	main: "#d16141", // deep orange
	contrastText: "#fff",
};

export const lightTheme = createTheme({
	palette: {
		mode: "light",
		primary: orange,
		background: {
			default: "#f9f9f9",
			paper: "#ffffff",
		},
	},
});

export const darkTheme = createTheme({
	palette: {
		mode: "dark",
		primary: orange,
		background: {
			default: "#121212",
			paper: "#1e1e1e",
		},
	},
});
