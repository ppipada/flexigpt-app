import { type ConversationMessage } from '@/spec/conversation';
import {
	type CompletionResponseBody,
	ContentItemKind,
	ImageDetail,
	type InferenceUsage,
	type ModelParam,
	OutputKind,
	type OutputUnion,
	type ReasoningContent,
	RoleEnum,
	Status,
	type ToolCall,
	type ToolOutput,
	type ToolOutputItemUnion,
} from '@/spec/inference';
import type { ProviderName } from '@/spec/modelpreset';
import {
	type ToolStoreChoice,
	type ToolStoreChoiceType,
	ToolStoreOutputKind,
	type ToolStoreOutputUnion,
	type UIToolCall,
	type UIToolOutput,
} from '@/spec/tool';

import { getUUIDv7 } from '@/lib/uuid_utils';

import { providerSetAPI } from '@/apis/baseapi';

import { extractPrimaryTextFromToolStoreOutputs, formatToolOutputSummary } from '@/chats/tools/tool_editor_utils';

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
		// console.log('history to completion', JSON.stringify(history, null, 2));
		const choiceMap = new Map<string, ToolStoreChoice>(
			(toolStoreChoices ?? []).map(choice => [choice.choiceID, choice])
		);

		const resp = await providerSetAPI.fetchCompletion(
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

		const inf = resp.inferenceResponse;
		const hasModelError = !!inf?.error;
		const hasOutputs = !!inf?.outputs && inf.outputs.length > 0;

		if (!hasModelError || hasOutputs) {
			const assistantMsg = buildAssistantMessageFromResponse(assistantPlaceholder.id, modelParams, resp, choiceMap);
			return { responseMessage: assistantMsg, rawResponse: resp };
		}

		// Error with no outputs at all -> fall back to existing "error stub".
		return getErrorStub(modelParams, assistantPlaceholder, resp, undefined);
	} catch (error) {
		if ((error as DOMException).name === 'AbortError') {
			throw error;
		}
		return getErrorStub(modelParams, assistantPlaceholder, undefined, error);
	}
}

function getErrorStub(
	modelParams: ModelParam,
	assistantPlaceholder: ConversationMessage,
	rawResponse: CompletionResponseBody | undefined,
	errorObj: any
) {
	assistantPlaceholder.modelParam = modelParams;
	assistantPlaceholder.status = Status.Failed;
	// Optionally keep the raw error on the message
	assistantPlaceholder.error = errorObj;
	const outText = (assistantPlaceholder.uiContent || '') + '\n\n>Got error in API processing.';
	assistantPlaceholder.uiContent = outText;
	assistantPlaceholder.outputs = [
		{
			kind: OutputKind.OutputMessage,
			outputMessage: {
				id: getUUIDv7(),
				role: RoleEnum.Assistant,
				status: Status.Failed,
				contents: [
					{
						kind: ContentItemKind.Text,
						textItem: {
							text: outText,
						},
					},
				],
			},
		},
	];

	// Prefer backend debugDetails if present
	let detailsMarkdown: string = '';
	if (rawResponse?.inferenceResponse) {
		detailsMarkdown =
			getDebugDetailsMarkdown(rawResponse.inferenceResponse.debugDetails, rawResponse.inferenceResponse.error) ?? '';
		assistantPlaceholder.debugDetails = rawResponse.inferenceResponse.debugDetails;
	}

	if (errorObj !== undefined && errorObj !== null) {
		detailsMarkdown = detailsMarkdown + '\n\n### Error\n\n' + getQuotedJSON(errorObj);
	}

	assistantPlaceholder.uiDebugDetails = detailsMarkdown;

	return { responseMessage: assistantPlaceholder, rawResponse };
}

export function getDebugDetailsMarkdown(debugObj?: any, errorObj?: any): string | undefined {
	const parts: string[] = [];

	const pushJSONBlock = (title: string, value: unknown) => {
		if (value === undefined) return;

		try {
			parts.push(title, getQuotedJSON(value));
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			parts.push(title, `\`[Failed to serialize ${title.replace(/^#+\s*/, '').toLowerCase()}: ${msg}]\``);
		}
	};

	// 1. Error object should always be first, if present
	if (errorObj !== undefined) {
		pushJSONBlock('### Error', errorObj);
	}

	// 2. Handle debug details
	if (debugObj !== undefined && debugObj !== null) {
		const isMap = debugObj instanceof Map;
		const isObjectLike = typeof debugObj === 'object' && !isMap;

		const asRecord: Record<string, unknown> | undefined = isObjectLike
			? (debugObj as Record<string, unknown>)
			: undefined;

		const hasKey = (key: string): boolean => {
			if (isMap) return (debugObj as Map<string, unknown>).has(key);
			if (asRecord) return key in asRecord;
			return false;
		};

		const getValue = (key: string): unknown => {
			if (isMap) return (debugObj as Map<string, unknown>).get(key);
			if (asRecord) return asRecord[key];
			return undefined;
		};

		const hasRequest = hasKey('requestDetails');
		const hasResponse = hasKey('responseDetails');
		const hasProvider = hasKey('providerResponse');
		const hasErrorDetails = hasKey('errorDetails');

		const hasStructuredKeys = hasRequest || hasResponse || hasProvider || hasErrorDetails;

		if (hasStructuredKeys) {
			// Order: error details (from debug) → request → response → provider
			if (hasErrorDetails) {
				pushJSONBlock('### Error details', getValue('errorDetails'));
			}

			if (hasRequest) {
				pushJSONBlock('### Request debug details', getValue('requestDetails'));
			}

			if (hasResponse) {
				pushJSONBlock('### Response debug details', getValue('responseDetails'));
			}

			if (hasProvider) {
				pushJSONBlock('### Provider response debug details', getValue('providerResponse'));
			}
		} else {
			// No special keys: fallback to original behavior (one block)
			pushJSONBlock('### Debug details', debugObj);
		}
	}

	if (parts.length === 0) {
		return undefined;
	}

	return parts.join('\n\n');
}

function buildAssistantMessageFromResponse(
	baseId: string,
	modelParams: ModelParam,
	resp: CompletionResponseBody,
	choiceMap: Map<string, ToolStoreChoice>
): ConversationMessage | undefined {
	const now = new Date();
	const id = baseId || now.toISOString();

	if (!resp.inferenceResponse) {
		return undefined;
	}

	const inf = resp.inferenceResponse;
	const outputs = inf.outputs ?? [];
	const usage: InferenceUsage | undefined = inf.usage;
	const error = inf.error;

	const { uiContent, uiReasoningContents, uiToolCalls, uiToolOutputs } = deriveUIFieldsFromOutputUnion(
		outputs,
		choiceMap
	);
	const debugDetails = inf.debugDetails;
	const uiDebugDetails = getDebugDetailsMarkdown(inf.debugDetails, inf.error);

	const s = error ? Status.Failed : Status.Completed;
	const msg: ConversationMessage = {
		id,
		createdAt: now,
		role: RoleEnum.Assistant,
		status: s,
		modelParam: modelParams,
		outputs,
		usage,
		error,
		debugDetails,
		uiContent,
		uiReasoningContents,
		uiToolCalls,
		uiToolOutputs,
		uiDebugDetails,
	};
	// console.log('assistant from message out', JSON.stringify(msg, null, 2));
	return msg;
}

export function deriveUIFieldsFromOutputUnion(
	outputs: OutputUnion[] | undefined,
	choiceMap: Map<string, ToolStoreChoice>
): {
	uiContent: string;
	uiReasoningContents?: ReasoningContent[];
	uiToolCalls?: UIToolCall[];
	uiToolOutputs?: UIToolOutput[];
} {
	if (!outputs || outputs.length === 0) {
		return { uiContent: '' };
	}

	const textParts: string[] = [];
	const reasoning: ReasoningContent[] = [];
	const toolCalls: UIToolCall[] = [];
	const toolOutputs: UIToolOutput[] = [];

	let toolCallMap: Map<string, ToolCall> | undefined;

	for (const o of outputs) {
		switch (o.kind) {
			case OutputKind.OutputMessage: {
				const msg = o.outputMessage;
				if (!msg || !msg.contents) break;
				for (const c of msg.contents) {
					if (c.kind === ContentItemKind.Text && c.textItem) {
						const t = c.textItem.text.trim();
						if (t) textParts.push(t);
					}
				}
				break;
			}

			case OutputKind.ReasoningMessage:
				if (o.reasoningMessage) {
					reasoning.push(o.reasoningMessage);
				}
				break;

			case OutputKind.FunctionToolCall: {
				const uiFunctionToolCall = deriveUIToolCallFromToolCall(o.functionToolCall, choiceMap);
				if (uiFunctionToolCall) toolCalls.push(uiFunctionToolCall);
				break;
			}
			case OutputKind.CustomToolCall: {
				const uiCustomToolCall = deriveUIToolCallFromToolCall(o.customToolCall, choiceMap);
				if (uiCustomToolCall) toolCalls.push(uiCustomToolCall);
				break;
			}
			case OutputKind.WebSearchToolCall: {
				const uiWebsearchToolCall = deriveUIToolCallFromToolCall(o.webSearchToolCall, choiceMap);
				if (uiWebsearchToolCall) toolCalls.push(uiWebsearchToolCall);
				break;
			}

			case OutputKind.WebSearchToolOutput: {
				const out = o.webSearchToolOutput;
				if (out) {
					if (!toolCallMap) {
						// outputs is definitely defined here because of the early return
						toolCallMap = buildToolCallMapFromOutputs(outputs);
					}
					// Only called when a real ToolOutput exists.
					toolOutputs.push(buildUIToolOutputFromToolOutput(out, choiceMap, toolCallMap));
				}
				break;
			}
		}
	}

	const content = textParts.join('\n\n');

	return {
		uiContent: content,
		uiReasoningContents: reasoning.length > 0 ? reasoning : undefined,
		uiToolCalls: toolCalls.length > 0 ? toolCalls : undefined,
		uiToolOutputs: toolOutputs.length > 0 ? toolOutputs : undefined,
	};
}

function deriveUIToolCallFromToolCall(
	toolCall: ToolCall | undefined,
	choiceMap: Map<string, ToolStoreChoice>
): UIToolCall | undefined {
	if (!toolCall) return undefined;

	const choiceID = toolCall.choiceID;
	if (!choiceID) return undefined;

	const toolStoreChoice = choiceMap.get(choiceID); // ToolStoreChoice | undefined
	if (!toolStoreChoice) return undefined;

	return {
		id: toolCall.id || toolCall.callID,
		callID: toolCall.callID,
		name: toolCall.name,
		arguments: toolCall.arguments ?? '',
		webSearchToolCallItems: toolCall.webSearchToolCallItems,
		type: toolCall.type as unknown as ToolStoreChoiceType,
		choiceID: toolCall.choiceID,
		// The LLM would consider the status of tool call as done as soon as it is in output,
		// for us the call is pending here and then it will run and move to final status.
		status: 'pending',
		toolStoreChoice,
	};
}

export function buildUIToolOutputFromToolOutput(
	out: ToolOutput,
	choiceMap: Map<string, ToolStoreChoice>,
	toolCallMap?: Map<string, ToolCall>
): UIToolOutput {
	const isError = out.isError;

	const toolStoreOutputs = mapToolOutputItemsToToolStoreOutputs(out.contents);
	const primaryText = extractPrimaryTextFromToolStoreOutputs(toolStoreOutputs);

	const summaryBase = formatToolOutputSummary(out.name);
	const summary = isError && primaryText ? `Error: ${primaryText.split('\n')[0].slice(0, 80)}` : summaryBase;

	const toolStoreChoice = choiceMap.get(out.choiceID);
	const call = toolCallMap?.get(out.callID);

	return {
		id: out.id,
		callID: out.callID,
		name: out.name,
		choiceID: out.choiceID,

		// ToolType and ToolStoreChoiceType share the same string enum values.
		type: out.type as unknown as ToolStoreChoiceType,

		summary,
		toolStoreOutputs,

		toolStoreChoice:
			toolStoreChoice ??
			({
				// Very defensive fallback; ideally you never hit this.
				choiceID: out.choiceID,
				bundleID: '',
				toolSlug: out.name,
				toolVersion: '',
				toolType: out.type as unknown as ToolStoreChoiceType,
			} as ToolStoreChoice),

		isError,
		errorMessage: isError ? primaryText : undefined,

		// Hydrate from the original call, if present.
		arguments: call?.arguments,
		webSearchToolCallItems: call?.webSearchToolCallItems,
	};
}

// Helper: map inference ToolOutput.contents -> UIToolOutput.toolStoreOutputs
function mapToolOutputItemsToToolStoreOutputs(contents?: ToolOutputItemUnion[]): ToolStoreOutputUnion[] | undefined {
	if (!contents?.length) return undefined;

	const outputs: ToolStoreOutputUnion[] = [];

	for (const item of contents) {
		switch (item.kind) {
			case ContentItemKind.Text: {
				const text = item.textItem?.text;
				if (text != null) {
					outputs.push({
						kind: ToolStoreOutputKind.Text,
						textItem: { text },
					});
				}
				break;
			}

			case ContentItemKind.Image: {
				const img = item.imageItem;
				if (img) {
					outputs.push({
						kind: ToolStoreOutputKind.Image,
						imageItem: {
							// ToolStoreOutputImage.detail is a string; keep enum value or default to "auto"
							detail: (img.detail ?? ImageDetail.Auto) as unknown as string,
							imageName: img.imageName ?? '',
							imageMIME: img.imageMIME ?? '',
							imageData: img.imageData ?? '',
						},
					});
				}
				break;
			}

			case ContentItemKind.File: {
				const file = item.fileItem;
				if (file) {
					outputs.push({
						kind: ToolStoreOutputKind.File,
						fileItem: {
							fileName: file.fileName ?? '',
							fileMIME: file.fileMIME ?? '',
							fileData: file.fileData ?? '',
						},
					});
				}
				break;
			}

			default:
				// Refusal / other kinds are ignored for tool-store outputs
				break;
		}
	}

	return outputs.length ? outputs : undefined;
}

function buildToolCallMapFromOutputs(outputs: OutputUnion[]): Map<string, ToolCall> {
	const map = new Map<string, ToolCall>();

	const addCall = (call?: ToolCall) => {
		if (call?.callID) map.set(call.callID, call);
	};

	for (const o of outputs) {
		switch (o.kind) {
			case OutputKind.FunctionToolCall:
				addCall(o.functionToolCall);
				break;
			case OutputKind.CustomToolCall:
				addCall(o.customToolCall);
				break;
			case OutputKind.WebSearchToolCall:
				addCall(o.webSearchToolCall);
				break;
		}
	}

	return map;
}

function getQuotedJSON(obj: any): string {
	return '```json\n' + JSON.stringify(obj, null, 2) + '\n```';
}
