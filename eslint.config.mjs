import { fixupPluginRules } from '@eslint/compat';
import { FlatCompat } from '@eslint/eslintrc';
import js from '@eslint/js';
import typescriptEslint from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import _import from 'eslint-plugin-import';
import prettier from 'eslint-plugin-prettier';
import globals from 'globals';
import path from 'path';
import tseslint from 'typescript-eslint';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
	baseDirectory: __dirname,
	recommendedConfig: js.configs.recommended,
	allConfig: js.configs.all,
});

export default tseslint.config(
	js.configs.recommended,
	tseslint.configs.recommended,
	{
		ignores: ['node_modules/**', 'dist/**', 'build/**'],
	},
	{
		files: ['**/*.{js,jsx,mjs,cjs,ts,tsx}'],
		plugins: {
			'@typescript-eslint': fixupPluginRules(typescriptEslint),
			import: fixupPluginRules(_import),
			prettier: fixupPluginRules(prettier),
		},
		languageOptions: {
			parser: tsParser,
			parserOptions: {
				ecmaVersion: 'latest',
				sourceType: 'module',
				project: './tsconfig.json',
				ecmaFeatures: {
					jsx: true,
				},
			},
			globals: {
				...globals.browser,
			},
		},
		settings: {
			'import/resolver': {
				typescript: {
					alwaysTryTypes: true,
					project: './tsconfig.json',
				},
			},
		},
		rules: {
			'import/no-unresolved': 'error',
			'prettier/prettier': [
				'error',
				{
					endOfLine: 'auto',
				},
			],
			'@typescript-eslint/consistent-type-imports': ['error'],
			'@typescript-eslint/no-explicit-any': 'off',
		},
	},
	// Use compat to add the remaining plugin configs
	...compat.config({
		extends: [
			'plugin:import/errors',
			'plugin:import/warnings',
			'plugin:import/typescript',
			'plugin:prettier/recommended',
		],
	})
);
