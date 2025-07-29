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

export namespace settingstore {
	
	export class AISetting {
	    isEnabled: boolean;
	    apiKey: string;
	    origin: string;
	    chatCompletionPathPrefix: string;
	
	    static createFrom(source: any = {}) {
	        return new AISetting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	        this.apiKey = source["apiKey"];
	        this.origin = source["origin"];
	        this.chatCompletionPathPrefix = source["chatCompletionPathPrefix"];
	    }
	}
	export class AddAISettingRequest {
	    ProviderName: string;
	    Body?: AISetting;
	
	    static createFrom(source: any = {}) {
	        return new AddAISettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.Body = this.convertValues(source["Body"], AISetting);
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
	export class AddAISettingResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new AddAISettingResponse(source);
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
	export class DeleteAISettingRequest {
	    ProviderName: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteAISettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	    }
	}
	export class DeleteAISettingResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteAISettingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
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
	    version: string;
	    aiSettings: Record<string, AISetting>;
	    app: AppSettings;
	
	    static createFrom(source: any = {}) {
	        return new SettingsSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
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
	export class SetAISettingAPIKeyRequestBody {
	    apiKey: string;
	
	    static createFrom(source: any = {}) {
	        return new SetAISettingAPIKeyRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	    }
	}
	export class SetAISettingAPIKeyRequest {
	    ProviderName: string;
	    Body?: SetAISettingAPIKeyRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetAISettingAPIKeyRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.Body = this.convertValues(source["Body"], SetAISettingAPIKeyRequestBody);
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
	
	export class SetAISettingAPIKeyResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetAISettingAPIKeyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SetAISettingAttrsRequestBody {
	    isEnabled?: boolean;
	    origin?: string;
	    chatCompletionPathPrefix?: string;
	
	    static createFrom(source: any = {}) {
	        return new SetAISettingAttrsRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	        this.origin = source["origin"];
	        this.chatCompletionPathPrefix = source["chatCompletionPathPrefix"];
	    }
	}
	export class SetAISettingAttrsRequest {
	    ProviderName: string;
	    Body?: SetAISettingAttrsRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetAISettingAttrsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.Body = this.convertValues(source["Body"], SetAISettingAttrsRequestBody);
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
	
	export class SetAISettingAttrsResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetAISettingAttrsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SetAppSettingsRequest {
	    Body?: AppSettings;
	
	    static createFrom(source: any = {}) {
	        return new SetAppSettingsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], AppSettings);
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
	export class SetAppSettingsResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetAppSettingsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace spec {
	
	export class APIResponseDetails {
	    data: any;
	    status: number;
	    headers: Record<string, any>;
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
	    headers?: Record<string, any>;
	    params?: Record<string, any>;
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
	
	
	export class ReasoningParams {
	    type: string;
	    level: string;
	    tokens: number;
	
	    static createFrom(source: any = {}) {
	        return new ReasoningParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.level = source["level"];
	        this.tokens = source["tokens"];
	    }
	}
	export class ModelPreset {
	    schemaVersion: string;
	    id: string;
	    name: string;
	    displayName: string;
	    shortCommand: string;
	    isEnabled: boolean;
	    stream?: boolean;
	    maxPromptLength?: number;
	    maxOutputLength?: number;
	    temperature?: number;
	    reasoning?: ReasoningParams;
	    systemPrompt?: string;
	    timeout?: number;
	    additionalParametersRawJSON?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    isBuiltIn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelPreset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.shortCommand = source["shortCommand"];
	        this.isEnabled = source["isEnabled"];
	        this.stream = source["stream"];
	        this.maxPromptLength = source["maxPromptLength"];
	        this.maxOutputLength = source["maxOutputLength"];
	        this.temperature = source["temperature"];
	        this.reasoning = this.convertValues(source["reasoning"], ReasoningParams);
	        this.systemPrompt = source["systemPrompt"];
	        this.timeout = source["timeout"];
	        this.additionalParametersRawJSON = source["additionalParametersRawJSON"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
	        this.isBuiltIn = source["isBuiltIn"];
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
	export class AddModelPresetRequest {
	    ProviderName: string;
	    ModelPresetID: string;
	    Body?: ModelPreset;
	
	    static createFrom(source: any = {}) {
	        return new AddModelPresetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.ModelPresetID = source["ModelPresetID"];
	        this.Body = this.convertValues(source["Body"], ModelPreset);
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
	export class AddModelPresetResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new AddModelPresetResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class AddProviderRequestBody {
	    apiKey: string;
	    origin: string;
	    chatCompletionPathPrefix: string;
	
	    static createFrom(source: any = {}) {
	        return new AddProviderRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	        this.origin = source["origin"];
	        this.chatCompletionPathPrefix = source["chatCompletionPathPrefix"];
	    }
	}
	export class AddProviderRequest {
	    Provider: string;
	    Body?: AddProviderRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new AddProviderRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = source["Provider"];
	        this.Body = this.convertValues(source["Body"], AddProviderRequestBody);
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
	
	export class AddProviderResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new AddProviderResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
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
	export class ConversationMessage {
	    id: string;
	    // Go type: time
	    createdAt?: any;
	    role: string;
	    content: string;
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
	export class ConversationListItem {
	    id: string;
	    sanatizedTitle: string;
	    // Go type: time
	    modifiedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new ConversationListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sanatizedTitle = source["sanatizedTitle"];
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
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
	
	export class ProviderPreset {
	    schemaVersion: string;
	    name: string;
	    displayName: string;
	    apiType: string;
	    isEnabled: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    isBuiltIn: boolean;
	    origin: string;
	    chatCompletionPathPrefix: string;
	    apiKeyHeaderKey: string;
	    defaultHeaders: Record<string, string>;
	    defaultModelPresetID: string;
	    modelPresets: Record<string, ModelPreset>;
	
	    static createFrom(source: any = {}) {
	        return new ProviderPreset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.apiType = source["apiType"];
	        this.isEnabled = source["isEnabled"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
	        this.isBuiltIn = source["isBuiltIn"];
	        this.origin = source["origin"];
	        this.chatCompletionPathPrefix = source["chatCompletionPathPrefix"];
	        this.apiKeyHeaderKey = source["apiKeyHeaderKey"];
	        this.defaultHeaders = source["defaultHeaders"];
	        this.defaultModelPresetID = source["defaultModelPresetID"];
	        this.modelPresets = this.convertValues(source["modelPresets"], ModelPreset, true);
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
	export class CreateProviderPresetRequest {
	    ProviderName: string;
	    Body?: ProviderPreset;
	
	    static createFrom(source: any = {}) {
	        return new CreateProviderPresetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.Body = this.convertValues(source["Body"], ProviderPreset);
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
	export class CreateProviderPresetResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new CreateProviderPresetResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
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
	export class DeleteModelPresetRequest {
	    ProviderName: string;
	    ModelPresetID: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteModelPresetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.ModelPresetID = source["ModelPresetID"];
	    }
	}
	export class DeleteModelPresetResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteModelPresetResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class DeletePromptBundleRequest {
	    BundleID: string;
	
	    static createFrom(source: any = {}) {
	        return new DeletePromptBundleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	    }
	}
	export class DeletePromptBundleResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeletePromptBundleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class DeletePromptTemplateRequest {
	    BundleID: string;
	    TemplateSlug: string;
	    Version: string;
	
	    static createFrom(source: any = {}) {
	        return new DeletePromptTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.TemplateSlug = source["TemplateSlug"];
	        this.Version = source["Version"];
	    }
	}
	export class DeletePromptTemplateResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeletePromptTemplateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class DeleteProviderPresetRequest {
	    ProviderName: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteProviderPresetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	    }
	}
	export class DeleteProviderPresetResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteProviderPresetResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class DeleteProviderRequest {
	    Provider: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteProviderRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = source["Provider"];
	    }
	}
	export class DeleteProviderResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteProviderResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class DeleteToolBundleRequest {
	    BundleID: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteToolBundleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	    }
	}
	export class DeleteToolBundleResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteToolBundleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class DeleteToolRequest {
	    BundleID: string;
	    ToolSlug: string;
	    Version: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteToolRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.ToolSlug = source["ToolSlug"];
	        this.Version = source["Version"];
	    }
	}
	export class DeleteToolResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new DeleteToolResponse(source);
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
	export class GetAllModelPresetsRequest {
	    ForceFetch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GetAllModelPresetsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ForceFetch = source["ForceFetch"];
	    }
	}
	export class PresetsSchema {
	    version: string;
	    providerPresets: Record<string, ProviderPreset>;
	
	    static createFrom(source: any = {}) {
	        return new PresetsSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.providerPresets = this.convertValues(source["providerPresets"], ProviderPreset, true);
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
	export class GetAllModelPresetsResponse {
	    Body?: PresetsSchema;
	
	    static createFrom(source: any = {}) {
	        return new GetAllModelPresetsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], PresetsSchema);
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
	export class ProviderParams {
	    name: string;
	    apiKey: string;
	    origin: string;
	    chatCompletionPathPrefix: string;
	    apiKeyHeaderKey: string;
	    defaultHeaders: Record<string, string>;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.apiKey = source["apiKey"];
	        this.origin = source["origin"];
	        this.chatCompletionPathPrefix = source["chatCompletionPathPrefix"];
	        this.apiKeyHeaderKey = source["apiKeyHeaderKey"];
	        this.defaultHeaders = source["defaultHeaders"];
	        this.type = source["type"];
	    }
	}
	export class GetConfigurationInfoResponseBody {
	    defaultProvider: string;
	    configuredProviders: ProviderParams[];
	    inbuiltProviderModels: Record<string, ProviderPreset>;
	
	    static createFrom(source: any = {}) {
	        return new GetConfigurationInfoResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultProvider = source["defaultProvider"];
	        this.configuredProviders = this.convertValues(source["configuredProviders"], ProviderParams);
	        this.inbuiltProviderModels = this.convertValues(source["inbuiltProviderModels"], ProviderPreset, true);
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
	export class GetPromptTemplateRequest {
	    BundleID: string;
	    TemplateSlug: string;
	    Version: string;
	
	    static createFrom(source: any = {}) {
	        return new GetPromptTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.TemplateSlug = source["TemplateSlug"];
	        this.Version = source["Version"];
	    }
	}
	export class PreProcessorCall {
	    id: string;
	    toolID: string;
	    args?: Record<string, any>;
	    saveAs: string;
	    pathExpr?: string;
	    onError?: string;
	
	    static createFrom(source: any = {}) {
	        return new PreProcessorCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.toolID = source["toolID"];
	        this.args = source["args"];
	        this.saveAs = source["saveAs"];
	        this.pathExpr = source["pathExpr"];
	        this.onError = source["onError"];
	    }
	}
	export class PromptVariable {
	    name: string;
	    type: string;
	    required: boolean;
	    source: string;
	    description?: string;
	    staticVal?: string;
	    toolID?: string;
	    enumValues?: string[];
	    default?: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptVariable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.required = source["required"];
	        this.source = source["source"];
	        this.description = source["description"];
	        this.staticVal = source["staticVal"];
	        this.toolID = source["toolID"];
	        this.enumValues = source["enumValues"];
	        this.default = source["default"];
	    }
	}
	export class MessageBlock {
	    id: string;
	    role: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new MessageBlock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	    }
	}
	export class PromptTemplate {
	    schemaVersion: string;
	    id: string;
	    slug: string;
	    isEnabled: boolean;
	    displayName: string;
	    description?: string;
	    tags?: string[];
	    blocks: MessageBlock[];
	    variables?: PromptVariable[];
	    preProcessors?: PreProcessorCall[];
	    version: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    isBuiltIn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.slug = source["slug"];
	        this.isEnabled = source["isEnabled"];
	        this.displayName = source["displayName"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.blocks = this.convertValues(source["blocks"], MessageBlock);
	        this.variables = this.convertValues(source["variables"], PromptVariable);
	        this.preProcessors = this.convertValues(source["preProcessors"], PreProcessorCall);
	        this.version = source["version"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
	        this.isBuiltIn = source["isBuiltIn"];
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
	export class GetPromptTemplateResponse {
	    Body?: PromptTemplate;
	
	    static createFrom(source: any = {}) {
	        return new GetPromptTemplateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], PromptTemplate);
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
	export class GetToolRequest {
	    BundleID: string;
	    ToolSlug: string;
	    Version: string;
	
	    static createFrom(source: any = {}) {
	        return new GetToolRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.ToolSlug = source["ToolSlug"];
	        this.Version = source["Version"];
	    }
	}
	export class HTTPResponse {
	    successCodes?: number[];
	    encoding?: string;
	    selector?: string;
	    errorMode?: string;
	
	    static createFrom(source: any = {}) {
	        return new HTTPResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.successCodes = source["successCodes"];
	        this.encoding = source["encoding"];
	        this.selector = source["selector"];
	        this.errorMode = source["errorMode"];
	    }
	}
	export class HTTPAuth {
	    type: string;
	    in?: string;
	    name?: string;
	    valueTemplate: string;
	
	    static createFrom(source: any = {}) {
	        return new HTTPAuth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.in = source["in"];
	        this.name = source["name"];
	        this.valueTemplate = source["valueTemplate"];
	    }
	}
	export class HTTPRequest {
	    method?: string;
	    urlTemplate: string;
	    query?: Record<string, string>;
	    headers?: Record<string, string>;
	    body?: string;
	    auth?: HTTPAuth;
	    timeoutMs?: number;
	
	    static createFrom(source: any = {}) {
	        return new HTTPRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.urlTemplate = source["urlTemplate"];
	        this.query = source["query"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.auth = this.convertValues(source["auth"], HTTPAuth);
	        this.timeoutMs = source["timeoutMs"];
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
	export class HTTPToolImpl {
	    request: HTTPRequest;
	    response: HTTPResponse;
	
	    static createFrom(source: any = {}) {
	        return new HTTPToolImpl(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.request = this.convertValues(source["request"], HTTPRequest);
	        this.response = this.convertValues(source["response"], HTTPResponse);
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
	export class GoToolImpl {
	    func: string;
	
	    static createFrom(source: any = {}) {
	        return new GoToolImpl(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.func = source["func"];
	    }
	}
	export class Tool {
	    schemaVersion: string;
	    id: string;
	    slug: string;
	    version: string;
	    displayName: string;
	    description?: string;
	    tags?: string[];
	    argSchema: number[];
	    outputSchema: number[];
	    type: string;
	    goImpl?: GoToolImpl;
	    httpImpl?: HTTPToolImpl;
	    isEnabled: boolean;
	    isBuiltIn: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Tool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.slug = source["slug"];
	        this.version = source["version"];
	        this.displayName = source["displayName"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.argSchema = source["argSchema"];
	        this.outputSchema = source["outputSchema"];
	        this.type = source["type"];
	        this.goImpl = this.convertValues(source["goImpl"], GoToolImpl);
	        this.httpImpl = this.convertValues(source["httpImpl"], HTTPToolImpl);
	        this.isEnabled = source["isEnabled"];
	        this.isBuiltIn = source["isBuiltIn"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
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
	export class GetToolResponse {
	    Body?: Tool;
	
	    static createFrom(source: any = {}) {
	        return new GetToolResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], Tool);
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
	    PageSize: number;
	    PageToken: string;
	
	    static createFrom(source: any = {}) {
	        return new ListConversationsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PageSize = source["PageSize"];
	        this.PageToken = source["PageToken"];
	    }
	}
	export class ListConversationsResponseBody {
	    conversationListItems: ConversationListItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListConversationsResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conversationListItems = this.convertValues(source["conversationListItems"], ConversationListItem);
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
	
	export class ListPromptBundlesRequest {
	    BundleIDs: string[];
	    IncludeDisabled: boolean;
	    PageSize: number;
	    PageToken: string;
	
	    static createFrom(source: any = {}) {
	        return new ListPromptBundlesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleIDs = source["BundleIDs"];
	        this.IncludeDisabled = source["IncludeDisabled"];
	        this.PageSize = source["PageSize"];
	        this.PageToken = source["PageToken"];
	    }
	}
	export class PromptBundle {
	    schemaVersion: string;
	    id: string;
	    slug: string;
	    displayName?: string;
	    description?: string;
	    isEnabled: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    isBuiltIn: boolean;
	    // Go type: time
	    softDeletedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new PromptBundle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.slug = source["slug"];
	        this.displayName = source["displayName"];
	        this.description = source["description"];
	        this.isEnabled = source["isEnabled"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
	        this.isBuiltIn = source["isBuiltIn"];
	        this.softDeletedAt = this.convertValues(source["softDeletedAt"], null);
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
	export class ListPromptBundlesResponseBody {
	    promptBundles: PromptBundle[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListPromptBundlesResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.promptBundles = this.convertValues(source["promptBundles"], PromptBundle);
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
	export class ListPromptBundlesResponse {
	    Body?: ListPromptBundlesResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new ListPromptBundlesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], ListPromptBundlesResponseBody);
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
	
	export class ListPromptTemplatesRequest {
	    BundleIDs: string[];
	    Tags: string[];
	    IncludeDisabled: boolean;
	    RecommendedPageSize: number;
	    PageToken: string;
	
	    static createFrom(source: any = {}) {
	        return new ListPromptTemplatesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleIDs = source["BundleIDs"];
	        this.Tags = source["Tags"];
	        this.IncludeDisabled = source["IncludeDisabled"];
	        this.RecommendedPageSize = source["RecommendedPageSize"];
	        this.PageToken = source["PageToken"];
	    }
	}
	export class PromptTemplateListItem {
	    bundleID: string;
	    bundleSlug: string;
	    templateSlug: string;
	    templateVersion: string;
	    isBuiltIn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PromptTemplateListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bundleID = source["bundleID"];
	        this.bundleSlug = source["bundleSlug"];
	        this.templateSlug = source["templateSlug"];
	        this.templateVersion = source["templateVersion"];
	        this.isBuiltIn = source["isBuiltIn"];
	    }
	}
	export class ListPromptTemplatesResponseBody {
	    promptTemplateListItems: PromptTemplateListItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListPromptTemplatesResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.promptTemplateListItems = this.convertValues(source["promptTemplateListItems"], PromptTemplateListItem);
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
	export class ListPromptTemplatesResponse {
	    Body?: ListPromptTemplatesResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new ListPromptTemplatesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], ListPromptTemplatesResponseBody);
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
	
	export class ListToolBundlesRequest {
	    BundleIDs: string[];
	    IncludeDisabled: boolean;
	    PageSize: number;
	    PageToken: string;
	
	    static createFrom(source: any = {}) {
	        return new ListToolBundlesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleIDs = source["BundleIDs"];
	        this.IncludeDisabled = source["IncludeDisabled"];
	        this.PageSize = source["PageSize"];
	        this.PageToken = source["PageToken"];
	    }
	}
	export class ToolBundle {
	    schemaVersion: string;
	    id: string;
	    slug: string;
	    displayName?: string;
	    description?: string;
	    isEnabled: boolean;
	    isBuiltIn: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    // Go type: time
	    softDeletedAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new ToolBundle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.slug = source["slug"];
	        this.displayName = source["displayName"];
	        this.description = source["description"];
	        this.isEnabled = source["isEnabled"];
	        this.isBuiltIn = source["isBuiltIn"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.modifiedAt = this.convertValues(source["modifiedAt"], null);
	        this.softDeletedAt = this.convertValues(source["softDeletedAt"], null);
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
	export class ListToolBundlesResponseBody {
	    toolBundles: ToolBundle[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListToolBundlesResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolBundles = this.convertValues(source["toolBundles"], ToolBundle);
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
	export class ListToolBundlesResponse {
	    Body?: ListToolBundlesResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new ListToolBundlesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], ListToolBundlesResponseBody);
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
	
	export class ListToolsRequest {
	    BundleIDs: string[];
	    Tags: string[];
	    IncludeDisabled: boolean;
	    RecommendedPageSize: number;
	    PageToken: string;
	
	    static createFrom(source: any = {}) {
	        return new ListToolsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleIDs = source["BundleIDs"];
	        this.Tags = source["Tags"];
	        this.IncludeDisabled = source["IncludeDisabled"];
	        this.RecommendedPageSize = source["RecommendedPageSize"];
	        this.PageToken = source["PageToken"];
	    }
	}
	export class ToolListItem {
	    bundleID: string;
	    bundleSlug: string;
	    toolSlug: string;
	    toolVersion: string;
	    isBuiltIn: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bundleID = source["bundleID"];
	        this.bundleSlug = source["bundleSlug"];
	        this.toolSlug = source["toolSlug"];
	        this.toolVersion = source["toolVersion"];
	        this.isBuiltIn = source["isBuiltIn"];
	    }
	}
	export class ListToolsResponseBody {
	    toolListItems: ToolListItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListToolsResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolListItems = this.convertValues(source["toolListItems"], ToolListItem);
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
	export class ListToolsResponse {
	    Body?: ListToolsResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new ListToolsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], ListToolsResponseBody);
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
	
	
	export class ModelParams {
	    name: string;
	    stream: boolean;
	    maxPromptLength: number;
	    maxOutputLength: number;
	    temperature?: number;
	    reasoning?: ReasoningParams;
	    systemPrompt: string;
	    timeout: number;
	    additionalParametersRawJSON?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.stream = source["stream"];
	        this.maxPromptLength = source["maxPromptLength"];
	        this.maxOutputLength = source["maxOutputLength"];
	        this.temperature = source["temperature"];
	        this.reasoning = this.convertValues(source["reasoning"], ReasoningParams);
	        this.systemPrompt = source["systemPrompt"];
	        this.timeout = source["timeout"];
	        this.additionalParametersRawJSON = source["additionalParametersRawJSON"];
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
	
	export class PatchPromptBundleRequestBody {
	    isEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PatchPromptBundleRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	    }
	}
	export class PatchPromptBundleRequest {
	    BundleID: string;
	    body?: PatchPromptBundleRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PatchPromptBundleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.body = this.convertValues(source["body"], PatchPromptBundleRequestBody);
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
	
	export class PatchPromptBundleResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PatchPromptBundleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PatchPromptTemplateRequestBody {
	    isEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PatchPromptTemplateRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	    }
	}
	export class PatchPromptTemplateRequest {
	    BundleID: string;
	    TemplateSlug: string;
	    Version: string;
	    body?: PatchPromptTemplateRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PatchPromptTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.TemplateSlug = source["TemplateSlug"];
	        this.Version = source["Version"];
	        this.body = this.convertValues(source["body"], PatchPromptTemplateRequestBody);
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
	
	export class PatchPromptTemplateResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PatchPromptTemplateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PatchToolBundleRequestBody {
	    isEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PatchToolBundleRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	    }
	}
	export class PatchToolBundleRequest {
	    BundleID: string;
	    body?: PatchToolBundleRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PatchToolBundleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.body = this.convertValues(source["body"], PatchToolBundleRequestBody);
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
	
	export class PatchToolBundleResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PatchToolBundleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PatchToolRequestBody {
	    isEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PatchToolRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isEnabled = source["isEnabled"];
	    }
	}
	export class PatchToolRequest {
	    BundleID: string;
	    ToolSlug: string;
	    Version: string;
	    body?: PatchToolRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PatchToolRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.ToolSlug = source["ToolSlug"];
	        this.Version = source["Version"];
	        this.body = this.convertValues(source["body"], PatchToolRequestBody);
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
	
	export class PatchToolResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PatchToolResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	
	
	
	
	
	
	
	
	export class PutConversationRequestBody {
	    title: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    modifiedAt: any;
	    messages: ConversationMessage[];
	
	    static createFrom(source: any = {}) {
	        return new PutConversationRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
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
	export class PutConversationRequest {
	    ID: string;
	    Body?: PutConversationRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PutConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Body = this.convertValues(source["Body"], PutConversationRequestBody);
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
	
	export class PutConversationResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PutConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PutMessagesToConversationRequestBody {
	    title: string;
	    messages: ConversationMessage[];
	
	    static createFrom(source: any = {}) {
	        return new PutMessagesToConversationRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
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
	export class PutMessagesToConversationRequest {
	    ID: string;
	    Body?: PutMessagesToConversationRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PutMessagesToConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Body = this.convertValues(source["Body"], PutMessagesToConversationRequestBody);
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
	
	export class PutMessagesToConversationResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PutMessagesToConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PutPromptBundleRequestBody {
	    slug: string;
	    displayName: string;
	    isEnabled: boolean;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new PutPromptBundleRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.displayName = source["displayName"];
	        this.isEnabled = source["isEnabled"];
	        this.description = source["description"];
	    }
	}
	export class PutPromptBundleRequest {
	    BundleID: string;
	    body?: PutPromptBundleRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PutPromptBundleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.body = this.convertValues(source["body"], PutPromptBundleRequestBody);
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
	
	export class PutPromptBundleResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PutPromptBundleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PutPromptTemplateRequestBody {
	    displayName: string;
	    isEnabled: boolean;
	    description?: string;
	    blocks: MessageBlock[];
	    tags?: string[];
	    variables?: PromptVariable[];
	    preProcessors?: PreProcessorCall[];
	
	    static createFrom(source: any = {}) {
	        return new PutPromptTemplateRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.isEnabled = source["isEnabled"];
	        this.description = source["description"];
	        this.blocks = this.convertValues(source["blocks"], MessageBlock);
	        this.tags = source["tags"];
	        this.variables = this.convertValues(source["variables"], PromptVariable);
	        this.preProcessors = this.convertValues(source["preProcessors"], PreProcessorCall);
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
	export class PutPromptTemplateRequest {
	    BundleID: string;
	    TemplateSlug: string;
	    Version: string;
	    body?: PutPromptTemplateRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PutPromptTemplateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.TemplateSlug = source["TemplateSlug"];
	        this.Version = source["Version"];
	        this.body = this.convertValues(source["body"], PutPromptTemplateRequestBody);
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
	
	export class PutPromptTemplateResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PutPromptTemplateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PutToolBundleRequestBody {
	    slug: string;
	    displayName: string;
	    isEnabled: boolean;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new PutToolBundleRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.displayName = source["displayName"];
	        this.isEnabled = source["isEnabled"];
	        this.description = source["description"];
	    }
	}
	export class PutToolBundleRequest {
	    BundleID: string;
	    body?: PutToolBundleRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PutToolBundleRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.body = this.convertValues(source["body"], PutToolBundleRequestBody);
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
	
	export class PutToolBundleResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PutToolBundleResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class PutToolRequestBody {
	    displayName: string;
	    description?: string;
	    tags?: string[];
	    isEnabled: boolean;
	    argSchema: number[];
	    outputSchema: number[];
	    type: string;
	    goImpl?: GoToolImpl;
	    httpImpl?: HTTPToolImpl;
	
	    static createFrom(source: any = {}) {
	        return new PutToolRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.description = source["description"];
	        this.tags = source["tags"];
	        this.isEnabled = source["isEnabled"];
	        this.argSchema = source["argSchema"];
	        this.outputSchema = source["outputSchema"];
	        this.type = source["type"];
	        this.goImpl = this.convertValues(source["goImpl"], GoToolImpl);
	        this.httpImpl = this.convertValues(source["httpImpl"], HTTPToolImpl);
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
	export class PutToolRequest {
	    BundleID: string;
	    ToolSlug: string;
	    Version: string;
	    body?: PutToolRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new PutToolRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.BundleID = source["BundleID"];
	        this.ToolSlug = source["ToolSlug"];
	        this.Version = source["Version"];
	        this.body = this.convertValues(source["body"], PutToolRequestBody);
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
	
	export class PutToolResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new PutToolResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	
	export class SearchConversationsRequest {
	    Query: string;
	    PageToken: string;
	    PageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchConversationsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Query = source["Query"];
	        this.PageToken = source["PageToken"];
	        this.PageSize = source["PageSize"];
	    }
	}
	export class SearchConversationsResponseBody {
	    conversationListItems: ConversationListItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchConversationsResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conversationListItems = this.convertValues(source["conversationListItems"], ConversationListItem);
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
	export class SearchConversationsResponse {
	    Body?: SearchConversationsResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new SearchConversationsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], SearchConversationsResponseBody);
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
	
	export class SearchPromptTemplatesRequest {
	    Query: string;
	    PageToken: string;
	    PageSize: number;
	    IncludeDisabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SearchPromptTemplatesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Query = source["Query"];
	        this.PageToken = source["PageToken"];
	        this.PageSize = source["PageSize"];
	        this.IncludeDisabled = source["IncludeDisabled"];
	    }
	}
	export class SearchPromptTemplatesResponseBody {
	    promptTemplateListItems: PromptTemplateListItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchPromptTemplatesResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.promptTemplateListItems = this.convertValues(source["promptTemplateListItems"], PromptTemplateListItem);
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
	export class SearchPromptTemplatesResponse {
	    Body?: SearchPromptTemplatesResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new SearchPromptTemplatesResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], SearchPromptTemplatesResponseBody);
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
	
	export class SearchToolsRequest {
	    Query: string;
	    PageToken: string;
	    PageSize: number;
	    IncludeDisabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SearchToolsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Query = source["Query"];
	        this.PageToken = source["PageToken"];
	        this.PageSize = source["PageSize"];
	        this.IncludeDisabled = source["IncludeDisabled"];
	    }
	}
	export class SearchToolsResponseBody {
	    toolListItems: ToolListItem[];
	    nextPageToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchToolsResponseBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolListItems = this.convertValues(source["toolListItems"], ToolListItem);
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
	export class SearchToolsResponse {
	    Body?: SearchToolsResponseBody;
	
	    static createFrom(source: any = {}) {
	        return new SearchToolsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Body = this.convertValues(source["Body"], SearchToolsResponseBody);
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
	
	export class SetDefaultModelPresetRequestBody {
	    ModelPresetID: string;
	
	    static createFrom(source: any = {}) {
	        return new SetDefaultModelPresetRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ModelPresetID = source["ModelPresetID"];
	    }
	}
	export class SetDefaultModelPresetRequest {
	    ProviderName: string;
	    Body?: SetDefaultModelPresetRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetDefaultModelPresetRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProviderName = source["ProviderName"];
	        this.Body = this.convertValues(source["Body"], SetDefaultModelPresetRequestBody);
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
	
	export class SetDefaultModelPresetResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetDefaultModelPresetResponse(source);
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
	export class SetProviderAPIKeyRequestBody {
	    apiKey: string;
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAPIKeyRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	    }
	}
	export class SetProviderAPIKeyRequest {
	    Provider: string;
	    Body?: SetProviderAPIKeyRequestBody;
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAPIKeyRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = source["Provider"];
	        this.Body = this.convertValues(source["Body"], SetProviderAPIKeyRequestBody);
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
	
	export class SetProviderAPIKeyResponse {
	
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAPIKeyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SetProviderAttributeRequestBody {
	    origin?: string;
	    chatCompletionPathPrefix?: string;
	
	    static createFrom(source: any = {}) {
	        return new SetProviderAttributeRequestBody(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.origin = source["origin"];
	        this.chatCompletionPathPrefix = source["chatCompletionPathPrefix"];
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
	
	

}

