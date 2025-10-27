import { type Dispatch, forwardRef, type SetStateAction, type SyntheticEvent, useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

const defaultTemperatureOptions = [0.0, 0.1, 0.5, 1.0];

// Subcomponent for Temperature Selection.
// Uses a string-based custom input and clamps the value to [0,1].
// Also displays four pre-set temperatures the user can click on.
type TemperatureDropdownProps = {
	temperature: number;
	setTemperature: (t: number) => void;
	isOpen: boolean;
	setIsOpen: Dispatch<SetStateAction<boolean>>;
};

export const TemperatureDropdown = forwardRef<HTMLDetailsElement, TemperatureDropdownProps>(
	function TemperatureDropdown({ temperature, setTemperature, isOpen, setIsOpen }, detailsRef) {
		// Local state for the custom text input (string-based).
		const [customTemp, setCustomTemp] = useState(temperature.toString());

		// Sync local string whenever "temperature" changes.
		useEffect(() => {
			setCustomTemp(temperature.toString());
		}, [temperature]);

		// Called on blur so we clamp to [0,1], fallback to 0.1 if invalid.
		function clampOnBlur() {
			let val = parseFloat(customTemp);
			if (isNaN(val)) {
				val = 0.1; // fallback if parse fails
			}
			// Clamp.
			val = Math.max(0, Math.min(1, val));
			setTemperature(val);

			// Close after user is done typing.
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
						className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-left text-nowrap shadow-none"
						title="Set Temperature"
					>
						<span className="mr-2 text-xs font-normal sm:hidden">Temp:</span>
						<span className="mr-2 hidden text-xs font-normal sm:inline">Temperature:</span>
						<span className="text-xs font-normal">{temperature.toFixed(2)} </span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 shrink-0 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 shrink-0 md:ml-2" />
						)}
					</summary>

					<ul className="dropdown-content menu bg-base-100 w-full rounded-xl p-4">
						{/* Default temperature options */}
						{defaultTemperatureOptions.map(tempVal => (
							<li
								key={tempVal}
								className="cursor-pointer text-xs"
								onClick={() => {
									setTemperature(tempVal);
									if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
										detailsRef.current.open = false;
									}
									setIsOpen(false);
								}}
							>
								<a className="m-0 items-center justify-between p-1">
									<span>{tempVal.toFixed(1)}</span>
									{temperature.toFixed(1) === tempVal.toFixed(1) && <FiCheck />}
								</a>
							</li>
						))}

						{/* Custom temperature input */}
						<li className="text-xs">
							<hr className="border-neutral/20 my-2 border-0 border-t p-0" />
							<label className="tooltip tooltip-top border-none outline-none">
								<div className="tooltip-content">
									<div className="text-xs">Custom value (0.0 - 1.0)</div>
								</div>
								<input
									type="text"
									name="temperature"
									className="input input-xs w-full"
									placeholder="Custom value (0.0 - 1.0)"
									value={customTemp}
									onChange={e => {
										setCustomTemp(e.target.value);
									}}
									onBlur={clampOnBlur}
									spellCheck="false"
								/>
							</label>
						</li>
					</ul>
				</details>
			</div>
		);
	}
);
