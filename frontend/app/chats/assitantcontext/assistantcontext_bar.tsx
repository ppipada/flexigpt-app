import React, { useCallback, useEffect, useRef, useState } from 'react';

import { FiSliders } from 'react-icons/fi';

import { type ReasoningLevel, ReasoningType } from '@/spec/modelpreset';

import { useCloseDetails } from '@/hooks/use_close_details';

import { type ChatOption, DefaultChatOptions, getChatInputOptions } from '@/apis/chatoption_helper';

import AdvancedParamsModal from '@/chats/assitantcontext/advanced_params_modal';
import DisablePreviousMessagesCheckbox from '@/chats/assitantcontext/disable_checkbox';
import ModelDropdown from '@/chats/assitantcontext/model_dropdown';
import ReasoningTokensDropdown, { HybridReasoningCheckbox } from '@/chats/assitantcontext/reasoning_hybrid';
import SingleReasoningDropdown from '@/chats/assitantcontext/reasoning_levels_dropdown';
import SystemPromptDropdown, {
	createSystemPromptItem,
	type SystemPromptItem,
} from '@/chats/assitantcontext/system_prompt';
import TemperatureDropdown from '@/chats/assitantcontext/temperature_dropdown';
import { useSetSystemPromptForChat } from '@/chats/events/set_system_prompt';

type AssistantContextBarProps = {
	onOptionsChange: (options: ChatOption) => void;
};

const AssistantContextBar: React.FC<AssistantContextBarProps> = ({ onOptionsChange }) => {
	const [selectedModel, setSelectedModel] = useState<ChatOption>(DefaultChatOptions);
	const [allOptions, setAllOptions] = useState<ChatOption[]>([DefaultChatOptions]);

	const [isHybridReasoningEnabled, setIsHybridReasoningEnabled] = useState(true);
	const [disablePreviousMessages, setDisablePreviousMessages] = useState(false);
	const [systemPrompts, setSystemPrompts] = useState<SystemPromptItem[]>([]);

	const loadInitialItems = useCallback(async () => {
		const r = await getChatInputOptions();
		setSelectedModel(r.default);
		setAllOptions(r.allOptions);
		setIsHybridReasoningEnabled(r.default.reasoning?.type === ReasoningType.HybridWithTokens);
		// Seed system prompts with current default if available
		const initialSP = r.default.systemPrompt.trim();
		if (initialSP) {
			setSystemPrompts(prev =>
				prev.some(i => i.prompt === initialSP) ? prev : [...prev, createSystemPromptItem(initialSP, { locked: true })]
			);
		}
	}, []);

	useEffect(() => {
		loadInitialItems();
	}, [loadInitialItems]);

	useEffect(() => {
		setIsHybridReasoningEnabled(selectedModel.reasoning?.type === ReasoningType.HybridWithTokens);
	}, [selectedModel]);

	useEffect(() => {
		const sp = selectedModel.systemPrompt.trim();
		if (sp) {
			setSystemPrompts(prev => (prev.some(i => i.prompt === sp) ? prev : [...prev, createSystemPromptItem(sp)]));
		}
	}, [selectedModel.systemPrompt]);

	const buildFinalOptions = useCallback((): ChatOption => {
		const base = { ...selectedModel, disablePreviousMessages };

		// remove reasoning when hybrid reasoning is toggled off
		if (selectedModel.reasoning?.type === ReasoningType.HybridWithTokens && !isHybridReasoningEnabled) {
			const modifiedOptions = { ...base };
			delete modifiedOptions.reasoning;
			return modifiedOptions;
		}

		return base;
	}, [selectedModel, disablePreviousMessages, isHybridReasoningEnabled]);

	useEffect(() => {
		onOptionsChange(buildFinalOptions());
	}, [buildFinalOptions, onOptionsChange]);

	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState(false);
	const [isSecondaryDropdownOpen, setIsSecondaryDropdownOpen] = useState(false);
	const [isSystemDropdownOpen, setIsSystemDropdownOpen] = useState(false);

	const modelDetailsRef = useRef<HTMLDetailsElement>(null);
	const secondaryDetailsRef = useRef<HTMLDetailsElement>(null);
	const systemDetailsRef = useRef<HTMLDetailsElement>(null);

	useCloseDetails({
		detailsRef: modelDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsModelDropdownOpen(false);
		},
	});

	useCloseDetails({
		detailsRef: secondaryDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsSecondaryDropdownOpen(false);
		},
	});

	useCloseDetails({
		detailsRef: systemDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsSystemDropdownOpen(false);
		},
	});

	const [isAdvancedModalOpen, setIsAdvancedModalOpen] = useState(false);

	const setTemperature = useCallback(
		(temp: number) => {
			const clampedTemp = Math.max(0, Math.min(1, temp));
			setSelectedModel(prev => ({ ...prev, temperature: clampedTemp }));
		},
		[setSelectedModel]
	);

	const setReasoningLevel = useCallback(
		(newLevel: ReasoningLevel) => {
			setSelectedModel(prev => ({
				...prev,
				reasoning: {
					type: ReasoningType.SingleWithLevels,
					level: newLevel,
					tokens: 1024,
				},
			}));
		},
		[setSelectedModel]
	);

	const setHybridTokens = useCallback(
		(tokens: number) => {
			setSelectedModel(prev => {
				if (!prev.reasoning || prev.reasoning.type !== ReasoningType.HybridWithTokens) return prev;
				return { ...prev, reasoning: { ...prev.reasoning, tokens } };
			});
		},
		[setSelectedModel]
	);

	const selectSystemPrompt = useCallback(
		(item: SystemPromptItem) => {
			setSelectedModel(prev => ({ ...prev, systemPrompt: item.prompt }));
		},
		[setSelectedModel]
	);

	const clearSystemPrompt = useCallback(() => {
		setSelectedModel(prev => ({ ...prev, systemPrompt: '' }));
	}, []);

	const editSystemPrompt = useCallback(
		(id: string, updatedPrompt: string) => {
			const updatedText = updatedPrompt.trim();
			if (!updatedText) return;
			setSystemPrompts(prev => {
				const oldItem = prev.find(i => i.id === id);
				if (!oldItem) return prev;
				const updated = prev.map(i =>
					i.id === id
						? {
								...i,
								prompt: updatedText,
								title: updatedText.length > 24 ? `${updatedText.slice(0, 24)}…` : updatedText || '(empty)',
							}
						: i
				);
				if ((selectedModel.systemPrompt || '').trim() === (oldItem.prompt || '').trim()) {
					setSelectedModel(sel => ({ ...sel, systemPrompt: updatedText }));
				}
				return updated;
			});
		},
		[selectedModel.systemPrompt]
	);

	const addSystemPrompt = useCallback(
		(item: SystemPromptItem) => {
			const p = item.prompt.trim();
			if (!p) return;
			setSystemPrompts(prev => (prev.some(i => i.prompt === p) ? prev : [...prev, item]));
			setSelectedModel(prev => ({ ...prev, systemPrompt: p }));
		},
		[setSelectedModel]
	);

	const removeSystemPrompt = useCallback((id: string) => {
		setSystemPrompts(prev => {
			const target = prev.find(i => i.id === id);
			if (target?.locked) return prev; // don't remove locked default
			return prev.filter(i => i.id !== id);
		});
	}, []);

	useSetSystemPromptForChat(prompt => {
		const p = (prompt || '').trim();
		if (!p) return;
		setSystemPrompts(prev => (prev.some(i => i.prompt === p) ? prev : [...prev, createSystemPromptItem(p)]));
		setSelectedModel(prev => ({ ...prev, systemPrompt: p }));
	});

	return (
		<div className="bg-base-200 mx-4 mb-1 flex items-center justify-between">
			<ModelDropdown
				ref={modelDetailsRef}
				selectedModel={selectedModel}
				setSelectedModel={setSelectedModel}
				allOptions={allOptions}
				isOpen={isModelDropdownOpen}
				setIsOpen={setIsModelDropdownOpen}
			/>

			{/* ---------------- Reasoning / temperature / checkboxes ------------ */}

			{selectedModel.reasoning?.type === ReasoningType.HybridWithTokens && (
				<HybridReasoningCheckbox
					isReasoningEnabled={isHybridReasoningEnabled}
					setIsReasoningEnabled={setIsHybridReasoningEnabled}
				/>
			)}

			{selectedModel.reasoning?.type === ReasoningType.HybridWithTokens ? (
				isHybridReasoningEnabled ? (
					<ReasoningTokensDropdown
						ref={secondaryDetailsRef}
						tokens={selectedModel.reasoning.tokens}
						setTokens={setHybridTokens}
						isOpen={isSecondaryDropdownOpen}
						setIsOpen={setIsSecondaryDropdownOpen}
					/>
				) : (
					<TemperatureDropdown
						ref={secondaryDetailsRef}
						temperature={selectedModel.temperature ?? 0.1}
						setTemperature={setTemperature}
						isOpen={isSecondaryDropdownOpen}
						setIsOpen={setIsSecondaryDropdownOpen}
					/>
				)
			) : selectedModel.reasoning?.type === ReasoningType.SingleWithLevels ? (
				<SingleReasoningDropdown
					ref={secondaryDetailsRef}
					reasoningLevel={selectedModel.reasoning.level}
					setReasoningLevel={setReasoningLevel}
					isOpen={isSecondaryDropdownOpen}
					setIsOpen={setIsSecondaryDropdownOpen}
				/>
			) : (
				<TemperatureDropdown
					ref={secondaryDetailsRef}
					temperature={selectedModel.temperature ?? 0.1}
					setTemperature={setTemperature}
					isOpen={isSecondaryDropdownOpen}
					setIsOpen={setIsSecondaryDropdownOpen}
				/>
			)}

			<SystemPromptDropdown
				ref={systemDetailsRef}
				prompts={systemPrompts}
				selectedPromptId={systemPrompts.find(i => i.prompt === (selectedModel.systemPrompt.trim() || ''))?.id}
				onSelect={selectSystemPrompt}
				onAdd={addSystemPrompt}
				onEdit={editSystemPrompt}
				onRemove={removeSystemPrompt}
				onClear={clearSystemPrompt}
				isOpen={isSystemDropdownOpen}
				setIsOpen={setIsSystemDropdownOpen}
			/>

			<DisablePreviousMessagesCheckbox
				disablePreviousMessages={disablePreviousMessages}
				setDisablePreviousMessages={setDisablePreviousMessages}
			/>

			{/* ---------------- Advanced params button ----------------------- */}
			<button
				type="button"
				className="btn btn-sm btn-ghost text-neutral-custom"
				onClick={() => {
					setIsAdvancedModalOpen(true);
				}}
				title="Set Advanced Params"
			>
				<FiSliders size={16} />
			</button>

			{/* ---------------- Advanced params modal -------------------------- */}
			{isAdvancedModalOpen && (
				<AdvancedParamsModal
					isOpen={isAdvancedModalOpen}
					onClose={() => {
						setIsAdvancedModalOpen(false);
					}}
					currentModel={selectedModel}
					onSave={(updatedModel: ChatOption) => {
						setSelectedModel(updatedModel);
						setIsAdvancedModalOpen(false);
					}}
				/>
			)}
		</div>
	);
};

export default AssistantContextBar;
