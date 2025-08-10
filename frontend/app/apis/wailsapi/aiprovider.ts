import type { ChatCompletionRequestMessage, CompletionResponse, IProviderSetAPI, ModelParams } from '@/spec/aiprovider';
import type { ProviderName } from '@/spec/modelpreset';

import { FetchCompletion } from '@/apis/wailsjs/go/main/ProviderSetWrapper';
import type { spec as wailsSpec } from '@/apis/wailsjs/go/models';
import { EventsOn } from '@/apis/wailsjs/runtime/runtime';

/**
 * @public
 */
export class WailsProviderSetAPI implements IProviderSetAPI {
	// Need an eventflow for getting completion.
	// Implemented that in main App Wrapper than aiprovider go package.
	// Wrapper redirects to providerSet after doing event handling
	async completion(
		provider: ProviderName,
		prompt: string,
		modelParams: ModelParams,
		prevMessages?: Array<ChatCompletionRequestMessage>,
		onStreamTextData?: (textData: string) => void,
		onStreamThinkingData?: (thinkingData: string) => void
	): Promise<CompletionResponse | undefined> {
		let textCallbackId: string = '';
		let thinkingCallbackId: string = '';

		if (onStreamTextData && onStreamThinkingData) {
			let prevTextData: string = '';
			textCallbackId = `text-${Date.now().toString()}-${Math.random().toString(36).substring(2, 9)}`;
			const cbText = (textData: string) => {
				const d = textData.trim();
				if (d !== prevTextData) {
					prevTextData = d;
					onStreamTextData(d);
				}
			};
			EventsOn(textCallbackId, cbText);

			let prevThinkingData: string = '';
			thinkingCallbackId = `thinking-${Date.now().toString()}-${Math.random().toString(36).substring(2, 9)}`;
			const cbThinking = (thinkingData: string) => {
				const d = thinkingData.trim();
				if (d !== prevThinkingData) {
					prevThinkingData = d;
					onStreamThinkingData(d);
				}
			};
			EventsOn(thinkingCallbackId, cbThinking);
		}

		const response = await FetchCompletion(
			provider,
			prompt,
			modelParams as wailsSpec.ModelParams,
			prevMessages ? ([...prevMessages] as wailsSpec.ChatCompletionRequestMessage[]) : [],
			textCallbackId,
			thinkingCallbackId
		);
		return response.Body as CompletionResponse;
	}
}
