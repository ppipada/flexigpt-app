import type { RefObject } from 'react';
import { useEffect, useState } from 'react';

export function useAtBottom(ref: RefObject<HTMLElement | null>, offset = 0) {
	const [isAtBottom, setIsAtBottom] = useState(false);
	const [isScrollable, setIsScrollable] = useState(false);

	useEffect(() => {
		const currentRef = ref.current;
		const handleScroll = () => {
			if (currentRef) {
				const { scrollTop, scrollHeight, clientHeight } = currentRef;
				setIsScrollable(scrollHeight > clientHeight);
				setIsAtBottom(scrollTop + clientHeight >= scrollHeight - offset - 10);
			}
		};

		// Add a throttling mechanism
		let throttleTimeout: NodeJS.Timeout | null = null;
		const throttledHandleScroll = () => {
			if (throttleTimeout === null) {
				throttleTimeout = setTimeout(() => {
					handleScroll();
					throttleTimeout = null;
				}, 100); // Adjust the timeout value as needed
			}
		};

		if (currentRef) {
			currentRef.addEventListener('scroll', throttledHandleScroll, {
				passive: true,
			});
		}
		handleScroll();

		return () => {
			if (currentRef) {
				currentRef.removeEventListener('scroll', throttledHandleScroll);
			}
			if (throttleTimeout) {
				clearTimeout(throttleTimeout);
			}
		};
	}, [ref, offset]);

	return { isAtBottom, isScrollable };
}
