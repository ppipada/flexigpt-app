import React, { useCallback, useEffect, useRef, useState } from 'react';

import { FiSliders } from 'react-icons/fi';

import { type ReasoningLevel, ReasoningType } from '@/spec/modelpreset';

import { useCloseDetails } from '@/hooks/use_close_details';

import { type ChatOption, DefaultChatOptions, getChatInputOptions } from '@/apis/chatoption_helper';

import AdvancedParamsModal from '@/chats/modelparams/advanced_params_modal';
import DisablePreviousMessagesCheckbox from '@/chats/modelparams/disable_checkbox';
import ModelDropdown from '@/chats/modelparams/model_dropdown';
import ReasoningTokensDropdown, { HybridReasoningCheckbox } from '@/chats/modelparams/reasoning_hybrid';
import SingleReasoningDropdown from '@/chats/modelparams/reasoning_levels_dropdown';
import TemperatureDropdown from '@/chats/modelparams/temperature_dropdown';

type ModelParamsBarProps = {
	/*  Emits the final, ready-to-use ChatOption every time something changes  */
	onOptionsChange: (options: ChatOption) => void;
};

const ModelParamsBar: React.FC<ModelParamsBarProps> = ({ onOptionsChange }) => {
	/* --------------------------------------------------------------------
	 * Internal state – everything that belongs only to the params-bar
	 * ------------------------------------------------------------------ */
	const [selectedModel, setSelectedModel] = useState<ChatOption>(DefaultChatOptions);
	const [allOptions, setAllOptions] = useState<ChatOption[]>([DefaultChatOptions]);

	const [isHybridReasoningEnabled, setIsHybridReasoningEnabled] = useState(true);
	const [disablePreviousMessages, setDisablePreviousMessages] = useState(false);

	/* --------------------------------------------------------------------
	 * Fetch initial model list / defaults
	 * ------------------------------------------------------------------ */
	const loadInitialItems = useCallback(async () => {
		const r = await getChatInputOptions();
		setSelectedModel(r.default);
		setAllOptions(r.allOptions);
		setIsHybridReasoningEnabled(r.default.reasoning?.type === ReasoningType.HybridWithTokens);
	}, []);

	useEffect(() => {
		loadInitialItems();
	}, [loadInitialItems]);

	/* --------------------------------------------------------------------
	 * Keep `isHybridReasoningEnabled` in sync with the model selection
	 * ------------------------------------------------------------------ */
	useEffect(() => {
		setIsHybridReasoningEnabled(selectedModel.reasoning?.type === ReasoningType.HybridWithTokens);
	}, [selectedModel]);

	/* --------------------------------------------------------------------
	 * Derive the final ChatOption & bubble it up whenever dependencies change
	 * ------------------------------------------------------------------ */
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

	/* --------------------------------------------------------------------
	 * Refs / open-state for dropdowns
	 * ------------------------------------------------------------------ */
	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState(false);
	const [isSecondaryDropdownOpen, setIsSecondaryDropdownOpen] = useState(false);
	const modelDetailsRef = useRef<HTMLDetailsElement>(null);
	const secondaryDetailsRef = useRef<HTMLDetailsElement>(null);

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

	/* --------------------------------------------------------------------
	 * Advanced params modal state
	 * ------------------------------------------------------------------ */
	const [isAdvancedModalOpen, setIsAdvancedModalOpen] = useState(false);

	/* --------------------------------------------------------------------
	 * Helpers for updating individual params
	 * ------------------------------------------------------------------ */
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

	/* --------------------------------------------------------------------
	 * Render
	 * ------------------------------------------------------------------ */
	return (
		<div className="flex items-center justify-between bg-base-200 mb-1 mx-8">
			{/* --------------------------- Model dropdown ----------------------- */}
			<div className="w-1/3">
				<ModelDropdown
					ref={modelDetailsRef}
					selectedModel={selectedModel}
					setSelectedModel={setSelectedModel}
					allOptions={allOptions}
					isOpen={isModelDropdownOpen}
					setIsOpen={setIsModelDropdownOpen}
				/>
			</div>

			{/* ---------------- Reasoning / temperature / checkboxes ------------ */}
			<div className="flex items-center justify-between w-2/3">
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

				<DisablePreviousMessagesCheckbox
					disablePreviousMessages={disablePreviousMessages}
					setDisablePreviousMessages={setDisablePreviousMessages}
				/>

				{/* ---------------- Advanced params button ----------------------- */}
				<button
					type="button"
					className="btn btn-sm btn-ghost mx-2 text-neutral-custom"
					onClick={() => {
						setIsAdvancedModalOpen(true);
					}}
					title="Set Advanced Params"
				>
					<FiSliders size={16} />
				</button>
			</div>

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

export default ModelParamsBar;
