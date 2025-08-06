import React, { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { TOOL_INVOKE_CHAR } from '@/spec/command';
import { type Tool, ToolType } from '@/spec/tool';

import { omitManyKeys } from '@/lib/obj_utils';
import { validateSlug, validateTags } from '@/lib/text_utils';

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

const TOOL_TYPE_OPTIONS = [
	{ value: ToolType.Go, label: 'Go' },
	{ value: ToolType.HTTP, label: 'HTTP' },
];

const AddEditToolModal: React.FC<AddEditToolModalProps> = ({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingTools,
}) => {
	const [formData, setFormData] = useState({
		displayName: '',
		slug: '',
		description: '',
		tags: '',
		isEnabled: true,
		type: ToolType.Go as ToolType,
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
		httpResponseEncoding: '',
		httpResponseSelector: '',
		httpResponseErrorMode: '',
		version: '1',
	});

	const [errors, setErrors] = useState<{
		displayName?: string;
		slug?: string;
		type?: string;
		argSchema?: string;
		outputSchema?: string;
		goFunc?: string;
		httpUrl?: string;
		tags?: string;
	}>({});

	const isEditMode = Boolean(initialData);

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
				argSchema: JSON.stringify(t.argSchema ?? {}, null, 2),
				outputSchema: JSON.stringify(t.outputSchema ?? {}, null, 2),
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
				httpResponseEncoding: t.httpImpl?.response.encoding ?? '',
				httpResponseSelector: t.httpImpl?.response.selector ?? '',
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
				type: ToolType.Go,
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
				httpResponseEncoding: '',
				httpResponseSelector: '',
				httpResponseErrorMode: '',
				version: '1',
			});
		}
		setErrors({});
	}, [isOpen, initialData]);

	// Validation
	const validateField = (field: keyof typeof errors, val: string) => {
		let newErrs = { ...errors };
		const v = val.trim();

		if (!v && ['displayName', 'slug', 'type'].includes(field)) {
			newErrs[field] = 'This field is required.';
		} else if (field === 'slug') {
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
			try {
				JSON.parse(val);
				newErrs = omitManyKeys(newErrs, [field]);
			} catch {
				newErrs[field] = 'Invalid JSON';
			}
		} else if (field === 'goFunc' && formData.type === ToolType.Go && !v) {
			newErrs.goFunc = 'Go function is required.';
		} else if (field === 'httpUrl' && formData.type === ToolType.HTTP && !v) {
			newErrs.httpUrl = 'HTTP URL is required.';
		} else {
			newErrs = omitManyKeys(newErrs, [field]);
		}
		setErrors(newErrs);
	};

	const handleInput = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;
		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (['displayName', 'slug', 'type', 'argSchema', 'outputSchema', 'goFunc', 'httpUrl', 'tags'].includes(name)) {
			validateField(name as keyof typeof errors, String(newVal));
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
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		validateField('displayName', formData.displayName);
		validateField('slug', formData.slug);
		validateField('type', formData.type);
		validateField('argSchema', formData.argSchema);
		validateField('outputSchema', formData.outputSchema);
		validateField('tags', formData.tags);

		if (formData.type === ToolType.Go) validateField('goFunc', formData.goFunc);
		if (formData.type === ToolType.HTTP) validateField('httpUrl', formData.httpUrl);

		if (!isAllValid) return;

		const tagsArr = formData.tags
			.split(',')
			.map(t => t.trim())
			.filter(Boolean);

		let goImpl = undefined;
		let httpImpl = undefined;

		if (formData.type === ToolType.Go) {
			goImpl = { func: formData.goFunc.trim() };
		} else {
			httpImpl = {
				request: {
					method: formData.httpMethod || 'GET',
					urlTemplate: formData.httpUrl,
					headers: formData.httpHeaders ? JSON.parse(formData.httpHeaders) : undefined,
					query: formData.httpQuery ? JSON.parse(formData.httpQuery) : undefined,
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
					encoding: formData.httpResponseEncoding || undefined,
					selector: formData.httpResponseSelector || undefined,
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
			argSchema: JSON.parse(formData.argSchema),
			outputSchema: JSON.parse(formData.outputSchema),
			goImpl,
			httpImpl,
			version: formData.version,
		});
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Tool' : 'Add Tool'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* form */}
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Display Name */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Display Name*</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="displayName"
								value={formData.displayName}
								onChange={handleInput}
								className={`input input-bordered w-full rounded-2xl ${errors.displayName ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
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
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Slug*</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="Lower-case, URL-friendly.">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<div className="relative">
								<span className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-custom">{TOOL_INVOKE_CHAR}</span>
								<input
									type="text"
									name="slug"
									value={formData.slug}
									onChange={handleInput}
									className={`input input-bordered w-full pl-8 rounded-2xl ${errors.slug ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									disabled={isEditMode && initialData?.tool.isBuiltIn}
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
								className="toggle toggle-accent rounded-full"
							/>
						</div>
					</div>

					{/* Type */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Type*</span>
						</label>
						<div className="col-span-9">
							<select
								name="type"
								value={formData.type}
								onChange={handleInput}
								className="select select-bordered w-full rounded-2xl"
								disabled={isEditMode}
							>
								{TOOL_TYPE_OPTIONS.map(opt => (
									<option key={opt.value} value={opt.value}>
										{opt.label}
									</option>
								))}
							</select>
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
					<div className="grid grid-cols-12 gap-2 items-start">
						<label className="label col-span-3">
							<span className="label-text text-sm">Description</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="description"
								value={formData.description}
								onChange={handleInput}
								className="textarea textarea-bordered w-full rounded-2xl h-20"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Arg Schema */}
					<div className="grid grid-cols-12 gap-2 items-start">
						<label className="label col-span-3">
							<span className="label-text text-sm">Arg Schema (JSON)*</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="JSON Schema for arguments">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="argSchema"
								value={formData.argSchema}
								onChange={handleInput}
								className={`textarea textarea-bordered w-full rounded-2xl h-24 ${errors.argSchema ? 'textarea-error' : ''}`}
								spellCheck="false"
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
					<div className="grid grid-cols-12 gap-2 items-start">
						<label className="label col-span-3">
							<span className="label-text text-sm">Output Schema (JSON)*</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="JSON Schema for output">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="outputSchema"
								value={formData.outputSchema}
								onChange={handleInput}
								className={`textarea textarea-bordered w-full rounded-2xl h-24 ${errors.outputSchema ? 'textarea-error' : ''}`}
								spellCheck="false"
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
						<div className="grid grid-cols-12 gap-2 items-center">
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
									className={`input input-bordered w-full rounded-2xl ${errors.goFunc ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
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
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">HTTP URL*</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpUrl"
										value={formData.httpUrl}
										onChange={handleInput}
										className={`input input-bordered w-full rounded-2xl ${errors.httpUrl ? 'input-error' : ''}`}
										spellCheck="false"
										autoComplete="off"
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
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">HTTP Method</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpMethod"
										value={formData.httpMethod}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Headers (JSON)</span>
								</label>
								<div className="col-span-9">
									<textarea
										name="httpHeaders"
										value={formData.httpHeaders}
										onChange={handleInput}
										className="textarea textarea-bordered w-full rounded-2xl h-16"
										spellCheck="false"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Query (JSON)</span>
								</label>
								<div className="col-span-9">
									<textarea
										name="httpQuery"
										value={formData.httpQuery}
										onChange={handleInput}
										className="textarea textarea-bordered w-full rounded-2xl h-16"
										spellCheck="false"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Body</span>
								</label>
								<div className="col-span-9">
									<textarea
										name="httpBody"
										value={formData.httpBody}
										onChange={handleInput}
										className="textarea textarea-bordered w-full rounded-2xl h-16"
										spellCheck="false"
									/>
								</div>
							</div>
							{/* Auth */}
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Auth Type</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpAuthType"
										value={formData.httpAuthType}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Auth Name</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpAuthName"
										value={formData.httpAuthName}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Auth Value Template</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpAuthValueTemplate"
										value={formData.httpAuthValueTemplate}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							{/* Response */}
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Success Codes (comma)</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpResponseCodes"
										value={formData.httpResponseCodes}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Encoding</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpResponseEncoding"
										value={formData.httpResponseEncoding}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Selector</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpResponseSelector"
										value={formData.httpResponseSelector}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
							<div className="grid grid-cols-12 gap-2 items-center">
								<label className="label col-span-3">
									<span className="label-text text-sm">Error Mode</span>
								</label>
								<div className="col-span-9">
									<input
										type="text"
										name="httpResponseErrorMode"
										value={formData.httpResponseErrorMode}
										onChange={handleInput}
										className="input input-bordered w-full rounded-2xl"
										spellCheck="false"
										autoComplete="off"
									/>
								</div>
							</div>
						</>
					)}

					{/* Tags */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Tags</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="tags"
								value={formData.tags}
								onChange={handleInput}
								className={`input input-bordered w-full rounded-2xl ${errors.tags ? 'input-error' : ''}`}
								placeholder="comma, separated, tags"
								spellCheck="false"
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
						<button type="button" className="btn rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-2xl" disabled={!isAllValid}>
							Save
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default AddEditToolModal;
