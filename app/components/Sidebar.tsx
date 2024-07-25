"use client";

import Link from "next/link";
import { ReactNode, useState } from "react";
import {
  FiCpu,
  FiHome,
  FiMenu,
  FiMessageSquare,
  FiSettings,
} from "react-icons/fi";

interface SidebarProps {
  children: ReactNode;
}

const Sidebar: React.FC<SidebarProps> = ({ children }) => {
  const [isDrawerOpen, setDrawerOpen] = useState(false);
  const toggle = () => setDrawerOpen(!isDrawerOpen);
  const setOpen = () => setDrawerOpen(true);
  return (
    <div className="drawer md:drawer-open h-screen">
      <input
        id="my-drawer"
        type="checkbox"
        className="drawer-toggle md:hidden"
        checked={isDrawerOpen}
        onChange={toggle}
      />
      <div className="drawer-content flex flex-col">
        {/* Hamburger menu button */}
        <label
          htmlFor="my-drawer"
          className="btn drawer-button md:hidden bg-transparent shadow-none fixed top-3 left-2 py-1 pl-1 pr-0 z-10"
        >
          <FiMenu className="w-6 h-6" aria-label="Open drawer" />
        </label>
        {/* Page content here */}
        <div className="flex-1 m-2 rounded-xl bg-base-200">{children}</div>
      </div>
      <div className="drawer-side z-20">
        <label htmlFor="my-drawer" className="drawer-overlay"></label>
        <ul className="menu justify-between h-full w-20 bg-base-300 text-base-content">
          <div>
            <li className="mt-16" title="Home" onClick={toggle}>
              <Link
                href="/"
                className="flex items-center justify-center rounded-lg"
                aria-label="ChatsHome"
              >
                <FiHome className="w-6 h-6" />
              </Link>
            </li>
            <li className="mt-4" title="Chats" onClick={toggle}>
              <Link
                href="/chats"
                className="flex items-center justify-center rounded-lg"
                aria-label="Chats"
              >
                <FiMessageSquare className="w-6 h-6" />
              </Link>
            </li>
            <li className="mt-4" title="Agents" onClick={toggle}>
              <Link
                href="/agents"
                className="flex items-center justify-center rounded-lg"
                aria-label="Agents"
              >
                <FiCpu className="w-6 h-6" />
              </Link>
            </li>
          </div>
          <li className="mb-16" title="Settings" onClick={toggle}>
            <Link
              href="/settings"
              className="flex items-center justify-center rounded-lg"
              aria-label="Settings"
            >
              <FiSettings className="w-6 h-6" />
            </Link>
          </li>
        </ul>
      </div>
    </div>
  );
};

export default Sidebar;
