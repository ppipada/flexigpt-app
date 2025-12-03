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
			<div className="tooltip tooltip-top" data-tip="Send only the new message to AI.">
				<label className="ml-2 flex items-center gap-2 space-x-2 truncate">
					<input
						type="checkbox"
						checked={disablePreviousMessages}
						onChange={e => {
							setDisablePreviousMessages(e.target.checked);
						}}
						className="checkbox checkbox-xs m-1 rounded-full"
						spellCheck="false"
					/>
					<span className="text-neutral-custom text-xs text-nowrap">Ignore chat</span>
				</label>
			</div>
		</div>
	);
}
