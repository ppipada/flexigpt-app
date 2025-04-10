import React, { useEffect, useState } from 'react';

import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/models/commands';
import type { PromptTemplate } from '@/models/promptmodel';

interface ModifyPromptTemplateProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (templateData: Partial<PromptTemplate>) => void;
	initialData?: Partial<PromptTemplate>;
	existingTemplates: PromptTemplate[];
}

const ModifyPromptTemplate: React.FC<ModifyPromptTemplateProps> = ({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingTemplates,
}) => {
	const [formData, setFormData] = useState<Partial<PromptTemplate>>({
		name: '',
		command: '',
		template: '',
		hasTools: false,
		hasDocStore: false,
		tokenCount: 0,
	});
	const [errors, setErrors] = useState<{ name?: string; command?: string; template?: string }>({});

	useEffect(() => {
		if (isOpen) {
			setFormData(
				initialData || { name: '', command: '', template: '', hasTools: false, hasDocStore: false, tokenCount: 0 }
			);
			setErrors({});
		}
	}, [isOpen, initialData]);

	const validateField = (name: string, value: string) => {
		const newErrors: { [key: string]: string | undefined } = {};

		if (!value.trim()) {
			newErrors[name] = 'This field is required';
		} else if (name === 'command') {
			if (value.startsWith(PROMPT_TEMPLATE_INVOKE_CHAR)) {
				newErrors[name] = `Command should not start with ${PROMPT_TEMPLATE_INVOKE_CHAR}`;
			} else {
				const isUnique = !existingTemplates.some(t => t.command === value && t.id !== initialData?.id);
				if (!isUnique) {
					newErrors[name] = 'This command is already in use';
				}
			}
		}

		// Populate newErrors with existing errors except for the current field if it's valid
		for (const key in errors) {
			if (Object.prototype.hasOwnProperty.call(errors, key) && key !== name) {
				newErrors[key] = errors[key as keyof typeof errors];
			}
		}

		setErrors(newErrors);
	};
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value } = e.target;
		setFormData(prev => ({ ...prev, [name]: value }));
		validateField(name, value);
	};

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		validateField('name', formData.name || '');
		validateField('command', formData.command || '');
		validateField('template', formData.template || '');

		if (Object.keys(errors).length === 0 && formData.name && formData.command && formData.template) {
			onSubmit(formData);
		}
	};

	if (!isOpen) return null;

	return (
		<div className="modal modal-open">
			<div className="modal-box rounded-2xl">
				<h3 className="font-bold text-lg">{initialData ? 'Edit Prompt Template' : 'Add New Prompt Template'}</h3>
				<form onSubmit={handleSubmit} className="mt-4">
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Name*
						</label>
						<input
							type="text"
							name="name"
							value={formData.name || ''}
							onChange={handleChange}
							className={`input rounded-2xl ${errors.name ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.name && <p className="text-error text-sm mt-1">{errors.name}</p>}
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Command*
						</label>
						<div className="relative">
							<span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-neutral">
								{PROMPT_TEMPLATE_INVOKE_CHAR}
							</span>
							<input
								type="text"
								name="command"
								value={formData.command || ''}
								onChange={handleChange}
								className={`input rounded-2xl pl-8 ${errors.command ? 'input-error' : ''}`}
								required
								spellCheck="false"
							/>
						</div>
						{errors.command && <p className="text-error text-sm mt-1">{errors.command}</p>}
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Template*
						</label>
						<textarea
							name="template"
							value={formData.template || ''}
							onChange={handleChange}
							className={`textarea rounded-2xl ${errors.template ? 'textarea-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.template && <p className="text-error text-sm mt-1">{errors.template}</p>}
					</fieldset>
					<div className="modal-action">
						<button type="button" className="btn btn-ghost rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-2xl"
							disabled={
								!!errors.name ||
								!!errors.command ||
								!!errors.template ||
								!formData.name ||
								!formData.command ||
								!formData.template
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
