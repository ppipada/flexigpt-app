import { ThemeProvider } from 'next-themes';

export function ThemeSwitchProvider({ children }: { children: React.ReactNode }) {
	return (
		<ThemeProvider attribute="data-theme" value={{ light: 'nord', dark: 'dracula' }} defaultTheme="system" enableSystem>
			{children}
		</ThemeProvider>
	);
}
