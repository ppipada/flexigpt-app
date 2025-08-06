import type { RefObject } from 'react';
import { useEffect, useRef, useState } from 'react';

export function useAtBottom(ref: RefObject<HTMLElement | null>, offset = 0) {
	const [isAtBottom, setIsAtBottom] = useState(false);
	const [isScrollable, setIsScrollable] = useState(false);

	// keep the handler in a ref so both the scroll listener and the
	// ResizeObserver can call the *same* function
	const handleScrollRef = useRef(() => {});

	useEffect(() => {
		const current = ref.current;
		if (!current) return;

		const handleScroll = () => {
			const { scrollTop, scrollHeight, clientHeight } = current;
			setIsScrollable(scrollHeight > clientHeight);
			setIsAtBottom(scrollTop + clientHeight >= scrollHeight - offset - 10);
		};

		// save it in the ref so it can be used inside ResizeObserver callback
		handleScrollRef.current = handleScroll;

		// --- scroll listener (throttled) ---------------------------------
		let throttleTimeout: NodeJS.Timeout | null = null;
		const throttledScroll = () => {
			if (throttleTimeout == null) {
				throttleTimeout = setTimeout(() => {
					handleScroll();
					throttleTimeout = null;
				}, 100);
			}
		};

		current.addEventListener('scroll', throttledScroll, { passive: true });

		// --- resize observer --------------------------------------------
		const resizeObserver = new ResizeObserver(() => {
			// no throttling here – this already fires at a sane rate
			handleScrollRef.current();
		});
		resizeObserver.observe(current);

		// run once on mount
		handleScroll();

		return () => {
			current.removeEventListener('scroll', throttledScroll);
			resizeObserver.disconnect();
			if (throttleTimeout) clearTimeout(throttleTimeout);
		};
	}, [ref, offset]);

	return { isAtBottom, isScrollable };
}
