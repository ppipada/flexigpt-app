import React, { forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { SingleBlockPlugin } from 'platejs';
import { Plate, PlateContent, usePlateEditor } from 'platejs/react';
import { FiSend } from 'react-icons/fi';

import { cssEscape } from '@/lib/text_utils';

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

import {
	compareEntryByPathDeepestFirst,
	EMPTY_VALUE,
	getFirstTemplateNodeWithPath,
	getTemplateNodesWithPath,
	getTemplateSelections,
	hasNonEmptyUserText,
	insertPlainTextAsSingleBlock,
	toPlainTextReplacingVariables,
} from '@/chats/inputeditor/slashtemplate/editor_utils';
import { buildUserInlineChildrenFromText } from '@/chats/inputeditor/slashtemplate/tempalte_variables_inline';
import { TemplateSlashKit } from '@/chats/inputeditor/slashtemplate/template_plugin';
import { getLastUserBlockContent } from '@/chats/inputeditor/slashtemplate/template_processing';
import { TemplateToolbars } from '@/chats/inputeditor/slashtemplate/template_toolbars';

export interface EditorTextInputHandle {
	focus: () => void;
}

interface EditorTextInputProps {
	isBusy: boolean;
	onSubmit: (text: string) => Promise<void>;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
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

		// doc version tick to re-run selection computations on any editor change (even if text string doesn't change)
		const [docVersion, setDocVersion] = useState(0);

		const lastPopulatedSelectionKeyRef = useRef<Set<string>>(new Set());

		const selectionInfo = useMemo(() => {
			const tplNodeWithPath = getFirstTemplateNodeWithPath(editor);
			const selections = getTemplateSelections(editor);
			const hasTemplate = selections.length > 0;

			let requiredCount = 0;
			let pendingTools = 0;
			let firstPendingVar: { name: string; selectionID?: string } | undefined = undefined;

			for (const s of selections) {
				requiredCount += s.requiredCount;
				pendingTools += s.toolsToRun.filter(t => t.status === 'pending').length;
				if (!firstPendingVar && s.requiredVariables.length > 0) {
					firstPendingVar = { name: s.requiredVariables[0], selectionID: s.selectionID };
				}
			}

			return {
				tplNodeWithPath,
				hasTemplate,
				requiredCount,
				pendingTools,
				firstPendingVar,
			};
		}, [editor, docVersion]);

		const { formRef, onKeyDown } = useEnterSubmit({
			isBusy,
			canSubmit: () => {
				const hasTemplate = selectionInfo.hasTemplate;
				if (hasTemplate) {
					return selectionInfo.requiredCount === 0 && selectionInfo.pendingTools === 0;
				}
				return hasNonEmptyUserText(editorRef.current);
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
			return hasNonEmptyUserText(editorRef.current);
		}, [isBusy, selectionInfo, docVersion]);

		// Populate editor with effective last-USER block for EACH template selection (once per selectionID)
		useEffect(() => {
			const populated = lastPopulatedSelectionKeyRef.current;
			const nodes = getTemplateNodesWithPath(editor);
			const insertedIds: string[] = [];

			// Process in reverse document order to keep captured paths valid
			const nodesRev = [...nodes].sort(compareEntryByPathDeepestFirst);

			for (const [tsenode, originalPath] of nodesRev) {
				// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
				if (!tsenode || !tsenode.selectionID) continue;
				const selectionID: string = tsenode.selectionID;
				if (populated.has(selectionID)) continue;

				// Build children: keep the selection chip, add parsed user text with variable pills
				const userText = getLastUserBlockContent(tsenode);
				const inlineChildren = buildUserInlineChildrenFromText(tsenode, userText);

				try {
					editor.tf.withoutNormalizing(() => {
						// Recompute a fresh path to guard against prior insertions shifting indices
						const pathArr = Array.isArray(originalPath) ? (originalPath as number[]) : [];

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

				populated.add(selectionID);
				insertedIds.push(selectionID);
			}

			// Focus first variable pill of the last inserted selection (if any)
			if (insertedIds.length > 0) {
				const focusId = insertedIds[insertedIds.length - 1];
				requestAnimationFrame(() => {
					try {
						const sel = contentRef.current?.querySelector(
							`span[data-template-variable][data-selection-id="${cssEscape(focusId)}"]`
						) as HTMLElement | null;
						if (sel && 'focus' in sel && typeof sel.focus === 'function') {
							sel.focus();
						} else {
							editor.tf.focus();
						}
					} catch {
						editor.tf.focus();
					}
				});
			}
		}, [editor, docVersion]);

		const handleSubmit = (e?: React.FormEvent) => {
			if (e) e.preventDefault();
			if (!isSendButtonEnabled || isSubmittingRef.current) {
				// If invalid, flash and focus first pending pill
				if (selectionInfo.hasTemplate && (selectionInfo.requiredCount > 0 || selectionInfo.pendingTools > 0)) {
					// ask the toolbar (rendered via plugin) to flash
					window.dispatchEvent(new CustomEvent('tpl-toolbar:flash'));

					// Focus first pending variable pill (if any)
					const fpv = selectionInfo.firstPendingVar;
					if (fpv?.name && contentRef.current) {
						const idSegment = fpv.selectionID ? `[data-selection-id="${cssEscape(fpv.selectionID)}"]` : '';
						const sel = contentRef.current.querySelector(
							`span[data-template-variable][data-var-name="${cssEscape(fpv.name)}"]${idSegment}`
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
			// Fire and hold guard until resolved, but keep "clear immediately".
			const submitPromise = onSubmit(textToSend).catch(() => {
				/* swallow, guard released below */
			});

			// Clear editor and state right away
			editor.tf.setValue(EMPTY_VALUE);
			lastPopulatedSelectionKeyRef.current.clear();
			editor.tf.focus();

			// Release guard only after onSubmit resolves
			submitPromise.finally(() => {
				isSubmittingRef.current = false;
			});
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
