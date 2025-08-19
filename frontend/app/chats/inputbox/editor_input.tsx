/* eslint-disable @typescript-eslint/no-unsafe-member-access */
import type { KeyboardEvent, RefObject } from 'react';
import React, { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';

import { SingleBlockPlugin, type Value } from 'platejs';
import { Plate, PlateContent, usePlateEditor } from 'platejs/react';
import { FiSend } from 'react-icons/fi';

import { AlignKit } from '@/components/editor/plugins/align_kit';
import { AutoformatKit } from '@/components/editor/plugins/auto_format_kit';
import { BasicBlocksKit } from '@/components/editor/plugins/basic_blocks_kit';
import { BasicMarksKit } from '@/components/editor/plugins/basic_marks_kit';
import { EmojiKit } from '@/components/editor/plugins/emoji_kit';
import { IndentKit } from '@/components/editor/plugins/indent_kit';
import { LineHeightKit } from '@/components/editor/plugins/line_height_kit';
import { ListKit } from '@/components/editor/plugins/list_kit';

export interface EditorTextInputHandle {
	focus: () => void;
}

interface EditorTextInputProps {
	isBusy: boolean;
	onSubmit: (text: string) => void;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
}

const EMPTY_VALUE: Value = [{ type: 'p', children: [{ text: '' }] }];

function expandTabs(line: string, tabSize = 2) {
	let out = '';
	let col = 0;
	for (const ch of line) {
		if (ch === '\t') {
			const n = tabSize - (col % tabSize);
			out += ' '.repeat(n);
			col += n;
		} else {
			out += ch;
			col = ch === '\n' ? 0 : col + 1;
		}
	}
	return out;
}

function insertPlainTextAsSingleBlock(editor: ReturnType<typeof usePlateEditor>, text: string, tabSize = 2) {
	if (!editor) {
		return;
	}
	const normalized = text.replace(/\r\n?/g, '\n');
	const lines = normalized.split('\n').map(l => expandTabs(l, tabSize));

	editor.tf.insertText(lines[0] ?? '');
	for (let i = 1; i < lines.length; i++) {
		if (editor.tf.insertSoftBreak) {
			editor.tf.insertSoftBreak();
		} else {
			// Fallback: literal newline works with white-space: break-spaces
			editor.tf.insertText('\n');
		}
		editor.tf.insertText(lines[i]);
	}
}

function useEnterSubmit() {
	const formRef = useRef<HTMLFormElement>(null);

	const onKeyDown = (
		e: KeyboardEvent<HTMLDivElement>,
		opts: {
			isBusy: boolean;
			getCanSubmit: () => boolean;
			editorRef: RefObject<ReturnType<typeof usePlateEditor> | null>;
		}
	) => {
		// Let Plate plugins (e.g., mentions) handle the key if they already prevented it
		if ((e as any).defaultPrevented) return;

		const native = e.nativeEvent as KeyboardEvent['nativeEvent'] & { isComposing?: boolean };
		const isComposing = native.isComposing;

		// Shift+Enter => insert a soft break (new line) instead of submitting
		if (e.key === 'Enter' && e.shiftKey && !isComposing) {
			// Try to insert a soft break if supported. Otherwise, allow default behavior.
			opts.editorRef.current?.tf?.insertSoftBreak?.();
			e.preventDefault();
			return;
		}

		// Enter (no shift) => submit
		if (e.key === 'Enter' && !e.shiftKey && !isComposing) {
			if (!opts.isBusy && opts.getCanSubmit()) {
				formRef.current?.requestSubmit();
			}
			e.preventDefault();
		}
	};

	return { formRef, onKeyDown };
}

const EditorTextInput = forwardRef<EditorTextInputHandle, EditorTextInputProps>(
	({ isBusy, onSubmit, setInputHeight }, ref) => {
		const editor = usePlateEditor({
			plugins: [
				SingleBlockPlugin,
				...BasicBlocksKit,
				...BasicMarksKit,
				...LineHeightKit,
				...AlignKit,
				...EmojiKit,
				...IndentKit,
				...ListKit,
				...AutoformatKit,
			],

			value: EMPTY_VALUE,
		});

		// Track plain text for enabling/disabling send button
		const [plainText, setPlainText] = useState<string>('');

		const isSubmittingRef = useRef<boolean>(false);
		const contentRef = useRef<HTMLDivElement | null>(null);
		const editorRef = useRef(editor);
		editorRef.current = editor; // keep a live ref for key handlers

		const { formRef, onKeyDown } = useEnterSubmit();

		// Height sync using ResizeObserver on content editable container
		useEffect(() => {
			if (!contentRef.current) return;

			const el = contentRef.current;
			const ro = new ResizeObserver(entries => {
				for (const entry of entries) {
					const rect = entry.target.getBoundingClientRect();
					setInputHeight(Math.ceil(rect.height));
				}
			});

			ro.observe(el);
			return () => {
				ro.disconnect();
			};
		}, [setInputHeight]);

		const isSendButtonEnabled = plainText.trim().length > 0;

		const handleSubmit = (e?: React.FormEvent) => {
			if (e) e.preventDefault();
			if (!isSendButtonEnabled || isSubmittingRef.current) return;

			isSubmittingRef.current = true;
			// const textToSend = editor.api.markdown.serialize().trim();
			const textToSend = editor.api.string([]);

			try {
				onSubmit(textToSend);
			} finally {
				// Clear editor and state after submitting
				// Prefer setValue over reset so we truly empty the content
				editor.tf.setValue(EMPTY_VALUE);
				// Focus back into the editor
				editor.tf.focus();
				isSubmittingRef.current = false;
			}
		};

		useImperativeHandle(ref, () => ({
			focus: () => {
				editor.tf.focus();
			},
		}));

		return (
			<form
				ref={formRef}
				onSubmit={handleSubmit}
				className="bg-base-100 border-base-300 focus-within:border-base-400 mx-2 flex items-center rounded-2xl border px-4"
			>
				<Plate
					editor={editor}
					onChange={() => {
						setPlainText(editor.api.string([]));
					}}
				>
					<PlateContent
						ref={contentRef}
						placeholder="Type message..."
						spellCheck={false}
						readOnly={isBusy}
						onKeyDown={e => {
							onKeyDown(e as unknown as KeyboardEvent<HTMLDivElement>, {
								isBusy,
								getCanSubmit: () => plainText.trim().length > 0,
								editorRef,
							});
						}}
						onPaste={e => {
							e.preventDefault();
							e.stopPropagation();
							const text = e.clipboardData.getData('text/plain');
							if (!text) return;
							insertPlainTextAsSingleBlock(editor, text);
							e.clipboardData.clearData('text/plain');
							e.clipboardData.clearData('text/html');
						}}
						className="max-h-64 min-h-[24px] flex-1 resize-none overflow-auto bg-transparent p-2 whitespace-break-spaces outline-none [tab-size:2] focus:outline-none"
						style={{
							fontSize: '14px',
							whiteSpace: 'break-spaces',
							tabSize: 2,
						}}
					/>
				</Plate>

				<button
					type="submit"
					className={`btn btn-md border-none !bg-transparent px-1 shadow-none ${
						!isSendButtonEnabled || isBusy ? 'btn-disabled' : ''
					}`}
					disabled={isBusy || !isSendButtonEnabled}
					aria-label="Send Message"
					title="Send Message"
				>
					<FiSend size={20} />
				</button>
			</form>
		);
	}
);

EditorTextInput.displayName = 'EditorTextInput';

export default EditorTextInput;
