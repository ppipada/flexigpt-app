import React, { forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { SingleBlockPlugin, type Value } from 'platejs';
import { Plate, PlateContent, type PlateEditor, usePlateEditor } from 'platejs/react';
import { FiSend } from 'react-icons/fi';

import { cssEscape, expandTabsToSpaces } from '@/lib/text_utils';

import { useEnterSubmit } from '@/hooks/use_enter_submit';

import { AlignKit } from '@/components/editor/plugins/align_kit';
import { AutoformatKit } from '@/components/editor/plugins/auto_format_kit';
import { BasicBlocksKit } from '@/components/editor/plugins/basic_blocks_kit';
import { BasicMarksKit } from '@/components/editor/plugins/basic_marks_kit';
import { EmojiKit } from '@/components/editor/plugins/emoji_kit';
import { FloatingToolbarKit } from '@/components/editor/plugins/floating_toolbar_kit';
import { IndentKit } from '@/components/editor/plugins/indent_kit';
import { LineHeightKit } from '@/components/editor/plugins/line_height_kit';
import { ListKit } from '@/components/editor/plugins/list_kit';
import { TabbableKit } from '@/components/editor/plugins/tabbable_kit';

import { TemplateEditModal } from '@/chats/inputeditor/slashtemplate/template_edit_modal';
import { TemplateFixedToolbar } from '@/chats/inputeditor/slashtemplate/template_fixed_toolbar';
import { TemplateSlashKit } from '@/chats/inputeditor/slashtemplate/template_plugin';
import {
	getLastUserBlockContent,
	type SelectedTemplateForRun,
} from '@/chats/inputeditor/slashtemplate/template_processing';
import {
	getFirstTemplateNodeWithPath,
	getTemplateSelections,
	KEY_TEMPLATE_SELECTION,
} from '@/chats/inputeditor/slashtemplate/template_selection_element';
import {
	buildUserInlineChildrenFromText,
	KEY_TEMPLATE_VARIABLE,
	toPlainTextReplacingVariables,
} from '@/chats/inputeditor/slashtemplate/variables_inline';

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
				...FloatingToolbarKit,
			],

			value: EMPTY_VALUE,
		});

		const isSubmittingRef = useRef<boolean>(false);
		const contentRef = useRef<HTMLDivElement | null>(null);
		const editorRef = useRef(editor);
		editorRef.current = editor; // keep a live ref for key handlers

		// Track plain text for enabling/disabling send button
		const [plainText, setPlainText] = useState<string>('');
		const [flashToolbar, setFlashToolbar] = useState(false);
		const [modalOpen, setModalOpen] = useState(false);

		// When a template is inserted the first time, populate editor with user last-block text + inline variable pills.
		const populatedFromTemplateRef = useRef<boolean>(false);

		const selectionInfo = useMemo(() => {
			// First (primary) template selection, if any
			const tplNodeWithPath = getFirstTemplateNodeWithPath(editor);
			const selections = getTemplateSelections(editor);
			let primary: SelectedTemplateForRun | undefined;
			let pendingTools = 0;
			if (selections.length > 0) {
				primary = selections[0];
				pendingTools = primary.toolsToRun.filter(t => t.status === 'pending').length;
			}

			return {
				primary,
				tplNodeWithPath,
				hasTemplate: Boolean(primary),
				requiredCount: primary?.requiredCount ?? 0,
				pendingTools,
				firstPendingVar: primary?.requiredVariables[0],
			};
		}, [editor]);

		const { formRef, onKeyDown } = useEnterSubmit({
			isBusy,
			canSubmit: () => {
				const hasTemplate = selectionInfo.hasTemplate;
				if (hasTemplate) {
					return selectionInfo.requiredCount === 0 && selectionInfo.pendingTools === 0;
				}
				return plainText.trim().length > 0;
			},

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

		const isSendButtonEnabled = useMemo(() => {
			if (isBusy) return false;
			if (selectionInfo.hasTemplate) {
				return selectionInfo.requiredCount === 0 && selectionInfo.pendingTools === 0;
			}
			return plainText.trim().length > 0;
		}, [isBusy, selectionInfo, plainText]);

		// Populate editor with effective last-USER block when a template is added (once)
		useEffect(() => {
			const tplNodeWithPath = selectionInfo.tplNodeWithPath;
			if (!tplNodeWithPath || populatedFromTemplateRef.current) return;
			const [tsenode] = tplNodeWithPath;
			// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
			if (!tsenode) return;

			// Build children: keep the selection node (for data + toolbar), add parsed user text with variable pills
			const userText = getLastUserBlockContent(tsenode);
			const inlineChildren = buildUserInlineChildrenFromText(tsenode, userText);
			const nextValue: Value = [
				{
					type: 'p',
					children: [
						// Keep the template selection element in-doc (hidden renderer) so toolbar can pick it up
						tsenode,
						...inlineChildren,
					],
				} as any,
			];

			editor.tf.setValue(nextValue);
			populatedFromTemplateRef.current = true;

			// Focus first variable pill, if any
			setTimeout(() => {
				try {
					const els = Array.from(contentRef.current?.querySelectorAll('span[data-template-variable]') ?? []);
					if (els.length > 0 && els[0] && 'focus' in els[0] && typeof els[0].focus === 'function') {
						els[0].focus();
					} else {
						editor.tf.focus();
					}
				} catch {
					editor.tf.focus();
				}
			}, 0);
		}, [editor, selectionInfo.tplNodeWithPath]);

		const handleSubmit = (e?: React.FormEvent) => {
			if (e) e.preventDefault();
			if (!isSendButtonEnabled || isSubmittingRef.current) {
				// If invalid, flash and focus first pending pill
				if (selectionInfo.hasTemplate && (selectionInfo.requiredCount > 0 || selectionInfo.pendingTools > 0)) {
					setFlashToolbar(true);
					setTimeout(() => {
						setFlashToolbar(false);
					}, 800);
					// Focus first pending variable pill (if any)
					const varName = selectionInfo.firstPendingVar;
					if (varName && contentRef.current) {
						const sel = contentRef.current.querySelector(
							`span[data-template-variable][data-var-name="${cssEscape(varName)}"]`
						);
						if (sel && 'focus' in sel && typeof sel.focus === 'function') {
							sel.focus();
						} else {
							// fallback focus
							editor.tf.focus();
						}
					} else {
						editor.tf.focus();
					}
				}
				return;
			}

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
				populatedFromTemplateRef.current = false;

				// Focus back into the editor
				editor.tf.focus();
				isSubmittingRef.current = false;
			}
		};

		function removeTemplate() {
			populatedFromTemplateRef.current = false;
			// Extra safety: remove any lingering selection/template nodes
			try {
				editor.tf.removeNodes({
					match: n => n.type === KEY_TEMPLATE_SELECTION || n.type === KEY_TEMPLATE_VARIABLE,
				});
			} catch {
				/* empty */
			}
			editor.tf.focus();
		}

		function expandTemplateToPlainText() {
			// Produce flat text keeping typed user content; variables replaced by value or {{name}}
			const flat = toPlainTextReplacingVariables(editor);
			editor.tf.setValue(EMPTY_VALUE);
			insertPlainTextAsSingleBlock(editor, flat);
			// Remove selection nodes so toolbar disappears
			try {
				editor.tf.removeNodes({
					match: n => n.type === KEY_TEMPLATE_SELECTION || n.type === KEY_TEMPLATE_VARIABLE,
				});
			} catch {
				/* empty */
			}
			populatedFromTemplateRef.current = false;
			editor.tf.focus();
		}

		useImperativeHandle(ref, () => ({
			focus: () => {
				editor.tf.focus();
			},
		}));

		return (
			<form
				ref={formRef}
				onSubmit={handleSubmit}
				className="bg-base-100 border-base-300 focus-within:border-base-400 mx-2 flex flex-col rounded-2xl border"
			>
				{/* Fixed toolbar (only when a template is present) */}
				{selectionInfo.hasTemplate && selectionInfo.primary && (
					<TemplateFixedToolbar
						selection={selectionInfo.primary}
						flashing={flashToolbar}
						onOpenModal={() => {
							setModalOpen(true);
						}}
						onRemove={removeTemplate}
						onFlatten={expandTemplateToPlainText}
					/>
				)}

				{/* Full-screen modal for power editing */}
				{selectionInfo.hasTemplate && selectionInfo.tplNodeWithPath && (
					<TemplateEditModal
						open={modalOpen}
						onClose={() => {
							setModalOpen(false);
						}}
						tsenode={selectionInfo.tplNodeWithPath[0]}
						editor={editor}
						path={selectionInfo.tplNodeWithPath[1]}
					/>
				)}

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
						className="max-h-64 min-h-[24px] w-full flex-1 resize-none overflow-auto bg-transparent px-4 py-2 whitespace-break-spaces outline-none [tab-size:2] focus:outline-none"
						style={{
							fontSize: '14px',
							whiteSpace: 'break-spaces',
							tabSize: 2,
						}}
					/>
				</Plate>

				<div className="flex w-full items-center justify-end px-2 pt-0 pb-1">
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
				</div>
			</form>
		);
	}
);

EditorTextInput.displayName = 'EditorTextInput';

export default EditorTextInput;
