import * as React from 'react';

import { NodeApi, type Path, type TElement, type TNode } from 'platejs';
import { type PlateEditor, useEditorRef } from 'platejs/react';

import { TemplateEditModal } from '@/chats/inputeditor/slashtemplate/template_edit_modal';
import { TemplateFixedToolbar } from '@/chats/inputeditor/slashtemplate/template_fixed_toolbar';
import type { TemplateSelectionElementNode } from '@/chats/inputeditor/slashtemplate/template_processing';
import {
	getTemplateNodesWithPath,
	getTemplateSelections,
	KEY_TEMPLATE_SELECTION,
} from '@/chats/inputeditor/slashtemplate/template_selection_element';
import { KEY_TEMPLATE_VARIABLE } from '@/chats/inputeditor/slashtemplate/variables_inline';

function useFlashSignal() {
	const [flashAll, setFlashAll] = React.useState(false);
	const timeoutRef = React.useRef<number | null>(null);

	React.useEffect(() => {
		const handler = () => {
			setFlashAll(true);
			if (timeoutRef.current) {
				window.clearTimeout(timeoutRef.current);
			}
			timeoutRef.current = window.setTimeout(() => {
				setFlashAll(false);
				timeoutRef.current = null;
			}, 800);
		};
		window.addEventListener('tpl-toolbar:flash', handler);
		return () => {
			window.removeEventListener('tpl-toolbar:flash', handler);
			if (timeoutRef.current) {
				window.clearTimeout(timeoutRef.current);
			}
		};
	}, []);

	return flashAll;
}

type TplKey = string; // path-based unique key

function pathKey(path: any): TplKey {
	return Array.isArray(path) ? path.join('.') : String(path ?? '');
}

function replaceVariablesForSelectionWithText(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string,
	templateVersion: string,
	pathOfSelection?: Path
) {
	// Replace all variable chips belonging to this selection with their current value (or {{name}} if empty)
	// Then remove the selection chip itself.
	// Note: we do not flatten other templates.
	const varEntries: Array<[TElement, Path]> = [];
	// Find the matching selection node to read variable values and to remove it at the end
	let tsNodeWithPath: [TemplateSelectionElementNode, Path] | undefined;

	if (pathOfSelection) {
		const got = NodeApi.get(editor, pathOfSelection);
		if (got && got.type === KEY_TEMPLATE_SELECTION) {
			tsNodeWithPath = [got as unknown as TemplateSelectionElementNode, pathOfSelection];
		}
	}
	if (!tsNodeWithPath) {
		for (const [el, p] of NodeApi.elements(editor)) {
			if (el.type === KEY_TEMPLATE_SELECTION) {
				const n = el as unknown as TemplateSelectionElementNode;
				if (n.bundleID === bundleID && n.templateSlug === templateSlug && n.templateVersion === templateVersion) {
					tsNodeWithPath = [n, p];
					break;
				}
			}
		}
	}

	const tsNode = tsNodeWithPath?.[0];
	const EMPTY_VARS: Record<string, unknown> = Object.freeze({}) as Record<string, unknown>;
	const vars: Record<string, unknown> = tsNode?.variables ?? EMPTY_VARS;

	const selectionID: string | undefined = tsNode?.selectionID;

	// Collect variable chips for this specific selection instance (prefer selectionId)
	for (const [el, p] of NodeApi.elements(editor)) {
		if (el.type === KEY_TEMPLATE_VARIABLE) {
			const sameInstance = selectionID
				? el.selectionID === selectionID
				: el.bundleID === bundleID && el.templateSlug === templateSlug && el.templateVersion === templateVersion;
			if (sameInstance) {
				varEntries.push([el, p]);
			}
		}
	}

	// Replace vars from deepest path to shallow to keep paths valid while mutating
	varEntries
		.sort((a, b) => {
			const pa = a[1] as number[];
			const pb = b[1] as number[];
			// deeper first
			return pb.length - pa.length || pb[pb.length - 1] - pa[pa.length - 1];
		})
		.forEach(([, path]) => {
			const nentry = NodeApi.get(editor, path)?.[0];
			if (!nentry) return;
			const name: string = (nentry as TElement).name as string;
			const value = vars[name];
			// eslint-disable-next-line @typescript-eslint/no-base-to-string
			const text = value !== undefined && value !== null && value !== '' ? String(value) : `{{${name}}}`;

			editor.tf.removeNodes({ at: path });
			editor.tf.insertNodes({ text }, { at: path });
		});

	// Remove only this selection chip
	if (tsNodeWithPath) {
		editor.tf.removeNodes({ at: tsNodeWithPath[1] });
	}
}

function removeSelection(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string,
	templateVersion: string,
	pathOfSelection?: any
) {
	editor.tf.withoutNormalizing(() => {
		// Identify target selection (get selectionId for precise removal)
		let targetSelection: [TemplateSelectionElementNode, Path] | undefined = undefined;
		if (pathOfSelection) {
			const got = NodeApi.get(editor, pathOfSelection);
			if (got && got.type === KEY_TEMPLATE_SELECTION) {
				targetSelection = [got as unknown as TemplateSelectionElementNode, pathOfSelection];
			}
		}
		if (!targetSelection) {
			for (const [el, p] of NodeApi.elements(editor)) {
				if (
					el.type === KEY_TEMPLATE_SELECTION &&
					el.bundleID === bundleID &&
					el.templateSlug === templateSlug &&
					el.templateVersion === templateVersion
				) {
					targetSelection = [el as unknown as TemplateSelectionElementNode, p];
					break;
				}
			}
		}
		const selectionID: string | undefined = targetSelection?.[0]?.selectionID;

		// Remove variable chips for this selection
		const entries: Array<[any, any]> = [];
		const it = NodeApi.elements(editor);
		for (const [el, p] of it) {
			if (
				el.type === KEY_TEMPLATE_VARIABLE &&
				(selectionID ? el.selectionID === selectionID : true) &&
				el.bundleID === bundleID &&
				el.templateSlug === templateSlug &&
				el.templateVersion === templateVersion
			) {
				entries.push([el, p]);
			}
		}
		const textEntries: Array<[TNode, Path]> = [];
		for (const [nnode, p] of NodeApi.nodes(editor)) {
			if (typeof nnode === 'object' && 'text' in nnode) {
				const ownerId = nnode.ownerSelectionID as string | undefined;
				if (ownerId && selectionID && ownerId === selectionID) {
					textEntries.push([nnode, p]);
				}
			}
		}

		// Remove deepest first
		[...entries, ...textEntries]
			.sort((a, b) => {
				const pa = a[1] as number[];
				const pb = b[1] as number[];
				return pb.length - pa.length || pb[pb.length - 1] - pa[pa.length - 1];
			})
			.forEach(([, p]) => {
				editor.tf.removeNodes({ at: p });
			});

		// Remove the selection chip
		if (targetSelection) {
			editor.tf.removeNodes({ at: targetSelection[1] });
		}
	});
	editor.tf.focus();
}

export function TemplateToolbars() {
	const editor = useEditorRef() as PlateEditor;
	const flashAll = useFlashSignal();

	// Build a stable mapping selection <-> its node+path by document order.
	const selections = getTemplateSelections(editor);
	const nodesWithPath = getTemplateNodesWithPath(editor);

	// There might be a mismatch if something unusual happened; pair by nearest match
	const used = new Set<string>();
	const items = selections.map(sel => {
		const entry = nodesWithPath.find(([n, p]) => {
			const ok =
				n.bundleID === sel.bundleID &&
				n.templateSlug === sel.templateSlug &&
				n.templateVersion === sel.templateVersion &&
				!used.has(pathKey(p));
			if (ok) used.add(pathKey(p));
			return ok;
		});
		const id = entry ? pathKey(entry[1]) : `${sel.bundleID}:${sel.templateSlug}:${sel.templateVersion}`;
		return { id, sel, nodeWithPath: entry };
	});

	// Single-open modal keyed by selection path
	const [openId, setOpenId] = React.useState<string | null>(null);

	if (items.length === 0) return null;

	return (
		<div className="border-base-300 bg-base-100/95 supports-[backdrop-filter]:bg-base-100/60 sticky top-0 left-0 z-50 w-full border-b backdrop-blur">
			{items.map(({ id, sel, nodeWithPath }) => {
				const [tsNode, tsPath] = nodeWithPath ?? [];
				const flashing = flashAll;

				return (
					<div key={id} className="w-full">
						<TemplateFixedToolbar
							selection={sel}
							flashing={flashing}
							onOpenModal={() => {
								setOpenId(id);
							}}
							onRemove={() => {
								removeSelection(editor, sel.bundleID, sel.templateSlug, sel.templateVersion, tsPath);
							}}
							onFlatten={() => {
								editor.tf.withoutNormalizing(() => {
									replaceVariablesForSelectionWithText(
										editor,
										sel.bundleID,
										sel.templateSlug,
										sel.templateVersion,
										tsPath
									);
								});
								editor.tf.focus();
							}}
						/>

						{tsNode ? (
							<TemplateEditModal
								open={openId === id}
								onClose={() => {
									setOpenId(null);
								}}
								tsenode={tsNode}
								editor={editor}
								path={tsPath}
							/>
						) : null}
					</div>
				);
			})}
		</div>
	);
}

// const KEY_TEMPLATE_TOOLBAR = 'template-fixed-toolbar';
// export const TemplateToolbarKit = [
// 	createPlatePlugin({
// 		key: KEY_TEMPLATE_TOOLBAR,
// 		render: {
// 			beforeEditable: () => {
// 				// Render a fixed toolbar before the editable area whenever a template selection exists
// 				return <TemplateToolbars />;
// 			},
// 		},
// 	}),
// ];
