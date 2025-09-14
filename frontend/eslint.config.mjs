import { FlatCompat } from "@eslint/eslintrc";
import { dirname } from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
	baseDirectory: __dirname,
});

const eslintConfig = [
	{
		ignores: ["node_modules/**", ".next/**", "out/**", "build/**", "next-env.d.ts"],
	},
	...compat.extends("next/core-web-vitals", "next/typescript"),
	{
		// Disable no unescaped-entities
		rules: {
			"react/no-unescaped-entities": "off",
			"no-console": "warn",
			"@typescript-eslint/no-unused-vars": "warn",
			"@typescript-eslint/no-unused-expressions": "warn",
		},
	},
];

export default eslintConfig;
