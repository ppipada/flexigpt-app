"use client";

import ThemeSwitch from "@/components/ThemeSwitch";
import React, { ChangeEvent, useState } from "react";

const SettingsPage: React.FC = () => {
  const [apiKey, setApiKey] = useState("");
  const [defaultModel, setDefaultModel] = useState("");

  const [defaultTemperature, setDefaultTemperature] = useState<number | string>(
    0.1
  );

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    // Allow the input to be empty
    if (value === "") {
      setDefaultTemperature("");
    } else {
      // Convert the input to a number and update the state
      setDefaultTemperature(Number(value));
    }
  };
  return (
    <div className="flex flex-col items-center justify-center p-2">
      <div className="w-11/12 lg:w-2/3 mb-8 mt-8">
        <div className="flex justify-center items-center h-full">
          <h1 className="text-2xl font-bold mb-16">Settings</h1>
        </div>

        <div className="bg-base-100 rounded-lg shadow-lg p-6 mb-6">
          <div className="grid grid-cols-12 items-center">
            <h3 className="col-span-3 text-xl font-semibold">Theme</h3>
            <div className="col-span-9">
              <ThemeSwitch />
            </div>
          </div>
        </div>

        <div className="bg-base-100 rounded-lg shadow-lg p-6">
          <h3 className="text-xl font-semibold mb-8">OpenAI</h3>
          <div className="grid grid-cols-12 gap-4 mb-4 items-center">
            <label className="col-span-3 text-sm font-medium">API Key</label>
            <input
              type="password"
              className="input input-bordered col-span-9 w-full"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-12 gap-4 mb-4 items-center">
            <label className="col-span-3 text-sm font-medium">
              Default Model
            </label>
            <input
              type="text"
              className="input input-bordered col-span-9 w-full"
              value={defaultModel}
              onChange={(e) => setDefaultModel(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-12 gap-4 mb-4 items-center">
            <label className="col-span-3 text-sm font-medium">
              Default Temperature
            </label>
            <input
              type="number"
              step="0.01"
              min="0"
              max="1"
              className="input input-bordered col-span-9 w-full"
              value={defaultTemperature}
              onChange={handleChange}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;
