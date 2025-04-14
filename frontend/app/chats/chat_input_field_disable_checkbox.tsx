/**
 * Subcomponent for "Disable previous messages" checkbox
 */
export default function DisablePreviousMessagesCheckbox(props: {
	disablePreviousMessages: boolean;
	setDisablePreviousMessages: React.Dispatch<React.SetStateAction<boolean>>;
}) {
	const { disablePreviousMessages, setDisablePreviousMessages } = props;

	return (
		<div className="flex w-full justify-center">
			<label className="flex space-x-2 text-neutral-400 overflow-hidden" title="Disable previous messages">
				<input
					type="checkbox"
					checked={disablePreviousMessages}
					onChange={e => {
						setDisablePreviousMessages(e.target.checked);
					}}
					className="checkbox checkbox-xs rounded-full"
					spellCheck="false"
				/>
				<span className="text-xs text-nowrap">Disable previous messages</span>
			</label>
		</div>
	);
}
