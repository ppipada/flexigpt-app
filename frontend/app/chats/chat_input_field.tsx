import type { ChangeEvent, KeyboardEvent } from 'react';
import React, { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiSend } from 'react-icons/fi';

import { type ChatOptions, DefaultChatOptions } from '@/models/settingmodel';

import { GetChatInputOptions } from '@/apis/settingstore_helper';

import { UseCloseDetails } from '@/lib/use_close_details';

/**
 * Four default pre-set temperatures.
 */
const defaultTemperatureOptions = [0.0, 0.1, 0.5, 1.0];

interface ChatInputFieldProps {
	onSend: (message: string, options: ChatOptions) => void;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
}

export interface ChatInputFieldHandle {
	getChatOptions: () => ChatOptions;
	focus: () => void;
}

// Custom hook for handling form submission on Enter key press
function useEnterSubmit(): {
	formRef: React.RefObject<HTMLFormElement | null>;
	onKeyDown: (event: KeyboardEvent<HTMLTextAreaElement>) => void;
} {
	const formRef = useRef<HTMLFormElement>(null);

	const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>): void => {
		if (event.key === 'Enter' && !event.shiftKey && !event.nativeEvent.isComposing) {
			if (formRef.current) {
				formRef.current.requestSubmit();
			}
			event.preventDefault();
		}
	};

	return { formRef, onKeyDown: handleKeyDown };
}

const MAX_HEIGHT = 360;

/**
 * Subcomponent for selecting the Model
 */
function ModelDropdown(props: {
	selectedModel: ChatOptions;
	setSelectedModel: React.Dispatch<React.SetStateAction<ChatOptions>>;
	allOptions: ChatOptions[];
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}) {
	const { selectedModel, setSelectedModel, allOptions, isOpen, setIsOpen, detailsRef } = props;

	return (
		<details
			ref={detailsRef}
			className="dropdown dropdown-top dropdown-end w-1/3"
			onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
				setIsOpen((event.currentTarget as HTMLDetailsElement).open);
			}}
			open={isOpen}
		>
			<summary
				className="btn btn-xs w-full text-left text-nowrap text-neutral-400 shadow-none border-none overflow-hidden"
				title="Select Model"
			>
				<div className="flex">
					<span className="sm:hidden">{selectedModel.title.substring(0, 8)}</span>
					<span className="hidden sm:inline">{selectedModel.title} </span>
					{isOpen ? (
						<FiChevronDown size={16} className="ml-1 md:ml-2" />
					) : (
						<FiChevronUp size={16} className="ml-1 md:ml-2" />
					)}
				</div>
			</summary>

			<ul className="dropdown-content menu bg-base-100 rounded-xl w-full">
				{allOptions.map(model => (
					<li
						key={`${model.provider}-${model.name}`}
						className="cursor-pointer text-xs"
						onClick={() => {
							setSelectedModel(model);
							if (detailsRef.current) {
								detailsRef.current.open = false;
							}
							setIsOpen(false);
						}}
					>
						<a className="justify-between items-center p-1 m-0">
							<span>{model.title}</span>
							{selectedModel.name === model.name && selectedModel.provider === model.provider && <FiCheck />}
						</a>
					</li>
				))}
			</ul>
		</details>
	);
}

/**
 * Subcomponent for Temperature Selection
 * -------------------------------------
 * Uses a string-based custom input and clamps the value to [0,1].
 * Also displays four pre-set temperatures the user can click on.
 */
function TemperatureDropdown(props: {
	temperature: number; // The current temperature value
	setTemperature: (temp: number) => void; // A function to update & clamp
	isOpen: boolean; // Whether details is open
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}) {
	const { temperature, setTemperature, isOpen, setIsOpen, detailsRef } = props;

	// local state for the custom text input (string-based)
	const [customTemp, setCustomTemp] = useState(temperature.toString());

	// Sync local string whenever "temperature" changes
	useEffect(() => {
		setCustomTemp(temperature.toString());
	}, [temperature]);

	// Called on blur so we clamp to [0,1], fallback to 0.1 if invalid
	function clampOnBlur() {
		let val = parseFloat(customTemp);
		if (isNaN(val)) {
			val = 0.1; // fallback if parse fails
		}
		// clamp
		val = Math.max(0, Math.min(1, val));
		setTemperature(val);

		// close after user is done typing
		if (detailsRef.current) {
			detailsRef.current.open = false;
		}
		setIsOpen(false);
	}

	return (
		<details
			ref={detailsRef}
			className="dropdown dropdown-top dropdown-end w-1/3"
			onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
				setIsOpen((event.currentTarget as HTMLDetailsElement).open);
			}}
			open={isOpen}
		>
			<summary
				className="btn btn-xs w-full text-left text-nowrap text-neutral-400 shadow-none border-none overflow-hidden"
				title="Set temperature"
			>
				<div className="flex">
					<span className="sm:hidden mr-2">Temp: </span>
					<span className="hidden sm:inline mr-2">Temperature: </span> {temperature.toFixed(2)}{' '}
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
					<hr className="p-0 my-2 border-0 border-t border-base-300" />
					<label className="tooltip tooltip-top outline-none border-none">
						<div className="tooltip-content">
							<div className="text-xs">Custom value (0.1 - 0.1 )</div>
						</div>
						<span>Custom</span>
						<input
							type="text"
							name="temperature"
							className="input input-xs w-full"
							placeholder="Custom value (0.1 - 0.1 )"
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
	);
}

/**
 * Subcomponent for "Disable previous messages" checkbox
 */
function DisablePreviousMessagesCheckbox(props: {
	disablePreviousMessages: boolean;
	setDisablePreviousMessages: React.Dispatch<React.SetStateAction<boolean>>;
}) {
	const { disablePreviousMessages, setDisablePreviousMessages } = props;

	return (
		<label
			className="flex items-center space-x-2 text-neutral-400 w-1/3 overflow-hidden"
			title="Disable previous messages"
		>
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
	);
}

/**
 * Main Chat Input Field
 * -------------------------------------
 * Renders ModelDropdown, TemperatureDropdown, DisablePreviousMessagesCheckbox,
 * and the message input form.
 */
const ChatInputField = forwardRef<ChatInputFieldHandle, ChatInputFieldProps>(({ onSend, setInputHeight }, ref) => {
	// For the main text area
	const [text, setText] = useState<string>('');
	const [isSendButtonEnabled, setIsSendButtonEnabled] = useState<boolean>(false);
	const inputRef = useRef<HTMLTextAreaElement>(null);
	const isSubmittingRef = useRef<boolean>(false);

	// Model state
	const [selectedModel, setSelectedModel] = useState<ChatOptions>(DefaultChatOptions);

	// Checkbox for "Disable previous messages"
	const [disablePreviousMessages, setDisablePreviousMessages] = useState<boolean>(false);

	// Model dropdown open state
	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState<boolean>(false);

	// Temperature dropdown open state
	const [isTemperatureDropdownOpen, setIsTemperatureDropdownOpen] = useState<boolean>(false);

	// All available models
	const [allOptions, setAllOptions] = useState<ChatOptions[]>([DefaultChatOptions]);

	// Refs for the 2 <details> elements
	const modelDetailsRef = useRef<HTMLDetailsElement>(null);
	const temperatureDetailsRef = useRef<HTMLDetailsElement>(null);

	// Close logic for model dropdown
	UseCloseDetails({
		detailsRef: modelDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsModelDropdownOpen(false);
		},
	});

	// Close logic for temperature dropdown
	UseCloseDetails({
		detailsRef: temperatureDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsTemperatureDropdownOpen(false);
		},
	});

	// Load initial model options
	const loadInitialItems = useCallback(async () => {
		const r = await GetChatInputOptions();
		setSelectedModel(r.default);
		setAllOptions(r.allOptions);
	}, []);

	useEffect(() => {
		loadInitialItems();
	}, [loadInitialItems]);

	// Enter key submission logic
	const { formRef, onKeyDown } = useEnterSubmit();

	// Automatically resize the textarea
	const autoResizeTextarea = useCallback(() => {
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			const newHeight = Math.min(inputRef.current.scrollHeight, MAX_HEIGHT);
			inputRef.current.style.height = `${newHeight}px`;
			setInputHeight((prevHeight: number) => (prevHeight !== newHeight ? newHeight : prevHeight));
		}
	}, [setInputHeight]);

	const handleTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
		const value = event.target.value;
		setText(value);
		setIsSendButtonEnabled(value.trim().length > 0);
		autoResizeTextarea();
	};

	useEffect(() => {
		autoResizeTextarea();
	}, [text, autoResizeTextarea]);

	const handleSubmit = (e?: React.FormEvent) => {
		if (e) e.preventDefault();
		if (text.trim().length === 0 || isSubmittingRef.current) return;

		isSubmittingRef.current = true;
		setIsSendButtonEnabled(false);

		onSend(text.trim(), {
			...selectedModel,
			disablePreviousMessages: disablePreviousMessages,
		});

		setText('');
		isSubmittingRef.current = false;

		// Reset + refocus
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, MAX_HEIGHT)}px`;
			setInputHeight(Math.min(inputRef.current.scrollHeight, MAX_HEIGHT));
			inputRef.current.focus();
		}
	};

	// Expose the function to get current chat options + focus
	useImperativeHandle(ref, () => ({
		getChatOptions: () => ({
			...selectedModel,
			disablePreviousMessages: disablePreviousMessages,
		}),
		focus: () => {
			if (inputRef.current) {
				inputRef.current.focus();
			}
		},
	}));

	// Clamps temperature to [0, 1] (this is triggered by the subcomponent on blur or selection)
	const setTemperature = (temp: number) => {
		const clampedTemp = Math.max(0, Math.min(1, temp));
		setSelectedModel(prev => ({
			...prev,
			temperature: clampedTemp,
		}));
	};

	return (
		<div className="relative">
			{/* Top bar with separate model dropdown, temperature dropdown, and disable checkbox */}
			<div className="flex items-center justify-between bg-base-200 gap-1 md:gap-8 mb-1 mx-4">
				<ModelDropdown
					selectedModel={selectedModel}
					setSelectedModel={setSelectedModel}
					allOptions={allOptions}
					isOpen={isModelDropdownOpen}
					setIsOpen={setIsModelDropdownOpen}
					detailsRef={modelDetailsRef}
				/>

				<TemperatureDropdown
					temperature={selectedModel.temperature ?? 0.1}
					setTemperature={setTemperature}
					isOpen={isTemperatureDropdownOpen}
					setIsOpen={setIsTemperatureDropdownOpen}
					detailsRef={temperatureDetailsRef}
				/>

				<DisablePreviousMessagesCheckbox
					disablePreviousMessages={disablePreviousMessages}
					setDisablePreviousMessages={setDisablePreviousMessages}
				/>
			</div>

			{/* Main input form for messages */}
			<form
				ref={formRef}
				onSubmit={handleSubmit}
				className="flex items-center bg-base-100 rounded-2xl border px-4 mx-2"
			>
				<textarea
					ref={inputRef}
					value={text}
					onChange={handleTextChange}
					onKeyDown={onKeyDown}
					placeholder="Type message..."
					className="flex-1 resize-none overflow-auto bg-transparent border-none outline-hidden placeholder-gray-400 min-h-[24px] max-h-[240px] p-2"
					rows={1}
					style={{ fontSize: '14px' }}
					spellCheck="false"
				/>
				<button
					type="submit"
					className={`btn btn-md !bg-transparent border-none shadow-none px-1 ${
						!isSendButtonEnabled ? 'btn-disabled' : ''
					}`}
					disabled={!isSendButtonEnabled}
					aria-label="Send message"
				>
					<FiSend size={24} />
				</button>
			</form>
		</div>
	);
});

ChatInputField.displayName = 'ChatInputField';

export default ChatInputField;
