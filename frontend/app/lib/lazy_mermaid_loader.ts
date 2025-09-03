import type mermaidDefault from 'mermaid';
import type { MermaidConfig } from 'mermaid';

type MermaidApi = typeof mermaidDefault;

let mermaidP: Promise<MermaidApi> | null = null;

export function loadMermaid(): Promise<MermaidApi> {
	if (!mermaidP) {
		mermaidP = (async () => {
			const mod = await import('mermaid');
			const mermaid = (mod as { default: MermaidApi }).default;

			mermaid.initialize({
				startOnLoad: false,
				securityLevel: 'loose',
				suppressErrorRendering: true,
				theme: 'default', // per-diagram theme is injected later
			});

			return mermaid;
		})();
	}

	return mermaidP;
}

export type { MermaidConfig };
