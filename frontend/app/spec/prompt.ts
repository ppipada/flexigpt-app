/**
 * @public
 */
export enum PromptRoleEnum {
	System = 'system',
	Developer = 'developer',
	User = 'user',
	Assistant = 'assistant',
}

/**
 * @public
 */
export enum VarType {
	String = 'string',
	Number = 'number',
	Boolean = 'boolean',
	Enum = 'enum',
	Date = 'date',
}

/**
 * @public
 */
export enum VarSource {
	User = 'user',
	Static = 'static',
}

export interface MessageBlock {
	id: string;
	role: PromptRoleEnum;
	content: string;
}

export interface PromptVariable {
	name: string;
	type: VarType;
	required: boolean;
	source: VarSource;
	description?: string;
	staticVal?: string;
	enumValues?: string[];
	default?: string;
}

export interface PromptTemplate {
	id: string;
	displayName: string;
	slug: string;
	isEnabled: boolean;
	description?: string;
	tags?: string[];
	blocks: MessageBlock[];
	variables?: PromptVariable[];
	version: string;
	createdAt: string;
	modifiedAt: string;
	isBuiltIn: boolean;
}

export interface PromptBundle {
	id: string;
	slug: string;
	displayName?: string;
	description?: string;
	isEnabled: boolean;
	createdAt: string;
	modifiedAt: string;
	isBuiltIn: boolean;
}

export interface PromptTemplateListItem {
	bundleID: string;
	bundleSlug: string;
	templateSlug: string;
	templateVersion: string;
	isBuiltIn: boolean;
}
