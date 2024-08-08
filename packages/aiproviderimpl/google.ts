import {
	APIFetchResponse,
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
	googleProviderInfo,
} from 'aiprovidermodel';
import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { AIAPI } from './completion_provider';

interface Content {
	role: string;
	parts: { text: string }[];
}

export class GoogleAPI extends AIAPI {
	constructor() {
		super(googleProviderInfo);
	}

	async completion(
		input: CompletionRequest,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		// return tempCodeString;
		// let messages: ChatCompletionRequestMessage[] = [{"role": "user", "content": "Hello!"}];
		if (!input.messages) {
			throw Error('No input messages found');
		}
		const request = this.createRequestData(input);
		let modelpath = `${this.providerInfo.chatCompletionPathPrefix}/${input.model}:generateContent`;
		if (input.stream) {
			modelpath = `${this.providerInfo.chatCompletionPathPrefix}/${input.model}:streamGenerateContent`;
		}

		const requestConfig: AxiosRequestConfig = {
			url: modelpath,
			method: 'POST',
			data: request,
		};
		if (input.stream) {
			return this.handleStreamingResponse(requestConfig, onStreamData);
		} else {
			return this.handleDirectResponse(requestConfig);
		}
	}

	private convertMessages(messages: ChatCompletionRequestMessage[]): Content[] {
		return messages.map(message => {
			let role: string = 'user';
			switch (message.role) {
				case ChatCompletionRoleEnum.user:
					role = 'user';
					break;
				case ChatCompletionRoleEnum.assistant:
					role = 'model';
					break;
			}
			return {
				role,
				parts: [
					{
						text: message.content || '', // If content is undefined, use an empty string
					},
				],
			};
		});
	}

	private createRequestData(input: CompletionRequest): Record<string, any> {
		if (!input.messages) {
			throw Error('No input messages found');
		}
		const generateConfig: Record<string, any> = {
			maxOutputTokens: input.maxOutputLength,
			temperature: input.temperature,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				if (!Object.prototype.hasOwnProperty.call(generateConfig, key)) {
					generateConfig[key] = input.additionalParameters[key];
				}
			}
		}

		const content = this.convertMessages(input.messages);
		// eslint-disable-next-line @typescript-eslint/naming-convention
		const request: Record<string, any> = {
			contents: content,
			generationConfig: generateConfig,
		};
		return request;
	}

	private parseFullResponse<T>(apiFetchData: APIFetchResponse<T>): CompletionResponse {
		if (!apiFetchData) {
			log.error('No API fetch data');
			return {};
		}
		const completionResponse: CompletionResponse = {
			requestDetails: apiFetchData.requestDetails,
			responseDetails: apiFetchData.responseDetails,
			errorDetails: apiFetchData.errorDetails,
		};

		const data = apiFetchData.data;
		if (typeof data !== 'object' || data === null) {
			const msg = 'Invalid data response. Expected an object.';
			// log.error(msg);
			if (completionResponse.errorDetails) {
				completionResponse.errorDetails.message += msg;
			} else {
				completionResponse.errorDetails = { message: msg };
			}
			return completionResponse;
		}
		let respText = '';
		if ('candidates' in data && Array.isArray(data.candidates) && data.candidates.length > 0) {
			if (
				'content' in data.candidates[0] &&
				'parts' in data.candidates[0].content &&
				Array.isArray(data.candidates[0].content.parts) &&
				data.candidates[0].content.parts.length > 0 &&
				'text' in data.candidates[0].content.parts[0]
			) {
				respText = data.candidates[0].content.parts[0].text as string;
			}
		}
		completionResponse.respContent = respText;

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

	private getStreamDoneResponse(respText: string, functionName: string, functionArgs: any): CompletionResponse {
		const completionResponse: CompletionResponse = {
			respContent: respText,
			functionName: functionName,
			functionArgs: functionArgs,
		};
		return completionResponse;
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
			const objParser = new StreamParser();

			const dataChunkProcessor = async (dataString: string) => {
				try {
					const lines = dataString.split('\n'); // Split the buffer by newline

					for (const line of lines) {
						// log.debug('got line:', line);
						const res = objParser.parse(line);
						if (res) {
							// log.debug('got obj', JSON.stringify(res, null, 2));
							const partCompletion = this.parseFullResponse(res);
							if (!partCompletion) {
								// log.debug('Didnt get a completion');
								continue;
							}
							const respNew = partCompletion.respContent || '';
							respText += respNew;
							await onStreamData(respNew);
						}
						if (objParser.isStreamComplete() || line === 'data: [DONE]') {
							const r = this.getStreamDoneResponse(respText, '', '');
							resolve(r);
						}
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
class StreamParser {
	private buffer: string = '';
	private braceDepth: number = 0;
	private bracketDepth: number = 0;
	private objectStart: boolean = false;
	private streamComplete: boolean = false;

	public parse(line: string): object | null {
		for (const char of line) {
			if (char === '{') {
				this.braceDepth++;
				if (this.braceDepth === 1 && !this.objectStart) {
					this.objectStart = true; // Mark the start of a new object
					this.buffer = char; // Start buffering this object
				} else if (this.objectStart) {
					this.buffer += char; // Continue buffering the current object
				}
			} else if (char === '}') {
				if (this.objectStart) {
					this.buffer += char; // Continue buffering until end of object
				}
				this.braceDepth--;
				if (this.braceDepth === 0 && this.objectStart) {
					this.objectStart = false; // Mark the end of an object
					try {
						// log.info('Accumulated buffer for parse', this.buffer);
						const parsedObject = JSON.parse(this.buffer);
						this.buffer = ''; // Reset buffer for the next object
						return parsedObject;
					} catch (e) {
						// Handle JSON parse error if needed
						log.error('JSON parse error:', e);
						this.buffer = ''; // Reset buffer
						return null;
					}
				}
			} else if (char === '[') {
				this.bracketDepth++;
				if (this.bracketDepth > 1 || this.bracketDepth === 0) {
					this.buffer += char; // Buffer nested arrays or inner objects
				}
			} else if (char === ']') {
				if (this.bracketDepth > 1 || this.bracketDepth === 0) {
					this.buffer += char; // Buffer nested arrays or inner objects
				}
				this.bracketDepth--;
				if (this.bracketDepth === 0) {
					this.streamComplete = true; // Mark the end of the stream
				}
			} else if (this.objectStart || this.bracketDepth > 1) {
				this.buffer += char; // Continue buffering the current object or nested arrays
			}
		}

		return null; // Object not complete yet
	}

	public isStreamComplete(): boolean {
		return this.streamComplete;
	}
}
