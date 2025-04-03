import { loadProviderSettings } from '@/backendapihelper/settings_helper';
import { FEATURE_FLAG_AGENTS } from '@/lib/features';
import { Link } from 'react-router';

export async function clientLoader() {
	// Wait for DOM content to be loaded and Wails runtime to be injected
	if (document.readyState !== 'complete' && document.readyState !== 'interactive') {
		await new Promise(resolve => {
			document.addEventListener('DOMContentLoaded', resolve, { once: true });
		});
	}

	// Now it's safe to call Wails backend functions
	return loadProviderSettings();
}

export default function Home() {
	return (
		<main className="flex flex-col h-full items-center justify-center p-24">
			<div className="flex flex-row items-center mb-10">
				<img src="/icon.png" alt="FlexiGPT Icon" width={64} height={64} />
				<h1 className="text-2xl font-bold m-8">FlexiGPT</h1>
			</div>
			<div className="flex flex-wrap gap-6 justify-center h-36">
				<Link to="/chats/">
					<div className="cursor-pointer text-center p-6 bg-base-100 rounded-lg shadow-lg transition-transform transform hover:scale-105 h-full">
						<h3 className="text-2xl font-semibold mb-3">Chat with AI</h3>
						<p>Interact with LLMs and get assistance.</p>
						<h3 className="text-2xl font-semibold mt-2">
							<span className="inline-block ml-4 transition-transform transform group-hover:translate-x-1">-&gt;</span>
						</h3>
					</div>
				</Link>
				{FEATURE_FLAG_AGENTS && (
					<Link to="/agents/">
						<div className="cursor-pointer p-6 text-center bg-base-100 rounded-lg shadow-lg transition-transform transform hover:scale-105 h-full">
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
