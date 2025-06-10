export const codeMessage = `
analyze and give a review of below code

import nlp from 'compromise';
import { removeStopwords } from 'stopword';

const MAX_LEN = 48;
const DEFAULT_TITLE = 'New conversation';


function isEnglish(text: string): boolean {
	const tokens = text.toLowerCase().match(/\b[a-z]{2,}\b/g) || [];
	if (!tokens.length) return false;

	const cleanedTokens: string[] = removeStopwords(tokens);
	const stopRatio = 1 - cleanedTokens.length / tokens.length;

	const asciiRatio = letters ? asciiLetters / letters : 0;

	return stopRatio > 0.05 && asciiRatio > 0.65;
}

function pickBestEnglish(sentences: string[]) {
	let best = '';
	let score = -Infinity;
	sentences.forEach((s, i) => {
		const pts = scoreEn(s, i, sentences.length);
		if (pts > score) {
			best = s;
			score = pts;
		}
	});
	return best;
}

	const doc = nlp(t);
	const uniq = new Set([...doc.nouns().out('array'), ...doc.verbs().out('array')]);
	pts += 0.2 * uniq.size;

	return pts;
}

// Non-English fallback
function pickBestGeneric(sentences: string[]) {
	const first = sentences[0] ?? '';
	const last = sentences[sentences.length - 1] ?? '';
	const short = [...sentences].sort((a, b) => a.length - b.length)[0] ?? '';
	return short.length <= 60 ? short : first.length <= 60 ? first : last;
}

// Final tidy-up
function finalise(raw: string) {

	if (!t) return DEFAULT_TITLE;

	if (t.length > MAX_LEN) {
		const cut = t.slice(0, MAX_LEN);
		const lastSpace = cut.lastIndexOf(' ');
		t = lastSpace > 25 ? cut.slice(0, lastSpace) : cut;
	}

	return sentenceCase(t);
}

function sentenceCase(str: string) {
	return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}

export function generateTitle(firstMessage: string): string {
	if (!firstMessage.trim()) return DEFAULT_TITLE;

	const body = stripCode(firstMessage).trim();

	// Already short? just polish & return
	if (body.length <= MAX_LEN) return finalise(body);

	const sentences = splitSentences(body);
	if (!sentences.length) return finalise(body.slice(0, MAX_LEN));

	const english = isEnglish(body);
	const best = english ? pickBestEnglish(sentences) : pickBestGeneric(sentences);

	return finalise(best);
}

try to find bugs too
`;
