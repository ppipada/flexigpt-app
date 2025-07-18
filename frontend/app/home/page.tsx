import { Link } from 'react-router';

import { FEATURE_FLAG_AGENTS } from '@/lib/features';

export async function clientLoader() {
	// Wait for DOM content to be loaded and Wails runtime to be injected
	if (document.readyState !== 'complete' && document.readyState !== 'interactive') {
		await new Promise(resolve => {
			document.addEventListener('DOMContentLoaded', resolve, { once: true });
		});
	}
	// Now it's safe to call Wails backend functions
}

export default function Home() {
	return (
		<main className="flex flex-col h-full items-center overflow-auto justify-start md:justify-center px-4 py-8">
			<div className="flex flex-row items-center mb-10">
				<img src="/icon.png" alt="FlexiGPT Icon" width={64} height={64} />
				<h1 className="text-2xl font-bold m-8">FlexiGPT</h1>
			</div>
			<div className="flex flex-col md:flex-row gap-6 h-36">
				<Link to="/chats/">
					<div className="cursor-pointer text-center p-6 bg-base-100 rounded-2xl shadow-lg transition-transform transform hover:scale-105">
						<h3 className="text-2xl font-semibold mb-3">Chat with AI</h3>
						<p>Interact with LLMs and get assistance.</p>
						<h3 className="text-2xl font-semibold mt-2">
							<span className="inline-block ml-4 transition-transform transform group-hover:translate-x-1">-&gt;</span>
						</h3>
					</div>
				</Link>
				{FEATURE_FLAG_AGENTS && (
					<Link to="/agents/">
						<div className="cursor-pointer p-6 mb-4 text-center bg-base-100 rounded-2xl shadow-lg transition-transform transform hover:scale-105">
							<h3 className="text-2xl font-semibold mb-3">Explore Agents</h3>
							<p>Discover agents and their functionalities.</p>
							<h3 className="text-2xl font-semibold mt-2">
								<span className="inline-block ml-4 transition-transform transform group-hover:translate-x-1">
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
