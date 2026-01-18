import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useSidebarCredits } from '../useSidebarCredits.js';
import { useCredits } from '../useCredits.js';

// Mock dependencies
vi.mock('../useCredits.js');

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useSidebarCredits', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should delegate to useCredits with correct refetchInterval', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: 100,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    const { result } = renderHook(() => useSidebarCredits({ interval: 5000, enabled: true }), {
      wrapper: createWrapper(),
    });

    expect(useCredits).toHaveBeenCalledWith({
      refetchInterval: 5000,
    });

    expect(result.current.balance).toBe(100);
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.isPolling).toBe(true);
  });

  it('should disable polling when enabled is false', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: 100,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    renderHook(() => useSidebarCredits({ interval: 5000, enabled: false }), {
      wrapper: createWrapper(),
    });

    expect(useCredits).toHaveBeenCalledWith({
      refetchInterval: false,
    });
  });

  it('should use default interval of 10000ms', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: 100,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    renderHook(() => useSidebarCredits(), {
      wrapper: createWrapper(),
    });

    expect(useCredits).toHaveBeenCalledWith({
      refetchInterval: 10000,
    });
  });

  it('should return proper API from useCredits', () => {
    const mockRefetch = vi.fn();
    vi.mocked(useCredits).mockReturnValue({
      balance: 150,
      loading: true,
      error: 'Test error',
      refetch: mockRefetch,
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    const { result } = renderHook(() => useSidebarCredits(), {
      wrapper: createWrapper(),
    });

    expect(result.current.balance).toBe(150);
    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBe('Test error');
    expect(result.current.refetch).toBe(mockRefetch);
  });

  it('should handle null balance from useCredits', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: null,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    const { result } = renderHook(() => useSidebarCredits(), {
      wrapper: createWrapper(),
    });

    expect(result.current.balance).toBe(0);
  });

  it('should provide deprecated startPolling with warning', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: 100,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

    const { result } = renderHook(() => useSidebarCredits(), {
      wrapper: createWrapper(),
    });

    result.current.startPolling();

    expect(warnSpy).toHaveBeenCalledWith(
      'useSidebarCredits.startPolling() is deprecated. Use useCredits({ refetchInterval }) instead'
    );

    warnSpy.mockRestore();
  });

  it('should provide deprecated stopPolling with warning', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: 100,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

    const { result } = renderHook(() => useSidebarCredits(), {
      wrapper: createWrapper(),
    });

    result.current.stopPolling();

    expect(warnSpy).toHaveBeenCalledWith(
      'useSidebarCredits.stopPolling() is deprecated. Use useCredits({ refetchInterval: false }) instead'
    );

    warnSpy.mockRestore();
  });

  it('should return lastFetchTime', () => {
    vi.mocked(useCredits).mockReturnValue({
      balance: 100,
      loading: false,
      error: null,
      refetch: vi.fn(),
      history: [],
      historyLoading: false,
      historyError: null,
      fetchCredits: vi.fn(),
      fetchHistory: vi.fn(),
      addCredits: vi.fn(),
      deductCredits: vi.fn(),
      refresh: vi.fn(),
      isAdding: false,
      isDeducting: false,
    });

    const { result } = renderHook(() => useSidebarCredits(), {
      wrapper: createWrapper(),
    });

    expect(result.current.lastFetchTime).toBeDefined();
    expect(typeof result.current.lastFetchTime).toBe('number');
  });
});
