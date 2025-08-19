import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { FiSearch, FiTrash2 } from 'react-icons/fi';

import type { ConversationSearchItem } from '@/spec/conversation';

import { formatDateAsString } from '@/lib/date_utils';
import { cleanSearchQuery } from '@/lib/text_utils';

import { conversationStoreAPI } from '@/apis/baseapi';

import { GroupedDropdown } from '@/components/date_grouped_dropdown';
import DeleteConfirmationModal from '@/components/delete_confirmation';

const CACHE_EXPIRY_TIME = 60_000; // 1 min

/** newest-first, title-matches before message-matches */
const sortResults = (a: SearchResult, b: SearchResult) => {
	const tA = new Date(a.searchConversation.modifiedAt).getTime();
	const tB = new Date(b.searchConversation.modifiedAt).getTime();

	if (tA !== tB) return tB - tA;
	if (a.matchType === b.matchType) return 0;
	return a.matchType === 'title' ? -1 : 1;
};

const conversationsToResults = (c: ConversationSearchItem[]): SearchResult[] =>
	c.map(conv => ({ searchConversation: conv, matchType: 'title' }));

interface SearchResult {
	searchConversation: ConversationSearchItem;
	matchType: 'title' | 'message';
	snippet?: string;
}

interface SearchState {
	query: string;
	results: SearchResult[];

	/* pagination */
	nextToken?: string;
	hasMore: boolean;

	loading: boolean;
	error?: string;

	/** `true` once the backend has been queried for *this* query        */
	searchedMessages: boolean;
}

interface SearchCacheEntry {
	results: SearchResult[];
	nextToken?: string;
	timestamp: number;
}

const searchCache = new Map<string, SearchCacheEntry>();

interface SearchDropdownProps {
	results: SearchResult[];
	loading: boolean;
	error?: string;
	hasMore: boolean;
	onLoadMore: () => void;
	focusedIndex: number;
	onPick: (item: ConversationSearchItem) => void;
	onDelete: (item: ConversationSearchItem) => void;
	query: string;
	showSearchAllHintShortQuery: boolean;
	currentConversationId: string;
}

const SearchDropdown: FC<SearchDropdownProps> = ({
	results,
	loading,
	error,
	hasMore,
	onLoadMore,
	focusedIndex,
	onPick,
	onDelete,
	query,
	showSearchAllHintShortQuery,
	currentConversationId,
}) => {
	const scrollRef = useRef<HTMLDivElement>(null);

	const shouldGroup = query.length === 0 || showSearchAllHintShortQuery;

	/* infinite scroll ------------------------------------------------ */
	const handleScroll = useCallback(() => {
		const c = scrollRef.current;
		if (!c || !hasMore || loading) return;
		if ((c.scrollTop + c.clientHeight) / c.scrollHeight >= 0.8) onLoadMore();
	}, [hasMore, loading, onLoadMore]);

	/* keep focused row in view -------------------------------------- */
	useEffect(() => {
		if (focusedIndex < 0) return;
		const el = scrollRef.current?.querySelector(`[data-index="${focusedIndex}"]`);
		el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
	}, [focusedIndex]);

	/* error ---------------------------------------------------------- */
	if (error) {
		return (
			<div className="bg-base-200 absolute right-0 left-0 mt-1 rounded-2xl p-4 text-center shadow-lg">
				<p className="text-error mb-2 text-sm">{error}</p>
				<button
					className="btn btn-sm btn-primary"
					onClick={() => {
						window.location.reload();
					}}
				>
					Retry
				</button>
			</div>
		);
	}

	/* status / hint bar --------------------------------------------- */
	const SEARCHED_ALL = 'Searched all titles & messages';
	const PRESS_ENTER_TO_SEARCH = 'Press Enter to search messages';
	const NO_CONVO = 'No conversations yet';

	const getTopBarContent = () => {
		let left = '';
		let right = '';

		if (loading && !results.length) {
			left = 'Searching...';
		} else if (!results.length && !loading) {
			if (query.length < 3) {
				if (showSearchAllHintShortQuery) {
					left = query ? `No title results for "${query} "` : NO_CONVO;
					right = PRESS_ENTER_TO_SEARCH;
				} else {
					left = query ? `No results for "${query}"` : NO_CONVO;
					right = SEARCHED_ALL;
				}
			} else {
				left = query ? `No results for "${query}"` : NO_CONVO;
				right = query ? SEARCHED_ALL : '';
			}
		} else {
			/* we have some results */
			if (query.length === 0) {
				left = 'Titles';
				right = 'Type to search';
			} else if (query.length < 3) {
				if (showSearchAllHintShortQuery) {
					left = 'Title matches';
					right = PRESS_ENTER_TO_SEARCH;
				} else {
					left = 'Title & message matches';
					right = SEARCHED_ALL;
				}
			} else {
				left = 'Title & message matches';
			}
		}
		return { left, right };
	};

	const { left: barLeft, right: barRight } = getTopBarContent();

	/* render --------------------------------------------------------- */
	return (
		<div className="relative">
			<div className="bg-base-200 absolute right-0 left-0 mt-1 overflow-hidden rounded-2xl shadow-lg">
				{/* sticky status / hint bar -------------------------------- */}
				<div className="text-neutral-custom border-base-300 sticky top-0 flex items-center justify-between border-b px-8 py-1 text-xs">
					<span className="truncate">{barLeft}</span>
					{barRight && <span className="shrink-0 pl-4">{barRight}</span>}
				</div>

				{/* results -------------------------------------------------- */}
				{!results.length && !loading ? (
					<div className="text-neutral-custom py-8 text-center text-sm">
						{query ? 'Try refining your search' : 'Start a conversation to see it here'}
					</div>
				) : (
					<div ref={scrollRef} className="max-h-[60vh] overflow-y-auto antialiased" onScroll={handleScroll}>
						{shouldGroup ? (
							<GroupedDropdown<SearchResult>
								items={results}
								focused={focusedIndex}
								getDate={r => new Date(r.searchConversation.modifiedAt)}
								getKey={r => r.searchConversation.id}
								getLabel={r => <span className="truncate">{r.searchConversation.title}</span>}
								onPick={r => {
									onPick(r.searchConversation);
								}}
								renderItemExtra={r => (
									<span className="inline-flex items-center gap-4">
										{r.matchType === 'message' && <span className="max-w-[12rem] truncate">{r.snippet}</span>}
										<span className="whitespace-nowrap">{formatDateAsString(r.searchConversation.modifiedAt)}</span>
										{/* delete (not for active conv) */}
										{r.searchConversation.id !== currentConversationId && (
											<FiTrash2
												size={14}
												className="text-neutral-custom hover:text-error shrink-0 cursor-pointer"
												onClick={e => {
													e.stopPropagation(); // don’t trigger onPick
													onDelete(r.searchConversation);
												}}
											/>
										)}
									</span>
								)}
							/>
						) : (
							<ul className="w-full text-sm">
								{results.map((r, idx) => {
									const isFocused = idx === focusedIndex;
									return (
										<li
											key={r.searchConversation.id}
											data-index={idx}
											onClick={() => {
												onPick(r.searchConversation);
											}}
											className={`hover:bg-base-100 flex cursor-pointer items-center justify-between px-12 py-2 ${
												isFocused ? 'bg-base-100' : ''
											}`}
										>
											<span className="truncate">{r.searchConversation.title}</span>

											<span className="text-neutral-custom hidden text-xs lg:block">
												<span className="inline-flex items-center gap-4">
													{r.matchType === 'message' && <span className="max-w-[12rem] truncate">{r.snippet}</span>}
													<span className="whitespace-nowrap">
														{formatDateAsString(r.searchConversation.modifiedAt)}
													</span>
													{/* delete (not for active conv) */}
													{r.searchConversation.id !== currentConversationId && (
														<FiTrash2
															size={14}
															className="text-neutral-custom hover:text-error shrink-0 cursor-pointer"
															onClick={e => {
																e.stopPropagation();
																onDelete(r.searchConversation);
															}}
														/>
													)}
												</span>
											</span>
										</li>
									);
								})}
							</ul>
						)}

						{/* footer / loader ------------------------------------ */}
						{loading && (
							<div className="flex items-center justify-center py-4">
								<span className="text-neutral-custom text-sm">
									{results.length ? 'Loading more...' : 'Searching...'}
								</span>
								<span className="loading loading-dots loading-sm" />
							</div>
						)}
						{!loading && !hasMore && results.length > 0 && query && (
							<div className="text-neutral-custom border-base-300 border-t py-1 text-center text-xs">
								End of results
							</div>
						)}
					</div>
				)}
			</div>
		</div>
	);
};

interface ChatSearchProps {
	onSelectConversation: (item: ConversationSearchItem) => Promise<void>;
	refreshKey: number;
	currentConversationId: string;
}

const ChatSearch: FC<ChatSearchProps> = ({ onSelectConversation, refreshKey, currentConversationId }) => {
	/* state ---------------------------------------------------------- */
	const [searchState, setSearchState] = useState<SearchState>({
		query: '',
		results: [],
		loading: false,
		hasMore: false,
		searchedMessages: false,
	});
	const [show, setShow] = useState(false);
	const [focusedIndex, setFocusedIndex] = useState(-1);
	const [deleteTarget, setDeleteTarget] = useState<ConversationSearchItem | null>(null);
	const [deleteLoading, setDeleteLoading] = useState(false);

	/** all conversations fetched so far (used for local filtering) */
	const [recentConversations, setRecentConversations] = useState<ConversationSearchItem[]>([]);

	/* refs ----------------------------------------------------------- */
	const abortControllerRef = useRef<AbortController | null>(null);
	const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);

	const loadRecentConversations = useCallback(async (token?: string, append = false) => {
		try {
			if (!append) setSearchState(p => ({ ...p, loading: true, error: undefined }));

			const { conversations, nextToken } = await conversationStoreAPI.listConversations(token, 20);
			const trimmedNext = nextToken?.trim() || '';

			setRecentConversations(prev => (append ? [...prev, ...conversations] : [...conversations]));

			/* update visible list **only** when we are in “recents” mode
							(i.e. an empty query) so we don’t clobber search results.  */
			setSearchState(prev => {
				if (prev.query !== '' && !append) {
					return {
						...prev,
						loading: false,
						nextToken: trimmedNext,
						hasMore: !!trimmedNext,
					};
				}

				const newRes = conversationsToResults(
					append ? [...prev.results.map(r => r.searchConversation), ...conversations] : conversations
				);

				return {
					...prev,
					results: newRes,
					loading: false,
					nextToken: trimmedNext,
					hasMore: !!trimmedNext,
					error: undefined,
					searchedMessages: false,
				};
			});
		} catch (e) {
			console.error(e);
			setSearchState(p => ({ ...p, loading: false, error: 'Failed to load conversations' }));
		}
	}, []);

	const performSearch = useCallback(async (rawQuery: string, token?: string, append = false) => {
		if (abortControllerRef.current && !append) abortControllerRef.current.abort();
		abortControllerRef.current = new AbortController();

		const query = cleanSearchQuery(rawQuery);

		/* if cleaner strips everything -> immediate empty result-set */
		if (query === '') {
			setSearchState(p => ({
				...p,
				results: [],
				loading: false,
				hasMore: false,
				nextToken: undefined,
				error: undefined,
				searchedMessages: true,
			}));
			return;
		}

		if (!append)
			setSearchState(p => ({
				...p,
				loading: true,
				error: undefined,
				hasMore: false,
				searchedMessages: false,
			}));

		try {
			const res = await conversationStoreAPI.searchConversations(query, token, 20);
			const nextToken = res.nextToken?.trim() || '';

			const newResults: SearchResult[] = res.conversations.map(searchConv => ({
				searchConversation: searchConv,
				matchType: searchConv.title.toLowerCase().includes(rawQuery.toLowerCase()) ? 'title' : 'message',
				snippet: '',
			}));

			setSearchState(p => ({
				...p,
				results: append ? [...p.results, ...newResults] : newResults,
				nextToken,
				hasMore: !!nextToken,
				loading: false,
				searchedMessages: true,
			}));

			/* cache (first page only) -------------------------------- */
			if (!append) {
				searchCache.set(query, {
					results: newResults,
					nextToken,
					timestamp: Date.now(),
				});
				/* keep cache small (5 entries) */
				while (searchCache.size > 5) {
					const oldest = [...searchCache.entries()].reduce((a, b) => (a[1].timestamp < b[1].timestamp ? a : b))[0];
					searchCache.delete(oldest);
				}
			}
		} catch (err) {
			if ((err as DOMException).name === 'AbortError') return;

			if (!abortControllerRef.current.signal.aborted) {
				setSearchState(p => ({
					...p,
					loading: false,
					error: 'Search failed. Please try again.',
					hasMore: false,
					searchedMessages: true,
				}));
			}
		}
	}, []);

	/* Local filter for 1-2 char queries -------------- */
	const filterLocalResults = useCallback(
		(q: string) => {
			const filtered = recentConversations.filter(c => c.title.toLowerCase().includes(q.toLowerCase()));
			setSearchState(p => ({
				...p,
				results: conversationsToResults(filtered),
				loading: false,
				/* keep pagination information so recents can still load   */
				hasMore: p.hasMore,
				nextToken: p.nextToken,
				error: undefined,
				searchedMessages: false,
			}));
		},
		[recentConversations]
	);

	const handleInputChange = useCallback(
		(e: ChangeEvent<HTMLInputElement>) => {
			const q = e.target.value;
			setSearchState(p => ({ ...p, query: q, searchedMessages: false }));
			setFocusedIndex(-1);

			/* re-open dropdown if hidden                                */
			if (!show && q.trim()) setShow(true);

			if (debounceTimeoutRef.current) clearTimeout(debounceTimeoutRef.current);

			/* empty query -> show recents                               */
			if (!q.trim()) {
				setSearchState(p => ({
					...p,
					results: conversationsToResults(recentConversations),
					loading: false,
					/* keep nextToken / hasMore so scrolling continues        */
					error: undefined,
					searchedMessages: false,
				}));
				return;
			}

			/* short query (1–2 chars) -> local only                    */
			if (q.length < 3) {
				filterLocalResults(q);
				return;
			}

			/* cache --------------------------------------------------- */
			const cleaned = cleanSearchQuery(q);
			const cached = searchCache.get(cleaned);
			if (cached && Date.now() - cached.timestamp < CACHE_EXPIRY_TIME) {
				setSearchState(p => ({
					...p,
					results: cached.results,
					nextToken: cached.nextToken,
					hasMore: !!cached.nextToken,
					searchedMessages: true,
					loading: false,
					error: undefined,
				}));
				return;
			}

			/* remote search (debounced) ------------------------------- */
			debounceTimeoutRef.current = setTimeout(() => performSearch(q), 300);
		},
		[show, recentConversations, filterLocalResults, performSearch]
	);

	/* ------------------------- focus / blur ------------------------ */
	const handleFocus = useCallback(async () => {
		setShow(true);
		if (!recentConversations.length) await loadRecentConversations();
		else if (!searchState.query) {
			setSearchState(p => ({
				...p,
				results: conversationsToResults(recentConversations),
			}));
		}
	}, [recentConversations, loadRecentConversations, searchState.query]);

	const handleBlur = useCallback(() => {
		setTimeout(() => {
			setShow(false);
		}, 150);
	}, []);

	const handlePick = useCallback(
		async (conv: ConversationSearchItem) => {
			await onSelectConversation(conv);
			setSearchState(p => ({ ...p, query: '', searchedMessages: false }));
			setShow(false);
			setFocusedIndex(-1);
			inputRef.current?.blur();
		},
		[onSelectConversation]
	);

	/* ------------ delete flow -------------------------------------- */
	const handleAskDelete = useCallback((conv: ConversationSearchItem) => {
		setDeleteTarget(conv);
	}, []);

	const handleDeleteConfirm = useCallback(async () => {
		if (!deleteTarget) return;
		setDeleteLoading(true);

		try {
			await conversationStoreAPI.deleteConversation(deleteTarget.id, deleteTarget.title);

			/* prune from local caches --------------------------------*/
			setRecentConversations(prev => prev.filter(c => c.id !== deleteTarget.id));
			setSearchState(prev => ({
				...prev,
				results: prev.results.filter(r => r.searchConversation.id !== deleteTarget.id),
			}));

			/* ---- prune *ALL* cached search pages --------------------- */
			searchCache.forEach((entry, key) => {
				const filtered = entry.results.filter(r => r.searchConversation.id !== deleteTarget.id);
				if (filtered.length === 0) {
					searchCache.delete(key);
				} else if (filtered.length !== entry.results.length) {
					searchCache.set(key, { ...entry, results: filtered });
				}
			});
			/* ---- (optional) refresh the first page of recents -------- */
			// Only if the dropdown is currently showing recents.
			if (searchState.query.trim() === '') {
				loadRecentConversations(); // pull a fresh page
			}
		} catch (err) {
			console.error(err);
		} finally {
			setDeleteLoading(false);
			setDeleteTarget(null);
		}
	}, [deleteTarget]);

	const showSearchAllHintShortQuery =
		searchState.query.length > 0 &&
		searchState.query.length < 3 &&
		!searchState.loading &&
		!searchState.searchedMessages;

	const isLocalMode = searchState.query.length === 0 || showSearchAllHintShortQuery;

	const handleLoadMore = useCallback(() => {
		if (searchState.loading || !searchState.hasMore || !searchState.nextToken) return;

		if (isLocalMode) {
			loadRecentConversations(searchState.nextToken, true);
		} else {
			performSearch(searchState.query, searchState.nextToken, true);
		}
	}, [isLocalMode, searchState, performSearch, loadRecentConversations]);

	const handleKeyDown = useCallback(
		(e: KeyboardEvent<HTMLInputElement>) => {
			if (!show) return;

			switch (e.key) {
				case 'ArrowDown':
					e.preventDefault();
					if (searchState.results.length) setFocusedIndex(i => (i + 1) % searchState.results.length);
					break;
				case 'ArrowUp':
					e.preventDefault();
					if (searchState.results.length)
						setFocusedIndex(i => (i - 1 + searchState.results.length) % searchState.results.length);
					break;
				case 'Enter':
					e.preventDefault();
					if (debounceTimeoutRef.current) {
						clearTimeout(debounceTimeoutRef.current);
						debounceTimeoutRef.current = null;
					}
					if (focusedIndex >= 0 && focusedIndex < searchState.results.length) {
						handlePick(searchState.results[focusedIndex].searchConversation);
					} else if (!searchState.loading && searchState.query.trim()) {
						/* allow any non-empty query                             */
						performSearch(searchState.query);
					}
					break;
				case 'Escape':
					setShow(false);
					setFocusedIndex(-1);
					if (searchState.query) setSearchState(p => ({ ...p, query: '', searchedMessages: false }));
					break;
			}
		},
		[show, searchState, focusedIndex, handlePick, performSearch]
	);

	useEffect(() => {
		searchCache.clear();
		loadRecentConversations(); // first page
	}, [refreshKey, loadRecentConversations]);

	useEffect(
		() => () => {
			abortControllerRef.current?.abort();
			if (debounceTimeoutRef.current) clearTimeout(debounceTimeoutRef.current);
		},
		[]
	);

	const orderedResults = useMemo(
		() => (isLocalMode ? [...searchState.results].sort(sortResults) : searchState.results),
		[searchState.results, isLocalMode]
	);

	return (
		<div>
			{/* input --------------------------------------------------- */}
			<div className="bg-base-100 border-base-300 focus-within:border-base-400 flex items-center rounded-2xl border p-2 transition-colors">
				<FiSearch size={20} className="text-neutral-custom mx-3 flex-shrink-0" />
				<input
					ref={inputRef}
					type="text"
					value={searchState.query}
					onChange={handleInputChange}
					onFocus={handleFocus}
					onBlur={handleBlur}
					onKeyDown={handleKeyDown}
					placeholder="Search conversations..."
					className="placeholder:text-neutral-custom w-full bg-transparent text-sm outline-none"
					spellCheck={false}
				/>
				{searchState.loading && <span className="loading loading-dots loading-sm" />}
			</div>

			{/* dropdown ------------------------------------------------ */}
			{show && (
				<SearchDropdown
					results={orderedResults}
					loading={searchState.loading}
					error={searchState.error}
					hasMore={searchState.hasMore}
					onLoadMore={handleLoadMore}
					focusedIndex={focusedIndex}
					onPick={handlePick}
					onDelete={handleAskDelete}
					query={searchState.query}
					showSearchAllHintShortQuery={showSearchAllHintShortQuery}
					currentConversationId={currentConversationId}
				/>
			)}
			{deleteTarget && (
				<DeleteConfirmationModal
					isOpen={!!deleteTarget}
					onClose={() => {
						if (!deleteLoading) {
							setDeleteTarget(null);
						}
					}}
					onConfirm={handleDeleteConfirm}
					title="Delete conversation?"
					message={`Are you sure you want to delete "${deleteTarget.title}"? This action cannot be undone.`}
					confirmButtonText={deleteLoading ? 'Deleting…' : 'Delete'}
				/>
			)}
		</div>
	);
};

export default ChatSearch;
