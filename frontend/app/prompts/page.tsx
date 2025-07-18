import { useState } from 'react';

import PromptTemplates from '@/prompts/prompt_bundles_page';
import Tools from '@/prompts/tool';

const Prompts: React.FC = () => {
	const [activeTab, setActiveTab] = useState('prompts');

	return (
		<div className="flex flex-col w-full h-full overflow-hidden">
			<div role="tablist" className="tabs tabs-bordered tabs-lg fixed">
				<a
					role="tab"
					className={`tab ${activeTab === 'prompts' ? 'tab-active font-bold' : 'font-bold'}`}
					onClick={() => {
						setActiveTab('prompts');
					}}
				>
					Prompts
				</a>
				<a
					role="tab"
					className={`tab ${activeTab === 'tools' ? 'tab-active font-bold' : 'font-bold'}`}
					onClick={() => {
						setActiveTab('tools');
					}}
				>
					Tools
				</a>
			</div>

			<div className="flex-1 overflow-y-auto mt-16 overscroll-y-contain">
				{activeTab === 'prompts' && <PromptTemplates />}
				{activeTab === 'tools' && <Tools />}
			</div>
		</div>
	);
};

export default Prompts;
