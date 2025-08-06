import type { ButtonHTMLAttributes, FC, RefObject } from 'react';

import { FiArrowDownCircle } from 'react-icons/fi';

interface ButtonScrollToBottomProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	scrollContainerRef: RefObject<HTMLElement | null>;
	iconSize: number;
	isAtBottom: boolean;
	isScrollable: boolean;
}

const ButtonScrollToBottom: FC<ButtonScrollToBottomProps> = ({
	scrollContainerRef,
	iconSize,
	isAtBottom,
	isScrollable,
	...props
}) => {
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
				<FiArrowDownCircle size={iconSize} />
			</button>
		)
	);
};

export default ButtonScrollToBottom;
