import { useEffect } from 'react';

import { isRouteErrorResponse, Links, Meta, Outlet, Scripts, ScrollRestoration } from 'react-router';

import { IS_WAILS_PLATFORM } from '@/lib/features';

import { initBuiltIns } from '@/apis/builtin_provider_cache';
import { initStartupTheme } from '@/apis/builtin_theme_cache';

import Sidebar from '@/components/sidebar';
import { ThemeSwitchProvider } from '@/components/theme';

import '@/globals.css';
import { ensureWorker } from '@/hooks/use_highlight';

import type { Route } from './+types/root';

export const meta: Route.MetaFunction = () => [
	{ title: 'FlexiGPT' },
	{ name: 'description', content: 'The FlexiGPT ecosystem agent' },
];

export function Layout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en" suppressHydrationWarning>
			<head>
				<meta charSet="utf-8" />
				<meta name="viewport" content="width=device-width, initial-scale=1" />
				{IS_WAILS_PLATFORM && (
					<>
						<meta name="wails-options" content="noautoinject" />
						<script src="/wails/ipc.js" />
						<script src="/wails/runtime.js" />
					</>
				)}
				<Meta />
				<Links />
			</head>
			<body className="h-full m-0 p-0 overflow-hidden antialiased">
				{children}
				<ScrollRestoration />
				<Scripts />
			</body>
		</html>
	);
}

export async function clientLoader() {
	// Wait for DOM content to be loaded and Wails runtime to be injected
	if (document.readyState !== 'complete' && document.readyState !== 'interactive') {
		await new Promise(resolve => {
			document.addEventListener('DOMContentLoaded', resolve, { once: true });
		});
	}
	// Now it's safe to call Wails backend functions
	await Promise.all([initBuiltIns(), initStartupTheme()]);
	// console.log('root builtins loaded');
}

// Important! Force the client loader to run during hydration and not just during ssr build.
clientLoader.hydrate = true as const;

export default function Root() {
	// Init worker on mount.
	useEffect(() => {
		if ('requestIdleCallback' in window) requestIdleCallback(() => ensureWorker());
		else setTimeout(() => ensureWorker(), 300);
	}, []);

	return (
		<ThemeSwitchProvider>
			<Sidebar>
				<Outlet />
			</Sidebar>
		</ThemeSwitchProvider>
	);
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
	let message = 'Oops!';
	let details = 'An unexpected error occurred.';
	let stack: string | undefined;

	if (isRouteErrorResponse(error)) {
		message = error.status === 404 ? '404' : 'Error';
		details = error.status === 404 ? 'The requested page could not be found.' : error.statusText || details;
	} else if (import.meta.env.DEV && error && error instanceof Error) {
		details = error.message;
		stack = error.stack;
	}

	return (
		<main className="pt-16 p-4 container mx-auto">
			<h1>{message}</h1>
			<p>{details}</p>
			{stack && (
				<pre className="w-full p-4 overflow-x-auto">
					<code>{stack}</code>
				</pre>
			)}
		</main>
	);
}

export function HydrateFallback() {
	return (
		<div id="loading-splash" className="flex justify-center items-center h-screen w-full flex-col gap-4">
			<div id="loading-splash-spinner" />
			<span className="loading loading-dots loading-xl text-primary-content"></span>
		</div>
	);
}
