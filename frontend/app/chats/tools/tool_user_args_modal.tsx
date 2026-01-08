import { type FormEvent, useEffect, useMemo, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiTool, FiX } from 'react-icons/fi';

import {
	getJSONObject,
	getPropertiesFromJSONSchema,
	getRequiredFromJSONSchema,
	type JSONObject,
	type JSONSchema,
} from '@/lib/jsonschema_utils';

import { computeToolUserArgsStatus } from '@/chats/tools/tool_editor_utils';

interface ToolUserArgsModalProps {
	isOpen: boolean;
	onClose: () => void;

	toolLabel: string;
	schema: JSONSchema | undefined;
	existingInstance?: string;
	onSave: (newInstance: string) => void;
}

type FormState = {
	rawJson: string;
};

// Shared type for "what args editor is currently open?"
export type ToolArgsTarget = { kind: 'attached'; selectionID: string } | { kind: 'conversation'; key: string };

export function ToolUserArgsModal({
	isOpen,
	onClose,
	toolLabel,
	schema,
	existingInstance,
	onSave,
}: ToolUserArgsModalProps) {
	const dialogRef = useRef<HTMLDialogElement | null>(null);
	const [form, setForm] = useState<FormState>({ rawJson: '' });
	const [error, setError] = useState<string | null>(null);

	// Derive simple hints from schema (memoized so deps are stable)
	const schemaObj = useMemo(() => getJSONObject(schema), [schema]);

	const properties = useMemo(() => getPropertiesFromJSONSchema(schema) ?? ({} as JSONObject), [schema]);

	const requiredKeys = useMemo(() => getRequiredFromJSONSchema(schema) ?? [], [schema]);

	const allKeys = useMemo(() => Object.keys(properties), [properties]);
	const optionalKeys = useMemo(() => allKeys.filter(k => !requiredKeys.includes(k)), [allKeys, requiredKeys]);

	// Initialize textarea content whenever the modal is opened.
	useEffect(() => {
		if (!isOpen) return;

		let initial = '';
		if (existingInstance && existingInstance.trim() !== '') {
			try {
				const parsed = JSON.parse(existingInstance);
				initial = JSON.stringify(parsed, null, 2);
			} catch {
				initial = existingInstance;
			}
		} else {
			const keys = Object.keys(properties);
			if (keys.length > 0) {
				const skeleton: Record<string, unknown> = {};
				for (const [key, prop] of Object.entries(properties)) {
					const propObj = getJSONObject(prop);
					if (propObj && Object.prototype.hasOwnProperty.call(propObj, 'default')) {
						skeleton[key] = propObj['default'];
					} else if (requiredKeys.includes(key)) {
						skeleton[key] = '';
					}
				}
				initial = JSON.stringify(skeleton, null, 2);
			} else {
				initial = existingInstance ?? '';
			}
		}

		setForm({ rawJson: initial });
		setError(null);
	}, [isOpen, existingInstance, properties, requiredKeys]);

	// Manage the native <dialog> lifecycle (open/close) based only on isOpen.
	useEffect(() => {
		if (!isOpen) return;

		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) {
			dialog.showModal();
		}

		return () => {
			// If unmounted while still open, ensure we close the dialog
			if (dialog.open) {
				dialog.close();
			}
		};
	}, [isOpen]);

	const handleDialogClose = () => {
		onClose();
	};

	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();
		e.stopPropagation();

		let raw = form.rawJson.trim();
		if (!raw) {
			// For now, require *something* when schema has required fields.
			if (requiredKeys.length > 0) {
				setError('This tool requires options. Provide a JSON object with the required keys.');
				return;
			}
			raw = '{}'; // treat empty as an empty object
		}

		let parsed: unknown;
		try {
			parsed = JSON.parse(raw);
		} catch (err) {
			setError((err as Error).message || 'Invalid JSON. Please fix it and try again.');
			return;
		}

		if (!parsed || typeof parsed !== 'object') {
			setError('Expected a JSON object.');
			return;
		}

		// Reuse the same status helper we use elsewhere for required validation.
		const status = computeToolUserArgsStatus(schema, JSON.stringify(parsed));
		if (status.hasSchema && !status.isSatisfied) {
			setError(`Missing required keys: ${status.missingRequired.join(', ')}. Populate them (non-empty) before saving.`);
			return;
		}

		// Pretty-print for storage; backend only cares that it's valid JSON.
		const pretty = JSON.stringify(parsed, null, 2);
		onSave(pretty);
		dialogRef.current?.close();
	};

	if (!isOpen) return null;

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-2xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* header */}
					<div className="mb-4 flex items-center justify-between gap-2">
						<h3 className="flex items-center gap-2 text-lg font-bold">
							<FiTool size={16} />
							<span>Tool options</span>
							<span className="badge badge-neutral">{toolLabel}</span>
						</h3>
						<button
							type="button"
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

					<form
						noValidate
						onSubmit={handleSubmit}
						className="space-y-4"
						onKeyDownCapture={e => {
							e.stopPropagation();
						}}
						onKeyUpCapture={e => {
							e.stopPropagation();
						}}
					>
						{/* Schema hints */}
						<div className="space-y-1 text-xs">
							{schemaObj ? (
								<>
									<div className="font-semibold">JSON schema</div>
									{requiredKeys.length > 0 ? (
										<div>
											<span className="font-semibold">Required keys:</span> {requiredKeys.join(', ')}
										</div>
									) : (
										<div>There are no required keys for this tool.</div>
									)}
									{optionalKeys.length > 0 && (
										<div>
											<span className="font-semibold">Optional keys:</span> {optionalKeys.join(', ')}
										</div>
									)}
								</>
							) : (
								<div>This tool does not define a user configuration schema.</div>
							)}
						</div>

						{/* JSON textarea */}
						<div>
							<label className="label p-1">
								<span className="label-text text-sm">Options (JSON)</span>
							</label>
							<textarea
								className={`textarea textarea-bordered w-full rounded-xl font-mono text-xs ${
									error ? 'textarea-error' : ''
								}`}
								rows={12}
								value={form.rawJson}
								onChange={e => {
									setForm({ rawJson: e.target.value });
									if (error) setError(null);
								}}
								spellCheck={false}
								placeholder={schemaObj ? '{\n  "key": "value"\n}' : 'JSON object; structure depends on this tool.'}
							/>
							<div className="mt-1 h-5 text-xs">
								{error && (
									<span className="text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {error}
									</span>
								)}
							</div>
						</div>

						{/* footer */}
						<div className="modal-action">
							<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
								Cancel
							</button>
							<button type="submit" className="btn btn-primary rounded-xl">
								Save
							</button>
						</div>
					</form>
				</div>
			</div>
			<form method="dialog" className="modal-backdrop">
				<button aria-label="Close" />
			</form>
		</dialog>,
		document.body
	);
}
