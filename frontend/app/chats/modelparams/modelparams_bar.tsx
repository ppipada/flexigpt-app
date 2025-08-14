import React, { useCallback, useRef, useState } from 'react';

import { FiSliders } from 'react-icons/fi';

import { type ReasoningLevel, ReasoningType } from '@/spec/modelpreset';

import { useCloseDetails } from '@/hooks/use_close_details';

import { type ChatOption } from '@/apis/chatoption_helper';

import AdvancedParamsModal from '@/chats/modelparams/advanced_params_modal';
import DisablePreviousMessagesCheckbox from '@/chats/modelparams/disable_checkbox';
import ModelDropdown from '@/chats/modelparams/model_dropdown';
import ReasoningTokensDropdown, { HybridReasoningCheckbox } from '@/chats/modelparams/reasoning_hybrid';
import SingleReasoningDropdown from '@/chats/modelparams/reasoning_levels_dropdown';
import TemperatureDropdown from '@/chats/modelparams/temperature_dropdown';

type ModelParamsBarProps = {
	selectedModel: ChatOption;
	setSelectedModel: React.Dispatch<React.SetStateAction<ChatOption>>;
	allOptions: ChatOption[];

	disablePreviousMessages: boolean;
	setDisablePreviousMessages: React.Dispatch<React.SetStateAction<boolean>>;

	isHybridReasoningEnabled: boolean;
	setIsHybridReasoningEnabled: React.Dispatch<React.SetStateAction<boolean>>;
};

const ModelParamsBar: React.FC<ModelParamsBarProps> = ({
	selectedModel,
	setSelectedModel,
	allOptions,
	disablePreviousMessages,
	setDisablePreviousMessages,
	isHybridReasoningEnabled,
	setIsHybridReasoningEnabled,
}) => {
	// Dropdown open state and refs
	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState(false);
	const [isSecondaryDropdownOpen, setIsSecondaryDropdownOpen] = useState(false);
	const modelDetailsRef = useRef<HTMLDetailsElement>(null);
	const secondaryDetailsRef = useRef<HTMLDetailsElement>(null);

	// Advanced params modal state
	const [isAdvancedModalOpen, setIsAdvancedModalOpen] = useState(false);

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

	// Helpers to update model params
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

	return (
		<div className="flex items-center justify-between bg-base-200 mb-1 mx-8">
			{/* Model dropdown */}
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

			{/* Middle section (reasoning / temperature) */}
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
						temperature={selectedModel.temperature ?? 0.1}
						setTemperature={setTemperature}
						isOpen={isSecondaryDropdownOpen}
						setIsOpen={setIsSecondaryDropdownOpen}
						ref={secondaryDetailsRef}
					/>
				)}

				<DisablePreviousMessagesCheckbox
					disablePreviousMessages={disablePreviousMessages}
					setDisablePreviousMessages={setDisablePreviousMessages}
				/>

				{/* Advanced params modal trigger */}
				<button
					type="button"
					className="btn btn-sm btn-ghost mx-2 text-neutral-custom"
					onClick={() => {
						setIsAdvancedModalOpen(true);
					}}
					title="Set Advanced Params"
				>
					<FiSliders color="text-neutral-custom" size={16} />
				</button>
			</div>

			{/* Advanced params modal */}
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
