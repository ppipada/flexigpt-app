import React, { useEffect, useState } from 'react';

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
		<div className="flex w-full justify-center mx-2">
			<label className="flex text-neutral-400 cursor-pointer">
				<input
					type="checkbox"
					className="checkbox checkbox-xs rounded-full"
					checked={isReasoningEnabled}
					onChange={e => {
						setIsReasoningEnabled(e.target.checked);
					}}
				/>
				<span className="text-xs ml-2">Hybrid Reasoning</span>
			</label>
		</div>
	);
}

export function ReasoningTokensDropdown(props: {
	tokens: number;
	setTokens: (tokens: number) => void;
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}) {
	const { tokens, setTokens, isOpen, setIsOpen, detailsRef } = props;

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

		if (detailsRef.current) {
			detailsRef.current.open = false;
		}
		setIsOpen(false);
	}

	return (
		<div className="flex w-full justify-center">
			<details
				ref={detailsRef}
				className="dropdown dropdown-top dropdown-end"
				onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
					setIsOpen((event.currentTarget as HTMLDetailsElement).open);
				}}
				open={isOpen}
			>
				<summary
					className="btn btn-xs text-left text-nowrap text-neutral-400 shadow-none border-none overflow-hidden"
					title="Set tokens"
				>
					<div className="flex">
						<span className="text-xs font-normal sm:hidden mr-2">Tokens: </span>
						<span className="text-xs font-normal hidden sm:inline mr-2">Effort Tokens: </span>{' '}
						<span className="text-xs font-normal">{tokens}</span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 md:ml-2" />
						)}
					</div>
				</summary>

				<ul className="dropdown-content menu bg-base-100 rounded-xl w-full p-4">
					{defaultTokenOptions.map(tk => (
						<li
							key={tk}
							className="cursor-pointer text-xs"
							onClick={() => {
								setTokens(tk);
								if (detailsRef.current) {
									detailsRef.current.open = false;
								}
								setIsOpen(false);
							}}
						>
							<a className="justify-between items-center p-1 m-0">
								<span>{tk}</span>
								{tokens === tk && <FiCheck />}
							</a>
						</li>
					))}

					<li className="text-xs">
						<hr className="p-0 my-2 border-0 border-t border-base-300" />
						<label className="tooltip tooltip-top outline-none border-none">
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
