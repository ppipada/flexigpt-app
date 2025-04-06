import type { ModelSetting } from '@/models/settingmodel';
import type { FC } from 'react';
import React, { useEffect, useState } from 'react';
import { FiX } from 'react-icons/fi';

interface ModifyModelModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (modelData: ModelSetting) => void;
	initialData?: ModelSetting;
	existingModels: ModelSetting[];
}

const ModifyModelModal: FC<ModifyModelModalProps> = ({ isOpen, onClose, onSubmit, initialData, existingModels }) => {
	const [formData, setFormData] = useState<Partial<ModelSetting>>({
		name: '',
		temperature: 0.1,
		isEnabled: true,
	});
	const [errors, setErrors] = useState<{ name?: string; temperature?: string }>({});

	useEffect(() => {
		if (isOpen) {
			setFormData(initialData || { name: '', temperature: 0.1, isEnabled: true });
			setErrors({});
		}
	}, [isOpen, initialData]);

	const validateField = (name: string, value: any) => {
		const newErrors: { [key: string]: string | undefined } = {};

		// Populate newErrors with existing errors except for the current field
		for (const key in errors) {
			if (Object.prototype.hasOwnProperty.call(errors, key) && key !== name) {
				newErrors[key] = errors[key as keyof typeof errors];
			}
		}

		if (name === 'name') {
			if (!value.trim()) {
				newErrors.name = 'Model name is required.';
			} else {
				const isDuplicate = existingModels.some(model => model.name === value && model.name !== initialData?.name);

				if (isDuplicate) {
					newErrors.name = 'Model name must be unique.';
				}
			}
		}

		if (name === 'temperature') {
			if (value === undefined || value === null || value === '') {
				newErrors.temperature = 'Temperature is required.';
			} else if (isNaN(value) || value < 0 || value > 1) {
				newErrors.temperature = 'Temperature must be between 0 and 1.';
			}
		}

		setErrors(newErrors);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value, type, checked } = e.target;
		let val: any = type === 'checkbox' ? checked : value;

		if (name === 'temperature' && val !== '') {
			val = parseFloat(val);
		}

		setFormData(prev => ({ ...prev, [name]: val }));
		validateField(name, val);
	};

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Validate all fields before submission
		validateField('name', formData.name || '');
		validateField('temperature', formData.temperature);

		if (Object.keys(errors).length === 0 && formData.name && formData.temperature !== undefined) {
			const modelData: ModelSetting = {
				name: formData.name,
				displayName: formData.name,
				isEnabled: formData.isEnabled ?? true,
				temperature: formData.temperature,
			};

			onSubmit(modelData);
		}
	};

	if (!isOpen) return null;

	return (
		<div className="modal modal-open">
			<div className="modal-box rounded-2xl w-full">
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{initialData ? 'Edit Model' : 'Add New Model'}</h3>
					<button className="btn btn-sm btn-circle btn-ghost" onClick={onClose}>
						<FiX size={16} />
					</button>
				</div>
				<form onSubmit={handleSubmit} className="mt-4">
					<fieldset className="form-control mb-4">
						<label className="label mb-2 text-sm" htmlFor="name">
							<span className="label-text">Model Name*</span>
						</label>
						<input
							type="text"
							id="name"
							name="name"
							value={formData.name || ''}
							onChange={handleChange}
							className={`input input-bordered rounded-2xl w-full ${errors.name ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.name && <p className="text-error text-sm mt-1">{errors.name}</p>}
					</fieldset>

					<fieldset className="form-control mb-6">
						<label className="label mb-2 text-sm" htmlFor="temperature">
							<span className="label-text">Temperature* (0.0 to 1.0)</span>
						</label>
						<input
							type="number"
							id="temperature"
							name="temperature"
							step="0.01"
							min="0"
							max="1"
							value={formData.temperature ?? ''}
							onChange={handleChange}
							className={`input input-bordered rounded-2xl w-full ${errors.temperature ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.temperature && <p className="text-error text-sm mt-1">{errors.temperature}</p>}
					</fieldset>

					<fieldset className="form-control mb-6">
						<div className="label cursor-pointer justify-start gap-2">
							<span className="label-text mr-4 text-sm">Enabled</span>
							<input
								type="checkbox"
								id="isEnabled"
								name="isEnabled"
								checked={formData.isEnabled ?? true}
								onChange={handleChange}
								className="toggle toggle-primary"
							/>
						</div>
					</fieldset>

					<div className="modal-action">
						<button type="button" className="btn btn-ghost rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-2xl"
							disabled={!!errors.name || !!errors.temperature || !formData.name || formData.temperature === undefined}
						>
							{initialData ? 'Save' : 'Add'}
						</button>
					</div>
				</form>
			</div>
		</div>
	);
};

export default ModifyModelModal;
