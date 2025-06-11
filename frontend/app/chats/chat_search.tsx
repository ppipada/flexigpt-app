import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { FiSearch } from 'react-icons/fi';

import type { ConversationItem } from '@/models/conversationmodel';

import { conversationStoreAPI } from '@/apis/baseapi';

import { formatDateAsString } from '@/lib/date_utils';

import { GroupedDropdown } from '@/components/date_grouped_dropdown';

const CACHE_EXPIRY_TIME = 60_000;

interface SearchResult {
	conversation: ConversationItem;
	matchType: 'title' | 'message';
	snippet?: string;
}

interface SearchState {
	query: string;
	results: SearchResult[];
	nextToken?: string;
	loading: boolean;
	error?: string;
	hasMore: boolean;
	searchedMessages: boolean;
}

interface SearchCacheEntry {
	results: SearchResult[];
	nextToken?: string;
	timestamp: number;
}

const searchCache = new Map<string, SearchCacheEntry>();

// newest-first, title-matches before message-matches
const sortResults = (a: SearchResult, b: SearchResult) => {
	const tA = new Date(a.conversation.createdAt).getTime();
	const tB = new Date(b.conversation.createdAt).getTime();

	if (tA !== tB) return tB - tA;
	if (a.matchType === b.matchType) return 0;
	return a.matchType === 'title' ? -1 : 1;
};

interface SearchDropdownProps {
	results: SearchResult[];
	loading: boolean;
	error?: string;
	hasMore: boolean;
	onLoadMore: () => void;
	focusedIndex: number;
	onPick: (item: ConversationItem) => void;
	query: string;
	showSearchAllHintShortQuery: boolean;
}

const SearchDropdown: FC<SearchDropdownProps> = ({
	results,
	loading,
	error,
	hasMore,
	onLoadMore,
	focusedIndex,
	onPick,
	query,
	showSearchAllHintShortQuery,
}) => {
	const scrollRef = useRef<HTMLDivElement>(null);

	/* ------------------------ infinite scroll ----------------------- */
	const handleScroll = useCallback(() => {
		const c = scrollRef.current;
		if (!c || !hasMore || loading) return;
		if ((c.scrollTop + c.clientHeight) / c.scrollHeight >= 0.8) onLoadMore();
	}, [hasMore, loading, onLoadMore]);

	/* --------------- keep focused row in view while typing ---------- */
	useEffect(() => {
		if (focusedIndex < 0) return;
		const el = scrollRef.current?.querySelector(`[data-index="${focusedIndex}"]`);
		el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
	}, [focusedIndex]);

	/* --------------------------- error ------------------------------ */
	if (error) {
		return (
			<div className="absolute left-0 right-0 mt-1 bg-base-200 rounded-2xl shadow-lg p-4 text-center z-50">
				<p className="text-error text-sm mb-2">{error}</p>
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

	const SEARCHED_ALL = 'Searched all titles & messages';
	const PRESS_ENTER_TO_SEARCH = 'Press Enter to search messages';
	const NO_CONVO = 'No conversations yet';

	/* ---------------- top-bar content helpers ---------------------- */
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
			// we have some results
			if (query.length === 0) {
				left = 'Titles';
				right = 'Type to search';
			} else if (query.length > 0 && query.length < 3) {
				left = 'Title matches';
				if (showSearchAllHintShortQuery) {
					right = PRESS_ENTER_TO_SEARCH;
				} else {
					right = SEARCHED_ALL;
				}
			} else {
				left = 'Title & message matches';
			}
		}

		return { left, right };
	};

	const { left: barLeft, right: barRight } = getTopBarContent();

	/* --------------------------- render ----------------------------- */
	return (
		<div className="absolute left-0 right-0 mt-1 bg-base-200 rounded-2xl shadow-lg overflow-hidden z-50">
			{/* ------- sticky status / hint bar (always visible) ---------- */}
			<div className="flex justify-between items-center px-8 py-1 text-xs text-neutral/60 border-b border-base-300 sticky top-0">
				<span className="truncate">{barLeft}</span>
				{barRight && <span className="pl-4 shrink-0">{barRight}</span>}
			</div>

			{/* -------------- empty state OR results list ---------------- */}
			{!results.length && !loading ? (
				<div className="py-8 text-center text-sm text-neutral/60">
					{query ? 'Try refining your search' : 'Start a conversation to see it here'}
				</div>
			) : (
				<div ref={scrollRef} className="max-h-[60vh] overflow-y-auto" onScroll={handleScroll}>
					<GroupedDropdown<SearchResult>
						items={results}
						focused={focusedIndex}
						getDate={r => new Date(r.conversation.createdAt)}
						getKey={r => r.conversation.id}
						getLabel={r => <span className="truncate">{r.conversation.title}</span>}
						onPick={r => {
							onPick(r.conversation);
						}}
						renderItemExtra={r => (
							<span className="inline-flex items-center gap-4">
								{r.matchType === 'message' && <span className="truncate max-w-[12rem]">{r.snippet}</span>}
								<span className="whitespace-nowrap">{formatDateAsString(r.conversation.createdAt)}</span>
							</span>
						)}
					/>

					{loading && (
						<div className="flex items-center justify-center py-4">
							<span className="text-sm text-neutral/60">{results.length ? 'Loading more…' : 'Searching…'}</span>
							<span className="loading loading-dots loading-md" />
						</div>
					)}
					{!loading && !hasMore && results.length > 0 && query && (
						<div className="text-center py-1 text-xs text-neutral/60 border-t border-base-300">End of results</div>
					)}
				</div>
			)}
		</div>
	);
};

interface ChatSearchProps {
	onSelectConversation: (item: ConversationItem) => Promise<void>;
	refreshKey: number;
}

const ChatSearch: FC<ChatSearchProps> = ({ onSelectConversation, refreshKey }) => {
	/* ----------------------------- state --------------------------- */
	const [searchState, setSearchState] = useState<SearchState>({
		query: '',
		results: [],
		loading: false,
		hasMore: false,
		searchedMessages: false,
	});
	const [show, setShow] = useState(false);
	const [focusedIndex, setFocusedIndex] = useState(-1);
	const [recentConversations, setRecentConversations] = useState<ConversationItem[]>([]);

	/* ----------------------------- refs ---------------------------- */
	const abortControllerRef = useRef<AbortController | null>(null);
	const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);
	const inputRef = useRef<HTMLInputElement>(null);

	/* --------------------------- helpers --------------------------- */
	const conversationsToResults = useCallback(
		(conversations: ConversationItem[]): SearchResult[] =>
			conversations.map(c => ({ conversation: c, matchType: 'title' })),
		[]
	);

	/* ------------------------ load recents ------------------------- */
	const loadRecentConversations = useCallback(async () => {
		try {
			setSearchState(p => ({ ...p, loading: true }));
			const { conversations } = await conversationStoreAPI.listConversations();
			setRecentConversations(conversations);

			if (!searchState.query) {
				setSearchState({
					...searchState,
					results: conversationsToResults(conversations),
					loading: false,
					hasMore: false,
					searchedMessages: false,
				});
			}
		} catch (e) {
			console.error(e);
			setSearchState(p => ({ ...p, loading: false, error: 'Failed to load conversations' }));
		}
	}, [searchState.query, conversationsToResults]);

	/* --------------------------- search ---------------------------- */
	const performSearch = useCallback(async (query: string, token?: string, append = false) => {
		if (abortControllerRef.current && !append) abortControllerRef.current.abort();
		abortControllerRef.current = new AbortController();

		if (!append)
			setSearchState(p => ({ ...p, loading: true, error: undefined, hasMore: false, searchedMessages: false }));

		try {
			const res = await conversationStoreAPI.searchConversations(query, token, 20);
			const conversations = res.conversations;
			const nextToken = res.nextToken?.trim() || '';
			const newResults: SearchResult[] = conversations.map(conv => ({
				conversation: conv,
				matchType: conv.title.toLowerCase().includes(query.toLowerCase()) ? 'title' : 'message',
				snippet: `...Matched a Message`,
			}));

			setSearchState(p => ({
				...p,
				results: append ? [...p.results, ...newResults] : newResults,
				nextToken,
				hasMore: !!nextToken && nextToken !== '',
				loading: false,
				searchedMessages: true,
			}));

			/* caching (first page only) */
			if (!append) {
				searchCache.set(query, {
					results: newResults,
					nextToken,
					timestamp: Date.now(),
				});
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

	/* -------------- local filter for 1–2 char queries -------------- */
	const filterLocalResults = useCallback(
		(q: string) => {
			const filtered = recentConversations.filter(c => c.title.toLowerCase().includes(q.toLowerCase()));
			setSearchState(p => ({
				...p,
				results: conversationsToResults(filtered),
				loading: false,
				hasMore: false,
				nextToken: undefined,
				error: undefined,
				searchedMessages: false,
			}));
		},
		[recentConversations, conversationsToResults]
	);

	/* ------------------------ input change ------------------------- */
	const handleInputChange = useCallback(
		(e: ChangeEvent<HTMLInputElement>) => {
			const q = e.target.value;
			setSearchState(p => ({ ...p, query: q, searchedMessages: false }));
			setFocusedIndex(-1);

			/* re-open dropdown if hidden */
			if (!show && q.trim()) setShow(true);

			if (debounceTimeoutRef.current) clearTimeout(debounceTimeoutRef.current);

			if (!q.trim()) {
				setSearchState(p => ({
					...p,
					results: conversationsToResults(recentConversations),
					loading: false,
					hasMore: false,
					nextToken: undefined,
					error: undefined,
					searchedMessages: false,
				}));
				return;
			}

			/* short query → only local filter */
			if (q.length < 3) {
				filterLocalResults(q);
				return;
			}

			/* check cache */
			const cached = searchCache.get(q);
			if (cached && Date.now() - cached.timestamp < CACHE_EXPIRY_TIME) {
				setSearchState(p => ({
					...p,
					results: cached.results,
					nextToken: cached.nextToken,
					hasMore: !!cached.nextToken && cached.nextToken !== '',
					searchedMessages: true,
					loading: false,
					error: undefined,
				}));
				return;
			}

			/* remote search (debounced) */
			debounceTimeoutRef.current = setTimeout(() => performSearch(q), 300);
		},
		[show, recentConversations, conversationsToResults, filterLocalResults, performSearch]
	);

	/* ------------------------- focus / blur ------------------------ */
	const handleFocus = useCallback(async () => {
		setShow(true);
		if (!recentConversations.length) await loadRecentConversations();
		else if (!searchState.query && !searchState.results.length) {
			setSearchState(p => ({
				...p,
				results: conversationsToResults(recentConversations),
			}));
		}
	}, [
		recentConversations,
		loadRecentConversations,
		searchState.query,
		searchState.results.length,
		conversationsToResults,
	]);

	const handleBlur = useCallback(() => {
		setTimeout(() => {
			setShow(false);
		}, 150);
	}, []);

	/* --------------------------- pick ------------------------------ */
	const handlePick = useCallback(
		async (conv: ConversationItem) => {
			await onSelectConversation(conv);
			setSearchState(p => ({ ...p, query: '', searchedMessages: false }));
			setShow(false);
			setFocusedIndex(-1);
			inputRef.current?.blur();
		},
		[onSelectConversation]
	);

	/* ------------------------ load more ---------------------------- */
	const handleLoadMore = useCallback(() => {
		if (searchState.hasMore && !searchState.loading && searchState.nextToken) {
			performSearch(searchState.query, searchState.nextToken, true);
		}
	}, [searchState, performSearch]);

	/* ---------------------- keyboard nav --------------------------- */
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
						handlePick(searchState.results[focusedIndex].conversation);
					} else if (!searchState.loading && searchState.query.trim()) {
						/* allow *any* non-empty query, even 1–2 chars */
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

	/* -------------------- mount / unmount -------------------------- */
	useEffect(() => {
		loadRecentConversations();
	}, [refreshKey]); // reload when refreshKey changes

	useEffect(
		() => () => {
			abortControllerRef.current?.abort();
			if (debounceTimeoutRef.current) clearTimeout(debounceTimeoutRef.current);
		},
		[]
	);

	/* -------------------- derived / memo --------------------------- */
	const orderedResults = useMemo(() => [...searchState.results].sort(sortResults), [searchState.results]);

	const showSearchAllHintShortQuery =
		searchState.query.length > 0 && // something was typed
		searchState.query.length < 3 && // short (1–2 chars)
		!searchState.loading && // not currently searching
		!searchState.searchedMessages; // backend not queried yet

	/* --------------------------- render ---------------------------- */
	return (
		<div className="relative">
			{/* ---------------- search input ---------------- */}
			<div className="flex items-center bg-base-100 py-3 px-1 rounded-2xl border border-base-300 focus-within:border-base-400 transition-colors">
				<FiSearch size={20} className="mx-3 text-neutral/60 flex-shrink-0" />
				<input
					ref={inputRef}
					type="text"
					value={searchState.query}
					onChange={handleInputChange}
					onFocus={handleFocus}
					onBlur={handleBlur}
					onKeyDown={handleKeyDown}
					placeholder="Search conversations..."
					className="w-full bg-transparent outline-none text-sm placeholder:text-neutral/60"
					spellCheck={false}
				/>
				{searchState.loading && <span className="loading loading-dots loading-md"></span>}
			</div>

			{/* ---------------- dropdown ------------------- */}
			{show && (
				<SearchDropdown
					results={orderedResults}
					loading={searchState.loading}
					error={searchState.error}
					hasMore={searchState.hasMore}
					onLoadMore={handleLoadMore}
					focusedIndex={focusedIndex}
					onPick={handlePick}
					query={searchState.query}
					showSearchAllHintShortQuery={showSearchAllHintShortQuery}
				/>
			)}
		</div>
	);
};

export default ChatSearch;
