import {
	ChatCompletionRequestMessage,
	ChatCompletionRoleEnum,
	CompletionRequest,
	CompletionResponse,
	openaiProviderInfo,
} from 'aiprovidermodel';
import { AxiosRequestConfig } from 'axios';
import { log } from 'logger';
import { AIAPI } from './completion_provider';
import { unescapeChars } from './provider_utils';

export class OpenAIAPI extends AIAPI {
	constructor() {
		super(openaiProviderInfo);
	}

	async completion(
		input: CompletionRequest,
		onStreamData?: (data: string) => Promise<void>
	): Promise<CompletionResponse | undefined> {
		if (!input.messages) {
			throw Error('No input messages found');
		}

		const requestData = this.createRequestData(input);
		const requestConfig: AxiosRequestConfig = {
			url: this.providerInfo.chatCompletionPathPrefix,
			method: 'POST',
			data: requestData,
		};

		if (input.stream) {
			return this.handleStreamingResponse(requestConfig, onStreamData);
		} else {
			return this.handleDirectResponse(requestConfig);
		}
	}

	private createRequestData(input: CompletionRequest): Record<string, any> {
		const request: Record<string, any> = {
			model: input.model,
			max_tokens: input.maxOutputLength,
			temperature: input.temperature,
			stream: input.stream || false,
			messages: input.messages,
		};

		if (input.additionalParameters) {
			for (const key in input.additionalParameters) {
				if (key !== 'systemPrompt' && !Object.prototype.hasOwnProperty.call(request, key)) {
					request[key] = input.additionalParameters[key];
				}
			}

			if (input.additionalParameters.systemPrompt) {
				const systemMessage: ChatCompletionRequestMessage = {
					role: ChatCompletionRoleEnum.system,
					content: input.additionalParameters.systemPrompt,
				};
				request.messages.unshift(systemMessage);
			}
		}

		return request;
	}

	private parseFullResponse(data: any): CompletionResponse | undefined {
		if (typeof data !== 'object' || data === null) {
			throw new Error('Invalid data response. Expected an object.' + data);
		}

		let respText = '';
		let functionName = '';
		let functionArgs: any;

		if ('choices' in data && Array.isArray(data.choices) && data.choices.length > 0) {
			const responseMessage = data.choices[0].message;
			respText = responseMessage?.content ? (responseMessage?.content as string) : '';

			if (
				'tool_calls' in responseMessage &&
				responseMessage['tool_calls'].length > 0 &&
				'function' in responseMessage['tool_calls'][0]
			) {
				functionName = responseMessage['tool_calls'][0]['function']['name'];
				respText += '\nFunction call:\nName:' + functionName;
				try {
					functionArgs = JSON.parse(unescapeChars(responseMessage['tool_calls'][0]['function']['arguments']));
				} catch (error) {
					log.error(
						'Error parsing function call arguments: ' +
							error +
							' ' +
							responseMessage['tool_calls'][0]['function']['arguments']
					);
					respText += '\nError in parsing returned args\n';
					functionArgs = responseMessage['tool_calls'][0]['function']['arguments'];
				}
				respText += '\nArgs: ' + JSON.stringify(functionArgs, null, 2);
			}
		}

		const completionResponse: CompletionResponse = {
			fullResponse: data,
			respContent: respText,
			functionName: functionName,
			functionArgs: functionArgs,
		};
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
		let parsedFunctionName: string | null = null;
		let parsedFunctionArgs: any | null = null;

		if ('choices' in data && Array.isArray(data.choices) && data.choices.length > 0) {
			const responseMessage = data.choices[0].delta;
			parsedText = responseMessage?.content ? (responseMessage?.content as string) : '';

			if (
				'tool_calls' in responseMessage &&
				responseMessage['tool_calls'].length > 0 &&
				'function' in responseMessage['tool_calls'][0]
			) {
				parsedFunctionName = responseMessage['tool_calls'][0]['function']['name'];
				try {
					parsedFunctionArgs = JSON.parse(unescapeChars(responseMessage['tool_calls'][0]['function']['arguments']));
				} catch (error) {
					log.error(
						'Error parsing function call arguments: ' +
							error +
							' ' +
							responseMessage['tool_calls'][0]['function']['arguments']
					);
				}
			}
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

		line = line.substring(6).trim(); // Remove 'data: ' prefix
		try {
			if (line.trim().toUpperCase() === '[DONE]') {
				const r = this.getStreamDoneResponse(respText, functionName, functionArgs);
				return { respNew: respNew, respFull: respText, fname: functionName, fargs: functionArgs, completeResponse: r };
			}
			// log.debug('line data', line);
			const dataObject = JSON.parse(line);
			// log.debug('json data', dataObject);
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
