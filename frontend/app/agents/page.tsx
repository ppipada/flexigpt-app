import type { FC } from 'react';

import PageFrame from '@/components/page_frame';

const AgentPage: FC = () => {
	return (
		<PageFrame>
			<div className="flex h-full w-full flex-col items-center justify-center">
				<h1 className="text-xl">Agent Page</h1>
			</div>
		</PageFrame>
	);
};

export default AgentPage;
