import { Link } from 'react-router';

import { FEATURE_FLAG_AGENTS } from '@/lib/features';

export default function Home() {
	return (
		<main className="flex h-full flex-col items-center justify-start overflow-auto px-4 py-8 md:justify-center">
			<div className="mb-10 flex flex-row items-center">
				<img src="/icon.png" alt="FlexiGPT Icon" width={64} height={64} />
				<h1 className="m-8 text-2xl font-bold">FlexiGPT</h1>
			</div>
			<div className="flex h-36 flex-col gap-6 md:flex-row">
				<Link to="/chats/">
					<div className="bg-base-100 transform cursor-pointer rounded-2xl p-6 text-center shadow-lg transition-transform hover:scale-105">
						<h3 className="mb-3 text-2xl font-semibold">Chat with AI</h3>
						<p>Interact with LLMs and get assistance.</p>
						<h3 className="mt-2 text-2xl font-semibold">
							<span className="ml-4 inline-block transform transition-transform group-hover:translate-x-1">-&gt;</span>
						</h3>
					</div>
				</Link>
				{FEATURE_FLAG_AGENTS && (
					<Link to="/agents/">
						<div className="bg-base-100 mb-4 transform cursor-pointer rounded-2xl p-6 text-center shadow-lg transition-transform hover:scale-105">
							<h3 className="mb-3 text-2xl font-semibold">Explore Agents</h3>
							<p>Discover agents and their functionalities.</p>
							<h3 className="mt-2 text-2xl font-semibold">
								<span className="ml-4 inline-block transform transition-transform group-hover:translate-x-1">
									-&gt;
								</span>
							</h3>
						</div>
					</Link>
				)}
			</div>
		</main>
	);
}
