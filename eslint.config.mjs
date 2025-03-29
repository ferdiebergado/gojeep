import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import css from "@eslint/css";

/** @type {import('eslint').Linter.Config[]} */
export default [
	{ files: ["./web/app/js/**/*.ts"] },
	{ languageOptions: { globals: globals.browser } },
	pluginJs.configs.recommended,
	...tseslint.configs.recommended,
	{
		files: ["./web/app/css/**/*.css"],
		plugins: {
			css,
		},
		language: "css/css",
		languageOptions: {
			tolerant: true,
		},
		rules: {
			"css/no-empty-blocks": "error",
		},
	},
];
