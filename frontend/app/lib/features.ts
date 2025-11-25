export const IS_WAILS_PLATFORM = import.meta.env.VITE_PLATFORM === 'wails';
export const FEATURE_FLAG_DOCUMENT_STORES = import.meta.env.VITE_FEATURE_DOCUMENT_STORES === 'true';
export const FEATURE_FLAG_AGENTS = import.meta.env.VITE_FEATURE_AGENTS === 'true';

function isMacLike(): boolean {
	if (typeof navigator === 'undefined') return false; // SSR / safety
	// eslint-disable-next-line @typescript-eslint/no-deprecated
	const platform = (navigator.platform || '').toLowerCase();
	const ua = navigator.userAgent || '';

	// Covers macOS, iOS (Safari in Wails/WebView will still report these)
	return platform.startsWith('mac') || /iphone|ipad|ipod/.test(platform) || /Mac OS X/.test(ua);
}

export const MOD_LABEL = isMacLike() ? '⌘' : 'Ctrl';
