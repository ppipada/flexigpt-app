import React, { FC, useEffect, useState } from 'react';

interface AdditionalSettingsProps {
	additionalSettings: Record<string, any>;
	onAdditionalSettingsChange: (value: Record<string, any>) => void;
}

const AdditionalSettings: FC<AdditionalSettingsProps> = ({ additionalSettings, onAdditionalSettingsChange }) => {
	const [jsonInput, setJsonInput] = useState<string>('');
	const [jsonError, setJsonError] = useState<string | null>(null);

	useEffect(() => {
		setJsonInput(JSON.stringify(additionalSettings, null, 2));
	}, [additionalSettings]);

	const handleJsonInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
		setJsonInput(e.target.value);
	};

	const handleJsonInputBlur = () => {
		try {
			const parsedValue = JSON.parse(jsonInput);
			setJsonError(null);
			onAdditionalSettingsChange(parsedValue);
		} catch (error) {
			setJsonError('Invalid JSON');
		}
	};

	return (
		<div className="grid grid-cols-12 gap-4 mb-4 items-center">
			<label className="col-span-3 text-sm font-medium text-left tooltip" data-tip="Additional Settings">
				Additional Settings
			</label>
			<textarea
				className="textarea textarea-bordered col-span-9 w-full rounded-lg px-4 py-2"
				rows={4}
				value={jsonInput}
				onChange={handleJsonInputChange}
				onBlur={handleJsonInputBlur}
			/>
			{jsonError && <span className="text-red-500 col-span-12">{jsonError}</span>}
		</div>
	);
};

export default AdditionalSettings;
