import React from "react";
import Typography from "@mui/material/Typography";

const ErrorMessage: React.FC<{ message: string }> = ({ message }) => {
	return (
		<>
			{message && (
				<Typography variant="h6" color="error">
					{message}
				</Typography>
			)}
		</>
	);
};

export default ErrorMessage;
