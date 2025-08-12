import React, { useState } from 'react';

interface ThinkingFenceProps {
	text: string;
	isStreaming: boolean;
	maxHeightClass?: string; // e.g. 'max-h-[50vh]' or 'max-h-60'
}

const ThinkingFence: React.FC<ThinkingFenceProps> = ({ text, isStreaming, maxHeightClass = 'max-h-[50vh]' }) => {
	const [open, setOpen] = useState(false);

	return (
		<details
			open={open}
			// reflect user toggles back into state
			onToggle={e => {
				setOpen((e.currentTarget as HTMLDetailsElement).open);
			}}
			className="group rounded shadow-none my-1 overflow-hidden bg-base-200/70"
		>
			<summary className="flex items-center gap-2 px-3 py-2 cursor-pointer select-none text-xs transition-colors">
				<span className="text-xs">Thinking</span>
				{/* simple chevron */}
				<svg
					className="w-3 h-3 ml-1 transition-transform group-open:rotate-90"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
				</svg>

				{isStreaming && <span className="loading loading-dots loading-xs ml-auto" />}
			</summary>

			{/* body */}
			<div className={`px-3 pb-3 whitespace-pre-wrap break-words ${maxHeightClass} overflow-y-auto text-xs`}>
				{text || '...'}
			</div>
		</details>
	);
};

export default ThinkingFence;
