/*  prompt_template_modify.tsx
 *  Lightweight “wizard” for creating / editing a prompt-template.
 *  Focuses on the most common metadata; advanced fields such as
 *  variables or pre-processors can be added later.
 */
import React, { useEffect, useState } from 'react';

import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/models/commands';
import { PromptRoleEnum, type PromptTemplate } from '@/models/promptmodel';

import { omitManyKeys } from '@/lib/obj_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

/* ---- local helper type identical to the one in the list view ---- */
interface TemplateItem {
	template: PromptTemplate;
	bundleID: string;
	templateSlug: string;
}

interface ModifyPromptTemplateProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (templateData: Partial<PromptTemplate>) => void;
	initialData?: TemplateItem;
	existingTemplates: TemplateItem[];
}

const ModifyPromptTemplate: React.FC<ModifyPromptTemplateProps> = ({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingTemplates,
}) => {
	/* --- form state --- */
	const [formData, setFormData] = useState({
		displayName: '',
		slug: '',
		description: '',
		content: '',
		tags: '',
		isEnabled: true,
	});
	const [errors, setErrors] = useState<{ displayName?: string; slug?: string; content?: string }>({});

	/* --- sync prop → state --- */
	useEffect(() => {
		if (isOpen) {
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
				setFormData({ displayName: '', slug: '', description: '', content: '', tags: '', isEnabled: true });
			}
			setErrors({});
		}
	}, [isOpen, initialData]);

	/* --- validation --- */
	const validateField = (name: string, value: string) => {
		let newErrors = { ...errors };

		if (!value.trim()) {
			newErrors[name as keyof typeof newErrors] = 'This field is required';
		} else if (name === 'slug') {
			if (value.startsWith(PROMPT_TEMPLATE_INVOKE_CHAR)) {
				newErrors.slug = `Slug should not start with ${PROMPT_TEMPLATE_INVOKE_CHAR}`;
			} else {
				const duplicate = existingTemplates.some(
					t => t.template.slug === value && t.template.id !== initialData?.template.id
				);
				if (duplicate) newErrors.slug = 'This slug is already in use';
				else delete newErrors.slug;
			}
		} else {
			newErrors = omitManyKeys(errors, [name as keyof typeof newErrors]);
		}
		setErrors(newErrors);
	};

	/* --- change handlers --- */
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value, type, checked } = e.target as any;
		const newVal = type === 'checkbox' ? checked : value;
		setFormData(prev => ({ ...prev, [name]: newVal }));
		if (['displayName', 'slug', 'content'].includes(name)) validateField(name, newVal);
	};

	/* --- submit --- */
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		validateField('displayName', formData.displayName);
		validateField('slug', formData.slug);
		validateField('content', formData.content);

		if (Object.keys(errors).length === 0 && formData.displayName && formData.slug && formData.content) {
			const tagsArr = formData.tags
				.split(',')
				.map(t => t.trim())
				.filter(Boolean);

			const payload: Partial<PromptTemplate> = {
				displayName: formData.displayName.trim(),
				slug: formData.slug.trim(),
				description: formData.description.trim() || undefined,
				isEnabled: formData.isEnabled,
				tags: tagsArr.length ? tagsArr : undefined,
				blocks: [
					{
						id: initialData?.template.blocks[0]?.id ?? getUUIDv7(),
						role: PromptRoleEnum.User,
						content: formData.content,
					},
				],
			};

			onSubmit(payload);
		}
	};

	/* --- early return --- */
	if (!isOpen) return null;

	/* --- render --- */
	return (
		<div className="modal modal-open">
			<div className="modal-box rounded-2xl">
				<h3 className="font-bold text-lg">{initialData ? 'Edit Prompt Template' : 'Add New Prompt Template'}</h3>

				<form onSubmit={handleSubmit} className="mt-4 space-y-3">
					{/* Display Name */}
					<fieldset>
						<label className="label">Display Name*</label>
						<input
							name="displayName"
							type="text"
							value={formData.displayName}
							onChange={handleChange}
							className={`input rounded-2xl w-full ${errors.displayName ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.displayName && <p className="text-error text-sm">{errors.displayName}</p>}
					</fieldset>

					{/* Slug */}
					<fieldset>
						<label className="label">Slug*</label>
						<div className="relative">
							<span className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral/60">
								{PROMPT_TEMPLATE_INVOKE_CHAR}
							</span>
							<input
								name="slug"
								type="text"
								value={formData.slug}
								onChange={handleChange}
								className={`input rounded-2xl pl-8 w-full ${errors.slug ? 'input-error' : ''}`}
								required
								spellCheck="false"
							/>
						</div>
						{errors.slug && <p className="text-error text-sm">{errors.slug}</p>}
					</fieldset>

					{/* Description */}
					<fieldset>
						<label className="label">Description</label>
						<textarea
							name="description"
							value={formData.description}
							onChange={handleChange}
							className="textarea rounded-2xl w-full"
							spellCheck="false"
						/>
					</fieldset>

					{/* Prompt Block Content */}
					<fieldset>
						<label className="label">Prompt Content*</label>
						<textarea
							name="content"
							value={formData.content}
							onChange={handleChange}
							className={`textarea rounded-2xl w-full ${errors.content ? 'textarea-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.content && <p className="text-error text-sm">{errors.content}</p>}
					</fieldset>

					{/* Tags */}
					<fieldset>
						<label className="label">Tags (comma separated)</label>
						<input
							name="tags"
							type="text"
							value={formData.tags}
							onChange={handleChange}
							className="input rounded-2xl w-full"
							spellCheck="false"
						/>
					</fieldset>

					{/* Enabled toggle */}
					<fieldset className="flex items-center gap-2">
						<input
							type="checkbox"
							name="isEnabled"
							checked={formData.isEnabled}
							onChange={handleChange}
							className="checkbox checkbox-primary"
						/>
						<label className="label cursor-pointer">Enabled</label>
					</fieldset>

					{/* Actions */}
					<div className="modal-action">
						<button type="button" className="btn btn-ghost rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-2xl"
							disabled={
								!!errors.displayName ||
								!!errors.slug ||
								!!errors.content ||
								!formData.displayName ||
								!formData.slug ||
								!formData.content
							}
						>
							Save
						</button>
					</div>
				</form>
			</div>
		</div>
	);
};

export default ModifyPromptTemplate;
