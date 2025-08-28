import React from 'react';

import type { PlateEditor } from 'platejs/react';
import { createPortal } from 'react-dom';
import { FiPlay } from 'react-icons/fi';

import { type PreProcessorCall, type PromptVariable, VarSource, VarType } from '@/spec/prompt';

import {
	computeEffectiveTemplate,
	computeRequirements,
	type TemplateSelectionElementNode,
	type ToolState,
} from '@/chats/inputeditor/slashtemplate/template_processing';

export function TemplateEditModal({
	open,
	onClose,
	tsenode,
	editor,
	path,
}: {
	open: boolean;
	onClose: () => void;
	tsenode: TemplateSelectionElementNode;
	editor: PlateEditor;
	path: any;
}) {
	const { template, blocks, variablesSchema, preProcessors } = computeEffectiveTemplate(tsenode);

	// Local form state
	const [displayName, setDisplayName] = React.useState<string>(
		tsenode.overrides?.displayName ?? template?.displayName ?? tsenode.templateSlug
	);
	const [description, setDescription] = React.useState<string>(
		tsenode.overrides?.description ?? template?.description ?? ''
	);
	const [tags, setTags] = React.useState<string>((tsenode.overrides?.tags ?? template?.tags ?? []).join(', '));

	// Editable message blocks (content)
	const [blockEdits, setBlockEdits] = React.useState(blocks);

	// Variable values
	const initialVarValues = React.useMemo(() => {
		const vals: Record<string, unknown> = { ...tsenode.variables };
		for (const v of variablesSchema) {
			const val = effectiveVarValueSafe(v, tsenode.variables, tsenode.toolStates);
			if (val !== undefined) vals[v.name] = val;
		}
		return vals;
	}, [variablesSchema, tsenode.variables, tsenode.toolStates]);

	const [varValues, setVarValues] = React.useState<Record<string, unknown>>(initialVarValues);

	// Preprocessor args and statuses
	const [toolArgs, setToolArgs] = React.useState<Record<string, Record<string, any>>>(() => {
		const curr: Record<string, Record<string, any>> = {};
		for (const p of preProcessors) {
			curr[p.id] = tsenode.toolStates?.[p.id]?.args ?? p.args ?? {};
		}
		return curr;
	});
	const [toolStatuses, setToolStatuses] = React.useState<Record<string, ToolState>>(tsenode.toolStates ?? {});

	React.useEffect(() => {
		if (!open) return;
		// Rehydrate on open in case node changed
		setDisplayName(tsenode.overrides?.displayName ?? template?.displayName ?? tsenode.templateSlug);
		setDescription(tsenode.overrides?.description ?? template?.description ?? '');
		setTags((tsenode.overrides?.tags ?? template?.tags ?? []).join(', '));
		setBlockEdits(blocks);
		const vals: Record<string, unknown> = { ...tsenode.variables };
		for (const v of variablesSchema) {
			const val = effectiveVarValueSafe(v, tsenode.variables, tsenode.toolStates);
			if (val !== undefined) vals[v.name] = val;
		}
		setVarValues(vals);

		const targs: Record<string, Record<string, any>> = {};
		for (const p of preProcessors) {
			targs[p.id] = tsenode.toolStates?.[p.id]?.args ?? p.args ?? {};
		}
		setToolArgs(targs);
		setToolStatuses(tsenode.toolStates ?? {});
	}, [open]);

	function saveAndClose() {
		const nextOverrides = {
			...tsenode.overrides,
			displayName,
			description,
			tags: tags
				.split(',')
				.map(s => s.trim())
				.filter(Boolean),
			blocks: blockEdits,
		};

		// Merge tool states with edited args
		const nextToolStates: Record<string, ToolState> = { ...(tsenode.toolStates ?? {}) };
		for (const p of preProcessors) {
			nextToolStates[p.id] = {
				...(nextToolStates[p.id] ?? { status: 'pending' }),
				args: toolArgs[p.id],
			};
		}

		editor.tf.setNodes(
			{
				overrides: nextOverrides,
				variables: varValues,
				toolStates: nextToolStates,
			},
			{ at: path }
		);

		onClose();
	}

	function markToolReady(preProc: PreProcessorCall) {
		const next = { ...toolStatuses };
		next[preProc.id] = {
			...(next[preProc.id] ?? {}),
			status: 'ready',
			args: toolArgs[preProc.id],
		};
		setToolStatuses(next);
	}

	const req = computeRequirements(variablesSchema, varValues, preProcessors, toolStatuses);
	if (!open) return null;

	return createPortal(
		<dialog className="modal modal-open" onClose={onClose}>
			<div className="modal-box max-w-3xl">
				<h3 className="flex items-center gap-2 text-lg font-bold">
					Edit Template
					<span className="badge badge-neutral">{displayName}</span>
				</h3>

				<div className="mt-4 space-y-6">
					<section>
						<h4 className="mb-2 font-semibold">Overview</h4>
						<div className="grid grid-cols-1 gap-3 md:grid-cols-2">
							<label className="form-control w-full">
								<span className="label-text">Display Name (local)</span>
								<input
									className="input input-bordered input-sm"
									value={displayName}
									onChange={e => {
										setDisplayName(e.target.value);
									}}
								/>
							</label>
							<label className="form-control w-full">
								<span className="label-text">Tags (comma-separated)</span>
								<input
									className="input input-bordered input-sm"
									value={tags}
									onChange={e => {
										setTags(e.target.value);
									}}
								/>
							</label>
						</div>
						<label className="form-control mt-2 w-full">
							<span className="label-text">Description (local)</span>
							<textarea
								className="textarea textarea-bordered"
								value={description}
								onChange={e => {
									setDescription(e.target.value);
								}}
							/>
						</label>
					</section>

					<section>
						<h4 className="mb-2 font-semibold">Variables</h4>
						<div className="space-y-3">
							{variablesSchema.length === 0 && (
								<div className="text-sm opacity-70">No variables defined for this template.</div>
							)}
							{variablesSchema.map(v => (
								<VariableEditorRow
									key={v.name}
									varDef={v}
									value={varValues[v.name]}
									onChange={val => {
										setVarValues(s => ({ ...s, [v.name]: val }));
									}}
								/>
							))}
						</div>
						<div className="mt-2 text-sm">
							{req.requiredCount > 0 ? (
								<span className="text-warning">Required remaining: {req.requiredVariables.join(', ')}</span>
							) : (
								<span className="text-success">All required variables provided.</span>
							)}
						</div>
					</section>

					<section>
						<h4 className="mb-2 font-semibold">Pre-processors</h4>
						<div className="space-y-3">
							{preProcessors.length === 0 && <div className="text-sm opacity-70">No pre-processors configured.</div>}
							{preProcessors.map(p => (
								<PreProcessorRow
									key={p.id}
									call={p}
									args={toolArgs[p.id]}
									status={toolStatuses[p.id].status}
									onArgsChange={next => {
										setToolArgs(s => ({
											...s,
											[p.id]: next,
										}));
									}}
									onRunNow={() => {
										markToolReady(p);
									}}
								/>
							))}
							{preProcessors.some(p => toolStatuses[p.id].status !== 'done') && (
								<div className="flex items-center gap-2 text-sm">
									<FiPlay />
									Marked tools as "Run now" will be returned to caller with args for execution.
								</div>
							)}
						</div>
					</section>

					<section>
						<h4 className="mb-2 font-semibold">Blocks (local override)</h4>
						<div className="space-y-3">
							{blockEdits.map((b, idx) => (
								<div key={b.id} className="rounded border p-2">
									<div className="mb-2 flex items-center gap-2 text-sm opacity-70">
										<span className="badge badge-outline">{b.role}</span>
										<span className="opacity-60">#{idx + 1}</span>
									</div>
									<textarea
										className="textarea textarea-bordered min-h-24 w-full"
										value={b.content}
										onChange={e => {
											setBlockEdits(arr => {
												const next = [...arr];
												next[idx] = { ...next[idx], content: e.target.value };
												return next;
											});
										}}
									/>
								</div>
							))}
						</div>
					</section>
				</div>

				<div className="modal-action">
					<button className="btn btn-ghost" onClick={onClose}>
						Cancel
					</button>
					<button className="btn btn-primary" onClick={saveAndClose}>
						Save
					</button>
				</div>
			</div>
			<form method="dialog" className="modal-backdrop" onSubmit={onClose}>
				<button>close</button>
			</form>
		</dialog>,
		document.body
	);
}

function VariableEditorRow({
	varDef,
	value,
	onChange,
}: {
	varDef: PromptVariable;
	value: unknown;
	onChange: (val: unknown) => void;
}) {
	const label = `${varDef.name}${varDef.required ? ' *' : ''}`;
	const help =
		varDef.description ??
		(varDef.source === VarSource.Static
			? `Static${varDef.staticVal ? `: ${varDef.staticVal}` : ''}`
			: varDef.source === VarSource.Tool
				? `From tool${varDef.default ? `, default: ${varDef.default}` : ''}`
				: `User${varDef.default ? `, default: ${varDef.default}` : ''}`);

	const common = (
		<div className="text-xs opacity-70">
			{help}
			{varDef.type === VarType.Enum && varDef.enumValues?.length ? ` | options: ${varDef.enumValues.join(', ')}` : ''}
		</div>
	);

	switch (varDef.type) {
		case VarType.Boolean:
			return (
				<label className="form-control w-full">
					<div className="label">
						<span className="label-text">{label}</span>
					</div>
					<input
						type="checkbox"
						className="toggle"
						checked={Boolean(value)}
						onChange={e => {
							onChange(e.target.checked);
						}}
					/>
					{common}
				</label>
			);
		case VarType.Number:
			return (
				<label className="form-control w-full">
					<div className="label">
						<span className="label-text">{label}</span>
					</div>
					<input
						type="number"
						className="input input-bordered input-sm"
						value={value === undefined || value === null ? '' : (value as string)}
						onChange={e => {
							const v = e.target.value;
							onChange(v === '' ? undefined : Number(v));
						}}
					/>
					{common}
				</label>
			);
		case VarType.Enum:
			return (
				<label className="form-control w-full">
					<div className="label">
						<span className="label-text">{label}</span>
					</div>
					<select
						className="select select-bordered select-sm"
						value={value === undefined || value === null ? '' : (value as string)}
						onChange={e => {
							onChange(e.target.value || undefined);
						}}
					>
						<option value="">-- select --</option>
						{(varDef.enumValues ?? []).map(opt => (
							<option value={opt} key={opt}>
								{opt}
							</option>
						))}
					</select>
					{common}
				</label>
			);
		case VarType.Date:
			return (
				<label className="form-control w-full">
					<div className="label">
						<span className="label-text">{label}</span>
					</div>
					<input
						type="date"
						className="input input-bordered input-sm"
						value={value ? (value as string) : ''}
						onChange={e => {
							onChange(e.target.value || undefined);
						}}
					/>
					{common}
				</label>
			);
		case VarType.String:
		default:
			return (
				<label className="form-control w-full">
					<div className="label">
						<span className="label-text">{label}</span>
					</div>
					<input
						className="input input-bordered input-sm"
						value={value === undefined || value === null ? '' : (value as string)}
						onChange={e => {
							onChange(e.target.value || undefined);
						}}
					/>
					{common}
				</label>
			);
	}
}

function PreProcessorRow({
	call,
	args,
	status,
	onArgsChange,
	onRunNow,
}: {
	call: PreProcessorCall;
	args?: Record<string, any>;
	status: ToolState['status'];
	onArgsChange: (next: Record<string, any>) => void;
	onRunNow: () => void;
}) {
	const [argsText, setArgsText] = React.useState<string>(JSON.stringify(args ?? {}, null, 2));
	const [jsonError, setJsonError] = React.useState<string | undefined>();

	React.useEffect(() => {
		setArgsText(JSON.stringify(args ?? {}, null, 2));
	}, [args]);

	function handleBlur() {
		try {
			const parsed = argsText.trim() ? JSON.parse(argsText) : {};
			setJsonError(undefined);
			onArgsChange(parsed);
		} catch {
			setJsonError('Invalid JSON');
		}
	}

	return (
		<div className="space-y-2 rounded border p-2">
			<div className="flex items-center justify-between">
				<div className="text-sm">
					<div className="flex items-center gap-2 font-medium">
						<span className="badge badge-outline">Tool</span>
						<span>{call.toolID}</span>
					</div>
					<div className="opacity-70">
						Save as: <code>{call.saveAs}</code>
						{call.pathExpr ? (
							<>
								{' '}
								| Path: <code>{call.pathExpr}</code>
							</>
						) : null}
						{call.onError ? <> | onError: {call.onError}</> : null}
					</div>
				</div>
				<div className="flex items-center gap-2">
					<span
						className={`badge badge-sm ${
							status === 'done' ? 'badge-success' : status === 'ready' ? 'badge-warning' : 'badge-neutral'
						}`}
						title="Status"
					>
						{status}
					</span>
					<button className="btn btn-xs" onClick={onRunNow} type="button" title="Mark to run now">
						<FiPlay className="mr-1" /> Run tool now
					</button>
				</div>
			</div>
			<div>
				<div className="label">
					<span className="label-text">Args (JSON)</span>
				</div>
				<textarea
					className={`textarea textarea-bordered w-full font-mono ${jsonError ? 'textarea-error' : ''}`}
					rows={6}
					value={argsText}
					onChange={e => {
						setArgsText(e.target.value);
					}}
					onBlur={handleBlur}
				/>
				{jsonError && <div className="text-error mt-1 text-xs">{jsonError}</div>}
			</div>
		</div>
	);
}

// Safe wrapper to avoid circular import of VarType/VarSource enums in UI file
function effectiveVarValueSafe(
	varDef: PromptVariable,
	userValues: Record<string, unknown>,
	toolStates?: Record<string, ToolState>
): unknown {
	// Inline minimal effective logic to avoid re-import if desired
	if (userValues[varDef.name] !== undefined && userValues[varDef.name] !== null) {
		return userValues[varDef.name];
	}
	if (varDef.source === VarSource.Static && varDef.staticVal) {
		return varDef.staticVal;
	}
	if (varDef.default !== undefined && varDef.default !== '') {
		return varDef.default;
	}
	// Try tool result if any attached
	if (varDef.source === VarSource.Tool && toolStates) {
		const hit = Object.values(toolStates).find(st => st.result !== undefined);
		if (hit?.result !== undefined) return hit.result;
	}
	return undefined;
}
