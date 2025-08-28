import { KEYS } from 'platejs';
import { useEditorReadOnly } from 'platejs/react';

import { MarkToolbarButton } from '@/components/editor/nodes/mark_toolbar_button';
import { ToolbarGroup } from '@/components/editor/nodes/toolbar';

export function FloatingToolbarButtons() {
	const readOnly = useEditorReadOnly();

	if (readOnly) return null;

	return (
		<ToolbarGroup>
			<MarkToolbarButton nodeType={KEYS.bold} tooltip="Bold (⌘+B)">
				<span className="font-bold">B</span>
			</MarkToolbarButton>

			<MarkToolbarButton nodeType={KEYS.italic} tooltip="Italic (⌘+I)">
				<span className="italic">I</span>
			</MarkToolbarButton>

			<MarkToolbarButton nodeType={KEYS.underline} tooltip="Underline (⌘+U)">
				<span className="underline underline-offset-2">U</span>
			</MarkToolbarButton>

			<MarkToolbarButton nodeType={KEYS.strikethrough} tooltip="Strikethrough (⌘+⇧+M)">
				<span className="line-through">S</span>
			</MarkToolbarButton>

			<MarkToolbarButton nodeType={KEYS.code} tooltip="Code (⌘+E)">
				<span className="font-mono text-[0.9em]">{'</>'}</span>
			</MarkToolbarButton>
		</ToolbarGroup>
	);
}
