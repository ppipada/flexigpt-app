export function HybridReasoningCheckbox({
	isReasoningEnabled,
	setIsReasoningEnabled,
}: {
	isReasoningEnabled: boolean;
	setIsReasoningEnabled: (enabled: boolean) => void;
}) {
	return (
		<div className="mx-2 flex w-full justify-center">
			<label className="text-neutral-custom flex cursor-pointer">
				<input
					type="checkbox"
					className="checkbox checkbox-xs rounded-full"
					checked={isReasoningEnabled}
					onChange={e => {
						setIsReasoningEnabled(e.target.checked);
					}}
				/>
				<span className="text-neutral-custom ml-2 text-xs">Hybrid Reasoning</span>
			</label>
		</div>
	);
}
