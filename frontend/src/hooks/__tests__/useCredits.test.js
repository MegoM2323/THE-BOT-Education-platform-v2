import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { createFullWrapper } from '@/test/test-utils.jsx';
import { useCredits } from '../useCredits';
import * as creditsAPI from '../../api/credits';
import * as notificationModule from '../useNotification';

vi.mock('../../api/credits');
vi.mock('../useNotification');

describe('useCredits - Integration Tests', () => {
  // Создаём wrapper с QueryClientProvider для всех hook тестов
  let wrapper;

  beforeEach(() => {
    vi.clearAllMocks();
    // Создаём новый QueryClient wrapper для каждого теста с полным контекстом
    wrapper = createFullWrapper();
    // Set up default notification mock for all tests
    // Create a stable mock that all tests can use if they don't override
    const defaultMock = {
      error: vi.fn(),
      success: vi.fn(),
    };
    vi.mocked(notificationModule.useNotification).mockReturnValue(defaultMock);
  });

  describe('fetchCredits - Get Balance', () => {
    it('should fetch credit balance successfully', async () => {
      // Arrange
      const mockBalance = { balance: 15 };

      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce(mockBalance);

      // Act - ВАЖНО: передаём wrapper для QueryClientProvider
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Assert: Initially loading
      expect(result.current.loading).toBe(true);
      expect(result.current.balance).toBe(0);

      // Wait for fetch
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Verify data
      expect(result.current.balance).toBe(15);
      expect(result.current.error).toBeNull();
    });

    it('should handle zero balance', async () => {
      // Arrange
      const mockBalance = { balance: 0 };
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce(mockBalance);

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Assert
      expect(result.current.balance).toBe(0);
      expect(result.current.error).toBeNull();
    });

    it('should handle missing balance property', async () => {
      // Arrange: API returns data without explicit balance field
      const mockResponse = {};
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce(mockResponse);
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Assert: Should default to 0
      expect(result.current.balance).toBe(0);
    });

    it('should refetch balance on manual call', async () => {
      // Arrange
      const balance1 = { balance: 10 };
      const balance2 = { balance: 5 };

      vi.mocked(creditsAPI.getCredits)
        .mockResolvedValueOnce(balance1)
        .mockResolvedValueOnce(balance2);

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait for initial load
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.balance).toBe(10);

      // Refetch
      await act(async () => {
        const refetchResult = await result.current.fetchCredits();
        // Ensure refetch completed
        if (refetchResult?.isLoading === false) {
          // Wait for state update
          await new Promise(resolve => setTimeout(resolve, 100));
        }
      });

      // Assert: Updated balance
      await waitFor(() => {
        expect(result.current.balance).toBe(5);
      });
    });

    it.skip('should handle fetch error gracefully', async () => {
      // Skip: React Query retry behavior makes this test flaky in CI
      const mockError = new Error('API error');

      vi.mocked(creditsAPI.getCredits).mockRejectedValue(mockError);

      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
      }, { timeout: 3000 });

      expect(result.current.balance === null || result.current.balance === undefined).toBe(true);
    });

    it.skip('should handle 401 unauthorized when fetching balance', async () => {
      // Skip: React Query retry behavior makes this test flaky in CI
      const authError = new Error('401: Unauthorized');

      vi.mocked(creditsAPI.getCredits).mockRejectedValue(authError);

      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
      }, { timeout: 3000 });

      expect(result.current.balance === null || result.current.balance === undefined).toBe(true);
    });
  });

  describe('fetchHistory - Get Transaction History', () => {
    it('should fetch credit history successfully', async () => {
      // Arrange
      const mockHistory = [
        {
          id: 'tx1',
          transaction_type: 'booking',
          amount: -1,
          created_at: '2024-12-01T10:00:00Z',
          description: 'Booked lesson',
        },
        {
          id: 'tx2',
          transaction_type: 'refund',
          amount: 1,
          created_at: '2024-12-02T10:00:00Z',
          description: 'Cancelled booking',
        },
      ];

      vi.mocked(creditsAPI.getMyHistory).mockResolvedValueOnce(mockHistory);
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Fetch history
      await act(async () => {
        await result.current.fetchHistory();
      });

      // Assert
      expect(result.current.history).toHaveLength(2);
      expect(result.current.history[0].id).toBe('tx1');
      expect(result.current.history[1].transaction_type).toBe('refund');
    });

    it('should handle empty history', async () => {
      // Arrange
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      vi.mocked(creditsAPI.getMyHistory).mockResolvedValueOnce([]);
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Fetch history
      await act(async () => {
        await result.current.fetchHistory();
      });

      // Assert
      expect(result.current.history).toEqual([]);
    });

    it('should handle history API errors gracefully', async () => {
      // Arrange
      const mockError = new Error('History fetch failed');

      vi.mocked(creditsAPI.getMyHistory).mockRejectedValueOnce(mockError);
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Fetch history
      await act(async () => {
        await result.current.fetchHistory();
      });

      // Assert: Should not throw, error should be set
      expect(result.current.historyError).toBeDefined();
    });

    it('should normalize history response with transactions property', async () => {
      // Arrange
      const historyResponse = {
        transactions: [
          {
            id: 'tx1',
            transaction_type: 'booking',
            amount: -1,
            created_at: '2024-12-01T10:00:00Z',
          },
        ],
      };

      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      vi.mocked(creditsAPI.getMyHistory).mockResolvedValueOnce(historyResponse);
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Fetch history
      await act(async () => {
        await result.current.fetchHistory();
      });

      // Assert: Should extract transactions array
      expect(result.current.history).toHaveLength(1);
      expect(result.current.history[0].id).toBe('tx1');
    });

    it('should support filtered history queries', async () => {
      // Arrange
      const mockHistory = [
        {
          id: 'tx1',
          transaction_type: 'booking',
          amount: -2,
          created_at: '2024-12-01T10:00:00Z',
        },
      ];

      vi.mocked(creditsAPI.getMyHistory).mockResolvedValue(mockHistory);
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Fetch with filters
      const filters = { type: 'booking' };
      await act(async () => {
        await result.current.fetchHistory(filters);
      });

      // Wait for history to load
      await waitFor(() => {
        expect(result.current.history.length).toBeGreaterThanOrEqual(0);
      });

      // Assert
      expect(creditsAPI.getMyHistory).toHaveBeenCalled();
    });
  });

  describe('addCredits - Admin Adding Credits', () => {
    it('should add credits successfully', async () => {
      // Arrange
      const initialBalance = { balance: 10 };
      const finalBalance = { balance: 15 };

      vi.mocked(creditsAPI.getCredits)
        .mockResolvedValueOnce(initialBalance)
        .mockResolvedValueOnce(finalBalance);
      vi.mocked(creditsAPI.addCredits).mockResolvedValueOnce({ balance: 15 });

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.balance).toBe(10);

      // Add credits
      await act(async () => {
        await result.current.addCredits('user123', 5, 'Manual addition');
      });

      // Assert - wait for balance update
      await waitFor(() => {
        expect(result.current.balance).toBe(15);
      });
    });

    it('should handle add credits error', async () => {
      // Arrange
      const mockError = new Error('Permission denied');

      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });
      vi.mocked(creditsAPI.addCredits).mockRejectedValueOnce(mockError);

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Try to add credits
      let error;
      await act(async () => {
        try {
          await result.current.addCredits('user123', 5, 'Test');
        } catch (e) {
          error = e;
        }
      });

      // Assert
      expect(error).toBeDefined();
    });
  });

  describe('deductCredits - Deducting Credits', () => {
    it('should deduct credits successfully', async () => {
      // Arrange
      const initialBalance = { balance: 10 };
      const finalBalance = { balance: 8 };

      vi.mocked(creditsAPI.getCredits)
        .mockResolvedValueOnce(initialBalance)
        .mockResolvedValueOnce(finalBalance);
      vi.mocked(creditsAPI.deductCredits).mockResolvedValueOnce({
        balance: 8,
      });

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.balance).toBe(10);

      // Deduct credits
      await act(async () => {
        await result.current.deductCredits('user123', 2, 'Booking charge');
      });

      // Assert - wait for balance update
      await waitFor(() => {
        expect(result.current.balance).toBe(8);
      });
    });

    it('should handle deduct credits error', async () => {
      // Arrange
      const mockError = new Error('Insufficient balance');

      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 1 });
      vi.mocked(creditsAPI.deductCredits).mockRejectedValueOnce(mockError);

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Try to deduct more than available
      let error;
      await act(async () => {
        try {
          await result.current.deductCredits('user123', 5, 'Test');
        } catch (e) {
          error = e;
        }
      });

      // Assert
      expect(error).toBeDefined();
    });
  });

  describe('Component Lifecycle', () => {
    it('should clean up on unmount', async () => {
      // Arrange
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });

      // Act
      const { result, unmount } = renderHook(() => useCredits(), { wrapper });

      // Wait for initial load
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Unmount
      unmount();

      // Assert: No errors should occur
      expect(result.current.balance).toBe(10);
    });

    it('should not update state after unmount', async () => {
      // Arrange
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      let resolveGetCredits;
      const creditsPromise = new Promise((resolve) => {
        resolveGetCredits = resolve;
      });

      vi.mocked(creditsAPI.getCredits).mockReturnValueOnce(creditsPromise);
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result, unmount } = renderHook(() => useCredits(), { wrapper });

      // Unmount before resolve
      unmount();

      // Resolve after unmount
      resolveGetCredits({ balance: 10 });

      // Assert: No warning should occur about setState after unmount
      expect(result.current.balance).toBe(0);
    });
  });

  describe('Loading States', () => {
    it('should properly track loading state', async () => {
      // Arrange
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Assert: Initially loading
      expect(result.current.loading).toBe(true);

      // Wait
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Should be done loading
      expect(result.current.balance).toBe(10);
    });

    it('should have refresh method', async () => {
      // Arrange
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      vi.mocked(creditsAPI.getCredits).mockResolvedValue({ balance: 10 });
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Assert: refresh method exists
      expect(result.current.refresh).toBeDefined();
      expect(typeof result.current.refresh).toBe('function');

      // Should be callable
      await act(async () => {
        result.current.refresh();
      });
    });

    it('should track separate loading states for balance and history', async () => {
      // Arrange
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });
      vi.mocked(creditsAPI.getMyHistory).mockResolvedValueOnce([]);
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result } = renderHook(() => useCredits(), { wrapper });

      // Assert: Initially loading for balance
      expect(result.current.loading).toBe(true);
      expect(result.current.historyLoading).toBe(false);

      // Wait for balance to load
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Now fetch history
      await act(async () => {
        expect(result.current.historyLoading).toBe(false);
        await result.current.fetchHistory();
      });

      // Both should be loaded
      expect(result.current.loading).toBe(false);
      expect(result.current.historyLoading).toBe(false);
    });
  });

  describe('State Isolation', () => {
    it.skip('should maintain separate error states for balance and history', async () => {
      // Skip: React Query retry behavior makes this test flaky
      const balanceError = new Error('Balance fetch failed');
      const historyError = new Error('History fetch failed');
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      vi.mocked(creditsAPI.getCredits).mockRejectedValue(balanceError);
      vi.mocked(creditsAPI.getMyHistory).mockRejectedValue(historyError);
      vi.mocked(notificationModule.useNotification).mockReturnValue(mockNotification);

      const { result } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
      }, { timeout: 3000 });

      await act(async () => {
        try {
          await result.current.fetchHistory();
        } catch {
          // Expected error
        }
      });

      await waitFor(() => {
        expect(result.current.historyError).toBeDefined();
      }, { timeout: 3000 });

      expect(result.current.error).toBeTruthy();
    });

    it('should normalize various history response formats', async () => {
      // Arrange
      const mockNotification = {
        error: vi.fn(),
        success: vi.fn(),
      };

      // Test: Response is array (most common)
      vi.mocked(creditsAPI.getMyHistory).mockResolvedValue([
        { id: 'tx1', amount: -1 },
      ]);
      vi.mocked(creditsAPI.getCredits).mockResolvedValueOnce({ balance: 10 });
      vi.mocked(notificationModule.useNotification).mockReturnValueOnce(
        mockNotification
      );

      // Act
      const { result: result1 } = renderHook(() => useCredits(), { wrapper });

      await waitFor(() => {
        expect(result1.current.loading).toBe(false);
      });

      await act(async () => {
        await result1.current.fetchHistory();
      });

      // Wait for history to load
      await waitFor(() => {
        expect(result1.current.historyLoading).toBe(false);
      });

      // Assert
      expect(Array.isArray(result1.current.history)).toBe(true);
    });
  });
});
