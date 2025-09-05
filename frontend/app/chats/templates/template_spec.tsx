import type {
	MessageBlock,
	PreProcessorCall,
	PreProcessorOnError,
	PromptTemplate,
	PromptVariable,
} from '@/spec/prompt';

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

export type ToolState = {
	args?: Record<string, any>;
	status: 'pending' | 'ready' | 'done' | 'error';
	result?: any;
	error?: string;
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
		preProcessors?: PreProcessorCall[];
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
	preProcessors: PreProcessorCall[];

	// Effective variable values for execution
	variableValues: Record<string, unknown>;

	// Requirements state
	requiredVariables: string[];
	requiredCount: number;

	// Preprocessor runs needed
	toolsToRun: Array<{
		id: string;
		toolID: string;
		args?: Record<string, any>;
		saveAs: string;
		pathExpr?: string;
		onError?: PreProcessorOnError;
		status: 'pending' | 'ready' | 'done' | 'error';
	}>;

	// Convenience
	isReady: boolean;
}
