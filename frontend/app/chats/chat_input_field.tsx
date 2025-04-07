import type { ModelOption } from '@/backendapihelper/settings_helper';
import { GetChatInputOptions } from '@/backendapihelper/settings_helper';
import { useCloseDetails } from '@/lib/useCloseDetails';
import { DefaultModelName, DefaultModelTitle, DefaultProviderName } from '@/models/aiprovidermodel';
import type { ChangeEvent, KeyboardEvent } from 'react';
import React, { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';
import { FiCheck, FiChevronDown, FiChevronUp, FiSend } from 'react-icons/fi';

// Define the structure of the model information and chat options

export interface ChatOptions {
	modelInfo: ModelOption;
	disablePreviousMessages: boolean;
}

export const DefaultModelOption: ModelOption = {
	title: DefaultModelTitle,
	provider: DefaultProviderName,
	name: DefaultModelName,
	temperature: 0.1,
};

interface ChatInputFieldProps {
	onSend: (message: string, options: ChatOptions) => void; // Updated to use ChatOptions
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

const ChatInputField = forwardRef<ChatInputFieldHandle, ChatInputFieldProps>(({ onSend, setInputHeight }, ref) => {
	const [text, setText] = useState<string>('');
	const [isSendButtonEnabled, setIsSendButtonEnabled] = useState<boolean>(false);
	const inputRef = useRef<HTMLTextAreaElement>(null);
	const isSubmittingRef = useRef<boolean>(false);
	const [selectedModel, setSelectedModel] = useState<ModelOption>(DefaultModelOption);
	const [disablePreviousMessages, setDisablePreviousMessages] = useState<boolean>(false);
	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState<boolean>(false);
	const [isTemperatureDropdownOpen, setIsTemperatureDropdownOpen] = useState<boolean>(false);
	const [allOptions, setAllOptions] = useState<ModelOption[]>([DefaultModelOption]);
	const modelDetailsRef = useRef<HTMLDetailsElement>(null);
	const temperatureDetailsRef = useRef<HTMLDetailsElement>(null);
	useCloseDetails({
		detailsRef: modelDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsModelDropdownOpen(false);
		},
	});
	useCloseDetails({
		detailsRef: temperatureDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsTemperatureDropdownOpen(false);
		},
	});

	const loadInitialItems = useCallback(async () => {
		const r = await GetChatInputOptions();
		setSelectedModel(r.default);
		setAllOptions(r.allOptions);
	}, []);

	useEffect(() => {
		loadInitialItems();
	}, [loadInitialItems]);

	const { formRef, onKeyDown } = useEnterSubmit();

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

		// Create the ChatOptions object
		const chatOptions: ChatOptions = {
			modelInfo: selectedModel,
			disablePreviousMessages: disablePreviousMessages,
		};

		onSend(text.trim(), chatOptions);

		setText('');
		isSubmittingRef.current = false;
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, MAX_HEIGHT)}px`;
			setInputHeight(Math.min(inputRef.current.scrollHeight, MAX_HEIGHT));
			inputRef.current.focus();
		}
	};

	// Expose the function to get current chat options
	useImperativeHandle(ref, () => ({
		getChatOptions: () => ({
			modelInfo: selectedModel,
			disablePreviousMessages: disablePreviousMessages,
		}),
		focus: () => {
			if (inputRef.current) {
				inputRef.current.focus();
			}
		},
	}));

	const setTemperature = (temp: number) => {
		setSelectedModel(prevModel => ({
			...prevModel,
			temperature: temp,
		}));
	};

	const temperatureOptions = Array.from({ length: 11 }, (_, i) => (i / 10).toFixed(1));

	return (
		<div className="relative">
			<div className="flex items-center justify-between bg-base-200 gap-1 md:gap-8 mb-1 mx-4">
				<details
					ref={modelDetailsRef}
					className="dropdown dropdown-top dropdown-end w-1/3"
					onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
						setIsModelDropdownOpen((event.currentTarget as HTMLDetailsElement).open);
					}}
				>
					<summary
						className="btn btn-xs w-full text-left text-nowrap text-neutral-400 shadow-none border-none overflow-hidden"
						title="Select Modal"
					>
						<div className="flex">
							<span className="sm:hidden">{selectedModel.title.substring(0, 8)}</span>
							<span className="hidden sm:inline">{selectedModel.title} </span>
							{isModelDropdownOpen ? (
								<FiChevronDown size={16} className="ml-1 md:ml-2" />
							) : (
								<FiChevronUp size={16} className="ml-1 md:ml-2" />
							)}
						</div>
					</summary>
					<ul className={`dropdown-content menu bg-base-100 rounded-box w-full`}>
						{allOptions.map((model, index) => (
							<li
								key={index}
								className="cursor-pointer text-xs"
								onClick={() => {
									setSelectedModel(model);
									if (modelDetailsRef.current) {
										modelDetailsRef.current.open = false;
									}
									setIsModelDropdownOpen(false);
								}}
							>
								<a className="justify-between items-center p-1 m-0">
									<span>{model.title}</span>
									{selectedModel.name && selectedModel.name === model.name && <FiCheck />}
								</a>
							</li>
						))}
					</ul>
				</details>

				<details
					ref={temperatureDetailsRef}
					className="dropdown dropdown-top dropdown-end w-1/3"
					onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
						setIsTemperatureDropdownOpen((event.currentTarget as HTMLDetailsElement).open);
					}}
				>
					<summary
						className="btn btn-xs w-full text-left text-nowrap text-neutral-400 shadow-none border-none overflow-hidden"
						title="Set temperature"
					>
						<div className="flex">
							<span className="sm:hidden">Temp:</span>
							<span className="hidden sm:inline">Temperature:</span> {selectedModel.temperature.toFixed(1)}{' '}
							{isTemperatureDropdownOpen ? (
								<FiChevronDown size={16} className="ml-1 md:ml-2" />
							) : (
								<FiChevronUp size={16} className="ml-1 md:ml-2" />
							)}
						</div>
					</summary>
					<ul className={`dropdown-content menu bg-base-100 rounded-box w-full`}>
						{temperatureOptions.map((temp, index) => (
							<li
								key={index}
								className="cursor-pointer text-xs"
								onClick={() => {
									setTemperature(parseFloat(temp));
									if (temperatureDetailsRef.current) {
										temperatureDetailsRef.current.open = false;
									}
									setIsTemperatureDropdownOpen(false);
								}}
							>
								<a className="justify-between items-center p-1 m-0">
									<span>{temp}</span>
									{selectedModel.temperature.toFixed(1) === temp && <FiCheck />}
								</a>
							</li>
						))}
					</ul>
				</details>

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
			</div>

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
