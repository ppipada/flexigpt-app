import type { Dispatch, SetStateAction } from 'react';

import type { PlateEditor } from 'platejs/react';

import type { JSONSchema } from '@/lib/jsonschema_utils';

import type { ConversationToolStateEntry } from '@/chats/tools/conversation_tools_chip';
import { computeToolUserArgsStatus, getToolNodesWithPath } from '@/chats/tools/tool_editor_utils';
import { type ToolArgsTarget, ToolUserArgsModal } from '@/chats/tools/tool_user_args_modal';

interface ToolArgsModalHostProps {
	editor: PlateEditor;
	conversationToolsState: ConversationToolStateEntry[];
	setConversationToolsState: Dispatch<SetStateAction<ConversationToolStateEntry[]>>;
	toolArgsTarget: ToolArgsTarget | null;
	setToolArgsTarget: Dispatch<SetStateAction<ToolArgsTarget | null>>;
	recomputeAttachedToolArgsBlocked: () => void;
}

export function ToolArgsModalHost({
	editor,
	conversationToolsState,
	setConversationToolsState,
	toolArgsTarget,
	setToolArgsTarget,
	recomputeAttachedToolArgsBlocked,
}: ToolArgsModalHostProps) {
	let isOpen = false;
	let toolLabel = '';
	let schema: JSONSchema | undefined;
	let existingInstance: string | undefined;
	let onSave: (newInstance: string) => void = () => {};

	if (toolArgsTarget?.kind === 'attached') {
		const all = getToolNodesWithPath(editor, false);
		const hit = all.find(([n]) => n.selectionID === toolArgsTarget.selectionID);
		if (hit) {
			isOpen = true;
			const [node] = hit;
			schema = node.toolSnapshot?.userArgSchema;
			toolLabel =
				node.toolSnapshot?.displayName && node.toolSnapshot.displayName.length > 0
					? node.toolSnapshot.displayName
					: node.toolSlug;
			existingInstance = node.userArgSchemaInstance;

			onSave = newInstance => {
				// Re-locate the node by selectionID to avoid stale paths.
				const allNow = getToolNodesWithPath(editor, false);
				const again = allNow.find(([n]) => n.selectionID === toolArgsTarget.selectionID);
				if (again) {
					const [, path] = again;
					editor.tf.setNodes({ userArgSchemaInstance: newInstance }, { at: path as any });
				}

				// Do *not* close here; the modal will call dialog.close(),
				// which triggers onClose -> setToolArgsTarget(null).
				recomputeAttachedToolArgsBlocked();
			};
		}
	} else if (toolArgsTarget?.kind === 'conversation') {
		const entry = conversationToolsState.find(e => e.key === toolArgsTarget.key);
		if (entry) {
			isOpen = true;
			const def = entry.toolDefinition;
			schema = def?.userArgSchema;
			toolLabel =
				entry.toolStoreChoice.displayName && entry.toolStoreChoice.displayName.length > 0
					? entry.toolStoreChoice.displayName
					: entry.toolStoreChoice.toolSlug;
			existingInstance = entry.toolStoreChoice.userArgSchemaInstance;

			onSave = newInstance => {
				setConversationToolsState(prev =>
					prev.map(e => {
						if (e.key !== toolArgsTarget.key) return e;

						const nextToolStoreChoice = {
							...e.toolStoreChoice,
							userArgSchemaInstance: newInstance,
						};

						const nextStatus =
							e.toolDefinition && e.toolDefinition.userArgSchema
								? computeToolUserArgsStatus(e.toolDefinition.userArgSchema, newInstance)
								: e.argStatus;

						return {
							...e,
							toolStoreChoice: nextToolStoreChoice,
							argStatus: nextStatus,
						};
					})
				);

				// This only recomputes attached-tool blocking; conversation-level
				// blocking is already handled by the useEffect in EditorArea.
				recomputeAttachedToolArgsBlocked();
			};
		}
	}

	const handleClose = () => {
		// Sync React state when the native <dialog> closes (ESC, backdrop, Cancel, Save).
		setToolArgsTarget(null);
	};

	return (
		<ToolUserArgsModal
			isOpen={isOpen}
			onClose={handleClose}
			toolLabel={toolLabel}
			schema={schema}
			existingInstance={existingInstance}
			onSave={onSave}
		/>
	);
}
