import { useEffect } from 'react';

type SetSystemPromptForChatDetail = {
	prompt: string;
};

const SET_SYSTEM_PROMPT_FOR_CHAT_EVENT = 'chat:set-system-prompt-for-chat';

export function dispatchSetSystemPromptForChat(prompt: string) {
	window.dispatchEvent(
		new CustomEvent<SetSystemPromptForChatDetail>(SET_SYSTEM_PROMPT_FOR_CHAT_EVENT, {
			detail: { prompt },
		})
	);
}

export function useSetSystemPromptForChat(handler: (prompt: string) => void) {
	useEffect(() => {
		const onEvt = (e: Event) => {
			const ce = e as CustomEvent<SetSystemPromptForChatDetail>;
			const p = ce.detail.prompt;
			handler(p);
		};
		window.addEventListener(SET_SYSTEM_PROMPT_FOR_CHAT_EVENT, onEvt as EventListener);
		return () => {
			window.removeEventListener(SET_SYSTEM_PROMPT_FOR_CHAT_EVENT, onEvt as EventListener);
		};
	}, [handler]);
}
