// eslint.config.js
import js from "@eslint/js";
import prettier from "eslint-config-prettier";
import importPlugin from "eslint-plugin-import";
import jsxA11y from "eslint-plugin-jsx-a11y";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import tseslint from "typescript-eslint";

export default [
	/* =========================
     Global ignores
     ========================= */
	{
		ignores: ["node_modules/**", ".next/**", "dist/**", "coverage/**", "*.config.*"],
	},

	/* =========================
     Base JS rules
     ========================= */
	js.configs.recommended,

	/* =========================
     TypeScript (NON type-aware, fast)
     ========================= */
	...tseslint.configs.recommended,

	/* =========================
     TypeScript + React (type-aware, scoped)
     ========================= */
	{
		files: ["src/**/*.{ts,tsx}"],
		languageOptions: {
			parserOptions: {
				project: "./tsconfig.eslint.json",
				tsconfigRootDir: import.meta.dirname,
			},
		},
		plugins: {
			react,
			"react-hooks": reactHooks,
			"jsx-a11y": jsxA11y,
			import: importPlugin,
		},
		settings: {
			react: { version: "detect" },
			"import/resolver": {
				typescript: {
					project: "./tsconfig.eslint.json",
				},
			},
		},
		rules: {
			/* =========================
         React
         ========================= */
			"react/react-in-jsx-scope": "off",
			"react/jsx-uses-react": "off",

			/* =========================
         Hooks
         ========================= */
			"react-hooks/rules-of-hooks": "error",
			"react-hooks/exhaustive-deps": "warn",

			/* =========================
         TypeScript
         ========================= */
			"@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_", varsIgnorePattern: "^_" }],
			"@typescript-eslint/no-explicit-any": "warn",

			/* =========================
         Imports
         ========================= */
			"import/order": [
				"warn",
				{
					groups: ["builtin", "external", "internal", "parent", "sibling", "index", "type"],
					"newlines-between": "always",
					alphabetize: { order: "asc", caseInsensitive: true },
				},
			],
		},
	},

	/* =========================
     Prettier (ALWAYS LAST)
     ========================= */
	prettier,
];
