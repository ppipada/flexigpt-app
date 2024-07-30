import { ReactNode } from 'react';

export enum ChatCompletionRoleEnum {
	system = 'system',
	user = 'user',
	assistant = 'assistant',
	function = 'function',
	feedback = 'feedback',
}

export interface Message {
	id: string;
	createdAt?: Date;

	role: ChatCompletionRoleEnum;
	content: string;
	timestamp?: string;
	name?: string;
	details?: string;
}

export interface Chat {
	id: string;
	title: string;
	createTime: Date;
	modifiedTime: Date;
	messages: Message[];
}

export interface User {
	id: string;
	role: ChatCompletionRoleEnum;
	icon: ReactNode;
}
