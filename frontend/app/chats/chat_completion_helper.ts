import {
	type Conversation,
	CONVERSATION_SCHEMA_VERSION,
	type ConversationMessage,
	type StoreConversation,
	type StoreConversationMessage,
} from '@/spec/conversation';
import {
	type CompletionResponseBody,
	ContentItemKind,
	type InferenceUsage,
	InputKind,
	type InputOutputContent,
	type InputOutputContentItemUnion,
	type InputUnion,
	type ModelParam,
	OutputKind,
	type OutputUnion,
	type ReasoningContent,
	RoleEnum,
	Status,
	type ToolOutput,
	type ToolOutputItemUnion,
	ToolType,
} from '@/spec/inference';
import type { ProviderName } from '@/spec/modelpreset';
import { type ToolStoreChoice, ToolStoreChoiceType, type UIToolCall, type UIToolOutput } from '@/spec/tool';

import { generateTitle } from '@/lib/title_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

import { providerSetAPI } from '@/apis/baseapi';

import { uiAttachmentToConversation } from '@/chats/attachments/attachment_editor_utils';
import type { EditorSubmitPayload } from '@/chats/chat_input_editor';
import { formatToolOutputSummary } from '@/chats/tools/tool_editor_utils';

export function initConversation(title = 'New Conversation'): Conversation {
	return {
		schemaVersion: CONVERSATION_SCHEMA_VERSION,
		id: getUUIDv7(),
		title: generateTitle(title).title,
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: [],
	};
}

export function initConversationMessage(role: RoleEnum): ConversationMessage {
	const d = new Date();
	return {
		id: d.toISOString(),
		createdAt: new Date(),
		role: role,
		status: Status.None,
		uiContent: '',
	};
}

export function deriveConversationToolsFromMessages(messages: ConversationMessage[]): ToolStoreChoice[] {
	for (let i = messages.length - 1; i >= 0; i -= 1) {
		const m = messages[i];
		if (m.role === RoleEnum.User && m.toolStoreChoices && m.toolStoreChoices.length > 0) {
			return m.toolStoreChoices;
		}
	}
	return [];
}

export function buildUserConversationMessageFromEditor(
	payload: EditorSubmitPayload,
	existingId?: string
): ConversationMessage {
	const now = new Date();
	const id = existingId ?? now.toISOString();

	const text = payload.text.trim();
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
		status: Status.None,
		contents,
	};

	const inputs: InputUnion[] = [
		{
			kind: InputKind.InputMessage,
			inputMessage,
		},
	];
	// console.log('inputs from editor', JSON.stringify(inputs, null, 2));

	// Attach tool outputs as FunctionToolOutput / CustomToolOutput events.
	for (const ui of payload.toolOutputs) {
		const infOut = buildToolOutputFromEditor(ui);
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

	const toolOutputs = payload.toolOutputs.length > 0 ? payload.toolOutputs : undefined;

	const msg: ConversationMessage = {
		id,
		createdAt: now,
		role: RoleEnum.User,
		status: Status.None,
		inputs,
		attachments,
		toolStoreChoices,
		uiContent: text,
		uiToolOutputs: toolOutputs,
	};
	// console.log('msg from editor', JSON.stringify(msg, null, 2));

	return msg;
}

function buildToolOutputFromEditor(ui: UIToolOutput): ToolOutput | undefined {
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
		isError: !!ui.isError,
		signature: undefined,
		contents,
		webSearchToolOutputItems: undefined,
	};
}

/**
 * Hydrate a stored conversation (from Go) into the full UI Conversation shape.
 * - Derives message.content, reasoningContents, toolCalls, toolOutputs.
 */
export function hydrateConversation(store: StoreConversation): Conversation {
	const hydratedMessages: ConversationMessage[] = store.messages.map(
		(m: StoreConversationMessage): ConversationMessage => {
			const role = m.role;
			let uiContent = '';
			let uiReasoningContents = undefined;
			let uiToolCalls = undefined;
			let uiToolOutputs = undefined;

			const outputs: OutputUnion[] | undefined = m.outputs;
			const inputs: InputUnion[] | undefined = m.inputs;

			if (role === RoleEnum.Assistant) {
				// Text + tool calls + tool outputs from model outputs.
				const derived = deriveUIFieldsFromOutputUnion(outputs);
				uiContent = derived.uiContent;
				uiReasoningContents = derived.uiReasoningContents;
				uiToolCalls = derived.uiToolCalls;
				// Prefer outputs from model response; fall back to any tool outputs in inputs.
				uiToolOutputs =
					derived.uiToolOutputs && derived.uiToolOutputs.length > 0
						? derived.uiToolOutputs
						: deriveUIToolOutputsFromInputUnion(inputs);
			} else if (role === RoleEnum.User) {
				// User content from inputs, tool outputs from inputs.
				uiContent = deriveUIContentFromInputUnion(inputs);
				uiToolOutputs = deriveUIToolOutputsFromInputUnion(inputs);
			} else {
				// System / developer / other roles: We do not show them in UI text bubble.
			}

			const uiDebugDetails = getDebugDetailsMarkdown(m.debugDetails, m.error);
			return {
				...(m as any),
				uiContent,
				uiReasoningContents,
				uiToolCalls,
				uiToolOutputs,
				uiDebugDetails,
			} as ConversationMessage;
		}
	);

	const allMesages = {
		...(store as any),
		messages: hydratedMessages,
	} as Conversation;
	// console.log('hydrate out', JSON.stringify(allMesages, null, 2));
	return allMesages;
}

// Derive UI-level tool outputs from tool-output events in inputs (Function/Custom).
function deriveUIToolOutputsFromInputUnion(inputs: InputUnion[] | undefined): UIToolOutput[] {
	if (!inputs || inputs.length === 0) return [];

	const uiOutputs: UIToolOutput[] = [];

	for (const iu of inputs) {
		let out: ToolOutput | undefined;

		if (iu.kind === InputKind.FunctionToolOutput && iu.functionToolOutput) {
			out = iu.functionToolOutput;
		} else if (iu.kind === InputKind.CustomToolOutput && iu.customToolOutput) {
			out = iu.customToolOutput;
		} else {
			continue;
		}

		const isError = out.isError;

		// Aggregate text from contents into a single rawOutput string.
		let raw = '';
		if (out.contents && out.contents.length > 0) {
			const parts: string[] = [];
			for (const item of out.contents) {
				if (item.kind === ContentItemKind.Text && item.textItem?.text) {
					parts.push(item.textItem.text);
				}
			}
			raw = parts.join('\n\n');
		}

		const summaryBase = formatToolOutputSummary(out.name);
		const summary = isError && raw ? `Error: ${raw.split('\n')[0].slice(0, 80)}` : summaryBase;

		const ui: UIToolOutput = {
			id: out.id,
			callID: out.callID,
			name: out.name,
			choiceID: out.choiceID,
			// ToolType and ToolStoreChoiceType share the same string enum values.
			type: out.type as unknown as ToolStoreChoiceType,
			summary,
			rawOutput: raw,
			isError,
			errorMessage: isError && raw ? raw : undefined,
			toolStoreChoice: undefined,
			// We cannot reliably reconstruct the original ToolStoreChoice or arguments from here.
			arguments: undefined,
			webSearchToolCallItems: undefined,
		};

		uiOutputs.push(ui);
	}

	return uiOutputs;
}

/**
 * Rebuild user-visible content from InputUnion(s) for a message.
 * Used primarily for user turns, since `content` is not persisted.
 */
function deriveUIContentFromInputUnion(inputs?: InputUnion[]): string {
	if (!inputs || inputs.length === 0) return '';

	for (const iu of inputs) {
		// We want only the first user input content.
		// Attachments that are processed as text etc are not shown in UI.
		if (iu.kind !== InputKind.InputMessage || iu.inputMessage?.role != RoleEnum.User || !iu.inputMessage.contents)
			continue;

		for (const c of iu.inputMessage.contents) {
			if (c.kind === ContentItemKind.Text && c.textItem?.text) {
				const t = c.textItem.text.trim();
				if (t) return t;
			}
		}
	}

	return '';
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
		// console.log('history to completion', JSON.stringify(history, null, 2));

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
			const assistantMsg = buildAssistantMessageFromResponse(assistantPlaceholder.id, modelParams, resp);
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

	const inf = resp.inferenceResponse;
	const outputs = inf.outputs ?? [];
	const usage: InferenceUsage | undefined = inf.usage;
	const error = inf.error;

	const { uiContent, uiReasoningContents, uiToolCalls, uiToolOutputs } = deriveUIFieldsFromOutputUnion(outputs);
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

function deriveUIFieldsFromOutputUnion(outputs: OutputUnion[] | undefined): {
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
					// refusal/image/file content items are not rendered inline in assistant text bubble
					// if (c.kind === ContentItemKind.Refusal && c.refusalItem) {
					// 	const t = c.refusalItem.refusal.trim();
					// 	if (t) textParts.push(t);
					// }
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
				if (o.functionToolCall) toolCalls.push(o.functionToolCall as unknown as UIToolCall);
				if (o.customToolCall) toolCalls.push(o.customToolCall as unknown as UIToolCall);
				if (o.webSearchToolCall) toolCalls.push(o.webSearchToolCall as unknown as UIToolCall);
				break;

			case OutputKind.WebSearchToolOutput:
				if (o.webSearchToolOutput) {
					toolOutputs.push(o.webSearchToolOutput as unknown as UIToolOutput);
				}
				break;
		}
	}

	const content = textParts.join('\n\n');

	const o = {
		uiContent: content,
		uiReasoningContents: reasoning.length > 0 ? reasoning : undefined,
		uiToolCalls: toolCalls.length > 0 ? toolCalls : undefined,
		uiToolOutputs: toolOutputs.length > 0 ? toolOutputs : undefined,
	};
	// console.log('ui from out', JSON.stringify(o, null, 2));
	return o;
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

function getDebugDetailsMarkdown(debugObj?: any, errorObj?: any): string | undefined {
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

function getQuotedJSON(obj: any): string {
	return '```json\n' + JSON.stringify(obj, null, 2) + '\n```';
}
