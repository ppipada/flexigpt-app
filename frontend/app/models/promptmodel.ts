export interface PromptTemplate {
	id: string;
	name: string;
	command: string;
	hasTools: boolean;
	hasDocStore: boolean;
	tokenCount: number;
}

export interface Tool {
	id: string;
	name: string;
	command: string;
	tokenCount: number;
}
