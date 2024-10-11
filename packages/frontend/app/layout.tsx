import Sidebar from '@/components/sidebar';
import { ThemeSwitchProvider } from '@/providers/theme_provider';
import '@/styles/globals.css';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import Script from 'next/script';
const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
	title: 'FlexiGPT UI',
	description: 'UI for FlexiGPT ecosystem',
};

export default function RootLayout({
	children,
}: Readonly<{
	children: React.ReactNode;
}>) {
	const isWailsPlatform = process.env.NEXT_PUBLIC_PLATFORM === 'wails';

	return (
		<html lang="en" suppressHydrationWarning>
			{isWailsPlatform && (
				<head>
					<>
						<meta name="wails-options" content="noautoinject" />
						<Script src="/wails/ipc.js" strategy="beforeInteractive" />
						<Script src="/wails/runtime.js" strategy="beforeInteractive" />
					</>
				</head>
			)}

			<body className={inter.className}>
				<ThemeSwitchProvider>
					<Sidebar>{children}</Sidebar>
				</ThemeSwitchProvider>
			</body>
		</html>
	);
}
