'use client';

import PromptTemplates from '@/prompts/prompttemplates';
import Tools from '@/prompts/tools';
import { useState } from 'react';

const Prompts: React.FC = () => {
	const [activeTab, setActiveTab] = useState('promptTemplates');

	return (
		<div className="container mx-auto p-4">
			<div role="tablist" className="tabs tabs-bordered tabs-lg">
				<a
					role="tab"
					className={`tab ${activeTab === 'promptTemplates' ? 'tab-active font-bold' : 'font-bold'}`}
					onClick={() => setActiveTab('promptTemplates')}
				>
					Prompt Templates
				</a>
				<a
					role="tab"
					className={`tab ${activeTab === 'tools' ? 'tab-active font-bold' : 'font-bold'}`}
					onClick={() => setActiveTab('tools')}
				>
					Tools
				</a>
			</div>

			<div>
				{activeTab === 'promptTemplates' && <PromptTemplates />}
				{activeTab === 'tools' && <Tools />}
			</div>
		</div>
	);
};

export default Prompts;
