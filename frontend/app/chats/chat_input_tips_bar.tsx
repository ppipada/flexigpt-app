import { useEffect, useMemo, useRef, useState } from 'react';

import { FiChevronDown, FiChevronUp, FiMoreHorizontal } from 'react-icons/fi';

type TipKey = 'lastWins' | 'remove' | 'collapse';

const tipsText: Record<TipKey, string> = {
	lastWins: 'Last system prompt wins: when multiple templates add system prompts, the last one added becomes active.',
	remove: 'Removing a template also removes the system prompt that came from it (if it was created by that template).',
	collapse:
		'Collapse to text decouples removes the system prompts, vars, tools. The current user block with placeholders for vars is inserted as plain text',
};

export function CommandTipsBar() {
	const [isOpen, setIsOpen] = useState(false);
	const detailsRef = useRef<HTMLDetailsElement>(null);

	const tips = useMemo(
		() => [
			{ key: 'lastWins' as TipKey, title: 'Last system prompt wins', description: tipsText.lastWins },
			{ key: 'remove' as TipKey, title: 'Removing a template', description: tipsText.remove },
			{ key: 'collapse' as TipKey, title: 'Collapse to text decouples', description: tipsText.collapse },
		],
		[]
	);

	// Close on outside click or Escape, like pattern used elsewhere
	useEffect(() => {
		const handlePointerDown = (e: MouseEvent | TouchEvent) => {
			if (!isOpen) return;
			const node = detailsRef.current;
			if (node && !node.contains(e.target as Node)) {
				node.open = false;
				setIsOpen(false);
			}
		};

		const handleKeyDown = (e: KeyboardEvent) => {
			if (!isOpen) return;
			if (e.key === 'Escape') {
				const node = detailsRef.current;
				if (node) node.open = false;
				setIsOpen(false);
			}
		};

		document.addEventListener('mousedown', handlePointerDown);
		document.addEventListener('touchstart', handlePointerDown, { passive: true });
		document.addEventListener('keydown', handleKeyDown);
		return () => {
			document.removeEventListener('mousedown', handlePointerDown);
			document.removeEventListener('touchstart', handlePointerDown);
			document.removeEventListener('keydown', handleKeyDown);
		};
	}, [isOpen]);

	return (
		<div className="bg-base-200 text-neutral-custom flex items-center justify-between rounded-xl px-2 py-0 text-xs">
			<span>
				Input tips: &nbsp;&nbsp; '/' invoke templates or commands &nbsp;&nbsp;&nbsp;&nbsp; '+' add tools
				&nbsp;&nbsp;&nbsp;&nbsp; '#' attach context (docs/files/url/git-pr/etc.)
			</span>

			<div className="flex items-center">
				<details
					ref={detailsRef}
					className="dropdown dropdown-top dropdown-end"
					open={isOpen}
					onToggle={e => {
						setIsOpen((e.currentTarget as HTMLDetailsElement).open);
					}}
				>
					<summary
						className="btn btn-xs text-neutral-custom flex items-center gap-2 overflow-hidden border-none text-left shadow-none"
						title="Additional input tips"
						aria-expanded={isOpen}
					>
						<span className="text-xs font-normal">Additional input tips</span>
						{isOpen ? <FiChevronDown size={14} className="shrink-0" /> : <FiChevronUp size={14} className="shrink-0" />}
					</summary>

					{/* Remove 'menu' class to avoid extra spacing; allow full height */}
					<ul className="dropdown-content bg-base-100 max-w-[80vw] min-w-[20vw] list-none overflow-visible rounded-xl p-0 shadow-xl">
						{tips.map(tip => (
							<li key={tip.key} className="m-0 p-0 text-xs">
								<div className="tooltip tooltip-top w-full" data-tip={tip.description}>
									<div className="hover:bg-base-200 flex cursor-default items-center gap-2 rounded-lg p-2">
										<span className="text-left font-medium">{tip.title}</span>
										<FiMoreHorizontal size={14} className="ml-auto opacity-60" aria-hidden="true" />
									</div>
								</div>
							</li>
						))}
					</ul>
				</details>
			</div>
		</div>
	);
}
