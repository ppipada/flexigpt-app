import { reactRouter } from '@react-router/dev/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import checker from 'vite-plugin-checker';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig(({ mode }) => {
	const isProd = mode === 'production';
	return {
		plugins: [
			tailwindcss(),
			reactRouter(),
			tsconfigPaths(),
			checker({
				typescript: true,
				eslint: {
					lintCommand: 'eslint . -c ./eslint.config.mjs',
					useFlatConfig: true,
				},
			}),
		],
		base: isProd ? '/frontend/dist/' : '/',

		// Add these configurations for better ESM support
		optimizeDeps: {
			include: ['shiki/bundle/full'],
			esbuildOptions: {
				target: 'esnext',
				supported: {
					bigint: true,
				},
			},
		},
		build: {
			outDir: 'dist',
			target: 'esnext',
			rollupOptions: {
				output: { format: 'es' },
			},
		},
		worker: {
			format: 'es',
		},
		// Handle specific problematic packages
		resolve: {
			mainFields: ['module', 'jsnext:main', 'jsnext'],
		},

		test: {
			globals: true, // use `describe/it/expect` without imports
			environment: 'jsdom', // DOM for React component tests
			setupFiles: './vitest.setup.ts',
			coverage: {
				reporter: ['text', 'html'],
				exclude: ['wailsjs', 'dist', 'vite.config.ts'],
			},
		},
	};
});
