import React from "react";
import { Box, Typography, Button } from "@mui/material";
import { useNavigate } from "react-router-dom";

const PageNotFound: React.FC = () => {
	const navigate = useNavigate();

	return (
		<Box
			sx={{
				display: "flex",
				flexDirection: "column",
				justifyContent: "center",
				alignItems: "center",
				height: "100vh",
				textAlign: "center",
				bgcolor: "#f5f5f5",
			}}
		>
			<Typography
				variant="h1"
				color="primary"
				sx={{ fontWeight: "bold" }}
			>
				404
			</Typography>
			<Typography variant="h5" color="textSecondary" sx={{ mb: 2 }}>
				Page Not Found
			</Typography>
			<Typography variant="body1" color="textSecondary" sx={{ mb: 4 }}>
				The page you are looking for does not exist or has been moved.
			</Typography>
			<Button
				variant="contained"
				color="primary"
				onClick={() => navigate("/")}
				sx={{ textTransform: "none" }}
			>
				Go Back to Home
			</Button>
		</Box>
	);
};

export default PageNotFound;
