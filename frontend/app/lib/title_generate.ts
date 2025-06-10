/********************************************************************
 *  Title generator – v2
 *  • paragraph-aware, code-aware, language-aware
 *  • ≤ 48 characters
 *******************************************************************/
import nlp from 'compromise';
import { removeStopwords } from 'stopword';

/* ─────────────────────────  Constants  ─────────────────────────── */
const MAX_LEN = 48;
const DEFAULT_TITLE = 'New conversation';

/**
 * Very light-weight “does this line look like code?” heuristic.
 * It is NOT 100 % accurate – it is only meant to be fast and
 * surprisingly good for common cases.
 */
function looksLikeCodeLine(line: string): boolean {
	const t = line.trim();
	if (!t) return false; // blank line

	/* ---------- 1.  Quick “certain” matches -------------------- */

	// single brace / bracket / semicolon lines:  }, );, ]
	if (/^[{}[\]()];?$/.test(t)) return true;

	// comment markers at line start ( //  #  /*  *  <!-- )
	if (/^\s*(\/\/|#|\/\*|\*|<!--)/.test(t)) return true;

	// import / class / def … at beginning of line
	if (/^\s*(import|export|package|namespace|using|class|interface|def|func|lambda|const|async)\b/.test(t)) return true;

	/* ---------- 2.  Count weaker indicators -------------------- */

	let score = 0;

	if (/[{}[$$;]/.test(t)) score++; // braces / semicolon
	if (/=>|->|::|\+\+|--|&&|\|\||\?\.|\?\?/.test(t)) score++; // multi-char operators
	if (/[=+\-*/%]=?\s*\w/.test(t)) score++; // assignments / operators
	if (/\w+\s*$[^)]*$/.test(t)) score++; // foo(…)  or  bar()
	if (/(['"`]).*\1/.test(t)) score++; // string literal

	/* ---------- 3.  Decision ----------------------------------- */

	// At least two weak indicators → probably code
	return score >= 2;
}

function isCodeParagraph(p: string): boolean {
	const lines = p.split(/\n/);
	const codeLike = lines.filter(looksLikeCodeLine).length;
	return codeLike >= lines.length / 2; // ≥ 50 % lines look like code
}

/* remove fenced, indented and inline code + crude HTML */
function stripCode(txt: string): string {
	return (
		txt
			// ``` … ```  or  ~~~ … ~~~
			.replace(/```[\s\S]*?```|~~~[\s\S]*?~~~/g, '')
			// 4-space / tab indented blocks
			.replace(/^(?: {4}|\t).*\n?/gm, '')
			// inline `code`
			.replace(/`[^`\n]*`/g, '')
			// very simple HTML tags
			.replace(/<\/?[a-z][^>]*>/gi, '')
	);
}

/* ───────────────────────  Splitters  ───────────────────────────── */
function splitParagraphs(txt: string): string[] {
	return txt
		.split(/\n{2,}/)
		.map(p => p.trim())
		.filter(Boolean);
}

function splitSentences(txt: string): string[] {
	return txt
		.split(/\n+/) // keep single \n as separator
		.flatMap(p => p.split(/(?<=[.!?])\s+/))
		.map(s => s.trim())
		.filter(Boolean);
}

/* ──────────────────  Candidate paragraph selection  ────────────── */
function selectCandidateParagraphs(paragraphs: string[]): string[] {
	const nonCode = paragraphs.filter(p => !isCodeParagraph(p));

	// ① first two + last two non-code paragraphs if possible
	if (nonCode.length >= 4) return [...nonCode.slice(0, 2), ...nonCode.slice(-2)];

	// ② otherwise take whatever non-code we have
	if (nonCode.length) return nonCode;

	// ③ nothing but code → fall back to very first paragraph
	return [paragraphs[0]];
}

/* ─────────────────────  Language detector  ─────────────────────── */
function isEnglish(text: string): boolean {
	const cleanedTxt = stripCode(text); // remove code again
	const tokens = cleanedTxt.toLowerCase().match(/\b[a-z]{2,}\b/g) || [];
	if (!tokens.length) return false;

	const noStop: string[] = removeStopwords(tokens);
	const stopRatio = 1 - noStop.length / tokens.length;

	return stopRatio > 0.05; // heuristic threshold
}

/* ─────────────────────  Scoring heuristics  ────────────────────── */
function scoreEn(s: string, idx: number, total: number): number {
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

function pickBestEnglish(sentences: string[]): string {
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

function pickBestGeneric(sentences: string[]): string {
	const first = sentences[0] ?? '';
	const last = sentences[sentences.length - 1] ?? '';
	const short = [...sentences].sort((a, b) => a.length - b.length)[0] ?? '';
	return short.length <= 60 ? short : first.length <= 60 ? first : last;
}

/* ───────────────────────  Final polish  ────────────────────────── */

function sentenceCase(str: string) {
	if (!str) return '';
	const match = str.match(/^(\s*)(\S)(.*)$/);
	if (!match) return str;
	const [, leading, first, rest] = match;
	return leading + first.toUpperCase() + rest;
}

function finalise(raw: string): string {
	let t = raw
		// salutations
		.replace(/^(hi|hello|hey|dear|greetings|good (morning|afternoon|evening))[,!:.\s-]+/i, '')
		// emails & phones
		.replace(/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}/gi, '')
		.replace(/\+?\d[\d\s.-]{7,}/g, '')
		// whitespace
		.replace(/\s+/g, ' ')
		.trim()
		// leading / trailing punctuation
		.replace(/^[\s:;,-]+|[\s:;,.?!-]+$/g, '');

	if (!t) return DEFAULT_TITLE;

	if (t.length > MAX_LEN) {
		const cut = t.slice(0, MAX_LEN);
		const lastSpace = cut.lastIndexOf(' ');
		t = lastSpace > 25 ? cut.slice(0, lastSpace) : cut;
	}

	return sentenceCase(t);
}

/* ──────────────────────────  Public API  ───────────────────────── */
export function generateTitle(firstMessage: string): string {
	if (!firstMessage.trim()) return DEFAULT_TITLE;

	const paragraphs = splitParagraphs(firstMessage.trim());
	if (!paragraphs.length) return DEFAULT_TITLE;

	/* ① choose candidate paragraphs */
	const candidates = selectCandidateParagraphs(paragraphs);
	// console.log('candidates', JSON.stringify(candidates, null, 2));

	/* ② collect sentences & drop code-looking lines */
	const sentences = candidates.flatMap(p => splitSentences(p)).filter(l => !looksLikeCodeLine(l));
	// console.log('sentences', JSON.stringify(sentences, null, 2));
	if (!sentences.length)
		// still nothing – just clean first paragraph
		return finalise(stripCode(paragraphs[0]));

	/* ③ language detection & scoring */
	const english = isEnglish(sentences.join(' '));
	const best = english ? pickBestEnglish(sentences) : pickBestGeneric(sentences);

	/* ④ final polish */
	return finalise(best);
}
