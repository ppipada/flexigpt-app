import { type KeyboardEvent as ReactKeyboardEvent, useCallback, useRef } from 'react';

type BoolOrGetter = boolean | (() => boolean);
const resolveBool = (v: BoolOrGetter) => (typeof v === 'function' ? (v as () => boolean)() : v);

interface EnterSubmitConfig {
	// can be a boolean or a getter (if computing busy state is cheaper lazily)
	isBusy: BoolOrGetter;
	// whether a submit is allowed (e.g., non-empty text)
	canSubmit: () => boolean;
	// if provided, used to insert a soft break on Shift+Enter
	insertSoftBreak?: () => void;
	// if provided, used instead of formRef.requestSubmit()
	onSubmitRequest?: () => void;
}

export function useEnterSubmit(config: EnterSubmitConfig) {
	const formRef = useRef<HTMLFormElement>(null);
	const configRef = useRef(config);
	configRef.current = config;

	const onKeyDown = useCallback((e: ReactKeyboardEvent<HTMLDivElement>) => {
		if (e.isDefaultPrevented()) return;

		const native = e.nativeEvent;
		const isComposing = (native as KeyboardEvent & { isComposing?: boolean }).isComposing;

		// Shift+Enter => soft break
		if (e.key === 'Enter' && e.shiftKey && !isComposing) {
			configRef.current.insertSoftBreak?.();
			e.preventDefault();
			return;
		}

		// Enter => submit
		if (e.key === 'Enter' && !e.shiftKey && !isComposing) {
			const busy = resolveBool(configRef.current.isBusy);
			if (!busy && configRef.current.canSubmit()) {
				if (configRef.current.onSubmitRequest) {
					configRef.current.onSubmitRequest();
				} else {
					formRef.current?.requestSubmit();
				}
			}
			e.preventDefault();
		}
	}, []);

	return { formRef, onKeyDown };
}
