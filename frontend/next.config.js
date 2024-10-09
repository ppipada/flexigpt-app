/** @type {import('next').NextConfig} */
const nextConfig = {
	distDir: 'dist',
	// basePath: process.env.NODE_ENV === 'production' ? '/frontend/dist' : undefined,
	images: {
		unoptimized: true,
	},
	output: 'export',
	trailingSlash: true,
};

export default nextConfig;
