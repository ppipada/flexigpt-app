import type { FC } from 'react';

type DisablePreviousMessagesCheckboxProps = {
	disablePreviousMessages: boolean;
	setDisablePreviousMessages: React.Dispatch<React.SetStateAction<boolean>>;
};

const DisablePreviousMessagesCheckbox: FC<DisablePreviousMessagesCheckboxProps> = ({
	disablePreviousMessages,
	setDisablePreviousMessages,
}) => {
	return (
		<div className="flex w-full justify-center">
			<label className="flex ml-4 space-x-2 overflow-hidden" title="Send Only This Message To AI">
				<input
					type="checkbox"
					checked={disablePreviousMessages}
					onChange={e => {
						setDisablePreviousMessages(e.target.checked);
					}}
					className="checkbox checkbox-xs rounded-full"
					spellCheck="false"
				/>
				<span className="text-xs text-neutral-custom text-nowrap">Ignore chat history</span>
			</label>
		</div>
	);
};

export default DisablePreviousMessagesCheckbox;
