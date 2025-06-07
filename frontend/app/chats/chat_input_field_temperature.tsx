import { useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

const defaultTemperatureOptions = [0.0, 0.1, 0.5, 1.0];

// Subcomponent for Temperature Selection.
// Uses a string-based custom input and clamps the value to [0,1].
// Also displays four pre-set temperatures the user can click on.
export default function TemperatureDropdown(props: {
	temperature: number;
	setTemperature: (temp: number) => void;
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}) {
	const { temperature, setTemperature, isOpen, setIsOpen, detailsRef } = props;

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
					className="btn btn-xs text-nowrap text-neutral/60 shadow-none border-none overflow-hidden"
					title="Set Temperature"
				>
					<div className="flex">
						<span className="text-xs font-normal sm:hidden mr-2">Temp: </span>
						<span className="text-xs font-normal hidden sm:inline mr-2">Temperature: </span>{' '}
						<span className="text-xs font-normal">{temperature.toFixed(2)} </span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 md:ml-2" />
						)}
					</div>
				</summary>

				<ul className="dropdown-content menu bg-base-100 rounded-xl w-full p-4">
					{/* Default temperature options */}
					{defaultTemperatureOptions.map(tempVal => (
						<li
							key={tempVal}
							className="cursor-pointer text-xs"
							onClick={() => {
								setTemperature(tempVal);
								if (detailsRef.current) {
									detailsRef.current.open = false;
								}
								setIsOpen(false);
							}}
						>
							<a className="justify-between items-center p-1 m-0">
								<span>{tempVal.toFixed(1)}</span>
								{temperature.toFixed(1) === tempVal.toFixed(1) && <FiCheck />}
							</a>
						</li>
					))}

					{/* Custom temperature input */}
					<li className="text-xs">
						<hr className="p-0 my-2 border-0 border-t border-neutral/20" />
						<label className="tooltip tooltip-top outline-none border-none">
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
