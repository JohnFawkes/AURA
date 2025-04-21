import React from "react";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import Box from "@mui/material/Box";

const Loader: React.FC<{
	loadingText?: string;
}> = ({ loadingText }) => {
	return (
		<Box
			sx={{
				display: "flex",
				flexDirection: "column",
				justifyContent: "center",
				alignItems: "center",
				height: "100vh", // Full viewport height
			}}
		>
			{loadingText && (
				<Typography
					variant="h6"
					sx={{ fontWeight: 400, marginBottom: 2 }}
				>
					{loadingText}
				</Typography>
			)}
			<CircularProgress size={50} thickness={4} />
		</Box>
	);
};

export default Loader;
