import { type Dispatch, forwardRef, type SetStateAction, type SyntheticEvent, useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

const defaultTokenOptions = [1024, 8192, 32000];

// chat_input_field_reasoning_checkbox.tsx

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

type ReasoningTokensDropdownProps = {
	tokens: number;
	setTokens: (tokens: number) => void;
	isOpen: boolean;
	setIsOpen: Dispatch<SetStateAction<boolean>>;
};

export const ReasoningTokensDropdown = forwardRef<HTMLDetailsElement, ReasoningTokensDropdownProps>(
	function ReasoningTokensDropdown({ tokens, setTokens, isOpen, setIsOpen }, detailsRef) {
		const [customTokens, setCustomTokens] = useState<string>(String(tokens));

		useEffect(() => {
			setCustomTokens(String(tokens));
		}, [tokens]);

		function clampTokensOnBlur() {
			let val = parseInt(customTokens, 10);
			if (isNaN(val)) {
				val = 1024;
			}
			if (val < 1024) {
				val = 1024;
			}
			setTokens(val);

			if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
				detailsRef.current.open = false;
			}
			setIsOpen(false);
		}

		return (
			<div className="flex w-full justify-center">
				<details
					ref={detailsRef}
					className="dropdown dropdown-top dropdown-end w-full justify-center"
					onToggle={(event: SyntheticEvent<HTMLElement>) => {
						setIsOpen((event.currentTarget as HTMLDetailsElement).open);
					}}
					open={isOpen}
				>
					<summary
						className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-center text-nowrap shadow-none"
						title="Set Tokens"
					>
						<span className="mr-2 text-xs font-normal sm:hidden">Tokens:</span>
						<span className="mr-2 hidden text-xs font-normal sm:inline">Effort Tokens:</span>
						<span className="text-xs font-normal">{tokens}</span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 shrink-0 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 shrink-0 md:ml-2" />
						)}
					</summary>

					<ul className="dropdown-content menu bg-base-100 w-full rounded-xl p-4">
						{defaultTokenOptions.map(tk => (
							<li
								key={tk}
								className="cursor-pointer text-xs"
								onClick={() => {
									setTokens(tk);
									if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
										detailsRef.current.open = false;
									}
									setIsOpen(false);
								}}
							>
								<a className="m-0 items-center justify-between p-1">
									<span>{tk}</span>
									{tokens === tk && <FiCheck />}
								</a>
							</li>
						))}

						<li className="text-xs">
							<hr className="border-neutral/20 my-2 border-0 border-t p-0" />
							<label className="tooltip tooltip-top border-none outline-none">
								<div className="tooltip-content">
									<div className="text-xs">Custom tokens (≥ 1024)</div>
								</div>
								<input
									type="text"
									className="input input-xs w-full"
									placeholder="Enter a custom integer ≥ 1024"
									value={customTokens}
									onChange={e => {
										setCustomTokens(e.target.value);
									}}
									onBlur={clampTokensOnBlur}
								/>
							</label>
						</li>
					</ul>
				</details>
			</div>
		);
	}
);
