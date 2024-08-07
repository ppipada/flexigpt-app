import {
	ALL_MODEL_INFO,
	anthropicProviderInfo,
	CompletionRequest,
	CompletionResponse,
	ModelName,
} from 'aiprovidermodel';
import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { AIAPI } from './completion_provider';

export class AnthropicAPI extends AIAPI {
	constructor() {
		super(anthropicProviderInfo);
	}

	async completion(
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}
		const request = this.createRequestData(input);
		const requestConfig: AxiosRequestConfig = {
			url: this.providerInfo.chatCompletionPathPrefix,
			method: 'POST',
			data: request,
		};
		if (input.stream) {
			return this.handleStreamingResponse(requestConfig, onStreamData);
		} else {
			return this.handleDirectResponse(requestConfig);
		}
	}

	private createRequestData(input: CompletionRequest): Record<string, any> {
		let maxTokens = input.maxOutputLength;
		if (!maxTokens) {
			if (input.model in ALL_MODEL_INFO) {
				maxTokens = ALL_MODEL_INFO[input.model as ModelName].maxOutputLength;
			} else {
				maxTokens = 4096;
			}
		}
		const request: Record<string, any> = {
			model: input.model,
			messages: input.messages,
			// eslint-disable-next-line @typescript-eslint/naming-convention
			max_tokens: maxTokens,
			temperature: input.temperature,
			stream: input.stream || false,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				// eslint-disable-next-line no-prototype-builtins
				if (key === 'systemPrompt' && typeof key === 'string') {
					request['system'] = input.additionalParameters[key];
					continue;
				}
				if (!Object.prototype.hasOwnProperty.call(request, key)) {
					request[key] = input.additionalParameters[key];
				}
			}
		}
		return request;
	}

	private parseFullResponse(data: any): CompletionResponse | undefined {
		if (typeof data !== 'object' || data === null) {
			throw new Error('Invalid data response. Expected an object.');
		}
		let respText = '';
		if ('content' in data && Array.isArray(data.content) && data.content.length > 0) {
			for (const resp of data.content) {
				respText += resp.text + '\n';
			}
		}
		const completionResponse: CompletionResponse = { fullResponse: data, respContent: respText };
		return completionResponse;
	}

	private async handleDirectResponse(requestConfig: AxiosRequestConfig): Promise<CompletionResponse | undefined> {
		try {
			const data = await this.apicaller.request(requestConfig);
			return this.parseFullResponse(data);
		} catch (error) {
			log.error('Error in completion request: ' + error);
			throw error;
		}
	}

	private parseStreamingData(data: any): {
		parsedText: string;
		parsedFunctionName: string | null;
		parsedFunctionArgs: any | null;
	} {
		let parsedText = '';
		// eslint-disable-next-line prefer-const
		let parsedFunctionName: string | null = null;
		// eslint-disable-next-line prefer-const
		let parsedFunctionArgs: any | null = null;

		if (!('content_block' in data) && !('delta' in data)) {
			return { parsedText, parsedFunctionName, parsedFunctionArgs };
		}

		if ('content_block' in data && 'text' in data['content_block']) {
			parsedText = data['content_block']['text'];
			return { parsedText, parsedFunctionName, parsedFunctionArgs };
		}

		if ('delta' in data && 'text' in data['delta']) {
			parsedText = data['delta']['text'];
			return { parsedText, parsedFunctionName, parsedFunctionArgs };
		}

		return { parsedText, parsedFunctionName, parsedFunctionArgs };
	}

	private getStreamDoneResponse(respText: string, functionName: string, functionArgs: any): CompletionResponse {
		const completionResponse: CompletionResponse = {
			fullResponse: undefined,
			respContent: respText,
			functionName: functionName,
			functionArgs: functionArgs,
		};
		return completionResponse;
	}

	private parseStreamLine(respText: string, functionName: string, functionArgs: any, line: string) {
		let respNew = '';
		if (!line.startsWith('data: ')) {
			return {
				respNew: respNew,
				respFull: respText,
				fname: functionName,
				fargs: functionArgs,
				completeResponse: undefined,
			};
		}
		if (line === 'data: [DONE]') {
			const r = this.getStreamDoneResponse(respText, functionName, functionArgs);
			return { respNew: respNew, respFull: respText, fname: functionName, fargs: functionArgs, completeResponse: r };
		}

		line = line.substring(6).trim(); // Remove 'data: ' prefix
		try {
			const dataObject = JSON.parse(line);
			// log.debug('json data', dataObject);
			if (!dataObject || !('type' in dataObject)) {
				// Non content related data
				return {
					respNew: respNew,
					respFull: respText,
					fname: functionName,
					fargs: functionArgs,
					completeResponse: undefined,
				};
			}
			const msgType = dataObject['type'].trim();
			if (msgType === 'message_stop') {
				const r = this.getStreamDoneResponse(respText, functionName, functionArgs);
				return { respNew: respNew, respFull: respText, fname: functionName, fargs: functionArgs, completeResponse: r };
			}

			if (msgType !== 'content_block_delta' && msgType !== 'content_block_start') {
				return {
					respNew: respNew,
					respFull: respText,
					fname: functionName,
					fargs: functionArgs,
					completeResponse: undefined,
				};
			}

			const { parsedText, parsedFunctionName, parsedFunctionArgs } = this.parseStreamingData(dataObject);
			// log.debug('parsed data', parsedText);
			respNew = parsedText;
			respText += parsedText || '';
			functionName = parsedFunctionName || functionName;
			functionArgs = parsedFunctionArgs || functionArgs;
		} catch (e) {
			log.error('Error in line data', line, JSON.stringify(e));
		}
		return {
			respNew: respNew,
			respFull: respText,
			fname: functionName,
			fargs: functionArgs,
			completeResponse: undefined,
		};
	}

	private async handleStreamingResponse(
		requestConfig: AxiosRequestConfig,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		if (!onStreamData) {
			throw new Error('Need a stream data handler');
		}

		return new Promise((resolve, reject) => {
			let respText = '';
			let functionName = '';
			let functionArgs: any;
			let buffer = ''; // Buffer to hold incomplete lines

			const dataChunkProcessor = async (dataString: string) => {
				try {
					buffer += dataString; // Accumulate data in the buffer
					const lines = buffer.split('\n'); // Split the buffer by newline
					buffer = lines.pop() || ''; // Keep the last partial line in the buffer

					for (const line of lines) {
						const r = this.parseStreamLine(respText, functionArgs, functionName, line);
						if (r.completeResponse) {
							resolve(r.completeResponse);
							return;
						}
						respText = r.respFull;
						functionName = r.fname;
						functionArgs = r.fargs;
						await onStreamData(r.respNew);
					}
				} catch (e) {
					reject(e);
				}
			};
			try {
				this.apicaller.requestStream(requestConfig, dataChunkProcessor);
			} catch (error) {
				log.error('Error in streaming completion request: ' + error);
				reject(error);
			}
		});
	}
}
