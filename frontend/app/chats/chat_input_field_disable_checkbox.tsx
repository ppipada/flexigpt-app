export default function DisablePreviousMessagesCheckbox(props: {
	disablePreviousMessages: boolean;
	setDisablePreviousMessages: React.Dispatch<React.SetStateAction<boolean>>;
}) {
	const { disablePreviousMessages, setDisablePreviousMessages } = props;

	return (
		<div className="flex w-full justify-center">
			<label className="flex ml-4 space-x-2 text-neutral/60 overflow-hidden" title="Send Only This Message To AI">
				<input
					type="checkbox"
					checked={disablePreviousMessages}
					onChange={e => {
						setDisablePreviousMessages(e.target.checked);
					}}
					className="checkbox checkbox-xs rounded-full"
					spellCheck="false"
				/>
				<span className="text-xs text-nowrap">Ignore chat history</span>
			</label>
		</div>
	);
}
