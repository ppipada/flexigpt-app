import nlp from 'compromise';
import { removeStopwords } from 'stopword';

const MAX_LEN = 48;
const DEFAULT_TITLE = 'New conversation';

function stripCode(txt: string) {
	return txt
		.replace(/```[\s\S]*?```/g, '') // fenced
		.replace(/`[^`]*`/g, '') // inline
		.replace(/<\/?[\w\s="'-]+>/g, ''); // crude HTML
}

function splitSentences(t: string) {
	return t
		.split(/\n{2,}/) // paragraphs
		.flatMap(p => p.split(/(?<=[.!?])\s+/))
		.map(s => s.trim())
		.filter(Boolean);
}

function isEnglish(text: string): boolean {
	const tokens = text.toLowerCase().match(/\b[a-z]{2,}\b/g) || [];
	if (!tokens.length) return false;

	const cleanedTokens: string[] = removeStopwords(tokens);
	const stopRatio = 1 - cleanedTokens.length / tokens.length;

	const letters = (text.match(/\p{L}/gu) || []).length;
	const asciiLetters = (text.match(/[A-Za-z]/g) || []).length;
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

function scoreEn(s: string, idx: number, total: number) {
	const t = s.trim();
	let pts = 0;
	if (t.includes('?')) pts += 3;
	if (/^(how|what|why|when|where|can|does|do|is|are|should|could|would|will)\b/i.test(t)) pts += 2;
	if (/\b(error|issue|bug|fail|help)\b/i.test(t)) pts += 1.5;
	if (idx === 0 || idx === total - 1) pts += 1;
	if (/^(\s*[-*]\s+|\s*\d+\.\s+)/.test(t)) pts += 1;
	if (t.length < 60) pts += 1;

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
	let t = raw
		.replace(/^(hi|hello|hey|dear|greetings|good (morning|afternoon|evening))[,!:.\s-]+/i, '')
		.replace(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi, '') // emails
		.replace(/\+?\d[\d\s.-]{7,}/g, '') // phones
		.replace(/\s+/g, ' ')
		.trim()
		.replace(/^[\s:;,-]+|[\s:;,.?!-]+$/g, '');

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
