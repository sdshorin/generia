import { useState, useEffect, useCallback, useRef } from 'react';
import { useInView } from 'react-intersection-observer';

interface FetchResponse<T> {
  items: T[];
  nextCursor?: string;
  hasMore: boolean;
}

interface InfiniteScrollOptions<T> {
  initialData?: T[];
  fetchItems: (limit: number, cursor: string) => Promise<FetchResponse<T>>;
  limit?: number;
}

export function useInfiniteScroll<T>({
  initialData = [],
  fetchItems,
  limit = 10,
}: InfiniteScrollOptions<T>) {
  const [items, setItems] = useState<T[]>(initialData);
  const [cursor, setCursor] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const isFirstRender = useRef(true);
  const requestInProgress = useRef(false);
  
  const { ref, inView } = useInView({
    threshold: 0.1,
    triggerOnce: false,
  });

  const loadMore = useCallback(async () => {
    if (isLoading || !hasMore || requestInProgress.current) return;
    
    requestInProgress.current = true;
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await fetchItems(limit, cursor);
      
      setItems(prev => [...prev, ...response.items]);
      setCursor(response.nextCursor || '');
      setHasMore(response.hasMore);
    } catch (err) {
      setError('Failed to load more items. Please try again.');
      console.error('Error loading more items:', err);
    } finally {
      setIsLoading(false);
      requestInProgress.current = false;
    }
  }, [fetchItems, cursor, limit, isLoading, hasMore]);

  // Reset everything
  const reset = useCallback(() => {
    setItems([]);
    setCursor('');
    setHasMore(true);
    setError(null);
  }, []);

  // Load more when the sentinel comes into view
  useEffect(() => {
    // Skip the initial effect run to prevent double-fetching
    if (isFirstRender.current) {
      isFirstRender.current = false;
      loadMore();
      return;
    }
    
    if (inView && !requestInProgress.current) {
      const timeoutId = setTimeout(() => {
        loadMore();
      }, 300); // Add debounce to prevent rapid-fire requests
      
      return () => clearTimeout(timeoutId);
    }
  }, [inView, loadMore]);

  return {
    items,
    isLoading,
    hasMore,
    error,
    loadMore,
    reset,
    sentinelRef: ref,
  };
}