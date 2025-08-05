import React, { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/spec/command';
import { PromptRoleEnum, type PromptTemplate } from '@/spec/prompt';

import { omitManyKeys } from '@/lib/obj_utils';
import { validateSlug, validateTags } from '@/lib/text_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

/* ---------- local helper types ---------- */
interface TemplateItem {
	template: PromptTemplate;
	bundleID: string;
	templateSlug: string;
}

interface ModifyPromptTemplateProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (templateData: Partial<PromptTemplate>) => void;
	initialData?: TemplateItem; // when editing
	existingTemplates: TemplateItem[];
}

/* ---------- component ---------- */

const ModifyPromptTemplate: React.FC<ModifyPromptTemplateProps> = ({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingTemplates,
}) => {
	/* ---------- form state ---------- */
	const [formData, setFormData] = useState({
		displayName: '',
		slug: '',
		description: '',
		content: '',
		tags: '',
		isEnabled: true,
	});

	const [errors, setErrors] = useState<{
		displayName?: string;
		slug?: string;
		content?: string;
		tags?: string;
	}>({});

	const isEditMode = Boolean(initialData);

	/* ---------- sync prop -> state ---------- */
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

	/* ---------- validation helpers ---------- */
	const validateField = (field: keyof typeof errors, val: string) => {
		let newErrs = { ...errors };
		const v = val.trim();

		/* required */
		if (!v) {
			newErrs[field] = 'This field is required.';
		} else if (field === 'slug') {
			if (v.startsWith(PROMPT_TEMPLATE_INVOKE_CHAR)) {
				newErrs.slug = `Do not prefix with "${PROMPT_TEMPLATE_INVOKE_CHAR}".`;
			} else {
				const err = validateSlug(v);
				if (err) {
					newErrs.slug = err;
				} else {
					const clash = existingTemplates.some(
						t => t.template.slug === v && t.template.id !== initialData?.template.id
					);
					if (clash) newErrs.slug = 'Slug already in use.';
					else newErrs = omitManyKeys(newErrs, ['slug']);
				}
			}
		} else if (field === 'tags') {
			const err = validateTags(val);
			if (err) newErrs.tags = err;
			else newErrs = omitManyKeys(newErrs, ['tags']);
		} else {
			newErrs = omitManyKeys(newErrs, [field]);
		}
		setErrors(newErrs);
	};

	/* ---------- generic change handler ---------- */
	const handleInput = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;

		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (name === 'displayName' || name === 'slug' || name === 'content' || name === 'tags') {
			validateField(name, String(newVal));
		}
	};

	/* ---------- overall validity ---------- */
	const isAllValid = useMemo(() => {
		const errs = Object.values(errors).some(Boolean);
		const filled = formData.displayName.trim() && formData.slug.trim() && formData.content.trim();
		return !errs && Boolean(filled);
	}, [errors, formData]);

	/* ---------- submit ---------- */
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		validateField('displayName', formData.displayName);
		validateField('slug', formData.slug);
		validateField('content', formData.content);
		validateField('tags', formData.tags);

		if (!isAllValid) return;

		const tagsArr = formData.tags
			.split(',')
			.map(t => t.trim())
			.filter(Boolean);

		onSubmit({
			displayName: formData.displayName.trim(),
			slug: formData.slug.trim(),
			description: formData.description.trim() || undefined,
			isEnabled: formData.isEnabled,
			tags: tagsArr.length ? tagsArr : undefined,
			/* We keep the simple single-block approach for now. */
			blocks: [
				{
					id: initialData?.template.blocks[0]?.id ?? getUUIDv7(),
					role: PromptRoleEnum.User,
					content: formData.content,
				},
			],
		});
	};

	/* ---------- early-return ---------- */
	if (!isOpen) return null;

	/* ---------- render ---------- */
	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Prompt Template' : 'Add Prompt Template'}</h3>
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
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip={`Without "${PROMPT_TEMPLATE_INVOKE_CHAR}" prefix`}
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<div className="relative">
								<span className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-custom">
									{PROMPT_TEMPLATE_INVOKE_CHAR}
								</span>
								<input
									type="text"
									name="slug"
									value={formData.slug}
									onChange={handleInput}
									className={`input input-bordered w-full pl-8 rounded-2xl ${errors.slug ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									disabled={isEditMode && initialData?.template.isBuiltIn}
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

					{/* Content */}
					<div className="grid grid-cols-12 gap-2 items-start">
						<label className="label col-span-3">
							<span className="label-text text-sm">Prompt Content*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="First (user) message - add advanced messages later"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="content"
								value={formData.content}
								onChange={handleInput}
								className={`textarea textarea-bordered w-full rounded-2xl h-32 ${
									errors.content ? 'textarea-error' : ''
								}`}
								spellCheck="false"
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

export default ModifyPromptTemplate;
