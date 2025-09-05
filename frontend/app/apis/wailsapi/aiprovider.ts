import type {
	ChatCompletionDataMessage,
	CompletionData,
	CompletionResponse,
	IProviderSetAPI,
	ModelParams,
} from '@/spec/aiprovider';
import type { ProviderName } from '@/spec/modelpreset';

import { BuildCompletionData, CancelCompletion, FetchCompletion } from '@/apis/wailsjs/go/main/ProviderSetWrapper';
import type { spec as wailsSpec } from '@/apis/wailsjs/go/models';
import { EventsOff, EventsOn } from '@/apis/wailsjs/runtime/runtime';

/**
 * @public
 */
export class WailsProviderSetAPI implements IProviderSetAPI {
	async buildCompletionData(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionDataMessage>
	): Promise<CompletionData> {
		const req = {
			Provider: provider,
			Body: {
				prompt: prompt,
				modelParams: modelParams as wailsSpec.ModelParams,
				prevMessages: prevMessages ? ([...prevMessages] as wailsSpec.ChatCompletionDataMessage[]) : [],
			} as wailsSpec.BuildCompletionDataRequestBody,
		};

		const resp = await BuildCompletionData(req as wailsSpec.BuildCompletionDataRequest);
		const respBody = resp.Body as wailsSpec.CompletionData;
		// console.log(JSON.stringify(respBody, undefined, 2));
		return respBody as CompletionData;
	}

	// Need an eventflow for getting completion.
	// Implemented that in main App Wrapper than aiprovider go package.
	// Wrapper redirects to providerSet after doing event handling
	async completion(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: ChatCompletionDataMessage[],
		requestId?: string,
		signal?: AbortSignal,
		onStreamTextData?: (text: string) => void,
		onStreamThinkingData?: (text: string) => void
	): Promise<CompletionResponse | undefined> {
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
			prompt,
			modelParams as wailsSpec.ModelParams,
			prevMessages ? ([...prevMessages] as wailsSpec.ChatCompletionDataMessage[]) : [],
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
			const respBody = resp.Body as wailsSpec.CompletionResponse;
			// console.log(JSON.stringify(respBody, undefined, 2));
			return respBody as CompletionResponse;
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
