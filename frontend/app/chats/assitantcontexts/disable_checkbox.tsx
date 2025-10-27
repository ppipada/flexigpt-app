import type { Dispatch, SetStateAction } from 'react';

type DisablePreviousMessagesCheckboxProps = {
	disablePreviousMessages: boolean;
	setDisablePreviousMessages: Dispatch<SetStateAction<boolean>>;
};

export function DisablePreviousMessagesCheckbox({
	disablePreviousMessages,
	setDisablePreviousMessages,
}: DisablePreviousMessagesCheckboxProps) {
	return (
		<div className="flex w-full items-center justify-center">
			<label className="ml-2 flex space-x-2 truncate" title="Send only the new message to AI">
				<input
					type="checkbox"
					checked={disablePreviousMessages}
					onChange={e => {
						setDisablePreviousMessages(e.target.checked);
					}}
					className="checkbox checkbox-xs m-1 rounded-full"
					spellCheck="false"
				/>
				<span className="text-neutral-custom m-1 text-xs text-nowrap">Ignore chat</span>
			</label>
		</div>
	);
}
