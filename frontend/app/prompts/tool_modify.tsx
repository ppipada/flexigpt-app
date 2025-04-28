import React, { useEffect, useState } from 'react';

import { TOOL_INVOKE_CHAR } from '@/models/commands';
import type { Tool } from '@/models/promptmodel';

interface ModifyToolProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (toolData: Partial<Tool>) => void;
	initialData?: Partial<Tool>;
	existingTools: Tool[];
}

const ModifyTool: React.FC<ModifyToolProps> = ({ isOpen, onClose, onSubmit, initialData, existingTools }) => {
	const [formData, setFormData] = useState<Partial<Tool>>({
		name: '',
		command: '',
		schema: '',
		inFunc: '',
	});
	const [errors, setErrors] = useState<{ name?: string; command?: string; schema?: string; inFunc?: string }>({});

	useEffect(() => {
		if (isOpen) {
			setFormData(initialData || { name: '', command: '', schema: '', inFunc: '' });
			setErrors({});
		}
	}, [isOpen, initialData]);

	const validateField = (name: string, value: string) => {
		const newErrors: { [key: string]: string | undefined } = {};

		if (!value.trim()) {
			newErrors[name] = 'This field is required';
		} else if (name === 'command') {
			if (value.startsWith(TOOL_INVOKE_CHAR)) {
				newErrors[name] = `Command should not start with ${TOOL_INVOKE_CHAR}`;
			} else {
				const isUnique = !existingTools.some(t => t.command === value && t.id !== initialData?.id);
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
		validateField('schema', formData.schema || '');
		validateField('inFunc', formData.inFunc || '');

		if (Object.keys(errors).length === 0 && formData.name && formData.command && formData.schema && formData.inFunc) {
			onSubmit(formData);
		}
	};

	if (!isOpen) return null;

	return (
		<div className="modal modal-open">
			<div className="modal-box rounded-2xl">
				<h3 className="font-bold text-lg">{initialData ? 'Edit Tool' : 'Add New Tool'}</h3>
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
							<span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-neutral/60">
								{TOOL_INVOKE_CHAR}
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
							Schema*
						</label>
						<textarea
							name="schema"
							value={formData.schema || ''}
							onChange={handleChange}
							className={`textarea rounded-2xl ${errors.schema ? 'textarea-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.schema && <p className="text-error text-sm mt-1">{errors.schema}</p>}
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Input Function*
						</label>
						<textarea
							name="inFunc"
							value={formData.inFunc || ''}
							onChange={handleChange}
							className={`textarea rounded-2xl ${errors.inFunc ? 'textarea-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.inFunc && <p className="text-error text-sm mt-1">{errors.inFunc}</p>}
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
								!!errors.schema ||
								!!errors.inFunc ||
								!formData.name ||
								!formData.command ||
								!formData.schema ||
								!formData.inFunc
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

export default ModifyTool;
