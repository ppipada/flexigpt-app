'use client';

import { useEffect, useState } from 'react';
import { FiEdit, FiFolder, FiPlus, FiServer, FiTrash2 } from 'react-icons/fi';

interface Collection {
	id: string;
	name: string;
	documentCount: number;
	chunkCount: number;
}

interface Server {
	id: string;
	name: string;
	url: string;
	collections: Collection[];
	status: 'online' | 'offline';
	description: string;
	dbName: string;
}

const fetchServers = async (): Promise<Server[]> => {
	await new Promise(resolve => setTimeout(resolve, 500));

	return [
		{
			id: '1',
			name: 'KB Server 1',
			url: 'https://kb1.example.com',
			collections: [
				{ id: '1', name: 'Collection 1', documentCount: 100, chunkCount: 1000 },
				{ id: '2', name: 'Collection 2', documentCount: 50, chunkCount: 500 },
			],
			status: 'online',
			description: 'Primary knowledge base server',
			dbName: 'kb_main',
		},
		{
			id: '2',
			name: 'KB Server 2',
			url: 'https://kb2.example.com',
			collections: [{ id: '3', name: 'Collection 3', documentCount: 75, chunkCount: 750 }],
			status: 'offline',
			description: 'Backup knowledge base server',
			dbName: 'kb_backup',
		},
	];
};

const DocumentStores: React.FC = () => {
	const [servers, setServers] = useState<Server[]>([]);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		const loadServers = async () => {
			try {
				const data = await fetchServers();
				setServers(data);
			} catch (error) {
				console.error('Failed to fetch servers:', error);
			} finally {
				setLoading(false);
			}
		};

		loadServers();
	}, []);

	if (loading) {
		return (
			<div className="flex justify-center items-center h-screen">
				<span className="loading loading-spinner loading-lg"></span>
			</div>
		);
	}

	return (
		<div className="container mx-auto p-4">
			{/* Heading Row */}
			<div className="grid grid-cols-3 items-center mb-6">
				<div></div>
				<h1 className="text-3xl font-bold text-center">Document Stores</h1>
				<div className="text-right">
					<button className="btn btn-warning rounded-2xl">
						<FiPlus /> Add Document Store
					</button>
				</div>
			</div>

			{/* Server Cards */}
			<div className="space-y-4">
				{servers.map(server => (
					<div key={server.id} className="card bg-base-100 shadow-xl rounded-2xl">
						<div className="card-body p-4">
							{/* Server Info Row */}
							<div className="flex items-center justify-between pr-2 pb-2">
								<div className="flex items-center space-x-2">
									<FiServer className="mr-2" />
									<span className="font-bold">
										{server.name} - {server.dbName}
									</span>
								</div>
								<div className={`badge-md ${server.status === 'online' ? 'badge-success' : 'badge-error'} rounded-2xl`}>
									{server.status}
								</div>
							</div>

							{/* Server URL and Action Buttons */}
							<div className="flex text-neutral items-center justify-between text-sm">
								<span>{server.url}</span>
								<div>
									<button className="btn btn-sm btn-ghost rounded-2xl">
										<FiPlus /> Add Collection
									</button>
									<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Edit Server">
										<FiEdit />
									</button>
									<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Delete Server">
										<FiTrash2 />
									</button>
								</div>
							</div>

							{/* Collections Table */}
							<div className="overflow-x-auto">
								<table className="table table-zebra w-full">
									<thead>
										<tr className="font-semibold text-sm px-4 py-0 m-0 bg-base-300">
											<th className="rounded-tl-2xl">Collection Name</th>
											<th>Documents</th>
											<th>Chunks</th>
											<th className="text-right rounded-tr-2xl pr-8">Actions</th>
										</tr>
									</thead>
									<tbody>
										{server.collections.map((collection, index) => (
											<tr key={collection.id} className="hover border-none shadow-none">
												<td className={index === server.collections.length - 1 ? 'rounded-bl-2xl' : ''}>
													<div className="flex items-center">
														<FiFolder className="mr-2" />
														{collection.name}
													</div>
												</td>
												<td>{collection.documentCount}</td>
												<td>{collection.chunkCount}</td>
												<td
													className={
														index === server.collections.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'
													}
												>
													<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Edit Collection">
														<FiEdit />
													</button>
													<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Delete Collection">
														<FiTrash2 />
													</button>
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
							{/* End of Collections */}
						</div>
					</div>
				))}
			</div>
		</div>
	);
};

export default DocumentStores;
