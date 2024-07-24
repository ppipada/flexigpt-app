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
          className="btn drawer-button md:hidden justify-start m-4"
        >
          <FiMenu className="w-6 h-6" aria-label="Open drawer" />
        </label>
        {/* Page content here */}
        <div className="flex-1 m-2 rounded bg-base-200">{children}</div>
      </div>
      <div className="drawer-side">
        <label htmlFor="my-drawer" className="drawer-overlay"></label>
        <ul className="menu justify-between h-full w-20 bg-base-300 text-base-content">
          <div>
            <li className="tooltip mt-16" data-tip="Home">
              <Link
                href="/"
                className="flex items-center justify-center rounded-lg"
                aria-label="ChatsHome"
              >
                <FiHome className="w-6 h-6" />
              </Link>
            </li>
            <li className="tooltip mt-4" data-tip="Chats">
              <Link
                href="/chats"
                className="flex items-center justify-center rounded-lg"
                aria-label="Chats"
              >
                <FiMessageSquare className="w-6 h-6" />
              </Link>
            </li>
            <li className="tooltip mt-4" data-tip="Agents">
              <Link
                href="/agents"
                className="flex items-center justify-center rounded-lg"
                aria-label="Agents"
              >
                <FiCpu className="w-6 h-6" />
              </Link>
            </li>
          </div>
          <li className="tooltip mb-16" data-tip="Settings">
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
