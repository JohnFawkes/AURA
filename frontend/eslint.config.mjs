import tseslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import reactPlugin from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";

export default [
	{
		ignores: ["node_modules/**", ".next/**", "out/**", "build/**", "next-env.d.ts", "tailwind.config.js"],
	},
	{
		files: ["**/*.{ts,tsx,js,jsx}"],
		languageOptions: {
			parser: tsParser,
			parserOptions: {
				ecmaVersion: 2020,
				sourceType: "module",
				ecmaFeatures: { jsx: true },
				project: "./tsconfig.json",
			},
		},
		plugins: {
			"@typescript-eslint": tseslint,
			"react-hooks": reactHooks,
			react: reactPlugin,
		},
		rules: {
			"react/no-unescaped-entities": "off",
			"no-console": "warn",
			"@typescript-eslint/no-unused-vars": "warn",
			"@typescript-eslint/no-unused-expressions": "warn",
			"react-hooks/exhaustive-deps": "warn",
			"react-hooks/rules-of-hooks": "error",
		},
	},
];
