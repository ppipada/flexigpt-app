import { type ChangeEvent, type FormEvent, useEffect, useMemo, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { TOOL_INVOKE_CHAR } from '@/spec/command';
import { type Tool, ToolType } from '@/spec/tool';

import { omitManyKeys } from '@/lib/obj_utils';
import { validateSlug, validateTags } from '@/lib/text_utils';
import { MessageEnterValidURL, validateUrlForInput } from '@/lib/url_utils';

import { Dropdown } from '@/components/dropdown';

interface ToolItem {
	tool: Tool;
	bundleID: string;
	toolSlug: string;
}

interface AddEditToolModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (toolData: Partial<Tool>) => void;
	initialData?: ToolItem; // when editing
	existingTools: ToolItem[];
}

const TOOL_TYPE_LABEL_GO = 'Go';
const TOOL_TYPE_LABEL_HTTP = 'HTTP';

type ErrorState = {
	displayName?: string;
	slug?: string;
	type?: string;
	argSchema?: string;
	outputSchema?: string;
	goFunc?: string;
	httpUrl?: string;
	tags?: string;
};

export function AddEditToolModal({ isOpen, onClose, onSubmit, initialData, existingTools }: AddEditToolModalProps) {
	const [formData, setFormData] = useState({
		displayName: '',
		slug: '',
		description: '',
		tags: '',
		isEnabled: true,
		type: ToolType.HTTP as ToolType,
		argSchema: '{}',
		outputSchema: '{}',
		goFunc: '',
		httpUrl: '',
		httpMethod: 'GET',
		httpHeaders: '{}',
		httpQuery: '{}',
		httpBody: '',
		httpAuthType: '',
		httpAuthName: '',
		httpAuthValueTemplate: '',
		httpResponseCodes: '',
		httpResponseErrorMode: '',
		version: '1',
	});

	const [errors, setErrors] = useState<ErrorState>({});
	const isEditMode = Boolean(initialData);

	const dialogRef = useRef<HTMLDialogElement | null>(null);
	const displayNameInputRef = useRef<HTMLInputElement | null>(null);
	const httpUrlInputRef = useRef<HTMLInputElement | null>(null);

	// Sync prop -> state
	useEffect(() => {
		if (!isOpen) return;

		if (initialData) {
			const t = initialData.tool;
			setFormData({
				displayName: t.displayName,
				slug: t.slug,
				description: t.description ?? '',
				tags: (t.tags ?? []).join(', '),
				isEnabled: t.isEnabled,
				type: t.type,
				argSchema: JSON.stringify(t.argSchema, null, 2),
				outputSchema: JSON.stringify(t.outputSchema, null, 2),
				goFunc: t.goImpl?.func ?? '',
				httpUrl: t.httpImpl?.request.urlTemplate ?? '',
				httpMethod: t.httpImpl?.request.method ?? 'GET',
				httpHeaders: JSON.stringify(t.httpImpl?.request.headers ?? {}, null, 2),
				httpQuery: JSON.stringify(t.httpImpl?.request.query ?? {}, null, 2),
				httpBody: t.httpImpl?.request.body ?? '',
				httpAuthType: t.httpImpl?.request.auth?.type ?? '',
				httpAuthName: t.httpImpl?.request.auth?.name ?? '',
				httpAuthValueTemplate: t.httpImpl?.request.auth?.valueTemplate ?? '',
				httpResponseCodes: (t.httpImpl?.response.successCodes ?? []).join(','),
				httpResponseErrorMode: t.httpImpl?.response.errorMode ?? '',
				version: t.version,
			});
		} else {
			setFormData({
				displayName: '',
				slug: '',
				description: '',
				tags: '',
				isEnabled: true,
				type: ToolType.HTTP,
				argSchema: '{}',
				outputSchema: '{}',
				goFunc: '',
				httpUrl: '',
				httpMethod: 'GET',
				httpHeaders: '{}',
				httpQuery: '{}',
				httpBody: '',
				httpAuthType: '',
				httpAuthName: '',
				httpAuthValueTemplate: '',
				httpResponseCodes: '',
				httpResponseErrorMode: '',
				version: '1',
			});
		}
		setErrors({});
	}, [isOpen, initialData]);

	// Open/close native dialog + focus first field
	useEffect(() => {
		if (!isOpen) return;
		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) {
			dialog.showModal();
		}

		window.setTimeout(() => {
			displayNameInputRef.current?.focus();
		}, 0);

		return () => {
			if (dialog.open) dialog.close();
		};
	}, [isOpen]);

	const handleDialogClose = () => {
		onClose();
	};

	const toolTypeDropdownItems = useMemo(
		() =>
			({
				[ToolType.Go]: { isEnabled: isEditMode, displayName: TOOL_TYPE_LABEL_GO },
				[ToolType.HTTP]: { isEnabled: true, displayName: TOOL_TYPE_LABEL_HTTP },
			}) as Record<ToolType, { isEnabled: boolean; displayName: string }>,
		[isEditMode]
	);

	// Validation
	const validateField = (field: keyof ErrorState, val: string, currentErrors: ErrorState): ErrorState => {
		let newErrs: ErrorState = { ...currentErrors };
		const v = val.trim();

		if (!v && ['displayName', 'slug', 'type'].includes(field)) {
			newErrs[field] = 'This field is required.';
			return newErrs;
		}

		if (field === 'slug') {
			if (v.startsWith(TOOL_INVOKE_CHAR)) {
				newErrs.slug = `Do not prefix with "${TOOL_INVOKE_CHAR}".`;
			} else {
				const err = validateSlug(v);
				if (err) {
					newErrs.slug = err;
				} else {
					const clash = existingTools.some(t => t.tool.slug === v && t.tool.id !== initialData?.tool.id);
					if (clash) newErrs.slug = 'Slug already in use.';
					else newErrs = omitManyKeys(newErrs, ['slug']);
				}
			}
		} else if (field === 'tags') {
			const err = validateTags(val);
			if (err) newErrs.tags = err;
			else newErrs = omitManyKeys(newErrs, ['tags']);
		} else if (field === 'argSchema' || field === 'outputSchema') {
			if (!val.trim()) {
				// Allow blank here; "required" is enforced via isAllValid
				newErrs = omitManyKeys(newErrs, [field]);
			} else {
				try {
					JSON.parse(val);
					newErrs = omitManyKeys(newErrs, [field]);
				} catch {
					newErrs[field] = 'Invalid JSON';
				}
			}
		} else if (field === 'goFunc' && formData.type === ToolType.Go) {
			if (!v) newErrs.goFunc = 'Go function is required.';
			else newErrs = omitManyKeys(newErrs, ['goFunc']);
		} else if (field === 'httpUrl' && formData.type === ToolType.HTTP) {
			// Use shared URL validator; HTTP URL is required when type=HTTP
			const { error } = validateUrlForInput(v, httpUrlInputRef.current, {
				required: true,
				requiredMessage: 'HTTP URL is required.',
			});

			if (error) newErrs.httpUrl = error;
			else newErrs = omitManyKeys(newErrs, ['httpUrl']);
		} else {
			newErrs = omitManyKeys(newErrs, [field]);
		}

		return newErrs;
	};

	const validateForm = (state: typeof formData): ErrorState => {
		let newErrs: ErrorState = {};
		newErrs = validateField('displayName', state.displayName, newErrs);
		newErrs = validateField('slug', state.slug, newErrs);
		newErrs = validateField('type', state.type, newErrs);
		newErrs = validateField('argSchema', state.argSchema, newErrs);
		newErrs = validateField('outputSchema', state.outputSchema, newErrs);
		newErrs = validateField('tags', state.tags, newErrs);
		if (state.type === ToolType.Go) {
			newErrs = validateField('goFunc', state.goFunc, newErrs);
		}
		if (state.type === ToolType.HTTP) {
			newErrs = validateField('httpUrl', state.httpUrl, newErrs);
		}
		return newErrs;
	};

	const handleInput = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;
		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (['displayName', 'slug', 'type', 'argSchema', 'outputSchema', 'goFunc', 'httpUrl', 'tags'].includes(name)) {
			setErrors(prev => validateField(name as keyof ErrorState, String(newVal), prev));
		}
	};

	// Overall validity
	const isAllValid = useMemo(() => {
		const errs = Object.values(errors).some(Boolean);
		const filled =
			formData.displayName.trim() &&
			formData.slug.trim() &&
			formData.argSchema.trim() &&
			formData.outputSchema.trim() &&
			(formData.type === ToolType.Go ? formData.goFunc.trim() : formData.httpUrl.trim());
		return !errs && Boolean(filled);
	}, [errors, formData]);

	// Submit
	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();

		const nextErrors = validateForm(formData);
		setErrors(nextErrors);
		if (Object.values(nextErrors).some(Boolean)) return;

		const tagsArr = formData.tags
			.split(',')
			.map(t => t.trim())
			.filter(Boolean);

		let goImpl = undefined;
		let httpImpl = undefined;

		if (formData.type === ToolType.Go) {
			goImpl = { func: formData.goFunc.trim() };
		} else {
			// Normalize & validate HTTP URL again to get canonical form
			const httpUrlInput = httpUrlInputRef.current;
			const { normalized: normalizedHttpUrl, error: httpUrlError } = validateUrlForInput(
				formData.httpUrl,
				httpUrlInput,
				{
					required: true,
					requiredMessage: 'HTTP URL is required.',
				}
			);

			if (!normalizedHttpUrl || httpUrlError) {
				setErrors(prev => ({
					...prev,
					httpUrl: httpUrlError ?? MessageEnterValidURL,
				}));
				httpUrlInput?.focus();
				return;
			}

			let headers: Record<string, string> | undefined;
			let query: Record<string, string> | undefined;

			try {
				headers = formData.httpHeaders ? JSON.parse(formData.httpHeaders) : undefined;
			} catch {
				setErrors(prev => ({ ...prev, argSchema: prev.argSchema }));
				return;
			}

			try {
				query = formData.httpQuery ? JSON.parse(formData.httpQuery) : undefined;
			} catch {
				setErrors(prev => ({ ...prev, argSchema: prev.argSchema }));
				return;
			}

			httpImpl = {
				request: {
					method: formData.httpMethod || 'GET',
					urlTemplate: normalizedHttpUrl,
					headers,
					query,
					body: formData.httpBody || undefined,
					auth: formData.httpAuthType
						? {
								type: formData.httpAuthType,
								name: formData.httpAuthName || undefined,
								valueTemplate: formData.httpAuthValueTemplate,
							}
						: undefined,
				},
				response: {
					successCodes: formData.httpResponseCodes
						? formData.httpResponseCodes
								.split(',')
								.map(s => Number(s.trim()))
								.filter(Boolean)
						: undefined,
					errorMode: formData.httpResponseErrorMode || undefined,
				},
			};
		}

		onSubmit({
			displayName: formData.displayName.trim(),
			slug: formData.slug.trim(),
			description: formData.description.trim() || undefined,
			isEnabled: formData.isEnabled,
			tags: tagsArr.length ? tagsArr : undefined,
			type: formData.type,
			argSchema: formData.argSchema,
			outputSchema: formData.outputSchema,
			goImpl,
			httpImpl,
			version: formData.version,
		});

		dialogRef.current?.close();
	};

	const onToolTypeChange = (key: ToolType) => {
		// Prevent selecting Go when adding a new tool
		if (!isEditMode && key === ToolType.Go) {
			return;
		}

		setFormData(prev => ({ ...prev, type: key }));
		setErrors(prev => validateField('type', key, prev));
	};

	if (!isOpen) return null;

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-3xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="text-lg font-bold">{isEditMode ? 'Edit Tool' : 'Add Tool'}</h3>
						<button
							type="button"
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

					{/* form */}
					<form noValidate onSubmit={handleSubmit} className="space-y-4">
						{/* Display Name */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Display Name*</span>
							</label>
							<div className="col-span-9">
								<input
									ref={displayNameInputRef}
									type="text"
									name="displayName"
									value={formData.displayName}
									onChange={handleInput}
									className={`input input-bordered w-full rounded-xl ${errors.displayName ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									aria-invalid={Boolean(errors.displayName)}
								/>
								{errors.displayName && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.displayName}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Slug */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Slug*</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip="Lower-case, URL-friendly.">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<div className="relative">
									<span className="text-neutral-custom absolute top-1/2 left-3 -translate-y-1/2">
										{TOOL_INVOKE_CHAR}
									</span>
									<input
										type="text"
										name="slug"
										value={formData.slug}
										onChange={handleInput}
										className={`input input-bordered w-full rounded-xl pl-8 ${errors.slug ? 'input-error' : ''}`}
										spellCheck="false"
										autoComplete="off"
										disabled={isEditMode && initialData?.tool.isBuiltIn}
										aria-invalid={Boolean(errors.slug)}
									/>
								</div>
								{errors.slug && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.slug}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Enabled toggle */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3 cursor-pointer">
								<span className="label-text text-sm">Enabled</span>
							</label>
							<div className="col-span-9">
								<input
									type="checkbox"
									name="isEnabled"
									checked={formData.isEnabled}
									onChange={handleInput}
									className="toggle toggle-accent"
								/>
							</div>
						</div>

						{/* Type */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Type*</span>
							</label>
							<div className={`col-span-9 ${isEditMode ? 'pointer-events-none opacity-60' : ''}`}>
								<Dropdown<ToolType>
									dropdownItems={toolTypeDropdownItems}
									selectedKey={formData.type}
									onChange={onToolTypeChange}
									filterDisabled={true}
									title="Select tool type"
									getDisplayName={k => toolTypeDropdownItems[k].displayName}
								/>
								{errors.type && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.type}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Description */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Description</span>
							</label>
							<div className="col-span-9">
								<textarea
									name="description"
									value={formData.description}
									onChange={handleInput}
									className="textarea textarea-bordered h-20 w-full rounded-xl"
									spellCheck="false"
								/>
							</div>
						</div>

						{/* Arg Schema */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Arg JSONSchema*</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip="JSON Schema for arguments">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<textarea
									name="argSchema"
									value={formData.argSchema}
									onChange={handleInput}
									className={`textarea textarea-bordered h-24 w-full rounded-xl ${
										errors.argSchema ? 'textarea-error' : ''
									}`}
									spellCheck="false"
									aria-invalid={Boolean(errors.argSchema)}
								/>
								{errors.argSchema && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.argSchema}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Output Schema */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Output JSONSchema*</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip="JSON Schema for output">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<textarea
									name="outputSchema"
									value={formData.outputSchema}
									onChange={handleInput}
									className={`textarea textarea-bordered h-24 w-full rounded-xl ${
										errors.outputSchema ? 'textarea-error' : ''
									}`}
									spellCheck="false"
									aria-invalid={Boolean(errors.outputSchema)}
								/>
								{errors.outputSchema && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.outputSchema}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Go Impl */}
						{formData.type === ToolType.Go && (
							<div className="grid grid-cols-12 items-center gap-2">
								<label className="label col-span-3">
									<span className="label-text text-sm">Go Func*</span>
									<span
										className="label-text-alt tooltip tooltip-right"
										data-tip="e.g. github.com/acme/flexigpt/tools.Weather"
									>
										<FiHelpCircle size={12} />
									</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="goFunc"
										value={formData.goFunc}
										onChange={handleInput}
										className={`input input-bordered w-full rounded-xl ${errors.goFunc ? 'input-error' : ''}`}
										spellCheck="false"
										autoComplete="off"
										aria-invalid={Boolean(errors.goFunc)}
									/>
									{errors.goFunc && (
										<div className="label">
											<span className="label-text-alt text-error flex items-center gap-1">
												<FiAlertCircle size={12} /> {errors.goFunc}
											</span>
										</div>
									)}
								</div>
							</div>
						)}

						{/* HTTP Impl */}
						{formData.type === ToolType.HTTP && (
							<>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">HTTP URL*</span>
									</label>
									<div className="col-span-9">
										<input
											ref={httpUrlInputRef}
											type="url"
											name="httpUrl"
											value={formData.httpUrl}
											onChange={handleInput}
											className={`input input-bordered w-full rounded-xl ${errors.httpUrl ? 'input-error' : ''}`}
											spellCheck="false"
											autoComplete="off"
											aria-invalid={Boolean(errors.httpUrl)}
											placeholder="https://api.example.com/endpoint OR api.example.com/endpoint"
										/>
										{errors.httpUrl && (
											<div className="label">
												<span className="label-text-alt text-error flex items-center gap-1">
													<FiAlertCircle size={12} /> {errors.httpUrl}
												</span>
											</div>
										)}
									</div>
								</div>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">HTTP Method</span>
									</label>
									<div className="col-span-9">
										<input
											type="text"
											name="httpMethod"
											value={formData.httpMethod}
											onChange={handleInput}
											className="input input-bordered w-full rounded-xl"
											spellCheck="false"
											autoComplete="off"
										/>
									</div>
								</div>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Headers (JSON)</span>
									</label>
									<div className="col-span-9">
										<textarea
											name="httpHeaders"
											value={formData.httpHeaders}
											onChange={handleInput}
											className="textarea textarea-bordered h-16 w-full rounded-xl"
											spellCheck="false"
										/>
									</div>
								</div>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Query (JSON)</span>
									</label>
									<div className="col-span-9">
										<textarea
											name="httpQuery"
											value={formData.httpQuery}
											onChange={handleInput}
											className="textarea textarea-bordered h-16 w-full rounded-xl"
											spellCheck="false"
										/>
									</div>
								</div>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Body</span>
									</label>
									<div className="col-span-9">
										<textarea
											name="httpBody"
											value={formData.httpBody}
											onChange={handleInput}
											className="textarea textarea-bordered h-16 w-full rounded-xl"
											spellCheck="false"
										/>
									</div>
								</div>
								{/* Auth */}
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Auth Type</span>
									</label>
									<div className="col-span-9">
										<input
											type="text"
											name="httpAuthType"
											value={formData.httpAuthType}
											onChange={handleInput}
											className="input input-bordered w-full rounded-xl"
											spellCheck="false"
											autoComplete="off"
										/>
									</div>
								</div>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Auth Name</span>
									</label>
									<div className="col-span-9">
										<input
											type="text"
											name="httpAuthName"
											value={formData.httpAuthName}
											onChange={handleInput}
											className="input input-bordered w-full rounded-xl"
											spellCheck="false"
											autoComplete="off"
										/>
									</div>
								</div>
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Auth Value Template</span>
									</label>
									<div className="col-span-9">
										<input
											type="text"
											name="httpAuthValueTemplate"
											value={formData.httpAuthValueTemplate}
											onChange={handleInput}
											className="input input-bordered w-full rounded-xl"
											spellCheck="false"
											autoComplete="off"
										/>
									</div>
								</div>
								{/* Response */}
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Success Codes (comma)</span>
									</label>
									<div className="col-span-9">
										<input
											type="text"
											name="httpResponseCodes"
											value={formData.httpResponseCodes}
											onChange={handleInput}
											className="input input-bordered w-full rounded-xl"
											spellCheck="false"
											autoComplete="off"
										/>
									</div>
								</div>

								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Error Mode</span>
									</label>
									<div className="col-span-9">
										<input
											type="text"
											name="httpResponseErrorMode"
											value={formData.httpResponseErrorMode}
											onChange={handleInput}
											className="input input-bordered w-full rounded-xl"
											spellCheck="false"
											autoComplete="off"
										/>
									</div>
								</div>
							</>
						)}

						{/* Tags */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Tags</span>
							</label>
							<div className="col-span-9">
								<input
									type="text"
									name="tags"
									value={formData.tags}
									onChange={handleInput}
									className={`input input-bordered w-full rounded-xl ${errors.tags ? 'input-error' : ''}`}
									placeholder="comma, separated, tags"
									spellCheck="false"
									aria-invalid={Boolean(errors.tags)}
								/>
								{errors.tags && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.tags}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* actions */}
						<div className="modal-action">
							<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
								Cancel
							</button>
							<button type="submit" className="btn btn-primary rounded-xl" disabled={!isAllValid}>
								Save
							</button>
						</div>
					</form>
				</div>
			</div>
			{/* NOTE: no modal-backdrop here: backdrop click should NOT close this modal */}
		</dialog>,
		document.body
	);
}
