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

// Convert Plate Value (nodes) to plain text for chat submission and button state.
function nodesToPlainText(nodes: Value): string {
	const walk = (node: any): string => {
		if (node.text !== undefined) {
			return String(node.text ?? '');
		}
		if (Array.isArray(node.children)) {
			return String(node.children.map(walk).join(''));
		}
		return '';
	};

	return nodes.map((n, i) => walk(n) + (i < nodes.length - 1 ? '\n' : '')).join('');
}

const EMPTY_VALUE: Value = [{ type: 'p', children: [{ text: '' }] }];

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
				...AutoformatKit,
				...EmojiKit,
				...IndentKit,
				...ListKit,
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
			const textToSend = plainText.trim();

			try {
				onSubmit(textToSend);
			} finally {
				// Clear editor and state after submitting
				// Prefer setValue over reset so we truly empty the content
				editor.tf.setValue(EMPTY_VALUE);
				setPlainText('');
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
				className="flex items-center bg-base-100 rounded-2xl border border-base-300 focus-within:border-base-400 px-4 mx-2"
			>
				<Plate
					editor={editor}
					onChange={({ value }) => {
						// Keep form state light: only compute a plain string for chat use
						setPlainText(nodesToPlainText(value));
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
						className="
              flex-1 overflow-auto resize-none bg-transparent outline-none focus:outline-none
              min-h-[24px] p-2
              max-h-64
            "
						style={{ fontSize: '14px' }}
					/>
				</Plate>

				<button
					type="submit"
					className={`btn btn-md !bg-transparent border-none shadow-none px-1 ${
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
