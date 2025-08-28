import React, { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';

import { SingleBlockPlugin, type Value } from 'platejs';
import { Plate, PlateContent, type PlateEditor, usePlateEditor } from 'platejs/react';
import { FiSend } from 'react-icons/fi';

import { expandTabsToSpaces } from '@/lib/text_utils';

import { useEnterSubmit } from '@/hooks/use_enter_submit';

import { AlignKit } from '@/components/editor/plugins/align_kit';
import { AutoformatKit } from '@/components/editor/plugins/auto_format_kit';
import { BasicBlocksKit } from '@/components/editor/plugins/basic_blocks_kit';
import { BasicMarksKit } from '@/components/editor/plugins/basic_marks_kit';
import { EmojiKit } from '@/components/editor/plugins/emoji_kit';
import { IndentKit } from '@/components/editor/plugins/indent_kit';
import { LineHeightKit } from '@/components/editor/plugins/line_height_kit';
import { ListKit } from '@/components/editor/plugins/list_kit';
import { TabbableKit } from '@/components/editor/plugins/tabbable_kit';

import { TemplateSlashKit } from '@/chats/inputeditor/slashtemplate/template_plugin';
import { getTemplateSelections } from '@/chats/inputeditor/slashtemplate/template_selection_element';

export interface EditorTextInputHandle {
	focus: () => void;
}

interface EditorTextInputProps {
	isBusy: boolean;
	onSubmit: (text: string) => void;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
}

const EMPTY_VALUE: Value = [{ type: 'p', children: [{ text: '' }] }];

function insertPlainTextAsSingleBlock(ed: ReturnType<typeof usePlateEditor>, text: string, tabSize = 2) {
	if (!ed) {
		return;
	}
	const editor = ed as PlateEditor;
	const normalized = text.replace(/\r\n?/g, '\n');
	const lines = normalized.split('\n').map(l => expandTabsToSpaces(l, tabSize));

	editor.tf.insertText(lines[0] ?? '');
	for (let i = 1; i < lines.length; i++) {
		editor.tf.insertSoftBreak();
		editor.tf.insertText(lines[i]);
	}
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
				...TabbableKit,
				...TemplateSlashKit,
			],

			value: EMPTY_VALUE,
		});

		// Track plain text for enabling/disabling send button
		const [plainText, setPlainText] = useState<string>('');

		const isSubmittingRef = useRef<boolean>(false);
		const contentRef = useRef<HTMLDivElement | null>(null);
		const editorRef = useRef(editor);
		editorRef.current = editor; // keep a live ref for key handlers

		const { formRef, onKeyDown } = useEnterSubmit({
			isBusy,
			canSubmit: () => plainText.trim().length > 0,
			insertSoftBreak: () => {
				editor.tf.insertSoftBreak();
			},
			// onSubmitRequest: () => formRef.current?.requestSubmit() // optional override
		});

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
			let textToSend = editor.api.string([]);
			const promptsToSend = getTemplateSelections(editor);
			if (promptsToSend.length > 0) {
				textToSend += '\n\nprompts:\n' + JSON.stringify(promptsToSend, null, 2);
			}
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
						onKeyDown={onKeyDown}
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
