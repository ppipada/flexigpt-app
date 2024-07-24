"use client";

import Link from "next/link";
import { ReactNode } from "react";
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
  return (
    <div className="drawer md:drawer-open h-screen">
      <input
        id="my-drawer"
        type="checkbox"
        className="drawer-toggle md:hidden"
      />
      <div className="drawer-content flex flex-col">
        {/* Hamburger menu button */}
        <label
          htmlFor="my-drawer"
          className="btn drawer-button md:hidden bg-transparent shadow-none fixed top-4 left-1 p-1 z-10"
        >
          <FiMenu className="w-6 h-6" aria-label="Open drawer" />
        </label>
        {/* Page content here */}
        <div className="flex-1 m-2 rounded bg-base-200">{children}</div>
      </div>
      <div className="drawer-side z-20">
        <label htmlFor="my-drawer" className="drawer-overlay"></label>
        <ul className="menu justify-between h-full w-20 bg-base-300 text-base-content">
          <div>
            <li className="mt-16" title="Home">
              <Link
                href="/"
                className="flex items-center justify-center rounded-lg"
                aria-label="ChatsHome"
              >
                <FiHome className="w-6 h-6" />
              </Link>
            </li>
            <li className="mt-4" title="Chats">
              <Link
                href="/chats"
                className="flex items-center justify-center rounded-lg"
                aria-label="Chats"
              >
                <FiMessageSquare className="w-6 h-6" />
              </Link>
            </li>
            <li className="mt-4" title="Agents">
              <Link
                href="/agents"
                className="flex items-center justify-center rounded-lg"
                aria-label="Agents"
              >
                <FiCpu className="w-6 h-6" />
              </Link>
            </li>
          </div>
          <li className="mb-16" title="Settings">
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
