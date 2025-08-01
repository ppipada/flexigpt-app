import type { FC, PropsWithChildren } from 'react';

interface LoaderProps {
	text?: string; // caption under the spinner
	full?: boolean; // occupy full viewport height (default true)
	className?: string; // extra classes supplied by caller
}

const Loader: FC<PropsWithChildren<LoaderProps>> = ({ text = 'Loading…', full = true, className = '' }) => {
	// build classes without clsx
	const base = 'flex flex-col items-center justify-center gap-4 w-full';
	const height = full ? 'h-screen' : 'h-full';
	const combined = `${base} ${height} ${className}`.trim();

	return (
		<div id="loading-splash" className={combined}>
			<span id="loading-splash-spinner" className="loading loading-dots loading-xl text-primary-content" />
			{text && <span className="text-sm opacity-80">{text}</span>}
		</div>
	);
};

export default Loader;
