import { useState, useEffect, useCallback, useRef } from 'react';
import { useInView } from 'react-intersection-observer';

interface InfiniteScrollOptions<T> {
  initialData?: T[];
  fetchItems: (page: number, limit: number) => Promise<T[]>;
  limit?: number;
}

export function useInfiniteScroll<T>({
  initialData = [],
  fetchItems,
  limit = 10,
}: InfiniteScrollOptions<T>) {
  const [items, setItems] = useState<T[]>(initialData);
  const [page, setPage] = useState(0);
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
      const newItems = await fetchItems(page, limit);
      
      if (newItems.length < limit) {
        setHasMore(false);
      }
      
      setItems(prev => [...prev, ...newItems]);
      setPage(prev => prev + 1);
    } catch (err) {
      setError('Failed to load more items. Please try again.');
      console.error('Error loading more items:', err);
    } finally {
      setIsLoading(false);
      requestInProgress.current = false;
    }
  }, [fetchItems, page, limit, isLoading, hasMore]);

  // Reset everything
  const reset = useCallback(() => {
    setItems([]);
    setPage(0);
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