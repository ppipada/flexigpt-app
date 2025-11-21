import type { MessageBlock, PromptTemplate, PromptVariable } from '@/spec/prompt';

export const KEY_TEMPLATE_SELECTION = 'templateSelection';
export const KEY_TEMPLATE_VARIABLE = 'templateVariable';
export const KEY_TEMPLATE_SLASH_COMMAND = 'templateSlash';
export const KEY_TEMPLATE_SLASH_INPUT = 'templateInput';

export type TemplateVariableElementNode = {
	type: typeof KEY_TEMPLATE_VARIABLE;
	bundleID: string;
	templateSlug: string;
	templateVersion: string;
	selectionID: string;
	name: string;
	// for layout only (computed again at render)
	required?: boolean;
	children: [{ text: '' }];
};

export enum ToolStatus {
	PENDING = 'pending',
	READY = 'ready',
	RUNNING = 'running',
	DONE = 'done',
	ERROR = 'error',
}

export type ToolState = {
	args?: Record<string, any>;
	status: ToolStatus;
	result?: any;
	error?: string;
	lastRunAt?: string;
};

export type TemplateSelectionElementNode = {
	type: typeof KEY_TEMPLATE_SELECTION;
	bundleID: string;
	templateSlug: string;
	templateVersion: string;
	selectionID: string;

	// User-provided variable values
	variables: Record<string, unknown>;

	// Captured template at insertion time
	templateSnapshot?: PromptTemplate;

	// Local per-chip overrides
	overrides?: {
		displayName?: string;
		description?: string;
		tags?: string[];
		blocks?: MessageBlock[];
		variables?: PromptVariable[];
	};

	// Tool run states per preprocessor id
	toolStates?: Record<string, ToolState>;

	// Slate text children
	children: [{ text: '' }];
};

export interface SelectedTemplateForRun {
	type: typeof KEY_TEMPLATE_SELECTION;
	bundleID: string;
	templateSlug: string;
	templateVersion: string;
	selectionID: string;

	// Final structures after applying local overrides
	template: PromptTemplate;
	blocks: MessageBlock[];
	variablesSchema: PromptVariable[];

	// Effective variable values for execution
	variableValues: Record<string, unknown>;

	// Requirements state
	requiredVariables: string[];
	requiredCount: number;

	// Convenience
	isReady: boolean;
}
