// hooks/useIsDarkMermaid.ts
import { useMemo } from 'react';

import { ALL_DARK_THEMES } from '@/spec/theme_consts';

import { useTheme } from '@/hooks/use_theme_provider';

export function useIsDarkMermaid(): boolean {
	const { theme: providerTheme } = useTheme();

	return useMemo(() => {
		/* “system” → fall back to prefers-color-scheme */
		if (providerTheme === 'system') {
			return window.matchMedia('(prefers-color-scheme: dark)').matches;
		}
		return ALL_DARK_THEMES.includes(providerTheme as any);
	}, [providerTheme]);
}
