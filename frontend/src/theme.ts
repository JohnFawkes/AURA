// src/theme.ts
import { createTheme } from "@mui/material/styles";

export const lightTheme = createTheme({
	palette: {
		mode: "light",
		primary: {
			main: "#e86143",
		},
		secondary: {
			main: "#ff9800",
		},
		error: {
			main: "#f44336",
		},
		background: {
			default: "#f3eeec", // Main background color
			paper: "#ece6e3", // Paper background color
		},
		text: {
			primary: "#000000",
			secondary: "#555555",
		},
	},
});

export const darkTheme = createTheme({
	palette: {
		mode: "dark",
		primary: {
			main: "#e86143",
		},
		secondary: {
			main: "#8c1c04",
		},
		error: {
			main: "#870e04",
		},
		background: {
			default: "#080401",
			paper: "#312c29",
		},
		text: {
			primary: "#ffffff",
			secondary: "#bbbbbb",
		},
	},
});
