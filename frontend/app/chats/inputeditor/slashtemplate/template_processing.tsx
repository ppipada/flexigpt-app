import type {
	MessageBlock,
	PreProcessorCall,
	PreProcessorOnError,
	PromptTemplate,
	PromptVariable,
} from '@/spec/prompt';
import { VarSource } from '@/spec/prompt';

export type ToolState = {
	args?: Record<string, any>;
	status: 'pending' | 'ready' | 'done' | 'error';
	result?: any;
	error?: string;
};

export type TemplateSelectionElementNode = {
	type: string; // should be KEY_TEMPLATE_SELECTION
	bundleID: string;
	templateSlug: string;
	templateVersion: string;

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
	type: 'templateSelection';
	bundleID: string;
	templateSlug: string;
	templateVersion: string;

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

export function buildInitialToolStates(template?: PromptTemplate): Record<string, ToolState> {
	const states: Record<string, ToolState> = {};
	if (template?.preProcessors?.length) {
		for (const p of template.preProcessors) {
			states[p.id] = {
				args: p.args ?? {},
				status: 'pending',
			};
		}
	}
	return states;
}

/**
 * Merge templateSnapshot with local overrides to produce effective template structures.
 */
export function computeEffectiveTemplate(el: TemplateSelectionElementNode): {
	template: PromptTemplate | undefined;
	blocks: MessageBlock[];
	variablesSchema: PromptVariable[];
	preProcessors: PreProcessorCall[];
} {
	const base = el.templateSnapshot;
	const blocks = el.overrides?.blocks ?? base?.blocks ?? [];
	const variablesSchema = el.overrides?.variables ?? base?.variables ?? [];
	const preProcessors = el.overrides?.preProcessors ?? base?.preProcessors ?? [];
	return { template: base, blocks, variablesSchema, preProcessors };
}

function effectiveVarValue(
	varDef: PromptVariable,
	userValues: Record<string, unknown>,
	toolStates?: Record<string, ToolState>
): unknown {
	// Local override always wins if present
	if (userValues[varDef.name] !== undefined && userValues[varDef.name] !== null) {
		return userValues[varDef.name];
	}

	// Source & defaults
	switch (varDef.source) {
		case VarSource.Static:
			if (varDef.staticVal !== undefined && varDef.staticVal !== '') {
				return varDef.staticVal;
			}
			break;
		case VarSource.Tool: {
			if (toolStates) {
				const hit = Object.values(toolStates).find(st => st.result !== undefined);
				if (hit?.result !== undefined) return hit.result;
			}
			break;
		}
		case VarSource.User:
		default:
			break;
	}

	// Fallback declared default
	if (varDef.default !== undefined && varDef.default !== '') return varDef.default;

	return undefined;
}

export function computeRequirements(
	variablesSchema: PromptVariable[],
	variableValues: Record<string, unknown>,
	preProcessors: PreProcessorCall[] = [],
	toolStates?: Record<string, ToolState>
) {
	const requiredNames: string[] = [];
	const values: Record<string, unknown> = { ...variableValues };

	// Fill effective values
	for (const v of variablesSchema) {
		const val = effectiveVarValue(v, variableValues, toolStates);
		values[v.name] = val;
	}

	// Identify required ones that remain empty
	for (const v of variablesSchema) {
		const val = values[v.name];
		if (v.required && (val === undefined || val === null || val === '')) {
			requiredNames.push(v.name);
		}
	}

	// Tools to run summary
	const toolsToRun = preProcessors.map(p => {
		const st = toolStates?.[p.id];
		return {
			id: p.id,
			toolID: p.toolID,
			args: st?.args ?? p.args,
			saveAs: p.saveAs,
			pathExpr: p.pathExpr,
			onError: p.onError,
			status: st?.status ?? 'pending',
		};
	});

	const pendingToolsCount = toolsToRun.filter(t => t.status !== 'done').length;

	return {
		variableValues: values,
		requiredVariables: requiredNames,
		requiredCount: requiredNames.length,
		toolsToRun,
		pendingToolsCount,
	};
}

export function makeSelectedTemplateForRun(tsenode: TemplateSelectionElementNode): SelectedTemplateForRun {
	const { template, blocks, variablesSchema, preProcessors } = computeEffectiveTemplate(tsenode);

	const effTemplate: PromptTemplate =
		template ??
		({
			id: '',
			displayName: tsenode.templateSlug,
			slug: tsenode.templateSlug,
			isEnabled: true,
			description: '',
			tags: [],
			blocks,
			variables: variablesSchema,
			preProcessors,
			version: tsenode.templateVersion,
			createdAt: new Date().toISOString(),
			modifiedAt: new Date().toISOString(),
			isBuiltIn: false,
		} as PromptTemplate);

	const req = computeRequirements(variablesSchema, tsenode.variables, preProcessors, tsenode.toolStates);

	return {
		type: 'templateSelection',
		bundleID: tsenode.bundleID,
		templateSlug: tsenode.templateSlug,
		templateVersion: tsenode.templateVersion,
		template: effTemplate,
		blocks,
		variablesSchema,
		preProcessors,
		variableValues: req.variableValues,
		requiredVariables: req.requiredVariables,
		requiredCount: req.requiredCount,
		toolsToRun: req.toolsToRun,
		isReady: req.requiredCount === 0 && req.toolsToRun.every(t => t.status === 'done'),
	};
}
