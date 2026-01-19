// ---- Large-text perf tuning ----
// If you type huge prompts (10k+ words), a single giant text leaf becomes slow.
// We keep ONE paragraph, but chunk its text into multiple text nodes ("leaves").
import type { Value } from 'platejs';
import type { PlateEditor } from 'platejs/react';

// We add a per-chunk prop so Slate doesn't merge them back together.
export const LARGE_TEXT_AUTOCHUNK_THRESHOLD_CHARS = 10000; // start chunking once draft grows beyond this
export const LARGE_TEXT_AUTODECHUNK_THRESHOLD_CHARS = 5000; // merge back once it shrinks below this
export const LARGE_TEXT_CHUNK_SIZE = 1600; // each leaf ~1.2k chars; adjust 800–2000 based on taste

type ChunkedTextNode = { text: string; __chunk: number };

function buildChunkedTextChildren(text: string, chunkSize: number): ChunkedTextNode[] {
	if (!text) return [{ text: '', __chunk: 0 }];
	const out: ChunkedTextNode[] = [];
	let i = 0;
	let idx = 0;
	while (i < text.length) {
		out.push({ text: text.slice(i, i + chunkSize), __chunk: idx++ });
		i += chunkSize;
	}
	return out;
}

export function buildSingleParagraphValue(text: string): Value {
	return [{ type: 'p', children: [{ text }] }];
}

export function buildSingleParagraphValueChunked(text: string, chunkSize: number): Value {
	return [{ type: 'p', children: buildChunkedTextChildren(text, chunkSize) }];
}

export function isCursorAtDocumentEnd(editor: PlateEditor): boolean {
	try {
		const sel = editor.selection;
		if (!sel) return false;
		const end = editor.api.end([]);
		return JSON.stringify(sel.anchor.path) === JSON.stringify(end?.path) && sel.anchor.offset === end?.offset;
	} catch {
		return false;
	}
}
