'use client';

import DeleteCollectionModal from '@/docstores/deletecollection';
import DeleteDocStoreModal from '@/docstores/deletestore';
import ModifyCollection from '@/docstores/modifycollection';
import ModifyDocStore from '@/docstores/modifystore';
import { DOCUMENT_COLLECTION_INVOKE_CHAR } from '@/models/commands';
import { Collection, DocStoreServer } from '@/models/docstoremodel';
import { useEffect, useState } from 'react';
import { FiEdit, FiFolder, FiPlus, FiServer, FiTrash2 } from 'react-icons/fi';

// Mock function to fetch servers
const fetchServers = async (): Promise<DocStoreServer[]> => {
	await new Promise(resolve => setTimeout(resolve, 10));

	return [
		{
			id: '1',
			name: 'KBsrv  1',
			url: 'https://kb1.example.com',
			collections: [
				{ id: '1', name: 'Collection 1', command: 'c1', documentCount: 100, chunkCount: 1000 },
				{ id: '2', name: 'Collection 2', command: 'c2', documentCount: 50, chunkCount: 500 },
			],
			status: 'online',
			description: 'Primary knowledge base server',
			dbName: 'kb_main',
		},
		{
			id: '2',
			name: 'KBsrv 2',
			url: 'https://kb2.example.com',
			collections: [{ id: '3', name: 'Collection 3', command: 'c3', documentCount: 75, chunkCount: 750 }],
			status: 'offline',
			description: 'Backup knowledge base server',
			dbName: 'kb_backup',
		},
	];
};

const DocumentStores: React.FC = () => {
	// State declarations
	const [servers, setServers] = useState<DocStoreServer[]>([]);
	const [loading, setLoading] = useState(true);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [selectedServer, setSelectedServer] = useState<DocStoreServer | null>(null);
	const [isDeleteCollectionModalOpen, setIsDeleteCollectionModalOpen] = useState(false);
	const [selectedCollection, setSelectedCollection] = useState<{
		serverId: string;
		collection: Collection | null;
	}>({
		serverId: '',
		collection: null,
	});
	const [isModifyCollectionModalOpen, setIsModifyCollectionModalOpen] = useState(false);
	const [error, setError] = useState<string | null>(null);

	// Effect to load servers on component mount
	useEffect(() => {
		const loadServers = async () => {
			try {
				const data = await fetchServers();
				setServers(data);
			} catch (error) {
				console.error('Failed to fetch servers:', error);
				setError('Failed to load servers. Please try again later.');
			} finally {
				setLoading(false);
			}
		};

		loadServers();
	}, []);

	// Handler for deleting a server
	const handleDeleteServer = (server: DocStoreServer) => {
		setSelectedServer(server);
		setIsDeleteModalOpen(true);
	};

	// Handler for editing a server
	const handleEditServer = (server: DocStoreServer) => {
		setSelectedServer(server);
		setIsModifyModalOpen(true);
	};

	// Handler for adding a new server
	const handleAddServer = () => {
		setSelectedServer(null);
		setIsModifyModalOpen(true);
	};

	// Confirmation handler for deleting a server
	const confirmDeleteDocStore = () => {
		if (selectedServer) {
			setServers(servers.filter(s => s.id !== selectedServer.id));
		}
		setIsDeleteModalOpen(false);
		setSelectedServer(null);
	};

	// Handler for modifying or adding a server
	const handleModifyDocStoreSubmit = (storeData: Partial<DocStoreServer>) => {
		if (selectedServer) {
			// Edit existing server
			setServers(servers.map(s => (s.id === selectedServer.id ? { ...s, ...storeData } : s)));
		} else {
			// Add new server
			const newServer: DocStoreServer = {
				id: Date.now().toString(),
				collections: [],
				status: 'offline',
				...storeData,
			} as DocStoreServer;
			setServers([...servers, newServer]);
		}
		setIsModifyModalOpen(false);
		setSelectedServer(null);
	};

	// Handler for deleting a collection
	const handleDeleteCollection = (serverId: string, collection: Collection) => {
		setSelectedCollection({ serverId, collection });
		setIsDeleteCollectionModalOpen(true);
	};

	// Confirmation handler for deleting a collection
	const confirmDeleteCollection = () => {
		if (selectedCollection.serverId && selectedCollection.collection) {
			setServers(
				servers.map(server =>
					server.id === selectedCollection.serverId
						? { ...server, collections: server.collections.filter(c => c.id !== selectedCollection.collection?.id) }
						: server
				)
			);
		}
		setIsDeleteCollectionModalOpen(false);
		setSelectedCollection({ serverId: '', collection: null });
	};

	// Handler for adding a new collection
	const handleAddCollection = (serverId: string) => {
		setSelectedCollection({ serverId, collection: null });
		setIsModifyCollectionModalOpen(true);
	};

	// Handler for editing a collection
	const handleEditCollection = (serverId: string, collection: Collection) => {
		setSelectedCollection({ serverId, collection });
		setIsModifyCollectionModalOpen(true);
	};

	// Handler for modifying or adding a collection
	const handleModifyCollectionSubmit = (collectionData: Partial<Collection>, serverId: string) => {
		setServers(
			servers.map(server => {
				if (server.id === serverId) {
					if (selectedCollection.collection) {
						// Edit existing collection
						return {
							...server,
							collections: server.collections.map(c =>
								c.id === selectedCollection.collection?.id ? { ...c, ...collectionData } : c
							),
						};
					} else {
						// Add new collection
						const newCollection: Collection = {
							id: Date.now().toString(),
							documentCount: 0,
							chunkCount: 0,
							...collectionData,
						} as Collection;
						return {
							...server,
							collections: [...server.collections, newCollection],
						};
					}
				}
				return server;
			})
		);
		setIsModifyCollectionModalOpen(false);
		setSelectedCollection({ serverId: '', collection: null });
	};

	// Loading state
	if (loading) {
		return (
			<div className="flex justify-center items-center h-screen">
				<span className="loading loading-spinner loading-lg"></span>
			</div>
		);
	}

	// Error state
	if (error) {
		return (
			<div className="flex justify-center items-center h-screen">
				<div className="alert alert-error">
					<span>{error}</span>
				</div>
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
					<button className="btn btn-lg btn-ghost rounded-2xl" onClick={handleAddServer}>
						<FiPlus size={24} />
					</button>
				</div>
			</div>

			{/* DocStoreServer Cards */}
			<div className="space-y-4">
				{servers.map(server => (
					<div key={server.id} className="card bg-base-100 shadow-xl rounded-2xl">
						<div className="card-body p-4">
							{/* DocStoreServer Info Row */}
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

							{/* DocStoreServer URL and Action Buttons */}
							<div className="flex text-neutral items-center justify-between text-sm">
								<span>{server.url}</span>
								<div>
									<button className="btn btn-sm btn-ghost rounded-2xl" onClick={() => handleAddCollection(server.id)}>
										<FiPlus /> Add Collection
									</button>
									<button className="btn btn-sm btn-ghost rounded-2xl" onClick={() => handleEditServer(server)}>
										<FiEdit />
									</button>
									<button className="btn btn-sm btn-ghost rounded-2xl" onClick={() => handleDeleteServer(server)}>
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
											<th>Use Command</th>
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
												<td>
													{DOCUMENT_COLLECTION_INVOKE_CHAR}
													{collection.command}
												</td>
												<td>{collection.documentCount}</td>
												<td>{collection.chunkCount}</td>
												<td
													className={
														index === server.collections.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'
													}
												>
													<button
														className="btn btn-sm btn-ghost rounded-2xl"
														aria-label="Edit Collection"
														onClick={() => handleEditCollection(server.id, collection)}
													>
														<FiEdit />
													</button>
													<button
														className="btn btn-sm btn-ghost rounded-2xl"
														aria-label="Delete Collection"
														onClick={() => handleDeleteCollection(server.id, collection)}
													>
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

			{/* Delete Confirmation Modal */}
			<DeleteDocStoreModal
				isOpen={isDeleteModalOpen}
				onClose={() => {
					setIsDeleteModalOpen(false);
					setSelectedServer(null);
				}}
				onConfirm={confirmDeleteDocStore}
				storeName={selectedServer?.name || ''}
			/>

			{/* Modify Store Modal */}
			<ModifyDocStore
				isOpen={isModifyModalOpen}
				onClose={() => setIsModifyModalOpen(false)}
				onSubmit={handleModifyDocStoreSubmit}
				initialData={selectedServer || undefined}
				existingServers={servers}
			/>

			{/* Delete Collection Modal */}
			<DeleteCollectionModal
				isOpen={isDeleteCollectionModalOpen}
				onClose={() => {
					setIsDeleteCollectionModalOpen(false);
					setSelectedCollection({ serverId: '', collection: null });
				}}
				onConfirm={confirmDeleteCollection}
				collectionName={selectedCollection?.collection?.name || ''}
			/>

			{/* Modify Collection Modal */}
			<ModifyCollection
				isOpen={isModifyCollectionModalOpen}
				onClose={() => setIsModifyCollectionModalOpen(false)}
				onSubmit={handleModifyCollectionSubmit}
				initialData={selectedCollection.collection || undefined}
				serverId={selectedCollection.serverId}
				existingCollections={servers.find(s => s.id === selectedCollection.serverId)?.collections || []}
			/>
		</div>
	);
};

export default DocumentStores;
