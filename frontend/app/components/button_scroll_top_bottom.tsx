import type { ButtonHTMLAttributes, RefObject } from 'react';

import { FiArrowDownCircle, FiArrowUpCircle } from 'react-icons/fi';

interface ButtonScrollTopBottomProps {
	scrollContainerRef: RefObject<HTMLElement | null>;
	iconSize: number;

	// Control visibility via CSS (not mount/unmount)
	showTop: boolean;
	showBottom: boolean;

	// Wrapper element (positioning/layout)
	className?: string;

	// Optional individual button props
	topButtonProps?: ButtonHTMLAttributes<HTMLButtonElement>;
	bottomButtonProps?: ButtonHTMLAttributes<HTMLButtonElement>;
}

export function ButtonScrollTopBottom({
	scrollContainerRef,
	iconSize,
	showTop,
	showBottom,
	className = '',
	topButtonProps,
	bottomButtonProps,
}: ButtonScrollTopBottomProps) {
	return (
		<div className={className}>
			{/* Scroll to top */}
			<button
				{...topButtonProps}
				aria-label={topButtonProps?.['aria-label'] ?? 'Scroll To Top'}
				title={topButtonProps?.title ?? 'Scroll To Top'}
				disabled={topButtonProps?.disabled ?? !showTop}
				onClick={event => {
					if (!showTop) return;

					// Let consumer hook into click; they can preventDefault to stop scroll
					topButtonProps?.onClick?.(event);
					if (event.defaultPrevented) return;

					const el = scrollContainerRef.current;
					if (el) {
						el.scrollTo({ top: 0, behavior: 'smooth' });
					}
				}}
				className={`transition-opacity duration-150 ${topButtonProps?.className ?? ''} ${
					showTop ? 'visible opacity-100' : 'pointer-events-none invisible opacity-0'
				}`}
			>
				<FiArrowUpCircle size={iconSize} />
			</button>

			{/* Scroll to bottom */}
			<button
				{...bottomButtonProps}
				aria-label={bottomButtonProps?.['aria-label'] ?? 'Scroll To Bottom'}
				title={bottomButtonProps?.title ?? 'Scroll To Bottom'}
				disabled={bottomButtonProps?.disabled ?? !showBottom}
				onClick={event => {
					if (!showBottom) return;

					bottomButtonProps?.onClick?.(event);
					if (event.defaultPrevented) return;

					const el = scrollContainerRef.current;
					if (el) {
						el.scrollTo({
							top: el.scrollHeight - el.clientHeight,
							behavior: 'smooth',
						});
					}
				}}
				className={`transition-opacity duration-150 ${bottomButtonProps?.className ?? ''} ${
					showBottom ? 'visible opacity-100' : 'pointer-events-none invisible opacity-0'
				}`}
			>
				<FiArrowDownCircle size={iconSize} />
			</button>
		</div>
	);
}
