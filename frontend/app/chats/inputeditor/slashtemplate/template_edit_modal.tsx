import React from 'react';

import type { PlateEditor } from 'platejs/react';
import { createPortal } from 'react-dom';
import { FiAlertCircle, FiHelpCircle, FiPlay, FiX } from 'react-icons/fi';

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
	// Compute current effective template for rendering
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
	const [varValues, setVarValues] = React.useState<Record<string, unknown>>(() => {
		const vals: Record<string, unknown> = { ...tsenode.variables };
		for (const v of variablesSchema) {
			const val = effectiveVarValueSafe(v, tsenode.variables, tsenode.toolStates);
			if (val !== undefined) vals[v.name] = val;
		}
		return vals;
	});

	// Preprocessor args and statuses
	const [toolArgs, setToolArgs] = React.useState<Record<string, Record<string, any>>>(() => {
		const curr: Record<string, Record<string, any>> = {};
		for (const p of preProcessors) {
			curr[p.id] = tsenode.toolStates?.[p.id]?.args ?? p.args ?? {};
		}
		return curr;
	});
	const [toolStatuses, setToolStatuses] = React.useState<Record<string, ToolState>>(tsenode.toolStates ?? {});

	// Track JSON validity of each tool args editor to safely enable/disable Save
	const [toolArgsValidity, setToolArgsValidity] = React.useState<Record<string, boolean>>(() => {
		const init: Record<string, boolean> = {};
		for (const p of preProcessors) init[p.id] = true;
		return init;
	});

	// Rehydrate form when opening or when the node reference changes
	React.useEffect(() => {
		if (!open) return;

		// Recompute effective template inside the effect to avoid depending on changing references
		const eff = computeEffectiveTemplate(tsenode);
		const t = eff.template;
		const blks = eff.blocks;
		const varsSchema = eff.variablesSchema;
		const tools = eff.preProcessors;

		setDisplayName(tsenode.overrides?.displayName ?? t?.displayName ?? tsenode.templateSlug);
		setDescription(tsenode.overrides?.description ?? t?.description ?? '');
		setTags((tsenode.overrides?.tags ?? t?.tags ?? []).join(', '));
		setBlockEdits(blks);

		const vals: Record<string, unknown> = { ...tsenode.variables };
		for (const v of varsSchema) {
			const val = effectiveVarValueSafe(v, tsenode.variables, tsenode.toolStates);
			if (val !== undefined) vals[v.name] = val;
		}
		setVarValues(vals);

		const targs: Record<string, Record<string, any>> = {};
		for (const p of tools) {
			targs[p.id] = tsenode.toolStates?.[p.id]?.args ?? p.args ?? {};
		}
		setToolArgs(targs);
		setToolStatuses(tsenode.toolStates ?? {});

		// Valid by default when hydrated from structured data
		const validity: Record<string, boolean> = {};
		for (const p of tools) validity[p.id] = true;
		setToolArgsValidity(validity);
	}, [open, tsenode]);

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

		// Merge tool states with edited args and current statuses
		const nextToolStates: Record<string, ToolState> = { ...(tsenode.toolStates ?? {}) };
		for (const p of preProcessors) {
			const prev = nextToolStates[p.id] ?? { status: 'pending' as ToolState['status'] };
			nextToolStates[p.id] = {
				...prev,
				args: toolArgs[p.id],
				// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
				status: toolStatuses[p.id]?.status ?? prev.status ?? 'pending',
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
	const hasToolJsonErrors = Object.values(toolArgsValidity).some(v => !v);

	if (!open) return null;

	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box max-h-[85vh] max-w-3xl overflow-auto rounded-2xl">
				{/* Header */}
				<div className="mb-4 flex items-center justify-between gap-2">
					<div className="flex items-center gap-2">
						<h3 className="text-lg font-bold">Edit Template</h3>
						<span className="badge badge-neutral">{displayName}</span>
					</div>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				<form
					onSubmit={e => {
						e.preventDefault();
						if (!hasToolJsonErrors) saveAndClose();
					}}
					className="space-y-6"
				>
					{/* Overview */}
					<section>
						<h4 className="text-base-content/70 mb-3 text-sm font-semibold tracking-wide uppercase">Overview</h4>
						<div className="space-y-3">
							<div className="grid grid-cols-12 items-center gap-3">
								<label className="label col-span-12 md:col-span-4">
									<span className="label-text text-sm">Display Name (local)</span>
									<span className="label-text-alt tooltip tooltip-right" data-tip="Local override; visible only here.">
										<FiHelpCircle size={12} />
									</span>
								</label>
								<div className="col-span-12 md:col-span-8">
									<input
										className="input input-bordered input-sm w-full rounded-xl"
										value={displayName}
										onChange={e => {
											setDisplayName(e.target.value);
										}}
										placeholder={template?.displayName ?? tsenode.templateSlug}
									/>
								</div>
							</div>

							<div className="grid grid-cols-12 items-center gap-3">
								<label className="label col-span-12 md:col-span-4">
									<span className="label-text text-sm">Tags</span>
									<span
										className="label-text-alt tooltip tooltip-right"
										data-tip="Comma-separated tags used for filtering."
									>
										<FiHelpCircle size={12} />
									</span>
								</label>
								<div className="col-span-12 md:col-span-8">
									<input
										className="input input-bordered input-sm w-full rounded-xl"
										value={tags}
										onChange={e => {
											setTags(e.target.value);
										}}
										placeholder="e.g. brainstorm, draft, review"
									/>
								</div>
							</div>

							<div className="grid grid-cols-12 items-start gap-3">
								<label className="label col-span-12 md:col-span-4">
									<span className="label-text text-sm">Description (local)</span>
									<span
										className="label-text-alt tooltip tooltip-right"
										data-tip="Local description for your reference."
									>
										<FiHelpCircle size={12} />
									</span>
								</label>
								<div className="col-span-12 md:col-span-8">
									<textarea
										className="textarea textarea-bordered w-full rounded-xl"
										value={description}
										onChange={e => {
											setDescription(e.target.value);
										}}
										placeholder={template?.description ?? 'Describe how this template should be used...'}
									/>
								</div>
							</div>
						</div>
					</section>

					<div className="divider before:bg-base-300 after:bg-base-300 my-0" />

					{/* Variables */}
					<section>
						<div className="mb-3 flex items-center justify-between">
							<h4 className="text-base-content/70 text-sm font-semibold tracking-wide uppercase">Variables</h4>
							{req.requiredCount > 0 ? (
								<div className="text-warning flex items-center gap-2 text-sm">
									<FiAlertCircle size={14} />
									<span>Required remaining: {req.requiredVariables.join(', ')}</span>
								</div>
							) : (
								<div className="badge badge-success badge-outline">All required variables provided</div>
							)}
						</div>

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
					</section>

					<div className="divider before:bg-base-300 after:bg-base-300 my-0" />

					{/* Pre-processors */}
					<section>
						<div className="mb-3 flex items-center justify-between">
							<h4 className="text-base-content/70 text-sm font-semibold tracking-wide uppercase">Pre-processors</h4>
							{hasToolJsonErrors ? (
								<div className="text-error flex items-center gap-2 text-sm">
									<FiAlertCircle size={14} />
									<span>Fix invalid JSON args before saving</span>
								</div>
							) : null}
						</div>

						<div className="space-y-3">
							{preProcessors.length === 0 && <div className="text-sm opacity-70">No pre-processors configured.</div>}
							{preProcessors.map(p => (
								<PreProcessorRow
									key={p.id}
									call={p}
									args={toolArgs[p.id]}
									// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
									status={toolStatuses[p.id]?.status ?? 'pending'}
									onArgsChange={next => {
										setToolArgs(s => ({ ...s, [p.id]: next }));
									}}
									onRunNow={() => {
										markToolReady(p);
									}}
									onValidityChange={valid => {
										setToolArgsValidity(prev => ({ ...prev, [p.id]: valid }));
									}}
								/>
							))}
							{preProcessors.some(p => toolStatuses[p.id].status !== 'done') && (
								<div className="flex items-center gap-2 text-sm opacity-80">
									<FiPlay />
									<span>Marked tools as "Run now" will be returned to the caller with args for execution.</span>
								</div>
							)}
						</div>
					</section>

					<div className="divider before:bg-base-300 after:bg-base-300 my-0" />

					{/* Blocks */}
					<section>
						<h4 className="text-base-content/70 mb-3 text-sm font-semibold tracking-wide uppercase">
							Blocks (local override)
						</h4>
						<div className="space-y-3">
							{blockEdits.map((b, idx) => (
								<div key={b.id} className="rounded-xl border p-3">
									<div className="mb-2 flex items-center gap-2 text-sm opacity-70">
										<span className="badge badge-outline">{b.role}</span>
										<span className="opacity-60">#{idx + 1}</span>
									</div>
									<textarea
										className="textarea textarea-bordered min-h-32 w-full rounded-xl"
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

					{/* Footer */}
					<div className="modal-action">
						<button type="button" className="btn rounded-xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-xl" disabled={hasToolJsonErrors}>
							Save
						</button>
					</div>
				</form>
			</div>

			<form
				method="dialog"
				className="modal-backdrop"
				onSubmit={e => {
					e.preventDefault();
					onClose();
				}}
			>
				<button aria-label="Close">close</button>
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
	const id = React.useId();
	const label = `${varDef.name}${varDef.required ? ' *' : ''}`;
	const sourceText =
		varDef.source === VarSource.Static
			? `Static${varDef.staticVal ? `: ${varDef.staticVal}` : ''}`
			: varDef.source === VarSource.Tool
				? `From tool${varDef.default ? `, default: ${varDef.default}` : ''}`
				: `User${varDef.default ? `, default: ${varDef.default}` : ''}`;
	const help = varDef.description ? `${varDef.description} (${sourceText})` : sourceText;

	const isDisabled = varDef.source === VarSource.Static;
	const commonHelp = (
		<div className="label">
			<span className="label-text-alt text-xs opacity-70">
				{help}
				{varDef.type === VarType.Enum && varDef.enumValues?.length ? ` | options: ${varDef.enumValues.join(', ')}` : ''}
			</span>
		</div>
	);

	const labelCol = (
		<label htmlFor={id} className="label col-span-12 md:col-span-4">
			<span className="label-text text-sm">{label}</span>
			<span className="label-text-alt tooltip tooltip-right" data-tip={help}>
				<FiHelpCircle size={12} />
			</span>
		</label>
	);

	switch (varDef.type) {
		case VarType.Boolean:
			return (
				<div className="grid grid-cols-12 items-center gap-3">
					{labelCol}
					<div className="col-span-12 md:col-span-8">
						<input
							id={id}
							type="checkbox"
							className="toggle toggle-accent rounded-full"
							checked={Boolean(value)}
							disabled={isDisabled}
							onChange={e => {
								onChange(e.target.checked);
							}}
						/>
						{commonHelp}
					</div>
				</div>
			);

		case VarType.Number:
			return (
				<div className="grid grid-cols-12 items-center gap-3">
					{labelCol}
					<div className="col-span-12 md:col-span-8">
						<input
							id={id}
							type="number"
							className="input input-bordered input-sm w-full rounded-xl"
							value={value === undefined || value === null ? '' : (value as number).toString()}
							disabled={isDisabled}
							onChange={e => {
								const v = e.target.value;
								onChange(v === '' ? undefined : Number(v));
							}}
							placeholder={varDef.default !== undefined ? varDef.default : ''}
						/>
						{commonHelp}
					</div>
				</div>
			);

		case VarType.Enum:
			return (
				<div className="grid grid-cols-12 items-center gap-3">
					{labelCol}
					<div className="col-span-12 md:col-span-8">
						<select
							id={id}
							className="select select-bordered select-sm w-full rounded-xl"
							value={value === undefined || value === null ? '' : (value as string)}
							disabled={isDisabled}
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
						{commonHelp}
					</div>
				</div>
			);

		case VarType.Date:
			return (
				<div className="grid grid-cols-12 items-center gap-3">
					{labelCol}
					<div className="col-span-12 md:col-span-8">
						<input
							id={id}
							type="date"
							className="input input-bordered input-sm w-full rounded-xl"
							value={value ? (value as string) : ''}
							disabled={isDisabled}
							onChange={e => {
								onChange(e.target.value || undefined);
							}}
						/>
						{commonHelp}
					</div>
				</div>
			);

		case VarType.String:
		default:
			return (
				<div className="grid grid-cols-12 items-center gap-3">
					{labelCol}
					<div className="col-span-12 md:col-span-8">
						<input
							id={id}
							className="input input-bordered input-sm w-full rounded-xl"
							value={value === undefined || value === null ? '' : (value as string)}
							disabled={isDisabled}
							onChange={e => {
								onChange(e.target.value || undefined);
							}}
							placeholder={varDef.default !== undefined ? varDef.default : ''}
						/>
						{commonHelp}
					</div>
				</div>
			);
	}
}

function PreProcessorRow({
	call,
	args,
	status,
	onArgsChange,
	onRunNow,
	onValidityChange,
}: {
	call: PreProcessorCall;
	args?: Record<string, any>;
	status?: ToolState['status'];
	onArgsChange: (next: Record<string, any>) => void;
	onRunNow: () => void;
	onValidityChange?: (valid: boolean) => void;
}) {
	const [argsText, setArgsText] = React.useState<string>(JSON.stringify(args ?? {}, null, 2));
	const [jsonError, setJsonError] = React.useState<string | undefined>();

	// Keep in sync when parent args change
	React.useEffect(() => {
		const next = JSON.stringify(args ?? {}, null, 2);
		setArgsText(next);
		setJsonError(undefined);
		onValidityChange?.(true);
	}, [args, onValidityChange]);

	// Validate as user types so parent can disable Save if invalid
	React.useEffect(() => {
		try {
			if (argsText.trim() === '') {
				setJsonError(undefined);
				onValidityChange?.(true);
			} else {
				JSON.parse(argsText);
				setJsonError(undefined);
				onValidityChange?.(true);
			}
		} catch {
			setJsonError('Invalid JSON');
			onValidityChange?.(false);
		}
	}, [argsText, onValidityChange]);

	function handleBlur() {
		// Commit only when valid or empty
		try {
			const parsed = argsText.trim() ? JSON.parse(argsText) : {};
			setJsonError(undefined);
			onValidityChange?.(true);
			onArgsChange(parsed);
		} catch {
			setJsonError('Invalid JSON');
			onValidityChange?.(false);
		}
	}

	return (
		<div className="space-y-2 rounded-xl border p-3">
			<div className="flex items-center justify-between gap-2">
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
							(status ?? 'pending') === 'done'
								? 'badge-success'
								: (status ?? 'pending') === 'ready'
									? 'badge-warning'
									: 'badge-neutral'
						}`}
						title="Status"
					>
						{status ?? 'pending'}
					</span>
					<button className="btn btn-xs rounded-xl" onClick={onRunNow} type="button" title="Mark to run now">
						<FiPlay className="mr-1" /> Run tool now
					</button>
				</div>
			</div>

			<div>
				<div className="label">
					<span className="label-text">Args (JSON)</span>
					<span className="label-text-alt tooltip tooltip-left" data-tip="JSON arguments to pass to the tool">
						<FiHelpCircle size={12} />
					</span>
				</div>
				<textarea
					className={`textarea textarea-bordered w-full rounded-xl font-mono ${jsonError ? 'textarea-error' : ''}`}
					rows={6}
					value={argsText}
					onChange={e => {
						setArgsText(e.target.value);
					}}
					onBlur={handleBlur}
					spellCheck={false}
				/>
				{jsonError && (
					<div className="text-error mt-1 flex items-center gap-1 text-xs">
						<FiAlertCircle size={12} /> {jsonError}
					</div>
				)}
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
	if (varDef.source === VarSource.Static && varDef.staticVal !== undefined) {
		return varDef.staticVal;
	}
	if (varDef.default !== undefined && varDef.default !== '') {
		return varDef.default;
	}
	if (toolStates) {
		const hit = Object.values(toolStates).find(st => st.result !== undefined);
		if (hit?.result !== undefined) return hit.result;
	}
	return undefined;
}
