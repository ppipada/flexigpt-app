/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { reactRouter } from '@react-router/dev/vite';
import tailwindcss from '@tailwindcss/vite';
import { visualizer } from 'rollup-plugin-visualizer';
import { defineConfig } from 'vite';
import checker from 'vite-plugin-checker';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig(({ mode }) => {
	const isProd = mode === 'production';
	// const analyze = process.env.ANALYZE === 'true' || !isProd;
	const analyze = false;
	const analyzerPlugin = visualizer({
		open: true,
		gzipSize: true,
		brotliSize: true,
		filename: 'dist/stats.html',
	});
	const rollupPlugins = [];
	if (analyze) {
		rollupPlugins.push(analyzerPlugin);
	}
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
			include: ['shiki', 'sprintf-js', 'mermaid'],
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
				plugins: rollupPlugins,
				output: {
					format: 'es',

					manualChunks(id) {
						// Only vendor modules
						if (!id.includes('node_modules')) return;

						// Cross-platform path normalization
						const n = id.replace(/\\/g, '/');

						// KaTeX (JS). Fonts are emitted separately as assets automatically.
						if (n.includes('/node_modules/katex/')) return 'libkatex';

						if (n.includes('/node_modules/compromise/')) return 'libcompromise';

						// Mermaid and common companions
						if (n.includes('/node_modules/mermaid/')) return 'libmermaid';

						if (/\/node_modules\/(unified|remark(?:-[^/]+)?|rehype(?:-[^/]+)?)\//.test(n)) {
							return 'libunified';
						}

						// PlateJS suite
						if (/\/node_modules\/(@udecode|@platejs|platejs)\//.test(n)) return 'libplate';
						if (/\/node_modules\/@emoji-mart\//.test(n)) return 'libemojimart';

						// Let Rollup decide otherwise
						return;
					},
				},
			},
		},

		worker: {
			format: 'es',
		},

		resolve: {
			// Prefer modern entry points; keep this if it works for you
			// You can also use: ['module', 'browser', 'exports', 'main']
			mainFields: ['module', 'jsnext:main', 'jsnext'],
		},

		test: {
			globals: true,
			environment: 'jsdom',
			setupFiles: './vitest.setup.ts',
			coverage: {
				reporter: ['text', 'html'],
				exclude: ['wailsjs', 'dist', 'vite.config.ts'],
			},
		},
	};
});
