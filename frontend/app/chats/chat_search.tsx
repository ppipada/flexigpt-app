import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';

import { FiLoader, FiSearch } from 'react-icons/fi';

import type { ConversationItem } from '@/models/conversationmodel';

import { conversationStoreAPI } from '@/apis/baseapi';

import { GroupedDropdown } from '@/components/date_grouped_dropdown';

const CACHE_EXPIRY_TIME = 60000;
// Types for search results with match information
interface SearchResult {
	conversation: ConversationItem;
	matchType: 'title' | 'message';
	snippet?: string; // For message matches
}

interface SearchState {
	query: string;
	results: SearchResult[];
	nextToken?: string;
	loading: boolean;
	error?: string;
	hasMore: boolean;
}

// Cache for search results (keep last 5 queries)
interface SearchCacheEntry {
	results: SearchResult[];
	nextToken?: string;
	timestamp: number;
}

const searchCache = new Map<string, SearchCacheEntry>();

// Format date for display
const formatDate = (d: Date): string => {
	return new Date(d).toLocaleDateString('en-US', {
		day: '2-digit',
		month: 'short',
		year: 'numeric',
	});
};

// Group results by match type
const groupResults = (results: SearchResult[]) => {
	const titleMatches = results.filter(r => r.matchType === 'title');
	const messageMatches = results.filter(r => r.matchType === 'message');
	return { titleMatches, messageMatches };
};

interface SearchDropdownProps {
	results: SearchResult[];
	focusedIndex: number;
	onPick: (item: ConversationItem) => void;
	loading: boolean;
	error?: string;
	hasMore: boolean;
	onLoadMore: () => void;
	query: string;
	showSearchAllHint: boolean;
}

const SearchDropdown: FC<SearchDropdownProps> = ({
	results,
	focusedIndex,
	onPick,
	loading,
	error,
	hasMore,
	onLoadMore,
	query,
	showSearchAllHint,
}) => {
	const scrollRef = useRef<HTMLDivElement>(null);

	// Handle scroll to load more
	const handleScroll = useCallback(() => {
		const container = scrollRef.current;
		if (!container || !hasMore || loading) return;

		const { scrollTop, scrollHeight, clientHeight } = container;
		const scrollPercentage = (scrollTop + clientHeight) / scrollHeight;

		if (scrollPercentage >= 0.8) {
			onLoadMore();
		}
	}, [hasMore, loading, onLoadMore]);

	// Auto-scroll focused item into view
	useEffect(() => {
		if (focusedIndex >= 0 && scrollRef.current) {
			const focusedElement = scrollRef.current.querySelector(`[data-index="${focusedIndex}"]`);
			if (focusedElement) {
				focusedElement.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
			}
		}
	}, [focusedIndex]);

	// Error state
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

	// Empty state
	if (!results.length && !loading) {
		return (
			<div className="absolute left-0 right-0 mt-1 bg-base-200 rounded-2xl shadow-lg p-8 text-center z-50">
				<div className="text-neutral/60 text-sm">{query ? `No results for "${query}"` : 'No conversations yet'}</div>
				{!query && <div className="text-xs text-neutral/40 mt-2">Start a conversation to see it here</div>}
			</div>
		);
	}

	const { titleMatches, messageMatches } = groupResults(results);
	let itemIndex = 0;

	return (
		<div className="absolute left-0 right-0 mt-1 bg-base-200 rounded-2xl shadow-lg z-50">
			<div ref={scrollRef} className="max-h-[60vh] overflow-y-auto" onScroll={handleScroll}>
				{/* Header for recent conversations */}
				{!query && results.length > 0 && (
					<div className="px-4 py-2 text-neutral/60 text-xs font-semibold uppercase tracking-wide border-b border-base-300">
						Recent conversations
					</div>
				)}

				{/* Title matches section */}
				{query && titleMatches.length > 0 && (
					<>
						<div className="px-4 py-2 text-neutral/60 text-xs font-semibold uppercase tracking-wide border-b border-base-300">
							Title matches
						</div>
						{titleMatches.map(result => {
							const currentIndex = itemIndex++;
							return (
								<SearchResultItem
									key={`title-${result.conversation.id}`}
									result={result}
									index={currentIndex}
									focused={currentIndex === focusedIndex}
									onPick={onPick}
								/>
							);
						})}
					</>
				)}

				{/* Message matches section */}
				{query && messageMatches.length > 0 && (
					<>
						<div className="px-4 py-2 text-neutral/60 text-xs font-semibold uppercase tracking-wide border-b border-base-300">
							Message matches
						</div>
						{messageMatches.map(result => {
							const currentIndex = itemIndex++;
							return (
								<SearchResultItem
									key={`message-${result.conversation.id}`}
									result={result}
									index={currentIndex}
									focused={currentIndex === focusedIndex}
									onPick={onPick}
								/>
							);
						})}
					</>
				)}

				{/* Recent conversations (no query) */}
				{!query && (
					<GroupedDropdown<SearchResult>
						items={results}
						focused={focusedIndex}
						getDate={r => new Date(r.conversation.createdAt)}
						getKey={r => r.conversation.id}
						getLabel={r => (
							<>
								<span className="truncate">{r.conversation.title}</span>
								{r.snippet && r.matchType === 'message' && (
									<span className="ml-1 text-neutral/60 line-clamp-1">{r.snippet}</span>
								)}
							</>
						)}
						onPick={r => {
							onPick(r.conversation);
						}}
						renderItemExtra={r => (
							<span className="text-xs text-neutral/60">{formatDate(r.conversation.createdAt)}</span>
						)}
					/>
				)}

				{/* Loading indicator */}
				{loading && (
					<div className="flex items-center justify-center py-4">
						<FiLoader className="animate-spin mr-2 text-primary" size={16} />
						<span className="text-sm text-neutral/60">{results.length > 0 ? 'Loading more...' : 'Searching...'}</span>
					</div>
				)}

				{/* End of results */}
				{!loading && !hasMore && results.length > 0 && query && (
					<div className="text-center py-3 text-xs text-neutral/60 border-t border-base-300">End of results</div>
				)}
			</div>

			{/* Search all messages hint */}
			{showSearchAllHint && (
				<div className="px-4 py-3 text-center text-xs text-neutral/60 border-t border-base-300 bg-base-100 rounded-b-2xl">
					Press <kbd className="kbd kbd-xs">Enter</kbd> to search all messages
				</div>
			)}
		</div>
	);
};

interface SearchResultItemProps {
	result: SearchResult;
	index: number;
	focused: boolean;
	onPick: (item: ConversationItem) => void;
}

const SearchResultItem: FC<SearchResultItemProps> = ({ result, index, focused, onPick }) => {
	const { conversation, matchType, snippet } = result;

	return (
		<div
			data-index={index}
			onClick={() => {
				onPick(conversation);
			}}
			className={`px-4 py-3 cursor-pointer transition-colors ${focused ? 'bg-base-100' : 'hover:bg-base-100/50'}`}
			title={conversation.title}
		>
			<div className="flex justify-between items-start gap-3">
				<div className="flex-1 min-w-0">
					<div className="text-sm font-normal truncate">{conversation.title}</div>
					{snippet && matchType === 'message' && (
						<div className="text-xs text-neutral/70 mt-1 line-clamp-1">{snippet}</div>
					)}
				</div>
				<div className="flex-shrink-0 text-xs text-neutral/60">{formatDate(conversation.createdAt)}</div>
			</div>
		</div>
	);
};

interface ChatSearchProps {
	onSelectConversation: (item: ConversationItem) => Promise<void>;
	refreshKey: number;
}

const ChatSearch: FC<ChatSearchProps> = ({ onSelectConversation, refreshKey }) => {
	const [searchState, setSearchState] = useState<SearchState>({
		query: '',
		results: [],
		loading: false,
		hasMore: false,
	});

	const [show, setShow] = useState(false);
	const [focusedIndex, setFocusedIndex] = useState(-1);
	const [recentConversations, setRecentConversations] = useState<ConversationItem[]>([]);

	// Refs for managing async operations
	const abortControllerRef = useRef<AbortController | null>(null);
	const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);

	// Convert conversations to search results
	const conversationsToResults = useCallback((conversations: ConversationItem[]): SearchResult[] => {
		return conversations.map(conv => ({
			conversation: conv,
			matchType: 'title' as const,
		}));
	}, []);

	// Load recent conversations
	const loadRecentConversations = useCallback(async () => {
		try {
			setSearchState(prev => ({ ...prev, loading: true }));
			const { conversations } = await conversationStoreAPI.listConversations();
			setRecentConversations(conversations);

			// Update search state if showing recent conversations
			if (!searchState.query) {
				setSearchState(prev => ({
					...prev,
					results: conversationsToResults(conversations),
					loading: false,
					hasMore: false,
				}));
			}
		} catch (error) {
			console.error('Failed to load recent conversations:', error);
			setSearchState(prev => ({
				...prev,
				loading: false,
				error: 'Failed to load conversations',
			}));
		}
	}, [searchState.query, conversationsToResults]);

	// Perform full-text search
	const performSearch = useCallback(async (query: string, token?: string, append = false) => {
		// Cancel previous request
		if (abortControllerRef.current && !append) {
			abortControllerRef.current.abort();
		}

		// Create new abort controller
		abortControllerRef.current = new AbortController();

		try {
			if (!append) {
				setSearchState(prev => ({ ...prev, loading: true, error: undefined }));
			}

			const { conversations, nextToken } = await conversationStoreAPI.searchConversations(query, token, 20);

			// Convert to search results with match type detection
			const results: SearchResult[] = conversations.map(conv => {
				const titleMatch = conv.title.toLowerCase().includes(query.toLowerCase());
				return {
					conversation: conv,
					matchType: titleMatch ? 'title' : 'message',
					snippet: titleMatch ? undefined : `...match found in conversation...`, // This would come from API in real implementation
				};
			});

			setSearchState(prev => ({
				...prev,
				results: append ? [...prev.results, ...results] : results,
				nextToken,
				hasMore: !!nextToken,
				loading: false,
			}));

			// Cache the results (only for new searches, not pagination)
			if (!append) {
				searchCache.set(query, {
					results,
					nextToken,
					timestamp: Date.now(),
				});

				// Keep only last 5 queries
				if (searchCache.size > 5) {
					// Find the oldest key
					let oldestKey: string | undefined;
					let oldestTimestamp = Infinity;

					for (const [key, entry] of searchCache.entries()) {
						if (entry.timestamp < oldestTimestamp) {
							oldestTimestamp = entry.timestamp;
							oldestKey = key;
						}
					}

					if (oldestKey !== undefined) {
						searchCache.delete(oldestKey);
					}
				}
			}
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
		} catch (error: any) {
			if (!abortControllerRef.current.signal.aborted) {
				setSearchState(prev => ({
					...prev,
					loading: false,
					error: 'Search failed. Please try again.', // error.message ||
				}));
			}
		}
	}, []);

	// Filter local results for short queries
	const filterLocalResults = useCallback(
		(query: string) => {
			const filtered = recentConversations.filter(conv => conv.title.toLowerCase().includes(query.toLowerCase()));

			setSearchState(prev => ({
				...prev,
				results: conversationsToResults(filtered),
				loading: false,
				hasMore: false,
				nextToken: undefined,
				error: undefined,
			}));
		},
		[recentConversations, conversationsToResults]
	);

	// Handle input change with debouncing
	const handleInputChange = useCallback(
		(e: ChangeEvent<HTMLInputElement>) => {
			const query = e.target.value;

			setSearchState(prev => ({ ...prev, query }));
			setFocusedIndex(-1);

			// Clear existing timeout
			if (debounceTimeoutRef.current) {
				clearTimeout(debounceTimeoutRef.current);
			}

			if (!query.trim()) {
				// Empty query - show recent conversations
				setSearchState(prev => ({
					...prev,
					results: conversationsToResults(recentConversations),
					loading: false,
					hasMore: false,
					nextToken: undefined,
					error: undefined,
				}));
				return;
			}

			if (query.length < 3) {
				// Short query - filter locally
				filterLocalResults(query);
				return;
			}

			// Check cache first
			const cached = searchCache.get(query);

			if (cached && Date.now() - cached.timestamp < CACHE_EXPIRY_TIME) {
				// 1 minute cache
				setSearchState(prev => ({
					...prev,
					results: cached.results,
					nextToken: cached.nextToken,
					hasMore: !!cached.nextToken,
					loading: false,
					error: undefined,
				}));
				return;
			}

			// Debounced search
			debounceTimeoutRef.current = setTimeout(() => {
				performSearch(query);
			}, 300);
		},
		[recentConversations, conversationsToResults, filterLocalResults, performSearch]
	);

	// Handle focus
	const handleFocus = useCallback(async () => {
		setShow(true);
		if (recentConversations.length === 0) {
			await loadRecentConversations();
		} else if (!searchState.query && searchState.results.length === 0) {
			setSearchState(prev => ({
				...prev,
				results: conversationsToResults(recentConversations),
			}));
		}
	}, [
		recentConversations,
		searchState.query,
		searchState.results.length,
		loadRecentConversations,
		conversationsToResults,
	]);

	// Handle blur
	const handleBlur = useCallback(() => {
		// Delay hiding to allow for click events
		setTimeout(() => {
			setShow(false);
		}, 150);
	}, []);

	// Handle item selection
	const handlePick = useCallback(
		async (conversation: ConversationItem) => {
			try {
				await onSelectConversation(conversation);
				setSearchState(prev => ({ ...prev, query: '' }));
				setShow(false);
				setFocusedIndex(-1);
			} catch (error) {
				console.error('Failed to select conversation:', error);
			}
		},
		[onSelectConversation]
	);

	// Handle load more
	const handleLoadMore = useCallback(() => {
		if (searchState.hasMore && !searchState.loading && searchState.nextToken) {
			performSearch(searchState.query, searchState.nextToken, true);
		}
	}, [searchState.hasMore, searchState.loading, searchState.nextToken, searchState.query, performSearch]);

	// Handle keyboard navigation
	const handleKeyDown = useCallback(
		(e: KeyboardEvent<HTMLInputElement>) => {
			if (!show) return;

			switch (e.key) {
				case 'ArrowDown':
					e.preventDefault();
					if (searchState.results.length > 0) {
						setFocusedIndex(prev => (prev + 1) % searchState.results.length);
					}
					break;

				case 'ArrowUp':
					e.preventDefault();
					if (searchState.results.length > 0) {
						setFocusedIndex(prev => (prev - 1 + searchState.results.length) % searchState.results.length);
					}
					break;

				case 'Enter':
					e.preventDefault();
					if (focusedIndex >= 0 && focusedIndex < searchState.results.length) {
						handlePick(searchState.results[focusedIndex].conversation);
					} else if (searchState.query.length >= 3) {
						// Force search if not already searching
						if (!searchState.loading) {
							performSearch(searchState.query);
						}
					}
					break;

				case 'Escape':
					setShow(false);
					setFocusedIndex(-1);
					if (searchState.query) {
						setSearchState(prev => ({ ...prev, query: '' }));
					}
					break;
			}
		},
		[show, searchState.results, searchState.query, searchState.loading, focusedIndex, handlePick, performSearch]
	);

	// Initialize recent conversations on mount
	useEffect(() => {
		loadRecentConversations();
	}, [refreshKey]);

	// Cleanup on unmount
	useEffect(() => {
		return () => {
			if (abortControllerRef.current) {
				abortControllerRef.current.abort();
			}
			if (debounceTimeoutRef.current) {
				clearTimeout(debounceTimeoutRef.current);
				debounceTimeoutRef.current = null;
			}
		};
	}, []);

	const showSearchAllHint = searchState.query.length > 0 && searchState.query.length < 3 && !searchState.loading;

	return (
		<div className="relative">
			<div className="flex items-center bg-base-100 py-3 px-1 rounded-2xl border border-base-300 focus-within:border-base-400 transition-colors">
				<FiSearch size={20} className="mx-3 text-neutral/60 flex-shrink-0" />
				<input
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
				{searchState.loading && searchState.results.length === 0 && (
					<FiLoader className="animate-spin mx-3 text-primary flex-shrink-0" size={16} />
				)}
			</div>

			{show && (
				<SearchDropdown
					results={searchState.results}
					focusedIndex={focusedIndex}
					onPick={handlePick}
					loading={searchState.loading}
					error={searchState.error}
					hasMore={searchState.hasMore}
					onLoadMore={handleLoadMore}
					query={searchState.query}
					showSearchAllHint={showSearchAllHint}
				/>
			)}
		</div>
	);
};

export default ChatSearch;
