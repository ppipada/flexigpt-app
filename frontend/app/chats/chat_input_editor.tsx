import { type FormEvent, forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { FiAlertTriangle, FiSend } from 'react-icons/fi';

import { SingleBlockPlugin, type Value } from 'platejs';
import { Plate, PlateContent, usePlateEditor } from 'platejs/react';

import { compareEntryByPathDeepestFirst } from '@/lib/path_utils';
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

import { AttachmentBottomBar } from '@/chats/attachments/attachment_bottom_bar';
import { getAttachedTools } from '@/chats/attachments/tool_editor_utils';
import { ToolPlusKit } from '@/chats/attachments/tool_plugin';
import { dispatchTemplateFlashEvent } from '@/chats/events/template_flash';
import {
	getFirstTemplateNodeWithPath,
	getTemplateNodesWithPath,
	getTemplateSelections,
	hasNonEmptyUserText,
	insertPlainTextAsSingleBlock,
	runAllReadyPreprocessors,
	toPlainTextReplacingVariables,
} from '@/chats/templates/template_editor_utils';
import { TemplateSlashKit } from '@/chats/templates/template_plugin';
import { getLastUserBlockContent } from '@/chats/templates/template_processing';
import { ToolStatus } from '@/chats/templates/template_spec';
import { TemplateToolbars } from '@/chats/templates/template_toolbars';
import { buildUserInlineChildrenFromText } from '@/chats/templates/template_variables_inline';

export interface EditorAreaHandle {
	focus: () => void;
}

const EDITOR_EMPTY_VALUE: Value = [{ type: 'p', children: [{ text: '' }] }];

interface EditorAreaProps {
	isBusy: boolean;
	onSubmit: (text: string) => Promise<void>;
}

export const EditorArea = forwardRef<EditorAreaHandle, EditorAreaProps>(function EditorArea({ isBusy, onSubmit }, ref) {
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
			...ToolPlusKit,
			...FloatingToolbarKit, // Keep floating formatting toolbar
		],

		value: EDITOR_EMPTY_VALUE,
	});

	const isSubmittingRef = useRef<boolean>(false);
	const contentRef = useRef<HTMLDivElement | null>(null);
	const editorRef = useRef(editor);
	editorRef.current = editor; // keep a live ref for key handlers

	// doc version tick to re-run selection computations on any editor change (even if text string doesn't change)
	const [docVersion, setDocVersion] = useState(0);
	const [submitError, setSubmitError] = useState<string | null>(null);

	const lastPopulatedSelectionKeyRef = useRef<Set<string>>(new Set());

	const selectionInfo = useMemo(() => {
		const tplNodeWithPath = getFirstTemplateNodeWithPath(editor);
		const selections = getTemplateSelections(editor);
		const hasTemplate = selections.length > 0;

		let requiredCount = 0;
		let pendingPreTools = 0;
		let firstPendingVar: { name: string; selectionID?: string } | undefined = undefined;

		for (const s of selections) {
			requiredCount += s.requiredCount;
			const notReady = s.toolsToRun.filter(t => t.status === ToolStatus.PENDING);
			pendingPreTools += notReady.length;
			if (!firstPendingVar) {
				if (s.requiredVariables.length > 0) {
					firstPendingVar = { name: s.requiredVariables[0], selectionID: s.selectionID };
				} else if (notReady.length > 0) {
					const unresolvedName = notReady[0]?.unresolved?.[0];
					if (unresolvedName) {
						firstPendingVar = { name: unresolvedName, selectionID: s.selectionID };
					}
				}
			}
		}

		return {
			tplNodeWithPath,
			hasTemplate,
			requiredCount,
			pendingPreTools,
			firstPendingVar,
		};
	}, [editor, docVersion]);

	const { formRef, onKeyDown } = useEnterSubmit({
		isBusy,
		canSubmit: () => {
			const hasTemplate = selectionInfo.hasTemplate;
			if (hasTemplate) {
				return selectionInfo.requiredCount === 0 && selectionInfo.pendingPreTools === 0;
			}
			return hasNonEmptyUserText(editorRef.current);
		},

		insertSoftBreak: () => {
			editor.tf.insertSoftBreak();
		},
	});

	const isSendButtonEnabled = useMemo(() => {
		if (isBusy) return false;
		if (selectionInfo.hasTemplate) {
			return selectionInfo.requiredCount === 0 && selectionInfo.pendingPreTools === 0;
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

	const handleSubmit = (e?: FormEvent) => {
		if (e) e.preventDefault();
		if (!isSendButtonEnabled || isSubmittingRef.current) {
			// If invalid, flash and focus first pending pill
			if (selectionInfo.hasTemplate && (selectionInfo.requiredCount > 0 || selectionInfo.pendingPreTools > 0)) {
				// ask the toolbar (rendered via plugin) to flash
				dispatchTemplateFlashEvent();

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

		setSubmitError(null);
		isSubmittingRef.current = true;

		(async () => {
			// If there are template selections, auto-run any ready tools before sending.
			const selections = getTemplateSelections(editor);
			const hasTpl = selections.length > 0;
			if (hasTpl) {
				const { ok, errors } = await runAllReadyPreprocessors(editor);
				if (!ok) {
					// Surface first error and abort send
					dispatchTemplateFlashEvent();
					const first = errors[0];
					const errorSummary = errors.map(e => `${e.saveAs || e.preprocId}: ${e.error || 'failed'}`).join('; ');
					setSubmitError(
						errorSummary ? `Tool run failed - ${errorSummary}` : 'Tool run failed. Please review the tool inputs.'
					);
					if (first.saveAs && contentRef.current) {
						const idSegment = first.selectionID ? `[data-selection-id="${cssEscape(first.selectionID)}"]` : '';
						const sel = contentRef.current.querySelector(
							`span[data-template-variable][data-var-name="${cssEscape(first.saveAs)}"]${idSegment}`
						);
						if (sel && 'focus' in sel && typeof sel.focus === 'function') {
							sel.focus();
						} else {
							editor.tf.focus();
						}
					}
					isSubmittingRef.current = false;
					return;
				}
			}

			const attachedTools = getAttachedTools(editor);
			const hasTools = attachedTools.length > 0;
			let textToSend = hasTpl ? toPlainTextReplacingVariables(editor) : editor.api.string([]);

			if (hasTools) {
				textToSend += '\n\ntools:\n' + JSON.stringify(attachedTools, null, 2);
			}

			try {
				await onSubmit(textToSend);
				setSubmitError(null);
			} finally {
				// Clear editor and state after successful preprocessor runs and submit
				editor.tf.setValue(EDITOR_EMPTY_VALUE);
				lastPopulatedSelectionKeyRef.current.clear();
				editor.tf.focus();
				isSubmittingRef.current = false;
			}
		})().catch(() => {
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
			className="bg-base-100 border-base-300 focus-within:border-base-400 mx-2 flex flex-col overflow-hidden rounded-2xl border"
		>
			{submitError ? (
				<div className="alert alert-error mx-4 mt-3 mb-1 flex items-start gap-2 text-sm" role="alert">
					<FiAlertTriangle size={16} className="mt-0.5" />
					<span>{submitError}</span>
				</div>
			) : null}
			<Plate
				editor={editor}
				onChange={() => {
					setDocVersion(v => v + 1);
					if (submitError) {
						setSubmitError(null);
					}
				}}
			>
				<TemplateToolbars />
				{/* Row: send button (left) + editor (right, flex-1) */}
				<div className="flex min-h-16 items-center gap-2 p-0">
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
						className="max-h-96 min-w-0 flex-1 resize-none overflow-auto bg-transparent px-4 py-2 whitespace-break-spaces outline-none [tab-size:2] focus:outline-none"
						style={{
							fontSize: '14px',
							whiteSpace: 'break-spaces',
							tabSize: 2,
							minHeight: '4rem',
						}}
					/>
					<button
						type="submit"
						className={`btn btn-circle btn-ghost shrink-0 ${!isSendButtonEnabled || isBusy ? 'btn-disabled' : ''}`}
						disabled={isBusy || !isSendButtonEnabled}
						aria-label="Send Message"
						title="Send Message"
					>
						<FiSend size={18} />
					</button>
				</div>
				<AttachmentBottomBar />
			</Plate>
		</form>
	);
});
