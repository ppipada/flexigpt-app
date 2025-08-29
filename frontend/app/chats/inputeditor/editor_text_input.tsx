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

import { TemplateSlashKit } from '@/chats/inputeditor/slashtemplate/template_plugin';
import {
	getLastUserBlockContent,
	type SelectedTemplateForRun,
} from '@/chats/inputeditor/slashtemplate/template_processing';
import {
	getFirstTemplateNodeWithPath,
	getTemplateSelections,
} from '@/chats/inputeditor/slashtemplate/template_selection_element';
import { TemplateToolbars } from '@/chats/inputeditor/slashtemplate/template_toolbar_kit';
import {
	buildUserInlineChildrenFromText,
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
				...FloatingToolbarKit, // Keep floating formatting toolbar
			],

			value: EMPTY_VALUE,
		});

		const isSubmittingRef = useRef<boolean>(false);
		const contentRef = useRef<HTMLDivElement | null>(null);
		const editorRef = useRef(editor);
		editorRef.current = editor; // keep a live ref for key handlers

		// Track plain text for enabling/disabling send button
		const [plainText, setPlainText] = useState<string>('');
		// doc version tick to re-run selection computations on any editor change (even if text string doesn't change)
		const [docVersion, setDocVersion] = useState(0);

		// When a template is inserted the first time, populate editor with user last-block text + inline variable pills.
		const populatedFromTemplateRef = useRef<boolean>(false);

		// Compute selection info from the editor value; re-run whenever the doc changes
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
				hasTemplate: selections.length > 0,
				requiredCount: primary?.requiredCount ?? 0,
				pendingTools,
				firstPendingVar: primary?.requiredVariables[0],
			};
		}, [editor, docVersion]);

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
			const [tsenode, tsPath] = tplNodeWithPath;
			// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
			if (!tsenode) return;

			// Build children: keep the selection node (for data + toolbar), add parsed user text with variable pills
			const userText = getLastUserBlockContent(tsenode);
			const inlineChildren = buildUserInlineChildrenFromText(tsenode, userText);

			// Insert inline children as siblings right after the selection chip; preserve existing content
			try {
				editor.tf.withoutNormalizing(() => {
					// tsPath points to the selection chip (inline void) inside the paragraph.
					// Insert after that index inside the same paragraph.
					const pathArr = Array.isArray(tsPath) ? (tsPath as number[]) : [];
					if (pathArr.length >= 2) {
						const blockPath = pathArr.slice(0, pathArr.length - 1); // parent paragraph path
						const indexAfter = pathArr[pathArr.length - 1] + 1;
						const atPath = [...blockPath, indexAfter] as any;
						editor.tf.insertNodes(inlineChildren, { at: atPath });
					} else {
						// Fallback: insert at start of first paragraph
						editor.tf.insertNodes(inlineChildren, { at: [0, 0] as any });
					}
				});
			} catch {
				// Last-resort fallback: insert at selection (or end)
				editor.tf.insertNodes(inlineChildren);
			}
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
					// ask the toolbar (rendered via plugin) to flash
					window.dispatchEvent(new CustomEvent('tpl-toolbar:flash'));

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
			const promptsToSend = getTemplateSelections(editor);
			const hasTpl = promptsToSend.length > 0;
			let textToSend = hasTpl ? toPlainTextReplacingVariables(editor) : editor.api.string([]);
			if (hasTpl) {
				textToSend += '\n\nprompts:\n' + JSON.stringify(promptsToSend, null, 2);
			}
			try {
				onSubmit(textToSend);
			} finally {
				// Clear editor and state after submitting
				editor.tf.setValue(EMPTY_VALUE);
				populatedFromTemplateRef.current = false;

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
				className="bg-base-100 border-base-300 focus-within:border-base-400 mx-2 flex flex-col rounded-2xl border"
			>
				<Plate
					editor={editor}
					onChange={() => {
						setPlainText(editor.api.string([]));
						setDocVersion(v => v + 1); // tick any time the doc changes
					}}
				>
					<TemplateToolbars />
					{/* Row: send button (left) + editor (right, flex-1) */}
					<div className="flex items-center gap-2 px-2 py-1">
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
							}}
							className="max-h-64 min-h-[24px] min-w-0 flex-1 resize-none overflow-auto bg-transparent px-4 py-2 whitespace-break-spaces outline-none [tab-size:2] focus:outline-none"
							style={{
								fontSize: '14px',
								whiteSpace: 'break-spaces',
								tabSize: 2,
							}}
						/>
						<button
							type="submit"
							className={`btn btn-circle btn-ghost shrink-0 self-end ${!isSendButtonEnabled || isBusy ? 'btn-disabled' : ''}`}
							disabled={isBusy || !isSendButtonEnabled}
							aria-label="Send Message"
							title="Send Message"
						>
							<FiSend size={18} />
						</button>
					</div>
				</Plate>
			</form>
		);
	}
);

EditorTextInput.displayName = 'EditorTextInput';

export default EditorTextInput;
