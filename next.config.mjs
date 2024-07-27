/** @type {import('next').NextConfig} */
const nextConfig = {
	// skipTrailingSlashRedirect: true,
	distDir: 'build',
	basePath: process.env.NODE_ENV === 'production' ? '/frontend/build' : undefined,
	// assetPrefix: process.env.NODE_ENV === 'production' ? '/frontend/build' : undefined,
	images: {
		unoptimized: true,
	},
	output: 'export',
	trailingSlash: true,
	skipTrailingSlashRedirect: true,
};

export default nextConfig;
