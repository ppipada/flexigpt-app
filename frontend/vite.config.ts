import eslintPlugin from '@nabla/vite-plugin-eslint';
import { reactRouter } from '@react-router/dev/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export default defineConfig(({ command, mode }) => {
	const isProd = mode === 'production';
	return {
		plugins: [tailwindcss(), reactRouter(), tsconfigPaths(), eslintPlugin()],
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
				output: {
					format: 'esm',
				},
			},
		},
		// Handle specific problematic packages
		resolve: {
			mainFields: ['module', 'jsnext:main', 'jsnext'],
		},
	};
});
