import type { ButtonHTMLAttributes, FC, RefObject } from 'react';

import { FiArrowDownCircle } from 'react-icons/fi';

interface ButtonScrollToBottomProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	scrollContainerRef: RefObject<HTMLElement | null>;
	iconSize: number;
	show: boolean; // new: control visibility via CSS, not mount/unmount
}

const ButtonScrollToBottom: FC<ButtonScrollToBottomProps> = ({
	scrollContainerRef,
	iconSize,
	show,
	className = '',
	...props
}) => {
	return (
		<button
			aria-label="Scroll To Bottom"
			title="Scroll To Bottom"
			disabled={!show}
			onClick={() => {
				if (!show) return;
				const el = scrollContainerRef.current;
				if (el) {
					el.scrollTo({ top: el.scrollHeight - el.clientHeight, behavior: 'smooth' });
				}
			}}
			className={`${className} transition-opacity duration-150 ${show ? 'visible opacity-100' : 'pointer-events-none invisible opacity-0'}`}
			{...props}
		>
			<FiArrowDownCircle size={iconSize} />
		</button>
	);
};

export default ButtonScrollToBottom;
