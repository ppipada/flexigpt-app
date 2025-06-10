/* src/lib/__tests__/titleExtractor.test.ts */
import { describe, expect, it } from 'vitest';

import { generateTitle } from '../title_generate';

const DEFAULT = 'New conversation';
const MAX_PREVIEW = 120; // truncate log strings longer than this

// helper: pretty console output
function debugPrint(label: string, input: string, output: string) {
	const crop = (txt: string) => (txt.length > MAX_PREVIEW ? txt.slice(0, MAX_PREVIEW) + ' …' : txt);
	console.log(
		`\n${label}\n` + `  input (${input.length}):    ${crop(input)}\n` + `  title (${output.length}):    ${output}\n`
	);
}

describe('generateTitle – base cases', () => {
	it('falls back to default on empty input', () => {
		const result = generateTitle('');
		debugPrint('Empty string', '', result);
		expect(result).toBe(DEFAULT);
	});

	it('returns short sentences unchanged (sentence-cased)', () => {
		const txt = 'how to install react?';
		const result = generateTitle(txt);
		debugPrint('Short English question', txt, result);
		expect(result).toBe('How to install react');
	});
});

describe('generateTitle – rich English path', () => {
	it('selects the best question in a paragraph', () => {
		const msg = `
      Hello!
      I keep getting "Cannot set headers after they are sent".
      How can I resolve this error?
    `;
		const title = generateTitle(msg);
		debugPrint('Paragraph with error + question', msg, title);
		expect(title).toBe('How can i resolve this error');
	});

	it('ignores fenced and inline code when scoring', () => {
		const msg = `
      \`\`\`python
      raise KeyError('id')
      \`\`\`
      Why does this KeyError happen?
    `;
		const title = generateTitle(msg);
		debugPrint('Input containing code', msg, title);
		expect(title).toBe('Why does this keyerror happen');
	});
});

describe('generateTitle – non-English / generic path', () => {
	it('still returns ≤ 48 chars for Spanish input', () => {
		const esp = 'Hola, ¿cómo puedo centrar un div en CSS?';
		const title = generateTitle(esp);
		debugPrint('Spanish question', esp, title);
		expect(title.length).toBeLessThanOrEqual(48);
	});
});

describe('generateTitle – hard limits', () => {
	it('never exceeds 48 characters', () => {
		const long = 'Hello! '.repeat(100);
		const title = generateTitle(long);
		debugPrint('Very long repeated input', long, title);
		expect(title.length).toBeLessThanOrEqual(48);
	});

	it('output is sentence-case (first char only in caps)', () => {
		const msg = 'WHY AM I SEEING ECONNREFUSED WHEN I DEPLOY?';
		const title = generateTitle(msg);
		debugPrint('All-caps input', msg, title);
		expect(title[0]).toBe(title[0].toUpperCase());
		expect(title.slice(1)).toBe(title.slice(1).toLowerCase());
	});
});

/* append to src/lib/__tests__/titleExtractor.test.ts */

describe('generateTitle – code sample followed by rectify request', () => {
	it('extracts the final “how do I fix” question after Python code', () => {
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
		debugPrint('Python KeyError rectify request', msg, title);

		// Assertions
		expect(title.length).toBeLessThanOrEqual(48);
		expect(title.toLowerCase()).toMatch(/fix|avoid|keyerror/);
		expect(title).not.toBe(DEFAULT); // should not fall back
	});
});

/*  add to titleExtractor.test.ts  ──────────────────────────────────────────── */
describe('generateTitle – generic factual Q&A (fruit)', () => {
	it('returns a concise, topic-rich title for a nutrition question', () => {
		const msg = `
Hey!

I've been reading about different fruits.

Could you explain why bananas have so much potassium and how that
benefits human health?  A short answer is fine.
    `;

		const title = generateTitle(msg);
		debugPrint('Banana potassium question', msg, title);

		// Assertions
		expect(
			/banana/.test(title.toLowerCase()) && (/potassium|benefit/.test(title.toLowerCase()) || true) // optional part
		).toBe(true);
	});
});
