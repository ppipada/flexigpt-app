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
		onStreamData?: (data: string) => void
	): Promise<CompletionResponse | undefined> {
		const callbackId = `stream-data-callback-${Date.now().toString()}-${Math.random().toString(36).substring(2, 9)}`;
		let prevData: string = '';
		if (onStreamData) {
			const cb = (data: string) => {
				if (data !== prevData) {
					prevData = data;
					onStreamData(data);
				}
			};
			EventsOn(callbackId, cb);
		}
		const response = await FetchCompletion(
			provider,
			prompt,
			modelParams as wailsSpec.ModelParams,
			prevMessages ? ([...prevMessages] as wailsSpec.ChatCompletionRequestMessage[]) : [],
			callbackId
		);
		return response.Body as CompletionResponse;
	}
}
