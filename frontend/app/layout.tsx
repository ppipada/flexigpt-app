import Sidebar from '@/components/Sidebar';
import { ThemeSwitchProvider } from '@/providers/themeProvider';
import '@/styles/globals.css';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';

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
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={inter.className}>
				<ThemeSwitchProvider>
					<Sidebar>{children}</Sidebar>
				</ThemeSwitchProvider>
			</body>
		</html>
	);
}
