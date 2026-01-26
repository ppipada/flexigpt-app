import { type ChangeEvent, type FormEvent, useEffect, useMemo, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { PromptRoleEnum, type PromptTemplate } from '@/spec/prompt';

import { omitManyKeys } from '@/lib/obj_utils';
import { validateSlug, validateTags } from '@/lib/text_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

interface TemplateItem {
	template: PromptTemplate;
	bundleID: string;
	templateSlug: string;
}

type ModalMode = 'add' | 'edit' | 'view';

interface AddEditPromptTemplateModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (templateData: Partial<PromptTemplate>) => void;
	initialData?: TemplateItem; // when editing/viewing
	existingTemplates: TemplateItem[];
	mode?: ModalMode;
}

type ErrorState = {
	displayName?: string;
	slug?: string;
	content?: string;
	tags?: string;
};

export function AddEditPromptTemplateModal({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingTemplates,
	mode,
}: AddEditPromptTemplateModalProps) {
	const effectiveMode: ModalMode = mode ?? (initialData ? 'edit' : 'add');
	const isViewMode = effectiveMode === 'view';
	const isEditMode = effectiveMode === 'edit';
	// const isAddMode = effectiveMode === 'add';

	const [formData, setFormData] = useState({
		displayName: '',
		slug: '',
		description: '',
		content: '',
		tags: '',
		isEnabled: true,
	});

	const [errors, setErrors] = useState<ErrorState>({});

	const dialogRef = useRef<HTMLDialogElement | null>(null);

	useEffect(() => {
		if (!isOpen) return;

		if (initialData) {
			setFormData({
				displayName: initialData.template.displayName,
				slug: initialData.template.slug,
				description: initialData.template.description ?? '',
				content: initialData.template.blocks[0]?.content ?? '',
				tags: (initialData.template.tags ?? []).join(', '),
				isEnabled: initialData.template.isEnabled,
			});
		} else {
			setFormData({
				displayName: '',
				slug: '',
				description: '',
				content: '',
				tags: '',
				isEnabled: true,
			});
		}
		setErrors({});
	}, [isOpen, initialData]);

	useEffect(() => {
		if (!isOpen) return;

		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) {
			dialog.showModal();
		}

		return () => {
			if (dialog.open) {
				dialog.close();
			}
		};
	}, [isOpen]);

	const handleDialogClose = () => {
		onClose();
	};

	const validateField = (field: keyof ErrorState, val: string, currentErrors: ErrorState): ErrorState => {
		let newErrs: ErrorState = { ...currentErrors };
		const v = val.trim();

		if (!v) {
			newErrs[field] = 'This field is required.';
			return newErrs;
		}

		if (field === 'slug') {
			const err = validateSlug(v);
			if (err) {
				newErrs.slug = err;
			} else {
				const clash = existingTemplates.some(t => t.template.slug === v && t.template.id !== initialData?.template.id);
				if (clash) newErrs.slug = 'Slug already in use.';
				else newErrs = omitManyKeys(newErrs, ['slug']);
			}
		} else if (field === 'tags') {
			const err = validateTags(val);
			if (err) newErrs.tags = err;
			else newErrs = omitManyKeys(newErrs, ['tags']);
		} else {
			newErrs = omitManyKeys(newErrs, [field]);
		}

		return newErrs;
	};

	const validateForm = (state: typeof formData): ErrorState => {
		let newErrs: ErrorState = {};
		newErrs = validateField('displayName', state.displayName, newErrs);
		newErrs = validateField('slug', state.slug, newErrs);
		newErrs = validateField('content', state.content, newErrs);

		// Fix: validate tags ONLY when non-empty
		if (state.tags.trim() !== '') {
			newErrs = validateField('tags', state.tags, newErrs);
		}

		return newErrs;
	};

	const handleInput = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;

		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (name === 'displayName' || name === 'slug' || name === 'content' || name === 'tags') {
			setErrors(prev => validateField(name as keyof ErrorState, String(newVal), prev));
		}
	};

	const isAllValid = useMemo(() => {
		if (isViewMode) return true;
		const requiredFilled = formData.displayName.trim() && formData.slug.trim() && formData.content.trim();
		const hasErrors = Object.keys(errors).length > 0;
		return Boolean(requiredFilled) && !hasErrors;
	}, [formData, errors, isViewMode]);

	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();
		e.stopPropagation();
		if (isViewMode) return;

		const trimmed = {
			displayName: formData.displayName.trim(),
			slug: formData.slug.trim(),
			description: formData.description.trim(),
			content: formData.content,
			tags: formData.tags,
			isEnabled: formData.isEnabled,
		};

		const newErrs = validateForm(trimmed);
		setErrors(newErrs);

		if (Object.keys(newErrs).length > 0) return;

		const tagsArr = trimmed.tags
			.split(',')
			.map(t => t.trim())
			.filter(Boolean);

		// Important: preserve existing blocks when editing, only update first block content
		const existingBlocks = initialData?.template.blocks ?? [];
		const blocks =
			isEditMode && existingBlocks.length > 0
				? existingBlocks.map((b, idx) => (idx === 0 ? { ...b, content: trimmed.content } : b))
				: [
						{
							id: getUUIDv7(),
							role: PromptRoleEnum.User,
							content: trimmed.content,
						},
					];

		onSubmit({
			displayName: trimmed.displayName,
			slug: trimmed.slug,
			description: trimmed.description || undefined,
			isEnabled: trimmed.isEnabled,
			tags: tagsArr.length ? tagsArr : undefined,
			blocks,
		});

		dialogRef.current?.close();
	};

	if (!isOpen) return null;

	const headerTitle =
		effectiveMode === 'view'
			? 'View Prompt Template'
			: effectiveMode === 'edit'
				? 'Edit Prompt Template'
				: 'Add Prompt Template';

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-3xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					<div className="mb-4 flex items-center justify-between">
						<h3 className="text-lg font-bold">{headerTitle}</h3>
						<button
							type="button"
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

					<form noValidate onSubmit={handleSubmit} className="space-y-4">
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Display Name*</span>
							</label>
							<div className="col-span-9">
								<input
									type="text"
									name="displayName"
									value={formData.displayName}
									onChange={handleInput}
									readOnly={isViewMode}
									className={`input input-bordered w-full rounded-xl ${errors.displayName ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									autoFocus={!isViewMode}
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

						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Slug*</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip="Short user friendly command">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<input
									type="text"
									name="slug"
									value={formData.slug}
									onChange={handleInput}
									className={`input input-bordered w-full rounded-xl ${errors.slug ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									disabled={isViewMode || isEditMode}
									aria-invalid={Boolean(errors.slug)}
								/>
								{errors.slug && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.slug}
										</span>
									</div>
								)}
							</div>
						</div>

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
									disabled={isViewMode}
								/>
							</div>
						</div>

						<div className="grid grid-cols-12 items-start gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Description</span>
							</label>
							<div className="col-span-9">
								<textarea
									name="description"
									value={formData.description}
									onChange={handleInput}
									readOnly={isViewMode}
									className="textarea textarea-bordered h-20 w-full rounded-xl"
									spellCheck="false"
								/>
							</div>
						</div>

						{/* Edit/Add content: still edits first block only */}
						{!isViewMode && (
							<div className="grid grid-cols-12 items-start gap-2">
								<label className="label col-span-3">
									<span className="label-text text-sm">Prompt Content*</span>
									<span
										className="label-text-alt tooltip tooltip-right"
										data-tip="First message content (advanced: view shows all blocks)"
									>
										<FiHelpCircle size={12} />
									</span>
								</label>
								<div className="col-span-9">
									<textarea
										name="content"
										value={formData.content}
										onChange={handleInput}
										className={`textarea textarea-bordered h-32 w-full rounded-xl ${errors.content ? 'textarea-error' : ''}`}
										spellCheck="false"
										aria-invalid={Boolean(errors.content)}
									/>
									{errors.content && (
										<div className="label">
											<span className="label-text-alt text-error flex items-center gap-1">
												<FiAlertCircle size={12} /> {errors.content}
											</span>
										</div>
									)}
								</div>
							</div>
						)}

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
									readOnly={isViewMode}
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

						{/* View mode: show full blocks + variables + meta */}
						{isViewMode && initialData?.template && (
							<>
								<div className="divider">Blocks</div>
								<div className="space-y-3">
									{initialData.template.blocks.map(b => (
										<div key={b.id} className="border-base-content/10 rounded-2xl border p-3">
											<div className="text-base-content/60 mb-2 text-xs font-semibold uppercase">Role: {b.role}</div>
											<textarea
												className="textarea textarea-bordered w-full rounded-xl"
												readOnly
												spellCheck="false"
												value={b.content}
											/>
										</div>
									))}
									{initialData.template.blocks.length === 0 && (
										<div className="text-base-content/70 text-sm">No blocks.</div>
									)}
								</div>

								<div className="divider">Variables</div>
								<pre className="bg-base-300 overflow-auto rounded-2xl p-3 text-xs">
									{JSON.stringify(initialData.template.variables ?? [], null, 2)}
								</pre>

								<div className="divider">Metadata</div>
								<div className="grid grid-cols-12 gap-2 text-sm">
									<div className="col-span-3 font-semibold">Version</div>
									<div className="col-span-9">{initialData.template.version}</div>

									<div className="col-span-3 font-semibold">Built-in</div>
									<div className="col-span-9">{initialData.template.isBuiltIn ? 'Yes' : 'No'}</div>

									<div className="col-span-3 font-semibold">Created</div>
									<div className="col-span-9">{initialData.template.createdAt}</div>

									<div className="col-span-3 font-semibold">Modified</div>
									<div className="col-span-9">{initialData.template.modifiedAt}</div>
								</div>
							</>
						)}

						<div className="modal-action">
							<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
								{isViewMode ? 'Close' : 'Cancel'}
							</button>
							{!isViewMode && (
								<button type="submit" className="btn btn-primary rounded-xl" disabled={!isAllValid}>
									Save
								</button>
							)}
						</div>
					</form>
				</div>
			</div>
		</dialog>,
		document.body
	);
}
