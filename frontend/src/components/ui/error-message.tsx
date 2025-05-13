import React from "react";

const ErrorMessage: React.FC<{ message: string }> = ({ message }) => {
	return (
		<div className="flex flex-col items-center justify-center mt-10">
			{message && (
				<p className="text-red-500 text-lg font-semibold text-center">
					{message}
				</p>
			)}
		</div>
	);
};

export default ErrorMessage;
