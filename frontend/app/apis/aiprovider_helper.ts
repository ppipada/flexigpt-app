/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import type { ConversationMessage } from '@/spec/conversation';
import type {
	CompletionResponseBody,
	InferenceUsage,
	InputOutputContent,
	InputOutputContentItemUnion,
	InputUnion,
	ModelParam,
	OutputUnion,
	ReasoningContent,
	ToolOutput,
	ToolOutputItemUnion,
} from '@/spec/inference';
import { ContentItemKind, InputKind, OutputKind, RoleEnum, Status, ToolType } from '@/spec/inference';
import { type ProviderName } from '@/spec/modelpreset';
import { type ToolStoreChoice, ToolStoreChoiceType, type UIToolCallChip, type UIToolOutput } from '@/spec/tool';

import { log, providerSetAPI } from '@/apis/baseapi';

import { uiAttachmentToConversation } from '@/chats/attachments/attachment_editor_utils';
import type { EditorSubmitPayload } from '@/chats/chat_input_editor';

/**
 * @public
 */
export function getQuotedJSON(obj: any): string {
	return '```json\n' + JSON.stringify(obj, null, 2) + '\n```';
}

function uiToolOutputToInferenceToolOutput(ui: UIToolOutput): ToolOutput | undefined {
	// Only function/custom; skip webSearch for now.
	if (ui.type !== ToolStoreChoiceType.Function && ui.type !== ToolStoreChoiceType.Custom) return undefined;

	const contents: ToolOutputItemUnion[] = [
		{
			kind: ContentItemKind.Text,
			textItem: { text: ui.rawOutput },
		},
	];

	return {
		type: ui.type as unknown as ToolType,
		choiceID: ui.choiceID,
		id: ui.id,
		callID: ui.callID,
		role: RoleEnum.Tool,
		status: Status.Completed,
		cacheControl: undefined,
		name: ui.name,
		isError: false,
		signature: undefined,
		contents,
		webSearchToolOutputItems: undefined,
	};
}

export function buildUserConversationMessageFromEditor(
	payload: EditorSubmitPayload,
	existingId?: string
): ConversationMessage {
	const now = new Date();
	const id = existingId ?? now.toISOString();

	const text = (payload.text ?? '').trim();
	const hasText = text.length > 0;

	const contents: InputOutputContentItemUnion[] = [];
	if (hasText) {
		contents.push({
			kind: ContentItemKind.Text,
			textItem: { text },
		});
	}

	const inputMessage: InputOutputContent = {
		id,
		role: RoleEnum.User,
		status: Status.Completed,
		contents,
	};

	const inputs: InputUnion[] = [
		{
			kind: InputKind.InputMessage,
			inputMessage,
		},
	];

	// Attach tool outputs as FunctionToolOutput / CustomToolOutput events.
	for (const ui of payload.toolOutputs) {
		const infOut = uiToolOutputToInferenceToolOutput(ui);
		if (!infOut) continue;

		if (infOut.type === ToolType.Function) {
			inputs.push({
				kind: InputKind.FunctionToolOutput,
				functionToolOutput: infOut,
			});
		} else if (infOut.type === ToolType.Custom) {
			inputs.push({
				kind: InputKind.CustomToolOutput,
				customToolOutput: infOut,
			});
		}
	}

	const attachments = payload.attachments.length > 0 ? payload.attachments.map(uiAttachmentToConversation) : undefined;

	const toolStoreChoices = payload.finalToolChoices.length > 0 ? payload.finalToolChoices : undefined;

	const msg: ConversationMessage = {
		id,
		createdAt: now,
		role: RoleEnum.User,
		status: Status.Completed,
		inputs,
		attachments,
		toolStoreChoices,
		content: text,
	};

	return msg;
}

function deriveUIFieldsFromOutputs(outputs: OutputUnion[] | undefined): {
	content: string;
	reasoningContents?: ReasoningContent[];
	toolCalls?: UIToolCallChip[];
	toolOutputs?: UIToolOutput[];
} {
	if (!outputs || outputs.length === 0) {
		return { content: '' };
	}

	const textParts: string[] = [];
	const reasoning: ReasoningContent[] = [];
	const toolCalls: UIToolCallChip[] = [];
	const toolOutputs: UIToolOutput[] = [];

	for (const o of outputs) {
		switch (o.kind) {
			case OutputKind.OutputMessage: {
				const msg = o.outputMessage;
				if (!msg || !msg.contents) break;
				for (const c of msg.contents) {
					if (c.kind === ContentItemKind.Text && c.textItem) {
						const t = (c.textItem.text ?? '').trim();
						if (t) textParts.push(t);
					}
					if (c.kind === ContentItemKind.Refusal && c.refusalItem) {
						const t = (c.refusalItem.refusal ?? '').trim();
						if (t) textParts.push(t);
					}
					// image/file content items are not rendered inline in text bubble
				}
				break;
			}

			case OutputKind.ReasoningMessage:
				if (o.reasoningMessage) {
					reasoning.push(o.reasoningMessage);
				}
				break;

			case OutputKind.FunctionToolCall:
			case OutputKind.CustomToolCall:
			case OutputKind.WebSearchToolCall:
				if (o.functionToolCall) toolCalls.push(o.functionToolCall as unknown as UIToolCallChip);
				if (o.customToolCall) toolCalls.push(o.customToolCall as unknown as UIToolCallChip);
				if (o.webSearchToolCall) toolCalls.push(o.webSearchToolCall as unknown as UIToolCallChip);
				break;

			case OutputKind.WebSearchToolOutput:
				if (o.webSearchToolOutput) {
					toolOutputs.push(o.webSearchToolOutput as unknown as UIToolOutput);
				}
				break;
		}
	}

	const content = textParts.join('\n\n');

	return {
		content,
		reasoningContents: reasoning.length > 0 ? reasoning : undefined,
		toolCalls: toolCalls.length > 0 ? toolCalls : undefined,
		toolOutputs: toolOutputs.length > 0 ? toolOutputs : undefined,
	};
}

function buildAssistantMessageFromResponse(
	baseId: string,
	modelParams: ModelParam,
	resp: CompletionResponseBody
): ConversationMessage | undefined {
	const now = new Date();
	const id = baseId || now.toISOString();

	if (!resp.inferenceResponse) {
		return undefined;
	}

	const outputs = resp.inferenceResponse.outputs ?? [];
	const usage: InferenceUsage | undefined = resp.inferenceResponse.usage;
	const error = resp.inferenceResponse.error;

	const { content, reasoningContents, toolCalls, toolOutputs } = deriveUIFieldsFromOutputs(outputs);

	const msg: ConversationMessage = {
		id,
		createdAt: now,
		role: RoleEnum.Assistant,
		status: error ? Status.Failed : Status.Completed,
		modelParam: modelParams,
		outputs,
		usage,
		error,
		// UI-only fields:
		content,
		reasoningContents,
		toolCalls,
		toolOutputs,
	};

	return msg;
}
export async function HandleCompletion(
	provider: ProviderName,
	modelParams: ModelParam,
	currentUserMsg: ConversationMessage,
	history: ConversationMessage[],
	toolStoreChoices: ToolStoreChoice[] | undefined,
	assistantPlaceholder: ConversationMessage,
	requestId?: string,
	signal?: AbortSignal,
	onStreamTextData?: (textData: string) => void,
	onStreamThinkingData?: (thinkingData: string) => void
): Promise<{
	responseMessage: ConversationMessage | undefined;
	rawResponse?: CompletionResponseBody;
}> {
	try {
		const resp = await providerSetAPI.completion(
			provider,
			modelParams,
			currentUserMsg,
			history,
			toolStoreChoices,
			requestId,
			signal,
			onStreamTextData,
			onStreamThinkingData
		);

		if (!resp) {
			return { responseMessage: undefined, rawResponse: undefined };
		}

		const assistantMsg = buildAssistantMessageFromResponse(assistantPlaceholder.id, modelParams, resp);
		return { responseMessage: assistantMsg, rawResponse: resp };
	} catch (error) {
		if ((error as DOMException).name === 'AbortError') {
			throw error;
		}
		const details = JSON.stringify(error, null, 2);
		log.error('provider completion failed', details);

		assistantPlaceholder.content = (assistantPlaceholder.content || '') + '\n\n>Got error in API processing.';
		assistantPlaceholder.details = details;

		return { responseMessage: assistantPlaceholder, rawResponse: undefined };
	}
}
