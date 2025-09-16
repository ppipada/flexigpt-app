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
					className="btn drawer-button fixed top-4 left-4 z-10 bg-transparent py-2 pr-0 pl-1 shadow-none md:hidden"
				>
					<FiMenu className="h-6 w-6" aria-label="Open" title="Open" />
				</label>
				{/* Page content here */}
				<div>{children}</div>
			</div>
			<div className="drawer-side z-10">
				<label htmlFor="my-drawer" className="drawer-overlay"></label>
				<ul className="menu bg-base-300 text-base-content ms-0 h-full w-12 justify-between ps-0">
					<div className="flex-col p-0">
						<li className="mt-16" title="Home" aria-label="Home" onClick={toggle}>
							<Link to="/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
								<FiHome size={24} />
							</Link>
						</li>
						<li className="mt-4" title="Chats" onClick={toggle} aria-label="Chats">
							<Link to="/chats/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
								<FiMessageSquare size={24} />
							</Link>
						</li>
						{FEATURE_FLAG_AGENTS && (
							<li className="mt-4" title="Agents" onClick={toggle} aria-label="Agents">
								<Link to="/agents/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
									<FiCpu size={24} />
								</Link>
							</li>
						)}
					</div>
					<div className="mb-8 flex-col p-0">
						{FEATURE_FLAG_DOCUMENT_STORES && (
							<li className="mt-4" title="Document Stores" onClick={toggle} aria-label="Document Stores">
								<Link to="/docstores/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
									<FiDatabase size={24} />
								</Link>
							</li>
						)}
						<li className="mt-4" title="Prompts" onClick={toggle} aria-label="Prompts">
							<Link to="/prompts/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
								<FiFilePlus size={24} />
							</Link>
						</li>
						<li className="mt-4" title="Model Presets" onClick={toggle} aria-label="Model Presets">
							<Link to="/modelpresets/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
								<FiSliders size={24} />
							</Link>
						</li>
						<li className="mt-4" title="Settings" onClick={toggle} aria-label="Settings">
							<Link to="/settings/" className="flex h-12 w-12 items-center justify-center rounded-lg p-0">
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
