export interface PromptTemplate {
	id: string;
	name: string;
	command: string;
	template: string;
	hasTools: boolean;
	hasDocStore: boolean;
	tokenCount: number;
}

export interface Tool {
	id: string;
	name: string;
	command: string;
	schema: string;
	inFunc: string;
	tokenCount: number;
}
