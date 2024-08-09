import { ModelName, ProviderName } from 'aiprovidermodel';
import React, {
	ChangeEvent,
	forwardRef,
	KeyboardEvent,
	useCallback,
	useEffect,
	useImperativeHandle,
	useRef,
	useState,
} from 'react';
import { FiCheck, FiChevronDown, FiChevronUp, FiSend } from 'react-icons/fi';

// Define the structure of the model information and chat options
interface ModelOption {
	title: string;
	provider: ProviderName;
	name: ModelName;
}

export interface ChatOptions {
	modelInfo?: {
		provider: ProviderName;
		model: ModelName;
		temperature: number;
	};
	disablePreviousMessages: boolean;
}

export interface ChatInputFieldProps {
	onSend: (message: string, options: ChatOptions) => void; // Updated to use ChatOptions
	setInputHeight: (height: number) => void;
}

export interface ChatInputFieldHandle {
	getChatOptions: () => ChatOptions;
}

const inputModels: ModelOption[] = [
	{ title: 'GPT4oMini', provider: ProviderName.OPENAI, name: ModelName.GPT_4O_MINI },
	{ title: 'GPT4o', provider: ProviderName.OPENAI, name: ModelName.GPT_4O },
	{ title: 'GPT4', provider: ProviderName.OPENAI, name: ModelName.GPT_4 },
];

const defaultTemperature = 0.1;
const getOptions = () => ({
	models: inputModels,
	defaultTemperature: defaultTemperature,
});

// Custom hook for handling form submission on Enter key press
function useEnterSubmit(): {
	formRef: React.RefObject<HTMLFormElement>;
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

const MAX_HEIGHT = 240;

const ChatInputField = forwardRef<ChatInputFieldHandle, ChatInputFieldProps>(({ onSend, setInputHeight }, ref) => {
	const [text, setText] = useState<string>('');
	const [isSendButtonEnabled, setIsSendButtonEnabled] = useState<boolean>(false);
	const inputRef = useRef<HTMLTextAreaElement>(null);
	const isSubmittingRef = useRef<boolean>(false);
	const [selectedModel, setSelectedModel] = useState<ModelOption>({
		title: 'GPT4oMini',
		provider: ProviderName.OPENAI,
		name: ModelName.GPT_4O_MINI,
	});
	const [temperature, setTemperature] = useState<number>(0.5);
	const [disablePreviousMessages, setDisablePreviousMessages] = useState<boolean>(false);
	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState<boolean>(false);
	const [isTemperatureDropdownOpen, setIsTemperatureDropdownOpen] = useState<boolean>(false);
	const { models, defaultTemperature } = getOptions();

	useEffect(() => {
		setTemperature(defaultTemperature);
	}, [defaultTemperature]);

	const { formRef, onKeyDown } = useEnterSubmit();

	const autoResizeTextarea = useCallback(() => {
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, MAX_HEIGHT)}px`;
			setInputHeight(Math.min(inputRef.current.scrollHeight, MAX_HEIGHT));
		}
	}, [setInputHeight]);

	const handleTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
		const value = event.target.value;
		setText(value);
		setIsSendButtonEnabled(value.trim().length > 0);
		autoResizeTextarea();
	};

	const handleSubmit = (e?: React.FormEvent) => {
		if (e) e.preventDefault();
		if (text.trim().length === 0 || isSubmittingRef.current) return;
		isSubmittingRef.current = true;
		setIsSendButtonEnabled(false);

		// Create the ChatOptions object
		const chatOptions: ChatOptions = {
			modelInfo: {
				provider: selectedModel.provider,
				model: selectedModel.name,
				temperature,
			},
			disablePreviousMessages,
		};

		onSend(text.trim(), chatOptions);

		setText('');
		isSubmittingRef.current = false;
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			inputRef.current.style.height = `${Math.min(inputRef.current.scrollHeight, MAX_HEIGHT)}px`;
			setInputHeight(Math.min(inputRef.current.scrollHeight, MAX_HEIGHT));
			inputRef.current?.focus();
		}
	};

	useEffect(() => {
		autoResizeTextarea();
	}, [text, autoResizeTextarea]);

	// Expose the function to get current chat options
	useImperativeHandle(ref, () => ({
		getChatOptions: () => ({
			modelInfo: {
				provider: selectedModel.provider,
				model: selectedModel.name,
				temperature,
			},
			disablePreviousMessages,
		}),
	}));

	const temperatureOptions = Array.from({ length: 11 }, (_, i) => (i / 10).toFixed(1));

	return (
		<div className="relative">
			<div className="flex items-center justify-between bg-base-200 gap-8 mb-1 mx-4">
				<div className="dropdown dropdown-top dropdown-end w-1/3">
					<label
						tabIndex={0}
						className="btn btn-xs w-full text-left text-neutral-400 shadow-none border-none"
						onClick={() => setIsModelDropdownOpen(!isModelDropdownOpen)}
					>
						{selectedModel.title}{' '}
						{isModelDropdownOpen ? <FiChevronDown className="ml-2" /> : <FiChevronUp className="ml-2" />}
					</label>
					<ul
						tabIndex={0}
						className={`dropdown-content menu bg-base-100 rounded-box w-full ${isModelDropdownOpen ? 'block' : 'hidden'}`}
						onClick={() => setIsModelDropdownOpen(false)}
					>
						{models.map((model, index) => (
							<li key={index} className="cursor-pointer text-xs" onClick={() => setSelectedModel(model)}>
								<a className="flex justify-between items-center p-1 m-0">
									<span>
										{model.title} ({model.provider})
									</span>
									{selectedModel.name === model.name && <FiCheck />}
								</a>
							</li>
						))}
					</ul>
				</div>

				<div className="dropdown dropdown-top dropdown-end w-1/3">
					<label
						tabIndex={0}
						className="btn btn-xs w-full text-left text-neutral-400 shadow-none border-none"
						onClick={() => setIsTemperatureDropdownOpen(!isTemperatureDropdownOpen)}
					>
						Temperature: {temperature}{' '}
						{isTemperatureDropdownOpen ? <FiChevronDown className="ml-2" /> : <FiChevronUp className="ml-2" />}
					</label>
					<ul
						tabIndex={0}
						className={`dropdown-content menu bg-base-100 rounded-box w-full ${isTemperatureDropdownOpen ? 'block' : 'hidden'}`}
						onClick={() => setIsTemperatureDropdownOpen(false)}
					>
						{temperatureOptions.map((temp, index) => (
							<li key={index} className="cursor-pointer text-xs" onClick={() => setTemperature(parseFloat(temp))}>
								<a className="flex justify-between items-center p-1 m-0">
									<span>{temp}</span>
									{temperature.toFixed(1) === temp && <FiCheck />}
								</a>
							</li>
						))}
					</ul>
				</div>

				<label className="flex items-center space-x-2 text-neutral-400 w-1/3">
					<input
						type="checkbox"
						checked={disablePreviousMessages}
						onChange={e => setDisablePreviousMessages(e.target.checked)}
						className="checkbox checkbox-xs rounded-full"
					/>
					<span className="text-xs">Disable previous messages</span>
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
					className="flex-1 resize-none overflow-hidden bg-transparent border-none outline-none placeholder-gray-400 min-h-[24px] max-h-[240px] p-2"
					rows={1}
					style={{ fontSize: '14px' }}
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
