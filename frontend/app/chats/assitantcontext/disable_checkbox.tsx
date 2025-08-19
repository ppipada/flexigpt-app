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
			<label className="ml-4 flex space-x-2 truncate" title="Send only the new message to AI">
				<input
					type="checkbox"
					checked={disablePreviousMessages}
					onChange={e => {
						setDisablePreviousMessages(e.target.checked);
					}}
					className="checkbox checkbox-xs rounded-full"
					spellCheck="false"
				/>
				<span className="text-neutral-custom text-xs text-nowrap">Ignore chat</span>
			</label>
		</div>
	);
};

export default DisablePreviousMessagesCheckbox;
