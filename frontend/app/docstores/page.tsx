import { useEffect, useState } from 'react';

import { FiEdit2, FiFolder, FiPlus, FiServer, FiTrash2 } from 'react-icons/fi';

import { DOCUMENT_COLLECTION_INVOKE_CHAR } from '@/spec/command';
import type { Collection, DocStore } from '@/spec/docstore';

import DeleteConfirmationModal from '@/components/delete_confirmation';
import PageFrame from '@/components/page_frame';

import ModifyCollection from '@/docstores/collection_modify';
import ModifyDocStore from '@/docstores/docstore_modify';

// Mock function to fetch docStores
const fetchDocStores = async (): Promise<DocStore[]> => {
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
			description: 'Primary knowledge base',
			dbName: 'kb_main',
		},
		{
			id: '2',
			name: 'KBsrv 2',
			url: 'https://kb2.example.com',
			collections: [{ id: '3', name: 'Collection 3', command: 'c3', documentCount: 75, chunkCount: 750 }],
			status: 'offline',
			description: 'Backup knowledge base',
			dbName: 'kb_backup',
		},
	];
};

const DocumentStores: React.FC = () => {
	// State declarations
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	const [docStores, setDocStores] = useState<DocStore[]>([]);
	const [isDeleteDocStoreModalOpen, setIsDeleteDocStoreModalOpen] = useState(false);
	const [isModifyDocStoreModalOpen, setIsModifyDocStoreModalOpen] = useState(false);
	const [selectedDocStore, setSelectedDocStore] = useState<DocStore | null>(null);

	const [isDeleteCollectionModalOpen, setIsDeleteCollectionModalOpen] = useState(false);
	const [selectedCollection, setSelectedCollection] = useState<{
		docStoreID: string;
		collection: Collection | null;
	}>({
		docStoreID: '',
		collection: null,
	});
	const [isModifyCollectionModalOpen, setIsModifyCollectionModalOpen] = useState(false);

	// Effect to load docStores on component mount
	useEffect(() => {
		const loadDocStores = async () => {
			try {
				const data = await fetchDocStores();
				setDocStores(data);
			} catch (error) {
				console.error('Failed to fetch doc stores:', error);
				setError('Failed to load doc stores. Please try again later.');
			} finally {
				setLoading(false);
			}
		};

		loadDocStores();
	}, []);

	const openDeleteDocStoreModal = (docStore: DocStore) => {
		setSelectedDocStore(docStore);
		setIsDeleteDocStoreModalOpen(true);
	};

	const closeDeleteDocStoreModal = () => {
		setIsDeleteDocStoreModalOpen(false);
		setSelectedDocStore(null);
	};

	const handleDeleteDocStore = () => {
		if (selectedDocStore) {
			setDocStores(docStores.filter(s => s.id !== selectedDocStore.id));
		}
		closeDeleteDocStoreModal();
	};

	// Handler for editing a docStore
	const handleEditDocStore = (docStore: DocStore) => {
		setSelectedDocStore(docStore);
		setIsModifyDocStoreModalOpen(true);
	};

	// Handler for adding a new docStore
	const handleAddDocStore = () => {
		setSelectedDocStore(null);
		setIsModifyDocStoreModalOpen(true);
	};

	// Handler for modifying or adding a docStore
	const handleModifyDocStoreSubmit = (storeData: Partial<DocStore>) => {
		if (selectedDocStore) {
			// Edit existing docStore
			setDocStores(docStores.map(s => (s.id === selectedDocStore.id ? { ...s, ...storeData } : s)));
		} else {
			// Add new docStore
			const newDocStore: DocStore = {
				id: Date.now().toString(),
				collections: [],
				status: 'offline',
				...storeData,
			} as DocStore;
			setDocStores([...docStores, newDocStore]);
		}
		setIsModifyDocStoreModalOpen(false);
		setSelectedDocStore(null);
	};

	const openDeleteCollectionModal = (docStoreID: string, collection: Collection) => {
		setSelectedCollection({ docStoreID, collection });
		setIsDeleteCollectionModalOpen(true);
	};

	const closeDeleteCollectionModal = () => {
		setIsDeleteCollectionModalOpen(false);
		setSelectedCollection({ docStoreID: '', collection: null });
	};

	// Confirmation handler for deleting a collection
	const handleDeleteCollection = () => {
		if (selectedCollection.docStoreID && selectedCollection.collection) {
			setDocStores(
				docStores.map(docStore =>
					docStore.id === selectedCollection.docStoreID
						? { ...docStore, collections: docStore.collections.filter(c => c.id !== selectedCollection.collection?.id) }
						: docStore
				)
			);
		}
		closeDeleteCollectionModal();
	};

	// Handler for adding a new collection
	const handleAddCollection = (docStoreID: string) => {
		setSelectedCollection({ docStoreID, collection: null });
		setIsModifyCollectionModalOpen(true);
	};

	// Handler for editing a collection
	const handleEditCollection = (docStoreID: string, collection: Collection) => {
		setSelectedCollection({ docStoreID, collection });
		setIsModifyCollectionModalOpen(true);
	};

	// Handler for modifying or adding a collection
	const handleModifyCollectionSubmit = (collectionData: Partial<Collection>, docStoreID: string) => {
		setDocStores(
			docStores.map(docStore => {
				if (docStore.id === docStoreID) {
					if (selectedCollection.collection) {
						// Edit existing collection
						return {
							...docStore,
							collections: docStore.collections.map(c =>
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
							...docStore,
							collections: [...docStore.collections, newCollection],
						};
					}
				}
				return docStore;
			})
		);
		setIsModifyCollectionModalOpen(false);
		setSelectedCollection({ docStoreID: '', collection: null });
	};

	// Loading state
	if (loading) {
		return (
			<div className="flex h-screen items-center justify-center">
				<span className="loading loading-spinner loading-sm"></span>
			</div>
		);
	}

	// Error state
	if (error) {
		return (
			<div className="flex h-screen items-center justify-center">
				<div className="alert alert-error">
					<span>{error}</span>
				</div>
			</div>
		);
	}

	return (
		<PageFrame>
			<div className="container mx-auto p-4">
				{/* Heading Row */}
				<div className="mb-6 grid grid-cols-3 items-center">
					<div></div>
					<h1 className="text-center text-3xl font-bold">Document Stores</h1>
					<div className="text-right">
						<button className="btn btn-lg btn-ghost rounded-2xl" onClick={handleAddDocStore}>
							<FiPlus size={20} />
						</button>
					</div>
				</div>

				{/* DocStore Cards */}
				<div className="space-y-4">
					{docStores.map(docStore => (
						<div key={docStore.id} className="card bg-base-100 rounded-2xl shadow-xl">
							<div className="card-body p-4">
								{/* DocStore Info Row */}
								<div className="flex items-center justify-between pr-2 pb-2">
									<div className="flex items-center space-x-2">
										<FiServer className="mr-2" />
										<span className="font-bold">
											{docStore.name} - {docStore.dbName}
										</span>
									</div>
									<div
										className={`badge-md ${docStore.status === 'online' ? 'badge-success' : 'badge-error'} rounded-2xl`}
									>
										{docStore.status}
									</div>
								</div>

								{/* DocStore URL and Action Buttons */}
								<div className="text-neutral-custom flex items-center justify-between text-sm">
									<span>{docStore.url}</span>
									<div>
										<button
											className="btn btn-sm btn-ghost rounded-2xl"
											onClick={() => {
												handleAddCollection(docStore.id);
											}}
										>
											<FiPlus /> Add Collection
										</button>
										<button
											className="btn btn-sm btn-ghost rounded-2xl"
											onClick={() => {
												handleEditDocStore(docStore);
											}}
										>
											<FiEdit2 />
										</button>
										<button
											className="btn btn-sm btn-ghost rounded-2xl"
											onClick={() => {
												openDeleteDocStoreModal(docStore);
											}}
										>
											<FiTrash2 />
										</button>
									</div>
								</div>

								{/* Collections Table */}
								<div className="overflow-x-auto">
									<table className="table-zebra table w-full">
										<thead>
											<tr className="bg-base-300 m-0 px-4 py-0 text-sm font-semibold">
												<th className="rounded-tl-2xl">Collection Name</th>
												<th>Use Command</th>
												<th>Documents</th>
												<th>Chunks</th>
												<th className="rounded-tr-2xl pr-8 text-right">Actions</th>
											</tr>
										</thead>
										<tbody>
											{docStore.collections.map((collection, index) => (
												<tr key={collection.id} className="hover:bg-base-300 border-none shadow-none">
													<td className={index === docStore.collections.length - 1 ? 'rounded-bl-2xl' : ''}>
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
															index === docStore.collections.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'
														}
													>
														<button
															className="btn btn-sm btn-ghost rounded-2xl"
															aria-label="Edit Collection"
															title="Edit Collection"
															onClick={() => {
																handleEditCollection(docStore.id, collection);
															}}
														>
															<FiEdit2 />
														</button>
														<button
															className="btn btn-sm btn-ghost rounded-2xl"
															aria-label="Delete Collection"
															title="Delete Collection"
															onClick={() => {
																openDeleteCollectionModal(docStore.id, collection);
															}}
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

				{/* DocStore Delete Confirmation Modal */}
				<DeleteConfirmationModal
					isOpen={isDeleteDocStoreModalOpen}
					onClose={closeDeleteDocStoreModal}
					onConfirm={handleDeleteDocStore}
					title="Delete Document Store"
					message={`Are you sure you want to delete the document store "${selectedDocStore?.name || ''}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>

				{/* Modify Store Modal */}
				<ModifyDocStore
					isOpen={isModifyDocStoreModalOpen}
					onClose={() => {
						setIsModifyDocStoreModalOpen(false);
					}}
					onSubmit={handleModifyDocStoreSubmit}
					initialData={selectedDocStore || undefined}
					existingDocStores={docStores}
				/>

				{/* Delete Collection Modal */}
				<DeleteConfirmationModal
					isOpen={isDeleteCollectionModalOpen}
					onClose={closeDeleteCollectionModal}
					onConfirm={handleDeleteCollection}
					title="Delete Collection"
					message={`Are you sure you want to delete the collection "${selectedCollection.collection?.name || ''}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>

				{/* Modify Collection Modal */}
				<ModifyCollection
					isOpen={isModifyCollectionModalOpen}
					onClose={() => {
						setIsModifyCollectionModalOpen(false);
					}}
					onSubmit={handleModifyCollectionSubmit}
					initialData={selectedCollection.collection || undefined}
					docStoreID={selectedCollection.docStoreID}
					existingCollections={docStores.find(s => s.id === selectedCollection.docStoreID)?.collections || []}
				/>
			</div>
		</PageFrame>
	);
};

export default DocumentStores;
