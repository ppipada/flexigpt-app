import Image from 'next/image';
import Link from 'next/link';

export default function Home() {
	return (
		<main className="flex flex-col h-full items-center justify-center p-24">
			<div className="flex flex-row items-center mb-10">
				<Image src="/icon.png" alt="FlexiUI Icon" width={64} height={64} />
				<h1 className="text-2xl font-bold m-8">FlexiGPT UI</h1>
			</div>
			<div className="grid gap-6 sm:grid-cols-2 h-36">
				<Link href="/chats">
					<div className="cursor-pointer text-center p-6 bg-base-100 rounded-lg shadow-lg transition-transform transform hover:scale-105 h-full">
						<h3 className="text-2xl font-semibold mb-3">Chat with AI</h3>
						<p>Interact with LLMs and get assistance.</p>
						<h3 className="text-2xl font-semibold mt-2">
							<span className="inline-block ml-4 transition-transform transform group-hover:translate-x-1">-&gt;</span>
						</h3>
					</div>
				</Link>
				<Link href="/agents">
					<div className="cursor-pointer p-6 text-center bg-base-100 rounded-lg shadow-lg transition-transform transform hover:scale-105 h-full">
						<h3 className="text-2xl font-semibold mb-3">Explore Agents</h3>
						<p>Discover agents and their functionalities.</p>
						<h3 className="text-2xl font-semibold mt-2">
							<span className="inline-block ml-4 transition-transform transform group-hover:translate-x-1">-&gt;</span>
						</h3>
					</div>
				</Link>
			</div>
		</main>
	);
}
