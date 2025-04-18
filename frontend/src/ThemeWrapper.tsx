import React, { useMemo, useState } from "react";
import { StyledEngineProvider, ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import App from "./App";
import { lightTheme, darkTheme } from "./theme";

const ThemeWrapper: React.FC = () => {
	const [darkMode, setDarkMode] = useState(false);

	const theme = useMemo(
		() => (darkMode ? darkTheme : lightTheme),
		[darkMode]
	);

	return (
		<StyledEngineProvider injectFirst>
			<ThemeProvider theme={theme}>
				<CssBaseline />
				<App darkMode={darkMode} setDarkMode={setDarkMode} />
			</ThemeProvider>
		</StyledEngineProvider>
	);
};

export default ThemeWrapper;
