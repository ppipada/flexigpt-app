import type { ReactNode } from 'react';
import { useState } from 'react';

import { FiCpu, FiDatabase, FiFilePlus, FiHome, FiMenu, FiMessageSquare, FiSettings, FiSliders } from 'react-icons/fi';
import { Link } from 'react-router';

import { FEATURE_FLAG_AGENTS, FEATURE_FLAG_DOCUMENT_STORES } from '@/lib/features';

interface SidebarProps {
	children: ReactNode;
}

const Sidebar: React.FC<SidebarProps> = ({ children }) => {
	const [isDrawerOpen, setDrawerOpen] = useState(false);
	const toggle = () => {
		setDrawerOpen(!isDrawerOpen);
	};

	// console.log(`${FEATURE_FLAG_AGENTS}`);
	// const setOpen = () => setDrawerOpen(true);
	return (
		<div className="drawer md:drawer-open h-screen">
			<input
				id="my-drawer"
				type="checkbox"
				className="drawer-toggle md:hidden"
				checked={isDrawerOpen}
				onChange={toggle}
				spellCheck="false"
			/>
			<div className="drawer-content flex flex-col overflow-auto">
				{/* Hamburger menu button */}
				<label
					htmlFor="my-drawer"
					className="btn drawer-button md:hidden bg-transparent shadow-none fixed top-4 left-4 py-2 pl-1 pr-0 z-10"
				>
					<FiMenu className="w-6 h-6" aria-label="Open" title="Open" />
				</label>
				{/* Page content here */}
				<div className="flex-1 m-2 rounded-xl bg-base-200">{children}</div>
			</div>
			<div className="drawer-side z-10">
				<label htmlFor="my-drawer" className="drawer-overlay"></label>
				<ul className="menu justify-between h-full w-12 ms-0 ps-0 bg-base-300 text-base-content">
					<div className="flex-col p-0">
						<li className="mt-16" title="Home" aria-label="Home" onClick={toggle}>
							<Link to="/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
								<FiHome size={24} />
							</Link>
						</li>
						<li className="mt-4" title="Chats" onClick={toggle} aria-label="Chats">
							<Link to="/chats/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
								<FiMessageSquare size={24} />
							</Link>
						</li>
						{FEATURE_FLAG_AGENTS && (
							<li className="mt-4" title="Agents" onClick={toggle} aria-label="Agents">
								<Link to="/agents/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
									<FiCpu size={24} />
								</Link>
							</li>
						)}
					</div>
					<div className="flex-col p-0 mb-8">
						{FEATURE_FLAG_DOCUMENT_STORES && (
							<li className="mt-4" title="Document Stores" onClick={toggle} aria-label="Document Stores">
								<Link to="/docstores/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
									<FiDatabase size={24} />
								</Link>
							</li>
						)}
						<li className="mt-4" title="Prompts" onClick={toggle} aria-label="Prompts">
							<Link to="/prompts/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
								<FiFilePlus size={24} />
							</Link>
						</li>
						<li className="mt-4" title="Model Presets" onClick={toggle} aria-label="Model Presets">
							<Link to="/modelpresets/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
								<FiSliders size={24} />
							</Link>
						</li>
						<li className="mt-4" title="Settings" onClick={toggle} aria-label="Settings">
							<Link to="/settings/" className="flex w-12 h-12 p-0 items-center justify-center rounded-lg">
								<FiSettings size={24} />
							</Link>
						</li>
					</div>
				</ul>
			</div>
		</div>
	);
};

export default Sidebar;
