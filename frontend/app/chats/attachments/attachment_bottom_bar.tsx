import { type ReactNode, type RefObject, useMemo } from 'react';

import { FiFileText, FiImage, FiPaperclip, FiSend, FiSquare, FiTool, FiX, FiZap } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, type MenuStore } from '@ariakit/react';
import { type PlateEditor, useEditorRef } from 'platejs/react';

import type { PromptTemplateListItem } from '@/spec/prompt';
import type { ToolListItem } from '@/spec/tool';

import { formatShortcut, type ShortcutConfig } from '@/lib/keyboard_shortcuts';

import { usePromptTemplates } from '@/hooks/use_template';
import { useTools } from '@/hooks/use_tool';

import { promptStoreAPI, toolStoreAPI } from '@/apis/baseapi';

import {
	type EditorAttachment,
	editorAttachmentKey,
	getEditorAttachmentPath,
} from '@/chats/attachments/editor_attachment_utils';
import {
	getToolNodesWithPath,
	insertToolSelectionNode,
	removeToolByKey,
	toolIdentityKey,
} from '@/chats/attachments/tool_editor_utils';
import { CommandTipsBar } from '@/chats/chat_command_tips_bar';
import { insertTemplateSelectionNode } from '@/chats/templates/template_editor_utils';

interface AttachmentBottomBarProps {
	attachments: EditorAttachment[];
	onAttachFiles: (mode: 'file' | 'image') => Promise<void> | void;
	onRemoveAttachment: (att: EditorAttachment) => void;
	isBusy: boolean;
	isSendButtonEnabled: boolean;
	templateMenuState: MenuStore;
	toolMenuState: MenuStore;
	attachmentMenuState: MenuStore;
	templateButtonRef: RefObject<HTMLButtonElement | null>;
	toolButtonRef: RefObject<HTMLButtonElement | null>;
	attachmentButtonRef: RefObject<HTMLButtonElement | null>;
	shortcutConfig: ShortcutConfig;

	onRequestStop: () => void;
}

interface PickerButtonProps {
	label: string;
	icon: ReactNode;
	buttonRef: RefObject<HTMLButtonElement | null>;
	menuState: MenuStore;
	shortcut?: string;
	disabled?: boolean;
}

function PickerButton({ label, icon, buttonRef, menuState, shortcut, disabled }: PickerButtonProps) {
	const tooltip = shortcut ? `${label} (${shortcut})` : label;
	return (
		<div className="tooltip tooltip-right" data-tip={tooltip}>
			<MenuButton
				ref={buttonRef}
				store={menuState}
				disabled={disabled}
				className="btn btn-ghost btn-circle btn-sm text-neutral-custom hover:text-base-content"
				aria-label={tooltip}
			>
				{icon}
			</MenuButton>
		</div>
	);
}

const menuClasses =
	'rounded-box bg-base-100 text-base-content z-50 max-h-72 min-w-[240px] overflow-y-auto border border-base-300 p-1 shadow-xl';

const menuItemClasses =
	'flex items-center gap-2 rounded-xl px-2 py-1 text-sm outline-none transition-colors ' +
	'hover:bg-base-200 data-[active-item]:bg-base-300';

/**
  Bottom bar for rendering attached items (templates/tools/files) and housing
  action buttons. Chips area is horizontally scrollable.
*/
export function AttachmentBottomBar({
	attachments,
	onAttachFiles,
	onRemoveAttachment,
	isBusy,
	isSendButtonEnabled,
	templateMenuState,
	toolMenuState,
	attachmentMenuState,
	templateButtonRef,
	toolButtonRef,
	attachmentButtonRef,
	shortcutConfig,
	onRequestStop,
}: AttachmentBottomBarProps) {
	const editor = useEditorRef() as PlateEditor;
	const shortcutLabels = useMemo(
		() => ({
			templates: formatShortcut(shortcutConfig.insertTemplate),
			tools: formatShortcut(shortcutConfig.insertTool),
			attachments: formatShortcut(shortcutConfig.insertAttachment),
		}),
		[shortcutConfig]
	);
	const { data: templateData, loading: templatesLoading } = usePromptTemplates();
	const { data: toolData, loading: toolsLoading } = useTools();

	const toolEntries = getToolNodesWithPath(editor);
	// const hasAttachments = attachments.length > 0 || toolEntries.length > 0;

	const attachedToolKeys = new Set(
		toolEntries.map(([node]) => toolIdentityKey(node.bundleID, node.bundleSlug, node.toolSlug, node.toolVersion))
	);

	const availableTools = toolsLoading
		? []
		: toolData.filter(
				it => !attachedToolKeys.has(toolIdentityKey(it.bundleID, it.bundleSlug, it.toolSlug, it.toolVersion))
			);

	const closeTemplateMenu = () => {
		templateMenuState.hide();
	};
	const closeToolMenu = () => {
		toolMenuState.hide();
	};
	const closeAttachmentMenu = () => {
		attachmentMenuState.hide();
	};

	const handleTemplatePick = async (item: PromptTemplateListItem) => {
		try {
			const tmpl = await promptStoreAPI.getPromptTemplate(item.bundleID, item.templateSlug, item.templateVersion);
			insertTemplateSelectionNode(editor, item.bundleID, item.templateSlug, item.templateVersion, tmpl);
		} catch {
			insertTemplateSelectionNode(editor, item.bundleID, item.templateSlug, item.templateVersion);
		} finally {
			closeTemplateMenu();
			editor.tf.focus();
		}
	};

	const handleToolPick = async (item: ToolListItem) => {
		try {
			const tool = await toolStoreAPI.getTool(item.bundleID, item.toolSlug, item.toolVersion);
			insertToolSelectionNode(
				editor,
				{
					bundleID: item.bundleID,
					bundleSlug: item.bundleSlug,
					toolSlug: item.toolSlug,
					toolVersion: item.toolVersion,
				},
				tool
			);
		} catch {
			insertToolSelectionNode(editor, {
				bundleID: item.bundleID,
				bundleSlug: item.bundleSlug,
				toolSlug: item.toolSlug,
				toolVersion: item.toolVersion,
			});
		} finally {
			closeToolMenu();
			editor.tf.focus();
		}
	};

	const handleAttachmentPick = async (mode: 'file' | 'image') => {
		await onAttachFiles(mode);
		closeAttachmentMenu();
		editor.tf.focus();
	};

	return (
		<div className="bg-base-200 w-full" data-attachments-bottom-bar aria-label="Templates, tools, and attachments">
			<div className="flex items-center gap-2 px-1 pt-1 pb-0 text-xs shadow-none">
				<div className="flex items-center gap-1">
					<PickerButton
						label="Insert template"
						icon={<FiZap size={16} />}
						buttonRef={templateButtonRef}
						menuState={templateMenuState}
						shortcut={shortcutLabels.templates}
					/>
					<Menu store={templateMenuState} gutter={8} className={menuClasses} data-menu-kind="templates" autoFocusOnShow>
						{templatesLoading ? (
							<div className={`${menuItemClasses} text-base-content/60 cursor-default`}>Loading templates…</div>
						) : templateData.length === 0 ? (
							<div className={`${menuItemClasses} text-base-content/60 cursor-default`}>No templates available</div>
						) : (
							templateData.map(item => (
								<MenuItem
									key={`${item.bundleID}-${item.templateSlug}-${item.templateVersion}`}
									onClick={() => {
										void handleTemplatePick(item);
									}}
									className={menuItemClasses}
								>
									<FiZap size={14} className="text-warning" />
									<span className="truncate">{item.templateSlug.replace(/[-_]/g, ' ')}</span>
									<span className="text-base-content/50 ml-auto text-[10px] uppercase" aria-hidden="true">
										{item.templateVersion}
									</span>
								</MenuItem>
							))
						)}
					</Menu>

					<PickerButton
						label="Add tool"
						icon={<FiTool size={16} />}
						buttonRef={toolButtonRef}
						menuState={toolMenuState}
						shortcut={shortcutLabels.tools}
					/>
					<Menu store={toolMenuState} gutter={8} className={menuClasses} data-menu-kind="tools" autoFocusOnShow>
						{toolsLoading ? (
							<div className={`${menuItemClasses} text-base-content/60 cursor-default`}>Loading tools…</div>
						) : availableTools.length === 0 ? (
							<div className={`${menuItemClasses} text-base-content/60 cursor-default`}>No additional tools</div>
						) : (
							availableTools.map(item => (
								<MenuItem
									key={`${item.bundleID}-${item.toolSlug}-${item.toolVersion}`}
									onClick={() => {
										void handleToolPick(item);
									}}
									className={menuItemClasses}
								>
									<FiTool size={14} />
									<span className="truncate">{item.toolSlug.replace(/[-_]/g, ' ')}</span>
									<span className="text-base-content/50 ml-auto text-[10px] uppercase" aria-hidden="true">
										{item.toolVersion}
									</span>
								</MenuItem>
							))
						)}
					</Menu>

					<PickerButton
						label="Attach files or images"
						icon={<FiPaperclip size={16} />}
						buttonRef={attachmentButtonRef}
						menuState={attachmentMenuState}
						shortcut={shortcutLabels.attachments}
					/>
					<Menu
						store={attachmentMenuState}
						gutter={8}
						className={menuClasses}
						data-menu-kind="attachments"
						autoFocusOnShow
					>
						<MenuItem
							onClick={() => {
								void handleAttachmentPick('file');
							}}
							className={menuItemClasses}
						>
							<FiFileText size={14} />
							<span>Attach file</span>
						</MenuItem>
						<MenuItem
							onClick={() => {
								void handleAttachmentPick('image');
							}}
							className={menuItemClasses}
						>
							<FiImage size={14} />
							<span>Attach image</span>
						</MenuItem>
					</Menu>
				</div>
				{/* Neutral tips bar  */}
				<CommandTipsBar shortcutConfig={shortcutConfig} />

				{/* Chips scroller */}
				<div className="no-scrollbar flex min-w-0 flex-1 items-center gap-2 overflow-x-auto py-0">
					{attachments.map(att => {
						const key = editorAttachmentKey(att);
						const label = att.label.length > 40 ? att.label.slice(0, 37) + '...' : att.label;
						const path = getEditorAttachmentPath(att);
						return (
							<div
								key={key}
								className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0 text-xs"
								title={`${att.kind} attachment: ${att.label}${path ? ` (${path})` : ''}`}
								data-attachment-chip="attachment"
							>
								<FiFileText />
								<span className="truncate">{label}</span>
								<button
									type="button"
									className="btn btn-ghost btn-xs text-error shrink-0 px-1 py-0 shadow-none"
									onClick={() => {
										onRemoveAttachment(att);
									}}
									title="Remove attachment"
									aria-label="Remove attachment"
								>
									<FiX />
								</button>
							</div>
						);
					})}

					{toolEntries.map(([node]) => {
						const n = node;
						const display = n.toolSnapshot?.displayName ?? n.toolSlug;
						const slug = `${n.bundleSlug ?? n.bundleID}/${n.toolSlug}@${n.toolVersion}`;
						const identityKey = toolIdentityKey(n.bundleID, n.bundleSlug, n.toolSlug, n.toolVersion);

						return (
							<div
								key={n.selectionID}
								className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0 text-xs"
								title={`Tool choice: ${display} (${slug})`}
								data-attachment-chip="tool"
								data-selection-id={n.selectionID}
							>
								<FiTool />
								<span className="truncate">{display.length > 36 ? display.slice(0, 36) + '...' : display}</span>
								<button
									type="button"
									className="btn btn-ghost btn-xs text-error shrink-0 px-1 py-0 shadow-none"
									onClick={() => {
										removeToolByKey(editor, identityKey);
									}}
									title="Remove tool choice"
									aria-label="Remove tool choice"
								>
									<FiX />
								</button>
							</div>
						);
					})}
				</div>

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
	);
}
