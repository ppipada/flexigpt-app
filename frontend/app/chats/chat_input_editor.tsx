/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import {
	type FormEvent,
	forwardRef,
	useCallback,
	useEffect,
	useImperativeHandle,
	useMemo,
	useRef,
	useState,
} from 'react';

import { FiAlertTriangle, FiEdit2, FiSend, FiSquare, FiX } from 'react-icons/fi';

import { useMenuStore } from '@ariakit/react';
import { SingleBlockPlugin, type Value } from 'platejs';
import { Plate, PlateContent, usePlateEditor } from 'platejs/react';

import { type Attachment, AttachmentKind, type AttachmentMode } from '@/spec/attachment';
import type { DirectoryAttachmentsResult } from '@/spec/backend';
import type { ToolChoice, ToolOutput } from '@/spec/tool';

import { type ShortcutConfig } from '@/lib/keyboard_shortcuts';
import { compareEntryByPathDeepestFirst } from '@/lib/path_utils';
import { cssEscape } from '@/lib/text_utils';

import { useEnterSubmit } from '@/hooks/use_enter_submit';

import { backendAPI } from '@/apis/baseapi';

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
import { AttachmentChipsBar } from '@/chats/attachments/attachment_chips_bar';
import {
	buildEditorAttachmentForLocalPath,
	buildEditorAttachmentForURL,
	type DirectoryAttachmentGroup,
	type EditorAttachment,
	editorAttachmentKey,
	MAX_FILES_PER_DIRECTORY,
} from '@/chats/attachments/editor_attachment_utils';
import {
	type EditorAttachedToolChoice,
	getAttachedTools,
	insertToolSelectionNode,
} from '@/chats/attachments/tool_editor_utils';
import { ToolPlusKit } from '@/chats/attachments/tool_plugin';
import { dispatchTemplateFlashEvent } from '@/chats/events/template_flash';
import {
	getFirstTemplateNodeWithPath,
	getTemplateNodesWithPath,
	getTemplateSelections,
	hasNonEmptyUserText,
	insertPlainTextAsSingleBlock,
	toPlainTextReplacingVariables,
} from '@/chats/templates/template_editor_utils';
import { TemplateSlashKit } from '@/chats/templates/template_plugin';
import { getLastUserBlockContent } from '@/chats/templates/template_processing';
import { TemplateToolbars } from '@/chats/templates/template_toolbars';
import { buildUserInlineChildrenFromText } from '@/chats/templates/template_variables_inline';

export interface EditorAreaHandle {
	focus: () => void;
	openTemplateMenu: () => void;
	openToolMenu: () => void;
	openAttachmentMenu: () => void;
	loadExternalMessage: (msg: EditorExternalMessage) => void;
	resetEditor: () => void;
}

export interface EditorExternalMessage {
	text: string;
	attachments?: Attachment[];
	toolChoices?: ToolChoice[];
}

export interface EditorSubmitPayload {
	text: string;
	attachedTools: EditorAttachedToolChoice[];
	attachments: EditorAttachment[];
	toolOutputs: ToolOutput[];
}

const EDITOR_EMPTY_VALUE: Value = [{ type: 'p', children: [{ text: '' }] }];

interface EditorAreaProps {
	isBusy: boolean;
	shortcutConfig: ShortcutConfig;
	onSubmit: (payload: EditorSubmitPayload) => Promise<void>;
	onRequestStop: () => void;
	editingMessageId: string | null;
	cancelEditing: () => void;
}

export const EditorArea = forwardRef<EditorAreaHandle, EditorAreaProps>(function EditorArea(
	{ isBusy, shortcutConfig, onSubmit, onRequestStop, editingMessageId, cancelEditing },
	ref
) {
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
			...FloatingToolbarKit,
		],
		value: EDITOR_EMPTY_VALUE,
	});

	const isSubmittingRef = useRef<boolean>(false);
	const contentRef = useRef<HTMLDivElement | null>(null);
	const editorRef = useRef(editor);
	editorRef.current = editor; // keep a live ref for key handlers
	const templateMenu = useMenuStore({ placement: 'top-start', focusLoop: true });
	const toolMenu = useMenuStore({ placement: 'top-start', focusLoop: true });
	const attachmentMenu = useMenuStore({ placement: 'top-start', focusLoop: true });
	const templateButtonRef = useRef<HTMLButtonElement | null>(null);
	const toolButtonRef = useRef<HTMLButtonElement | null>(null);
	const attachmentButtonRef = useRef<HTMLButtonElement | null>(null);

	// doc version tick to re-run selection computations on any editor change
	const [docVersion, setDocVersion] = useState(0);
	const [submitError, setSubmitError] = useState<string | null>(null);
	const [attachments, setAttachments] = useState<EditorAttachment[]>([]);
	const [directoryGroups, setDirectoryGroups] = useState<DirectoryAttachmentGroup[]>([]);

	const closeAllMenus = useCallback(() => {
		templateMenu.hide();
		toolMenu.hide();
		attachmentMenu.hide();
	}, [templateMenu, toolMenu, attachmentMenu]);

	const focusFirstMenuItem = (kind: 'templates' | 'tools' | 'attachments') => {
		// Allow the menu to mount/portal before querying DOM
		requestAnimationFrame(() => {
			const menuRoot = document.querySelector<HTMLElement>(`[data-menu-kind="${kind}"]`);
			if (!menuRoot) return;
			const firstItem = menuRoot.querySelector<HTMLElement>('[role="menuitem"]');
			firstItem?.focus();
		});
	};

	const openTemplatePicker = useCallback(() => {
		closeAllMenus();
		templateMenu.show();
		focusFirstMenuItem('templates');
	}, [closeAllMenus, templateMenu]);

	const openToolPicker = useCallback(() => {
		closeAllMenus();
		toolMenu.show();
		focusFirstMenuItem('tools');
	}, [closeAllMenus, toolMenu]);

	const openAttachmentPicker = useCallback(() => {
		closeAllMenus();
		attachmentMenu.show();
		focusFirstMenuItem('attachments');
	}, [closeAllMenus, attachmentMenu]);

	const lastPopulatedSelectionKeyRef = useRef<Set<string>>(new Set());

	const selectionInfo = useMemo(() => {
		const tplNodeWithPath = getFirstTemplateNodeWithPath(editor);
		const selections = getTemplateSelections(editor);
		const hasTemplate = selections.length > 0;

		let requiredCount = 0;
		let firstPendingVar: { name: string; selectionID?: string } | undefined = undefined;

		for (const s of selections) {
			requiredCount += s.requiredCount;

			if (!firstPendingVar) {
				if (s.requiredVariables.length > 0) {
					firstPendingVar = { name: s.requiredVariables[0], selectionID: s.selectionID };
				}
			}
		}

		return {
			tplNodeWithPath,
			hasTemplate,
			requiredCount,
			firstPendingVar,
		};
	}, [editor, docVersion]);

	const { formRef, onKeyDown } = useEnterSubmit({
		isBusy,
		canSubmit: () => {
			const hasTemplate = selectionInfo.hasTemplate;
			if (hasTemplate) {
				return selectionInfo.requiredCount === 0;
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
			return selectionInfo.requiredCount === 0;
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
			if (selectionInfo.hasTemplate && selectionInfo.requiredCount > 0) {
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

			const textToSend = hasTpl ? toPlainTextReplacingVariables(editor) : editor.api.string([]);
			const attachedTools = getAttachedTools(editor);
			const payload: EditorSubmitPayload = {
				text: textToSend,
				attachedTools: attachedTools,
				attachments: attachments,
				toolOutputs: [],
			};

			try {
				await onSubmit(payload);
				setSubmitError(null);
			} finally {
				// Clear editor and state after successful preprocessor runs and submit
				editor.tf.setValue(EDITOR_EMPTY_VALUE);
				setAttachments([]);
				setDirectoryGroups([]);

				lastPopulatedSelectionKeyRef.current.clear();
				editor.tf.focus();
				isSubmittingRef.current = false;
			}
		})().catch(() => {
			isSubmittingRef.current = false;
		});
	};

	const resetEditor = useCallback(() => {
		closeAllMenus();
		setSubmitError(null);
		lastPopulatedSelectionKeyRef.current.clear();
		isSubmittingRef.current = false;

		editor.tf.setValue(EDITOR_EMPTY_VALUE);
		setAttachments([]);
		setDirectoryGroups([]);
		// Let Plate onChange bump docVersion; no need to do it here.
		editor.tf.focus();
	}, [closeAllMenus, editor]);

	const loadExternalMessage = useCallback(
		(incoming: EditorExternalMessage) => {
			closeAllMenus();
			setSubmitError(null);
			lastPopulatedSelectionKeyRef.current.clear();
			isSubmittingRef.current = false;

			// 1) Reset document to plain text paragraphs.
			const plain = incoming.text ?? '';
			const paragraphs = plain.split(/\r?\n/);
			const value: Value =
				paragraphs.length === 0
					? EDITOR_EMPTY_VALUE
					: paragraphs.map(line => ({
							type: 'p',
							children: [{ text: line }],
						}));

			editor.tf.setValue(value);

			// 2) Rebuild attachments as EditorAttachment[]
			setAttachments(() => {
				if (!incoming.attachments || incoming.attachments.length === 0) return [];
				const next: EditorAttachment[] = [];
				const seen = new Set<string>();

				for (const att of incoming.attachments) {
					let ui: EditorAttachment | undefined = undefined;

					if (att.kind === AttachmentKind.url) {
						// URL attachment
						if (att.urlRef) {
							ui = buildEditorAttachmentForURL(att);
						} else {
							continue;
						}
					} else if (att.kind === AttachmentKind.file || att.kind === AttachmentKind.image) {
						// File/image/etc. – same type we originally got from backend.
						ui = buildEditorAttachmentForLocalPath(att);
					}

					if (!ui) continue;

					const key = editorAttachmentKey(ui);
					if (seen.has(key)) continue;
					seen.add(key);
					next.push(ui);
				}
				return next;
			});

			// We don’t attempt to reconstruct directoryGroups; show flat chips instead.
			setDirectoryGroups([]);

			// 3) Re-add tool choices as tool selection nodes.
			//    The doc we just set has no tool nodes, so we can just insert new ones.

			for (const choice of incoming.toolChoices ?? []) {
				insertToolSelectionNode(editor, {
					bundleID: choice.bundleID,
					toolSlug: choice.toolSlug,
					toolVersion: choice.toolVersion,
				});
			}

			editor.tf.focus();
		},
		[closeAllMenus, editor]
	);
	useImperativeHandle(ref, () => ({
		focus: () => {
			editor.tf.focus();
		},
		openTemplateMenu: () => {
			openTemplatePicker();
		},
		openToolMenu: () => {
			openToolPicker();
		},
		openAttachmentMenu: () => {
			openAttachmentPicker();
		},
		loadExternalMessage,
		resetEditor,
	}));

	const handleAttachFiles = async () => {
		let results: Attachment[];
		try {
			results = await backendAPI.openMultipleFilesAsAttachments(true);
		} catch {
			return;
		}

		if (!results || results.length === 0) return;

		setAttachments(prev => {
			const existing = new Set(prev.map(editorAttachmentKey));
			const next: EditorAttachment[] = [...prev];

			for (const r of results) {
				const att = buildEditorAttachmentForLocalPath(r);
				if (!att) {
					console.error('invalid attachment result');
					continue;
				}
				const key = editorAttachmentKey(att);
				if (existing.has(key)) continue;
				existing.add(key);
				next.push(att);
			}
			return next;
		});
		editor.tf.focus();
	};

	const handleAttachDirectory = async () => {
		let result: DirectoryAttachmentsResult;
		try {
			result = await backendAPI.openDirectoryAsAttachments(MAX_FILES_PER_DIRECTORY);
		} catch {
			// Backend canceled or errored; nothing to do.
			return;
		}

		if (!result || !result.dirPath) return;

		const { dirPath, attachments, overflowDirs } = result;

		if ((!attachments || attachments.length === 0) && (!overflowDirs || overflowDirs.length === 0)) {
			// Nothing readable / allowed in this folder.
			return;
		}

		const folderLabel = dirPath.trim().split(/[\\/]/).pop() || dirPath.trim();

		const groupId = crypto.randomUUID() ?? `dir-${Date.now()}-${Math.random().toString(16).slice(2)}`;

		// First, add or reuse attachments.
		const attachmentKeysForGroup: string[] = [];
		const ownedAttachmentKeysForGroup: string[] = [];

		// Deduplicate paths/keys within this single directory attach
		const seenKeysForGroup = new Set<string>();

		setAttachments(prev => {
			const existing = new Map<string, EditorAttachment>();
			for (const att of prev) {
				existing.set(editorAttachmentKey(att), att);
			}

			const added: EditorAttachment[] = [];

			for (const r of attachments ?? []) {
				const att = buildEditorAttachmentForLocalPath(r);
				if (!att) {
					console.error('invalid attachment result');
					continue;
				}
				const key = editorAttachmentKey(att);

				// Skip duplicates within this folder selection
				if (seenKeysForGroup.has(key)) continue;
				seenKeysForGroup.add(key);

				attachmentKeysForGroup.push(key);

				if (!existing.has(key)) {
					existing.set(key, att);
					added.push(att);
					ownedAttachmentKeysForGroup.push(key);
				}
			}

			return [...prev, ...added];
		});

		// Then, record the directory group that references those attachments.
		setDirectoryGroups(prev => [
			...prev,
			{
				id: groupId,
				dirPath,
				label: folderLabel,
				attachmentKeys: attachmentKeysForGroup,
				ownedAttachmentKeys: ownedAttachmentKeysForGroup,

				overflowDirs: overflowDirs ?? [],
			},
		]);

		editor.tf.focus();
	};

	const handleAttachURL = async (rawUrl: string) => {
		const trimmed = rawUrl.trim();
		if (!trimmed) return;

		const bAtt = await backendAPI.openURLAsAttachment(trimmed);
		if (!bAtt) return;
		const att = buildEditorAttachmentForURL(bAtt);
		const key = editorAttachmentKey(att);

		setAttachments(prev => {
			const existing = new Set(prev.map(editorAttachmentKey));
			if (existing.has(key)) return prev;
			return [...prev, att];
		});
		editor.tf.focus();
	};

	const handleChangeAttachmentMode = (att: EditorAttachment, newMode: AttachmentMode) => {
		const targetKey = editorAttachmentKey(att);
		setAttachments(prev => prev.map(a => (editorAttachmentKey(a) === targetKey ? { ...a, mode: newMode } : a)));
		editor.tf.focus();
	};

	const handleRemoveAttachment = (att: EditorAttachment) => {
		const targetKey = editorAttachmentKey(att);

		setAttachments(prev => prev.filter(a => editorAttachmentKey(a) !== targetKey));

		// Also detach from any directory groups (and drop empty groups)
		setDirectoryGroups(prevGroups => {
			const updated = prevGroups.map(g => ({
				...g,
				attachmentKeys: g.attachmentKeys.filter(k => k !== targetKey),
				ownedAttachmentKeys: g.ownedAttachmentKeys.filter(k => k !== targetKey),
			}));
			return updated.filter(g => g.attachmentKeys.length > 0 || g.overflowDirs.length > 0);
		});
	};

	const handleRemoveDirectoryGroup = (groupId: string) => {
		setDirectoryGroups(prevGroups => {
			const groupToRemove = prevGroups.find(g => g.id === groupId);
			if (!groupToRemove) return prevGroups;

			const remainingGroups = prevGroups.filter(g => g.id !== groupId);

			// Keys owned by other groups (so we don't delete shared attachments).
			const keysOwnedByOtherGroups = new Set<string>();
			for (const g of remainingGroups) {
				for (const key of g.ownedAttachmentKeys) {
					keysOwnedByOtherGroups.add(key);
				}
			}

			if (groupToRemove.ownedAttachmentKeys.length > 0) {
				setAttachments(prevAttachments =>
					prevAttachments.filter(att => {
						const key = editorAttachmentKey(att);
						if (!groupToRemove.ownedAttachmentKeys.includes(key)) return true;
						// If other groups still own this attachment, keep it.
						if (keysOwnedByOtherGroups.has(key)) return true;
						// Otherwise, drop it when this folder is removed.
						return false;
					})
				);
			}

			return remainingGroups;
		});
	};

	const handleRemoveOverflowDir = (groupId: string, dirPath: string) => {
		setDirectoryGroups(prevGroups => {
			const updated = prevGroups.map(g =>
				g.id !== groupId
					? g
					: {
							...g,
							overflowDirs: g.overflowDirs.filter(od => od.dirPath !== dirPath),
						}
			);
			return updated.filter(g => g.attachmentKeys.length > 0 || g.overflowDirs.length > 0);
		});
	};

	const handleCancelEditing = useCallback(() => {
		resetEditor();
		cancelEditing();
	}, [cancelEditing]);

	return (
		<form
			ref={formRef}
			onSubmit={handleSubmit}
			className="mx-0 flex w-full max-w-full min-w-0 flex-col overflow-x-hidden overflow-y-visible"
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

					// Auto-cancel editing when there's no text and no tools.
					const hasText = hasNonEmptyUserText(editorRef.current);
					const hasTools = getAttachedTools(editorRef.current).length > 0;
					const isEmpty = !hasText && !hasTools;

					// Only do this while editing an older message.
					if (isEmpty && editingMessageId) {
						// IMPORTANT: do NOT call resetEditor here; we only exit edit mode.
						cancelEditing();
					}
				}}
			>
				<div className="bg-base-100 border-base-200 w-full max-w-full min-w-0 overflow-hidden rounded-2xl border">
					<TemplateToolbars />
					{editingMessageId && (
						<div className="flex items-center justify-end gap-2 pt-1 pr-3 pb-0 text-xs">
							<div className="flex items-center gap-2">
								<FiEdit2 size={14} />
								<span>
									Editing an earlier message. Sending will replace it and drop all later messages from the conversation.
								</span>
							</div>
							<button
								type="button"
								className="btn btn-circle btn-neutral btn-xs shrink-0"
								onClick={handleCancelEditing}
								title="Cancel Edit"
							>
								<FiX size={14} />
							</button>
						</div>
					)}
					{/* Row: editor with send/stop button on the right */}
					<div className="flex min-h-20 min-w-0 gap-2 px-1 py-0">
						<PlateContent
							ref={contentRef}
							placeholder="Type message..."
							spellCheck={false}
							readOnly={isBusy}
							onKeyDown={e => {
								onKeyDown(e); // from useEnterSubmit
							}}
							onPaste={e => {
								e.preventDefault();
								e.stopPropagation();
								const text = e.clipboardData.getData('text/plain');
								if (!text) return;
								insertPlainTextAsSingleBlock(editor, text);
							}}
							className="max-h-96 min-w-0 flex-1 resize-none overflow-auto bg-transparent p-2 wrap-break-word whitespace-break-spaces outline-none [tab-size:2] focus:outline-none"
							style={{
								fontSize: '14px',
								whiteSpace: 'break-spaces',
								tabSize: 2,
								minHeight: '4rem',
							}}
						/>

						{/* Send / Stop button anchored at bottom-right */}
						<div className="flex items-end pr-1 pb-2">
							{isBusy ? (
								<button
									type="button"
									className="btn btn-circle btn-neutral btn-sm shrink-0"
									onClick={onRequestStop}
									title="Stop response"
									aria-label="Stop response"
								>
									<FiSquare size={20} />
								</button>
							) : (
								<button
									type="submit"
									className={`btn btn-circle btn-neutral btn-sm shrink-0 ${!isSendButtonEnabled ? 'btn-disabled' : ''}`}
									disabled={!isSendButtonEnabled}
									aria-label="Send message"
									title="Send message"
								>
									<FiSend size={20} />
								</button>
							)}
						</div>
					</div>
					{/* Chips bar for attachments & tools (scrollable, only when there are chips) */}
					<div className="flex w-full min-w-0 overflow-x-hidden">
						<AttachmentChipsBar
							attachments={attachments}
							directoryGroups={directoryGroups}
							onRemoveAttachment={handleRemoveAttachment}
							onChangeAttachmentMode={handleChangeAttachmentMode}
							onRemoveDirectoryGroup={handleRemoveDirectoryGroup}
							onRemoveOverflowDir={handleRemoveOverflowDir}
						/>
					</div>
				</div>

				{/* Bottom bar for template/tool/attachment pickers + tips menus */}
				<AttachmentBottomBar
					onAttachFiles={handleAttachFiles}
					onAttachDirectory={handleAttachDirectory}
					onAttachURL={handleAttachURL}
					templateMenuState={templateMenu}
					toolMenuState={toolMenu}
					attachmentMenuState={attachmentMenu}
					templateButtonRef={templateButtonRef}
					toolButtonRef={toolButtonRef}
					attachmentButtonRef={attachmentButtonRef}
					shortcutConfig={shortcutConfig}
				/>
			</Plate>
		</form>
	);
});
