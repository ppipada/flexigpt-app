// components/ThemeSwitch.tsx
import { useEffect, useState } from 'react';

import { useTheme } from 'next-themes';
import { FiMoon, FiSun } from 'react-icons/fi';

export default function ThemeSwitch() {
	const [mounted, setMounted] = useState(false);
	const { setTheme, resolvedTheme } = useTheme();

	useEffect(() => {
		setMounted(true);
	}, []);

	if (!mounted)
		return (
			<div className="flex items-center justify-center w-8 h-5">
				<FiSun size={24} />
			</div>
		);

	return (
		<div className="flex items-center">
			<FiSun
				className="flex items-center justify-center w-8 h-5"
				size={24}
				onClick={() => {
					setTheme('light');
				}}
			/>
			<label className="p-1 justify-between" title="Switch Light/Dark mode">
				<input
					type="checkbox"
					className="toggle toggle-primary rounded-full mt-1 h-6"
					checked={resolvedTheme === 'dark'}
					onChange={() => {
						setTheme(resolvedTheme === 'dark' ? 'light' : 'dark');
					}}
					spellCheck="false"
				/>
			</label>
			<FiMoon
				className="flex items-center justify-center w-8 h-5"
				size={24}
				onClick={() => {
					setTheme('dark');
				}}
			/>
		</div>
	);
}
