import type { ButtonHTMLAttributes, FC, RefObject } from 'react';
import { useEffect, useState } from 'react';

import { FiArrowDownCircle } from 'react-icons/fi';

function useAtBottom(ref: RefObject<HTMLElement | null>, offset = 0) {
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

interface ButtonScrollToBottomProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	scrollContainerRef: RefObject<HTMLElement | null>;
	size: number;
}

const ButtonScrollToBottom: FC<ButtonScrollToBottomProps> = ({ scrollContainerRef, size, ...props }) => {
	const { isAtBottom, isScrollable } = useAtBottom(scrollContainerRef);

	return (
		isScrollable &&
		!isAtBottom && (
			<button
				aria-label="Scroll To Bottom"
				title="Scroll To Bottom"
				disabled={isAtBottom}
				onClick={() => {
					if (scrollContainerRef.current) {
						scrollContainerRef.current.scrollTo({
							top: scrollContainerRef.current.scrollHeight,
							behavior: 'smooth',
						});
					}
				}}
				{...props}
			>
				<FiArrowDownCircle size={size} />
			</button>
		)
	);
};

export default ButtonScrollToBottom;
