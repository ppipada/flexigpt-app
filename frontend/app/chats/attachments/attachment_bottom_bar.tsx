import { type ReactNode, type RefObject, useMemo } from 'react';

import { FiFilePlus, FiLink, FiPaperclip, FiTool, FiUpload } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, type MenuStore } from '@ariakit/react';
import { type PlateEditor, useEditorRef } from 'platejs/react';

import type { PromptTemplateListItem } from '@/spec/prompt';
import type { ToolListItem } from '@/spec/tool';

import { formatShortcut, type ShortcutConfig } from '@/lib/keyboard_shortcuts';

import { usePromptTemplates } from '@/hooks/use_template';
import { useTools } from '@/hooks/use_tool';

import { promptStoreAPI, toolStoreAPI } from '@/apis/baseapi';

import { getToolNodesWithPath, insertToolSelectionNode, toolIdentityKey } from '@/chats/attachments/tool_editor_utils';
import { CommandTipsBar } from '@/chats/chat_command_tips_bar';
import { insertTemplateSelectionNode } from '@/chats/templates/template_editor_utils';

interface AttachmentBottomBarProps {
	onAttachFiles: () => Promise<void> | void;
	onAttachURL: (url: string) => Promise<void> | void;

	templateMenuState: MenuStore;
	toolMenuState: MenuStore;
	attachmentMenuState: MenuStore;
	templateButtonRef: RefObject<HTMLButtonElement | null>;
	toolButtonRef: RefObject<HTMLButtonElement | null>;
	attachmentButtonRef: RefObject<HTMLButtonElement | null>;
	shortcutConfig: ShortcutConfig;
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
  Bottom bar for template/tool/attachment buttons and tips menus.
  The chips scroller now lives in a separate bar inside the editor.
*/
export function AttachmentBottomBar({
	onAttachFiles,
	onAttachURL,

	templateMenuState,
	toolMenuState,
	attachmentMenuState,
	templateButtonRef,
	toolButtonRef,
	attachmentButtonRef,
	shortcutConfig,
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

	const handleAttachmentPick = async () => {
		await onAttachFiles();
		closeAttachmentMenu();
		editor.tf.focus();
	};

	const handleAttachmentPickURL = async () => {
		const raw = window.prompt('Paste a link or URL to attach');
		const url = raw?.trim();
		if (!url) return;
		await onAttachURL(url);
		closeAttachmentMenu();
		editor.tf.focus();
	};

	return (
		<div
			className="bg-base-200 w-full overflow-hidden"
			data-attachments-bottom-bar
			aria-label="Templates, tools, and attachments"
		>
			<div className="flex items-center gap-2 px-1 py-0 text-xs shadow-none">
				{/* Left: template / tool / attachment pickers */}
				<div className="flex items-center gap-1">
					<PickerButton
						label="Attach files or links"
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
								void handleAttachmentPick();
							}}
							className={menuItemClasses}
						>
							<FiUpload size={14} />
							<span>From your computer...</span>
						</MenuItem>
						<MenuItem
							onClick={() => {
								void handleAttachmentPickURL();
							}}
							className={menuItemClasses}
						>
							<FiLink size={14} />
							<span>From a link or URL...</span>
						</MenuItem>
					</Menu>
					<PickerButton
						label="Insert template"
						icon={<FiFilePlus size={16} />}
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
									<FiFilePlus size={14} className="text-warning" />
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
				</div>

				{/* Right: keyboard shortcuts & tips menus */}
				<div className="ml-auto flex items-center gap-1">
					<CommandTipsBar shortcutConfig={shortcutConfig} />
				</div>
			</div>
		</div>
	);
}
