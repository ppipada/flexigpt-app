export namespace frontend {
	
	export class FileFilter {
	    DisplayName: string;
	    Pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DisplayName = source["DisplayName"];
	        this.Pattern = source["Pattern"];
	    }
	}

}

export namespace spec {
	
	export class AISetting {
	    isEnabled: boolean;
	    apiKey: string;
	    defaultModel: string;
	    defaultTemperature: number;
	    defaultOrigin: string;
	    additionalSettings: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new AISetting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	        this.apiKey = source["apiKey"];
	        this.defaultModel = source["defaultModel"];
	        this.defaultTemperature = source["defaultTemperature"];
	        this.defaultOrigin = source["defaultOrigin"];
	        this.additionalSettings = source["additionalSettings"];
	    }
	}
	export class APIResponseDetails {
	    data: any;
	    status: number;
	    headers: {[key: string]: any};
	    requestDetails?: APIRequestDetails;
	
	    static createFrom(source: any = {}) {
	        return new APIResponseDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.status = source["status"];
	        this.headers = source["headers"];
	        this.requestDetails = this.convertValues(source["requestDetails"], APIRequestDetails);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class APIRequestDetails {
	    url?: string;
	    method?: string;
	    headers?: {[key: string]: any};
	    params?: {[key: string]: any};
	    data?: any;
	    timeout?: number;
	    curlCommand?: string;
	
	    static createFrom(source: any = {}) {
	        return new APIRequestDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.method = source["method"];
	        this.headers = source["headers"];
	        this.params = source["params"];
	        this.data = source["data"];
	        this.timeout = source["timeout"];
	        this.curlCommand = source["curlCommand"];
	    }
	}
	export class APIErrorDetails {
	    message: string;
	    requestDetails?: APIRequestDetails;
	    responseDetails?: APIResponseDetails;
	
	    static createFrom(source: any = {}) {
	        return new APIErrorDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.requestDetails = this.convertValues(source["requestDetails"], APIRequestDetails);
	        this.responseDetails = this.convertValues(source["responseDetails"], APIResponseDetails);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class ConversationMessage {
	    id: string;
	    // Go type: time
	    createdAt?: any;
	    role: string;
	    content: string;
	    timestamp?: string;
	    name?: string;
	    details?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversationMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = source["timestamp"];
	        this.name = source["name"];
	        this.details = source["details"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AddMessageToConversationRequestBody {
	    title: string;
	    newMessage: ConversationMessage;
	
	    static createFrom(source: any = {}) {
	        return new AddMessageToConversationRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.newMessage = this.convertValues(source["newMessage"], ConversationMessage);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AddMessageToConversationRequest {
	    ID: string;
	    Body?: AddMessageToConversationRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new AddMessageToConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Body = this.convertValues(source["Body"], AddMessageToConversationRequestBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class AddMessageToConversationResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new AddMessageToConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class AppSettings {
	    defaultProvider: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultProvider = source["defaultProvider"];
	    }
	}
	export class ChatCompletionFunctions {
	    name: string;
	    description?: string;
	    parameters?: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new ChatCompletionFunctions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.parameters = source["parameters"];
	    }
	}
	export class ChatCompletionRequestMessageFunctionCall {
	    name?: string;
	    arguments?: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatCompletionRequestMessageFunctionCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.arguments = source["arguments"];
	    }
	}
	export class ChatCompletionRequestMessage {
	    role: string;
	    content?: string;
	    name?: string;
	    functionCall?: ChatCompletionRequestMessageFunctionCall;
	
	    static createFrom(source: any = {}) {
	        return new ChatCompletionRequestMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.name = source["name"];
	        this.functionCall = this.convertValues(source["functionCall"], ChatCompletionRequestMessageFunctionCall);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CompletionRequest {
	    model: string;
	    messages?: ChatCompletionRequestMessage[];
	    temperature: number;
	    maxPromptLength: number;
	    stream: boolean;
	    systemPrompt?: string;
	    maxOutputLength?: number;
	    functions?: ChatCompletionFunctions[];
	    functionCall?: any;
	    suffix?: string;
	    timeout?: number;
	    additionalParameters?: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new CompletionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.messages = this.convertValues(source["messages"], ChatCompletionRequestMessage);
	        this.temperature = source["temperature"];
	        this.maxPromptLength = source["maxPromptLength"];
	        this.stream = source["stream"];
	        this.systemPrompt = source["systemPrompt"];
	        this.maxOutputLength = source["maxOutputLength"];
	        this.functions = this.convertValues(source["functions"], ChatCompletionFunctions);
	        this.functionCall = source["functionCall"];
	        this.suffix = source["suffix"];
	        this.timeout = source["timeout"];
	        this.additionalParameters = source["additionalParameters"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CompletionResponse {
	    requestDetails?: APIRequestDetails;
	    responseDetails?: APIResponseDetails;
	    errorDetails?: APIErrorDetails;
	    respContent?: string;
	    functionName?: string;
	    functionArgs?: any;
	
	    static createFrom(source: any = {}) {
	        return new CompletionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestDetails = this.convertValues(source["requestDetails"], APIRequestDetails);
	        this.responseDetails = this.convertValues(source["responseDetails"], APIResponseDetails);
	        this.errorDetails = this.convertValues(source["errorDetails"], APIErrorDetails);
	        this.respContent = source["respContent"];
	        this.functionName = source["functionName"];
	        this.functionArgs = source["functionArgs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Conversation {
	    id: string;
	    title: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    messages: ConversationMessage[];
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
	        this.messages = this.convertValues(source["messages"], ConversationMessage);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConversationItem {
	    id: string;
	    title: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ConversationItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class DeleteConversationRequest {
	    ID: string;
	    Title: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Title = source["Title"];
	    }
	}
	export class DeleteConversationResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class FetchCompletionResponse {
	    Body?: CompletionResponse;
	
	    static createFrom(source: any = {}) {
	        return new FetchCompletionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], CompletionResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetAllSettingsRequest {
	    ForceFetch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GetAllSettingsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ForceFetch = source["ForceFetch"];
	    }
	}
	export class SettingsSchema {
	    aiSettings: {[key: string]: AISetting};
	    app: AppSettings;
	
	    static createFrom(source: any = {}) {
	        return new SettingsSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.aiSettings = this.convertValues(source["aiSettings"], AISetting, true);
	        this.app = this.convertValues(source["app"], AppSettings);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetAllSettingsResponse {
	    Body?: SettingsSchema;
	
	    static createFrom(source: any = {}) {
	        return new GetAllSettingsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], SettingsSchema);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetConfigurationInfoRequest {
	
	
	    static createFrom(source: any = {}) {
	        return new GetConfigurationInfoRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class GetConfigurationInfoResponseBody {
	    configurationInfo: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new GetConfigurationInfoResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configurationInfo = source["configurationInfo"];
	    }
	}
	export class GetConfigurationInfoResponse {
	    Body?: GetConfigurationInfoResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new GetConfigurationInfoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], GetConfigurationInfoResponseBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class GetConversationRequest {
	    ID: string;
	    Title: string;
	
	    static createFrom(source: any = {}) {
	        return new GetConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Title = source["Title"];
	    }
	}
	export class GetConversationResponse {
	    Body?: Conversation;
	
	    static createFrom(source: any = {}) {
	        return new GetConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], Conversation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetDefaultProviderRequest {
	
	
	    static createFrom(source: any = {}) {
	        return new GetDefaultProviderRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class GetDefaultProviderResponseBody {
	    defaultProvider: string;
	
	    static createFrom(source: any = {}) {
	        return new GetDefaultProviderResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultProvider = source["defaultProvider"];
	    }
	}
	export class GetDefaultProviderResponse {
	    Body?: GetDefaultProviderResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new GetDefaultProviderResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], GetDefaultProviderResponseBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ListConversationsRequest {
	    Token: string;
	
	    static createFrom(source: any = {}) {
	        return new ListConversationsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Token = source["Token"];
	    }
	}
	export class ListConversationsResponseBody {
	    conversationItems: ConversationItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListConversationsResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conversationItems = this.convertValues(source["conversationItems"], ConversationItem);
	        this.nextPageToken = source["nextPageToken"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ListConversationsResponse {
	    Body?: ListConversationsResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new ListConversationsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], ListConversationsResponseBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MakeCompletionRequestBody {
	    prompt: string;
	    prevMessages: ChatCompletionRequestMessage[];
	    inputParams: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new MakeCompletionRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.prevMessages = this.convertValues(source["prevMessages"], ChatCompletionRequestMessage);
	        this.inputParams = source["inputParams"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MakeCompletionRequest {
	    Provider: string;
	    Body?: MakeCompletionRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new MakeCompletionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = source["Provider"];
	        this.Body = this.convertValues(source["Body"], MakeCompletionRequestBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MakeCompletionResponse {
	    Body?: CompletionRequest;
	
	    static createFrom(source: any = {}) {
	        return new MakeCompletionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], CompletionRequest);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveConversationRequest {
	    Body?: Conversation;
	
	    static createFrom(source: any = {}) {
	        return new SaveConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], Conversation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SaveConversationResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SaveConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SetDefaultProviderRequestBody {
	    provider: string;
	
	    static createFrom(source: any = {}) {
	        return new SetDefaultProviderRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	    }
	}
	export class SetDefaultProviderRequest {
	    Body?: SetDefaultProviderRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetDefaultProviderRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], SetDefaultProviderRequestBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SetDefaultProviderResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetDefaultProviderResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SetProviderAttributeRequestBody {
	    apiKey?: string;
	    defaultModel?: string;
	    defaultTemperature?: number;
	    defaultOrigin?: string;
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAttributeRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	        this.defaultModel = source["defaultModel"];
	        this.defaultTemperature = source["defaultTemperature"];
	        this.defaultOrigin = source["defaultOrigin"];
	    }
	}
	export class SetProviderAttributeRequest {
	    Provider: string;
	    Body?: SetProviderAttributeRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAttributeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = source["Provider"];
	        this.Body = this.convertValues(source["Body"], SetProviderAttributeRequestBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SetProviderAttributeResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAttributeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SetSettingRequestBody {
	    value: any;
	
	    static createFrom(source: any = {}) {
	        return new SetSettingRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	    }
	}
	export class SetSettingRequest {
	    Key: string;
	    Body?: SetSettingRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetSettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Body = this.convertValues(source["Body"], SetSettingRequestBody);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class SetSettingResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetSettingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

