// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {spec} from '../models';

export function AddProvider(arg1:spec.AddProviderRequest):Promise<spec.AddProviderResponse>;

export function DeleteProvider(arg1:spec.DeleteProviderRequest):Promise<spec.DeleteProviderResponse>;

export function FetchCompletion(arg1:string,arg2:string,arg3:spec.ModelParams,arg4:Array<spec.ChatCompletionRequestMessage>,arg5:string):Promise<spec.FetchCompletionResponse>;

export function SetProviderAPIKey(arg1:spec.SetProviderAPIKeyRequest):Promise<spec.SetProviderAPIKeyResponse>;
