// components/ThemeSwitch.tsx

"use client";

import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { FiMoon, FiSun } from "react-icons/fi";

export default function ThemeSwitch() {
  const [mounted, setMounted] = useState(false);
  const { setTheme, resolvedTheme } = useTheme();

  useEffect(() => setMounted(true), []);

  if (!mounted)
    return (
      <div className="flex items-center justify-center w-10 h-10">
        <FiSun size={24} />
      </div>
    );

  return (
    <div className="flex items-center">
      <FiSun size={24} onClick={() => setTheme("light")} />
      <label
        className="tooltip ml-2 mr-2 p-0 justify-between"
        data-tip="Switch Light/Dark mode"
      >
        <input
          type="checkbox"
          className="toggle toggle-primary rounded-full mt-1"
          checked={resolvedTheme === "dark"}
          onChange={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
        />
      </label>
      <FiMoon size={24} onClick={() => setTheme("dark")} />
    </div>
  );
}
