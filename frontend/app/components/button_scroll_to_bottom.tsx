import type { ButtonHTMLAttributes, FC, RefObject } from 'react';

import { FiArrowDownCircle } from 'react-icons/fi';

import { useAtBottom } from '@/hooks/use_at_bottom';

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
