/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { useEffect, useMemo, useState } from 'react';

import mermaid, { type MermaidConfig } from 'mermaid';

import { ALL_DARK_THEMES } from '@/spec/theme_consts';

import { useTheme } from '@/hooks/use_theme_provider';

export function useIsDarkMermaid(): boolean {
	const { theme: providerTheme } = useTheme();
	const [prefersDark, setPrefersDark] = useState(false);
	useEffect(() => {
		if (typeof window === 'undefined' || !window.matchMedia) return;
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const update = () => {
			setPrefersDark(mq.matches);
		};
		update();
		mq.addEventListener?.('change', update);
		return () => {
			mq.removeEventListener?.('change', update);
		};
	}, []);

	return useMemo(() => {
		/* “system” → fall back to prefers-color-scheme */
		if (providerTheme === 'system') {
			return prefersDark;
		}
		return ALL_DARK_THEMES.includes(providerTheme as any);
	}, [providerTheme, prefersDark]);
}

type RenderResult = Awaited<ReturnType<typeof mermaid.render>>;

let queue: Promise<unknown> = Promise.resolve();
let lastInitKey: string | null = null;

function makeInitKey(config: MermaidConfig): string {
	// Only include parts that impact SVG output
	const keyObj = {
		theme: config.theme,
		securityLevel: config.securityLevel,
		suppressErrorRendering: config.suppressErrorRendering,
		// themeVariables is common for controlling background etc.
		themeVariables: config.themeVariables ?? null,
	};
	return JSON.stringify(keyObj);
}

export function renderMermaidQueued(id: string, code: string, config: MermaidConfig): Promise<RenderResult> {
	const task = async () => {
		const initKey = makeInitKey(config);
		if (initKey !== lastInitKey) {
			mermaid.initialize(config);
			lastInitKey = initKey;
		}
		return mermaid.render(id, code);
	};

	const result = queue.then(task);
	// Keep queue alive even if one render fails
	queue = result.catch(() => undefined);
	return result;
}
