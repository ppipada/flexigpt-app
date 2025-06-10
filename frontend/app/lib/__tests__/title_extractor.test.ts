import { describe, expect, it } from 'vitest';

import { generateTitle } from '../title_generate';
import { codeMessage } from './sample_code_message';

const DEFAULT = 'New conversation';
const MAX_PREVIEW = 120; // truncate long log lines

/* ─── debug printer ──────────────────────────────────────────────── */
function debugPrint(label: string, input: string, output: string) {
	const crop = (t: string) => (t.length > MAX_PREVIEW ? t.slice(0, MAX_PREVIEW) + ' …' : t);
	console.log(
		`\n${label}\n` + `  input (${input.length}):  ${crop(input)}\n` + `  title (${output.length}):  ${output}\n`
	);
}

/* ─── base behaviour ─────────────────────────────────────────────── */
describe('generateTitle - base cases', () => {
	it('falls back to default on empty input', () => {
		const res = generateTitle('');
		debugPrint('Empty string', '', res);
		expect(res).toBe(DEFAULT);
	});

	it('keeps short single-sentence prompts (sentence-cased)', () => {
		const txt = 'how to install react?';
		const res = generateTitle(txt);
		debugPrint('Short question', txt, res);
		expect(res).toBe('How to install react');
	});
});

/* ─── rich English examples ──────────────────────────────────────── */
describe('generateTitle - rich English path', () => {
	it('picks the best question in a paragraph block', () => {
		const msg = `
Hello!
I keep getting "Cannot set headers after they are sent".
How can I resolve this error?
`;
		const title = generateTitle(msg);
		debugPrint('Error paragraph', msg, title);
		expect(title).toBe('How can I resolve this error');
	});

	it('ignores fenced/inline code while scoring', () => {
		const msg = `
\`\`\`python
raise KeyError('id')
\`\`\`
Why does this KeyError happen?
`;
		const title = generateTitle(msg);
		debugPrint('Code + question', msg, title);
		expect(title).toBe('Why does this KeyError happen');
	});
});

/* ─── non-English fallback ───────────────────────────────────────── */
describe('generateTitle - non-English / generic path', () => {
	it('still returns ≤48 chars for Spanish input', () => {
		const esp = 'Hola, ¿cómo puedo centrar un div en CSS?';
		const title = generateTitle(esp);
		debugPrint('Spanish question', esp, title);
		expect(title.length).toBeLessThanOrEqual(48);
	});
});

/* ─── hard limits & casing ───────────────────────────────────────── */
describe('generateTitle - hard limits', () => {
	it('never exceeds 48 characters', () => {
		const long = 'Hello! '.repeat(100);
		const title = generateTitle(long);
		debugPrint('Long repeated input', long, title);
		expect(title.length).toBeLessThanOrEqual(48);
	});

	it('output is sentence-case', () => {
		const msg = 'WHY AM I SEEING ECONNREFUSED WHEN I DEPLOY?';
		const title = generateTitle(msg);
		debugPrint('All-caps input', msg, title);
		expect(title[0]).toBe(title[0].toUpperCase());
		expect(title.slice(1)).toBe(title.slice(1).toUpperCase());
	});
});

/* ─── code sample → rectify request ─────────────────────────────── */
describe('generateTitle - code sample + rectify request', () => {
	it('extracts the final “how do I fix” question', () => {
		const msg = `
Hi team,

Below code crashes:

\`\`\`python
data = {'a': 1}
print(data['b'])   # KeyError here
\`\`\`

It raises KeyError: 'b'.

How can I fix the code so it prints 0 when the key is missing, without using try/except?
`;
		const title = generateTitle(msg);
		debugPrint('KeyError rectify', msg, title);
		expect(title.length).toBeLessThanOrEqual(48);
		expect(title.toLowerCase()).toMatch(/fix|avoid|keyerror/);
		expect(title).not.toBe(DEFAULT);
	});
});

/* ─── generic factual Q&A (fruit) ───────────────────────────────── */
describe('generateTitle - generic factual Q&A (fruit)', () => {
	it('produces a concise, topic-rich title', () => {
		const msg = `
Hey!

I've been reading about different fruits.

Could you explain why bananas have so much potassium and how that
benefits human health?  A short answer is fine.
`;
		const title = generateTitle(msg);
		debugPrint('Banana question', msg, title);
		expect(/benefits/.test(title.toLowerCase())).toBe(true);
	});
});

/* ─── paragraph-aware edge cases ─────────────────────────────────── */
it('ignores plain code lines inside a paragraph', () => {
	const msg = `
analyze the code below
processData()
return result
please list any obvious bugs?
`;
	const title = generateTitle(msg);
	debugPrint('Plain code lines', msg, title);
	expect(title).toBe('Please list any obvious bugs');
});

it('multi-line short prompts keep both lines', () => {
	const msg = 'give me a quote\ngive author details too';
	const title = generateTitle(msg);
	debugPrint('Short multi-line', msg, title);
	expect(title).toBe('Give me a quote');
});

it('prefers the last paragraph question over earlier notes', () => {
	const msg = `
Random note at top.

someFunction();
anotherCall();

Finally, how can I optimise this without recursion?
`;
	const title = generateTitle(msg);
	debugPrint('Last paragraph wins', msg, title);
	expect(title.toLowerCase()).toMatch(/optimise|optimize/);
});

it('min actual lines around code', () => {
	const title = generateTitle(codeMessage);
	debugPrint('Plain code lines', codeMessage, title);
	expect(title).toBe('Analyze and give a review of below code');
});
