import monokai from '@shikijs/themes/monokai';
import { bundledLanguages, getSingletonHighlighter } from 'shiki';

export const SupportedLanguages = Object.keys(bundledLanguages);

export function isLanguageSupported(lang: string) {
	return lang in bundledLanguages;
}

export const highlighter = await getSingletonHighlighter({
	themes: [monokai],
	langs: SupportedLanguages,
});
