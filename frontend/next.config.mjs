/** @type {import('next').NextConfig} */
const nextConfig = {
	distDir: 'build',
	// basePath: process.env.NODE_ENV === 'production' ? '/frontend/build' : undefined,
	images: {
		unoptimized: true,
	},
	output: 'export',
	trailingSlash: true,
};

export default nextConfig;
