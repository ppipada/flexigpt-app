import { useState } from 'react';

interface ThinkingFenceProps {
	detailsSummaryText: string;
	text: string;
	isBusy: boolean;
	maxHeightClass?: string; // e.g. 'max-h-[50vh]' or 'max-h-60'
}

export function ThinkingFence({
	detailsSummaryText,
	text,
	isBusy,
	maxHeightClass = 'max-h-[50vh]',
}: ThinkingFenceProps) {
	const [open, setOpen] = useState(false);

	return (
		<details
			open={open}
			// reflect user toggles back into state
			onToggle={e => {
				setOpen((e.currentTarget as HTMLDetailsElement).open);
			}}
			className="group bg-base-200/70 my-1 overflow-hidden rounded shadow-none"
		>
			<summary className="flex cursor-pointer items-center gap-2 px-3 py-2 text-xs transition-colors select-none">
				<span className="text-xs">{detailsSummaryText}</span>
				{/* simple chevron */}
				<svg
					className="ml-1 h-3 w-3 transition-transform group-open:rotate-90"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
				</svg>

				{isBusy && <span className="loading loading-dots loading-xs ml-auto" />}
			</summary>

			{/* body */}
			<div className={`px-3 pb-3 wrap-break-word whitespace-pre-wrap ${maxHeightClass} overflow-y-auto text-xs`}>
				{text || '...'}
			</div>
		</details>
	);
}
