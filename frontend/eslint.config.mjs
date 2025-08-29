import js from '@eslint/js';
import prettier from 'eslint-plugin-prettier';
import eslintPluginPrettierRecommended from 'eslint-plugin-prettier/recommended';
import react from 'eslint-plugin-react';
import globals from 'globals';
import path from 'path';
import tseslint, { configs } from 'typescript-eslint';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export default tseslint.config(
	{
		ignores: ['**/dist/**', '**/app/apis/wailsjs/**', '**/.react-router/**'],
	},
	js.configs.recommended,
	configs.strictTypeChecked,
	eslintPluginPrettierRecommended,

	// ...tailwind.configs['flat/recommended'],
	{
		files: ['**/*.{js,jsx,mjs,ts,tsx}'],
		plugins: {
			prettier,
			react,
		},
		languageOptions: {
			parserOptions: {
				ecmaFeatures: {
					jsx: true,
				},
				projectService: true,
				tsconfigRootDir: import.meta.dirname,
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
			// Add React settings
			react: {
				version: 'detect',
			},
			// tailwindcss: {
			// 	config: path.resolve(__dirname, './app/globals.css'),
			// },
		},
		rules: {
			'no-restricted-imports': [
				'error',
				{
					patterns: [
						{
							group: ['./', '../'],
							message: 'Relative imports are not allowed.',
						},
					],
				},
			],
			'@typescript-eslint/consistent-type-imports': 'error',
			'@typescript-eslint/no-explicit-any': 'off',
			'@typescript-eslint/no-unsafe-assignment': 'off',
			'@typescript-eslint/no-unsafe-call': 'off',
			'@typescript-eslint/no-unsafe-argument': 'off',
			'@typescript-eslint/restrict-template-expressions': 'off',
			'@typescript-eslint/prefer-promise-reject-errors': 'off',
			'@typescript-eslint/require-await': 'off',
			'@typescript-eslint/no-floating-promises': 'off',
			'@typescript-eslint/no-misused-promises': 'off',
			'@typescript-eslint/no-unused-vars': [
				'error',
				{
					argsIgnorePattern: '^_',
					varsIgnorePattern: '^_',
					caughtErrorsIgnorePattern: '^_',
				},
			],
			'prettier/prettier': [
				'error',
				{
					endOfLine: 'auto',
				},
				{
					usePrettierrc: true,
					fileInfoOptions: {
						// Use path.resolve to get the absolute path to the .prettierignore file
						ignorePath: path.resolve(__dirname, '../.prettierignore'),
					},
				},
			],
			// // There is some issue with daisyui classname detection with eslint tailwind 4 beta
			// 'tailwindcss/no-custom-classname': 'off',
			// // We use prettier for classname order
			// 'tailwindcss/classnames-order': 'off',
		},
	}
);
