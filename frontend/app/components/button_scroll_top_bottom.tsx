import type { ButtonHTMLAttributes, RefObject } from 'react';

import { FiArrowDownCircle, FiArrowUpCircle } from 'react-icons/fi';

interface ButtonScrollToBottomProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	scrollContainerRef: RefObject<HTMLElement | null>;
	iconSize: number;
	show: boolean; // new: control visibility via CSS, not mount/unmount
}

export function ButtonScrollToBottom({
	scrollContainerRef,
	iconSize,
	show,
	className = '',
	...props
}: ButtonScrollToBottomProps) {
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
}

interface ButtonScrollToTopProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	scrollContainerRef: RefObject<HTMLElement | null>;
	iconSize: number;
	show: boolean; // new: control visibility via CSS, not mount/unmount
}

export function ButtonScrollToTop({
	scrollContainerRef,
	iconSize,
	show,
	className = '',
	...props
}: ButtonScrollToTopProps) {
	return (
		<button
			aria-label="Scroll To Top"
			title="Scroll To Top"
			disabled={!show}
			onClick={() => {
				if (!show) return;
				const el = scrollContainerRef.current;
				if (el) {
					el.scrollTo({ top: 10, behavior: 'smooth' });
				}
			}}
			className={`${className} transition-opacity duration-150 ${show ? 'visible opacity-100' : 'pointer-events-none invisible opacity-0'}`}
			{...props}
		>
			<FiArrowUpCircle size={iconSize} />
		</button>
	);
}
