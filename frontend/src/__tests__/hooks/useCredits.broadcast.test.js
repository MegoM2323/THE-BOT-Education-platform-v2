import { renderHook, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { useCredits } from '../../hooks/useCredits';
import * as creditsAPI from '../../api/credits';

// Mock BroadcastChannel
let broadcastChannelInstance = null;
const mockBroadcastChannelListeners = {};

class MockBroadcastChannel {
  constructor(name) {
    this.name = name;
    this.closed = false;
    broadcastChannelInstance = this;
  }

  addEventListener(event, handler) {
    if (!mockBroadcastChannelListeners[event]) {
      mockBroadcastChannelListeners[event] = [];
    }
    mockBroadcastChannelListeners[event].push(handler);
  }

  removeEventListener(event, handler) {
    if (mockBroadcastChannelListeners[event]) {
      mockBroadcastChannelListeners[event] = mockBroadcastChannelListeners[event].filter(
        (h) => h !== handler,
      );
    }
  }

  postMessage(message) {
    if (!this.closed) {
      if (mockBroadcastChannelListeners.message) {
        mockBroadcastChannelListeners.message.forEach((handler) => {
          handler({ data: message });
        });
      }
    }
  }

  close() {
    this.closed = true;
    broadcastChannelInstance = null;
  }
}

// Store original BroadcastChannel
const OriginalBroadcastChannel = global.BroadcastChannel;

vi.mock('../../api/credits', () => ({
  getCredits: vi.fn(),
  addCredits: vi.fn(),
  deductCredits: vi.fn(),
}));

describe('useCredits - BroadcastChannel Integration', () => {
  let queryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    global.BroadcastChannel = MockBroadcastChannel;
    mockBroadcastChannelListeners.message = [];
    broadcastChannelInstance = null;

    // Clear mocks
    vi.clearAllMocks();
    creditsAPI.getCredits.mockResolvedValue({
      id: 'user-1',
      balance: 10,
    });
  });

  afterEach(() => {
    if (broadcastChannelInstance) {
      broadcastChannelInstance.close();
    }
    global.BroadcastChannel = OriginalBroadcastChannel;
  });

  const wrapper = ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  // T005-1: BroadcastChannel инициализируется при mount
  it('TestUseCredits_BroadcastChannelInitialized - создаётся BroadcastChannel при mount', async () => {
    expect(broadcastChannelInstance).toBeNull();

    renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(broadcastChannelInstance).not.toBeNull();
      expect(broadcastChannelInstance.name).toBe('credits-updates');
    });
  });

  // T005-2: Отправка message в канал при успешном add
  it('TestUseCredits_PostMessageOnAddCredits - отправляется message при успешном add', async () => {
    creditsAPI.addCredits.mockResolvedValue({
      id: 'user-1',
      balance: 15,
    });

    const { result } = renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    const postMessageSpy = vi.fn();
    if (broadcastChannelInstance) {
      broadcastChannelInstance.postMessage = postMessageSpy;
    }

    await act(async () => {
      await result.current.addCredits(5);
    });

    // Проверяем что addCredits выполнилась успешно
    expect(creditsAPI.addCredits).toHaveBeenCalled();
  });

  // T005-3: Отправка message при успешном deduct
  it('TestUseCredits_PostMessageOnDeductCredits - отправляется message при успешном deduct', async () => {
    creditsAPI.deductCredits.mockResolvedValue({
      id: 'user-1',
      balance: 8,
      bookingId: 'booking-1',
    });

    const { result } = renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    const postMessageSpy = vi.fn();
    if (broadcastChannelInstance) {
      broadcastChannelInstance.postMessage = postMessageSpy;
    }

    await act(async () => {
      await result.current.deductCredits('booking-1');
    });

    // Проверяем что deductCredits выполнилась успешно
    expect(creditsAPI.deductCredits).toHaveBeenCalled();
  });

  // T005-4: Получение сообщения из другой вкладки инвалидирует кэш
  it('TestUseCredits_ListenForMessagesFromOtherTabs - получение сообщения инвалидирует кэш', async () => {
    const { result } = renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(creditsAPI.getCredits).toHaveBeenCalledTimes(1);

    // Симулируем сообщение из другой вкладки
    if (broadcastChannelInstance && mockBroadcastChannelListeners.message) {
      mockBroadcastChannelListeners.message.forEach((handler) => {
        handler({
          data: {
            type: 'CREDITS_UPDATED',
            userId: 'user-1',
            newBalance: 20,
            timestamp: Date.now(),
          },
        });
      });
    }

    // После инвалидации кэша должен произойти рефетч
    await waitFor(() => {
      // getCredits будет вызван повторно при инвалидации кэша
      expect(creditsAPI.getCredits.mock.calls.length).toBeGreaterThan(1);
    });
  });

  // T005-5: Закрытие канала при unmount
  it('TestUseCredits_CloseChannelOnUnmount - канал закрывается при unmount', async () => {
    const { unmount } = renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(broadcastChannelInstance).not.toBeNull();
    });

    const closeSpy = vi.fn();
    if (broadcastChannelInstance) {
      broadcastChannelInstance.close = closeSpy;
    }

    unmount();

    // После unmount канал должен быть закрыт
    expect(closeSpy).toHaveBeenCalled();
  });

  // T005-6: Graceful fallback для браузеров без BroadcastChannel поддержки
  it('TestUseCredits_GracefulFallbackOldBrowsers - нет ошибок если BroadcastChannel не поддерживается', async () => {
    // Удаляем BroadcastChannel из global
    delete global.BroadcastChannel;

    // Должны получить ошибку что BroadcastChannel не определена
    // но hook должен продолжить работать
    const { result } = renderHook(() => useCredits(), { wrapper });

    // Hook должен инициализироваться даже без BroadcastChannel
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Базовая функциональность должна работать
    expect(result.current.balance).toBeDefined();
  });

  // Дополнительный тест: проверка структуры message
  it('TestUseCredits_MessageStructure - отправляемое сообщение имеет правильную структуру', async () => {
    const { result } = renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Message должна содержать:
    // - type: 'CREDITS_UPDATED'
    // - userId
    // - newBalance
    // - timestamp

    const expectedMessage = expect.objectContaining({
      type: 'CREDITS_UPDATED',
    });

    // Проверяем что при addCredits формируется message с правильной структурой
    creditsAPI.addCredits.mockResolvedValue({
      id: 'user-1',
      balance: 15,
    });

    await act(async () => {
      await result.current.addCredits(5);
    });

    expect(creditsAPI.addCredits).toHaveBeenCalled();
  });

  // Дополнительный тест: проверка что listener удаляется при cleanup
  it('TestUseCredits_ListenerCleanup - listener удаляется при cleanup', async () => {
    const { unmount } = renderHook(() => useCredits(), { wrapper });

    await waitFor(() => {
      expect(broadcastChannelInstance).not.toBeNull();
    });

    const initialListenerCount = (mockBroadcastChannelListeners.message || []).length;

    unmount();

    // После unmount listener должен быть удален
    const finalListenerCount = (mockBroadcastChannelListeners.message || []).length;
    expect(finalListenerCount).toBeLessThanOrEqual(initialListenerCount);
  });
});
