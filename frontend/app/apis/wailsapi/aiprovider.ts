import type {
	ChatCompletionDataMessage,
	FetchCompletionData,
	FetchCompletionResponseBody,
	IProviderSetAPI,
	ModelParams,
} from '@/spec/aiprovider';
import type { Attachment } from '@/spec/attachment';
import type { ProviderName } from '@/spec/modelpreset';
import type { ToolChoice } from '@/spec/tool';

import { BuildCompletionData, CancelCompletion, FetchCompletion } from '@/apis/wailsjs/go/main/ProviderSetWrapper';
import type { attachment as wailsAttachment, spec as wailsSpec } from '@/apis/wailsjs/go/models';
import { EventsOff, EventsOn } from '@/apis/wailsjs/runtime/runtime';

/**
 * @public
 */
export class WailsProviderSetAPI implements IProviderSetAPI {
	async buildCompletionData(
		provider: ProviderName,
		modelParams: ModelParams,
		currentMessage: ChatCompletionDataMessage,
		prevMessages?: Array<ChatCompletionDataMessage>,
		toolChoices?: Array<ToolChoice>,
		attachments?: Array<Attachment>
	): Promise<FetchCompletionData> {
		const req = {
			Provider: provider,
			Body: {
				modelParams: modelParams as wailsSpec.ModelParams,
				currentMessage: currentMessage as wailsSpec.ChatCompletionDataMessage,
				prevMessages: prevMessages ? ([...prevMessages] as wailsSpec.ChatCompletionDataMessage[]) : [],
				toolChoices: toolChoices ? ([...toolChoices] as wailsSpec.ToolChoice[]) : [],
				attachments: attachments ? ([...attachments] as wailsAttachment.Attachment[]) : [],
			} as wailsSpec.BuildCompletionDataRequestBody,
		};

		const resp = await BuildCompletionData(req as wailsSpec.BuildCompletionDataRequest);
		const respBody = resp.Body as wailsSpec.FetchCompletionData;
		// console.log(JSON.stringify(respBody, undefined, 2));
		return respBody as FetchCompletionData;
	}

	// Need an eventflow for getting completion.
	// Implemented that in main App Wrapper than aiprovider go package.
	// Wrapper redirects to providerSet after doing event handling
	async completion(
		provider: ProviderName,
		completionData: FetchCompletionData,
		requestId?: string,
		signal?: AbortSignal,
		onStreamTextData?: (text: string) => void,
		onStreamThinkingData?: (text: string) => void
	): Promise<FetchCompletionResponseBody | undefined> {
		let textCallbackId = '';
		let thinkingCallbackId = '';

		if (onStreamTextData && onStreamThinkingData) {
			const uid = requestId ?? `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

			textCallbackId = `text-${uid}`;
			thinkingCallbackId = `thinking-${uid}`;

			let lastText = '';
			const textCb = (t: string) => {
				const d = t.trim();
				if (d !== lastText) {
					lastText = d;
					onStreamTextData(d);
				}
			};
			EventsOn(textCallbackId, textCb);

			let lastThinking = '';
			const thinkingCb = (t: string) => {
				const d = t.trim();
				if (d !== lastThinking) {
					lastThinking = d;
					onStreamThinkingData(d);
				}
			};
			EventsOn(thinkingCallbackId, thinkingCb);
		}

		const responsePromise = FetchCompletion(
			provider,
			completionData as wailsSpec.FetchCompletionData,
			textCallbackId,
			thinkingCallbackId,
			requestId ?? ''
		);

		const abortPromise: Promise<never> = new Promise((_, reject) => {
			// Already aborted before we even start?
			if (signal?.aborted) {
				reject(new DOMException('Aborted', 'AbortError'));
				return;
			}

			const abortHandler = () => {
				/* Detach server-side and local listeners */
				this.cancelCompletion(requestId ?? '').catch(() => {});
				EventsOff(textCallbackId);
				EventsOff(thinkingCallbackId);

				reject(new DOMException('Aborted', 'AbortError'));
			};

			signal?.addEventListener('abort', abortHandler, { once: true });
		});

		try {
			const resp = await Promise.race([responsePromise, abortPromise]);
			const respBody = resp.Body as wailsSpec.FetchCompletionResponseBody;
			// console.log(JSON.stringify(respBody, undefined, 2));
			return respBody as FetchCompletionResponseBody;
		} finally {
			/* Always clean up – even when the race rejected with AbortError */
			if (textCallbackId) EventsOff(textCallbackId);
			if (thinkingCallbackId) EventsOff(thinkingCallbackId);
		}
	}

	async cancelCompletion(requestId: string): Promise<void> {
		if (!requestId) return;
		try {
			await CancelCompletion(requestId);
		} catch {
			/* Swallow any Go-side error; we only care that the signal aborts */
		}
	}
}
