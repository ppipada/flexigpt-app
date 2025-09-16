import { useState } from 'react';

import PageFrame from '@/components/page_frame';

import PromptTemplates from '@/prompts/prompt_bundles_page';
import Tools from '@/prompts/tool_bundles_page';

const Prompts: React.FC = () => {
	const [activeTab, setActiveTab] = useState('prompts');

	return (
		<PageFrame>
			<div className="flex h-full w-full flex-col">
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

				<div className="mt-16 flex-1 overflow-y-auto overscroll-y-contain">
					{activeTab === 'prompts' && <PromptTemplates />}
					{activeTab === 'tools' && <Tools />}
				</div>
			</div>
		</PageFrame>
	);
};

export default Prompts;
