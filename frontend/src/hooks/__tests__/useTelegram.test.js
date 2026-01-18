import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useTelegram } from '../useTelegram';
import * as telegramAPI from '../../api/telegram';
import * as useNotificationModule from '../useNotification';

// Мокаем API и notification hook
vi.mock('../../api/telegram');
vi.mock('../useNotification');

describe('useTelegram', () => {
  // Mock notification функций
  const mockNotification = {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Мокаем useNotification hook
    vi.spyOn(useNotificationModule, 'useNotification').mockReturnValue(mockNotification);
  });

  describe('fetchLinkStatus', () => {
    it('должен загружать статус привязки при монтировании', async () => {
      const mockLinkStatus = {
        linked: true,
        telegram_id: 123456789,
        username: 'testuser',
        subscribed: true,
      };

      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue(mockLinkStatus);

      const { result } = renderHook(() => useTelegram());

      // Начальное состояние
      expect(result.current.loading).toBe(true);
      expect(result.current.linkStatus).toBe(null);

      // Ждем завершения загрузки
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Проверяем загруженные данные
      expect(result.current.linkStatus).toEqual(mockLinkStatus);
      expect(result.current.error).toBe(null);
      expect(telegramAPI.getMyTelegramLink).toHaveBeenCalledTimes(1);
    });

    it('должен обрабатывать ошибки загрузки статуса', async () => {
      const errorMessage = 'Network error';
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockRejectedValue(
        new Error(errorMessage)
      );

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBe(errorMessage);
    });

    it('должен перезагружать статус при вызове refetchLinkStatus', async () => {
      const mockLinkStatus1 = { linked: false };
      const mockLinkStatus2 = { linked: true, telegram_id: 123456789 };

      vi.spyOn(telegramAPI, 'getMyTelegramLink')
        .mockResolvedValueOnce(mockLinkStatus1)
        .mockResolvedValueOnce(mockLinkStatus2);

      const { result } = renderHook(() => useTelegram());

      // Первая загрузка
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
      expect(result.current.linkStatus).toEqual(mockLinkStatus1);

      // Перезагрузка
      act(() => {
        result.current.refetchLinkStatus();
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
      expect(result.current.linkStatus).toEqual(mockLinkStatus2);
      expect(telegramAPI.getMyTelegramLink).toHaveBeenCalledTimes(2);
    });
  });

  describe('generateToken', () => {
    it('должен успешно генерировать токен', async () => {
      const mockTokenData = {
        token: 'test-token-12345',
        bot_username: 'TestBot',
      };

      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({ linked: false });
      vi.spyOn(telegramAPI, 'generateLinkToken').mockResolvedValue(mockTokenData);

      const { result } = renderHook(() => useTelegram());

      // Ждем начальной загрузки
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Генерируем токен
      let tokenData;
      await act(async () => {
        tokenData = await result.current.generateToken();
      });

      expect(tokenData).toEqual(mockTokenData);
      expect(result.current.token).toBe('test-token-12345');
      expect(result.current.botUsername).toBe('TestBot');
      expect(mockNotification.success).toHaveBeenCalledWith(
        'Токен создан. Откройте бота в Telegram для привязки'
      );
    });

    it('должен обрабатывать ошибки генерации токена', async () => {
      const errorMessage = 'Already linked';
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({ linked: false });
      vi.spyOn(telegramAPI, 'generateLinkToken').mockRejectedValue(
        new Error(errorMessage)
      );

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Пытаемся сгенерировать токен
      await act(async () => {
        try {
          await result.current.generateToken();
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });

    it('должен перезагружать статус при ошибке генерации токена', async () => {
      const errorMessage = 'Already linked';
      const updatedLinkStatus = {
        linked: true,
        telegram: {
          telegram_id: 123456789,
          username: 'testuser',
        },
      };

      // Первая загрузка: не привязан
      vi.spyOn(telegramAPI, 'getMyTelegramLink')
        .mockResolvedValueOnce({ linked: false })
        .mockResolvedValueOnce(updatedLinkStatus);

      vi.spyOn(telegramAPI, 'generateLinkToken').mockRejectedValue(
        new Error(errorMessage)
      );

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Проверяем начальное состояние
      expect(result.current.linkStatus).toEqual({ linked: false });

      // Пытаемся сгенерировать токен (получаем ошибку)
      await act(async () => {
        try {
          await result.current.generateToken();
        } catch (err) {
          // Ожидаем ошибку
        }
      });

      // Проверяем что статус был перезагружен и обновлён
      await waitFor(() => {
        expect(result.current.linkStatus).toEqual(updatedLinkStatus);
      });

      // Проверяем что getMyTelegramLink вызван дважды:
      // 1. При монтировании
      // 2. После ошибки generateToken
      expect(telegramAPI.getMyTelegramLink).toHaveBeenCalledTimes(2);
    });

    it('должен устанавливать loading state во время генерации', async () => {
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({ linked: false });
      vi.spyOn(telegramAPI, 'generateLinkToken').mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve({ token: 'test' }), 100))
      );

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Начинаем генерацию
      act(() => {
        result.current.generateToken();
      });

      // Проверяем что loading = true
      expect(result.current.loading).toBe(true);

      // Ждем завершения
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
    });
  });

  describe('unlinkAccount', () => {
    it('должен успешно отвязывать аккаунт', async () => {
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({
        linked: true,
        telegram_id: 123456789,
      });
      vi.spyOn(telegramAPI, 'unlinkTelegram').mockResolvedValue();

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Проверяем начальное состояние
      expect(result.current.linkStatus.linked).toBe(true);

      // Отвязываем аккаунт
      await act(async () => {
        await result.current.unlinkAccount();
      });

      expect(result.current.linkStatus).toEqual({ linked: false });
      expect(result.current.token).toBe(null);
      expect(result.current.botUsername).toBe(null);
      expect(mockNotification.success).toHaveBeenCalledWith('Telegram успешно отвязан');
      expect(telegramAPI.unlinkTelegram).toHaveBeenCalledTimes(1);
    });

    it('должен обрабатывать ошибки отвязки', async () => {
      const errorMessage = 'Server error';
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({
        linked: true,
        telegram_id: 123456789,
      });
      vi.spyOn(telegramAPI, 'unlinkTelegram').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Пытаемся отвязать
      await act(async () => {
        try {
          await result.current.unlinkAccount();
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });

    it('должен устанавливать loading state во время отвязки', async () => {
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({
        linked: true,
        telegram_id: 123456789,
      });
      vi.spyOn(telegramAPI, 'unlinkTelegram').mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Начинаем отвязку
      act(() => {
        result.current.unlinkAccount();
      });

      // Проверяем что loading = true
      expect(result.current.loading).toBe(true);

      // Ждем завершения
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
    });
  });

  describe('state management', () => {
    it('должен корректно управлять состоянием error', async () => {
      vi.spyOn(telegramAPI, 'getMyTelegramLink')
        .mockRejectedValueOnce(new Error('First error'))
        .mockResolvedValueOnce({ linked: false });

      const { result } = renderHook(() => useTelegram());

      // Первая загрузка с ошибкой
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
      expect(result.current.error).toBe('First error');

      // Повторная загрузка успешная
      await act(async () => {
        await result.current.refetchLinkStatus();
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Ошибка должна быть сброшена
      expect(result.current.error).toBe(null);
    });

    it('должен сбрасывать ошибку вручную через clearError', async () => {
      const errorMessage = 'Test error';
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockRejectedValue(
        new Error(errorMessage)
      );

      const { result } = renderHook(() => useTelegram());

      // Ждем загрузки с ошибкой
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Проверяем что ошибка установлена
      expect(result.current.error).toBe(errorMessage);

      // Очищаем ошибку вручную
      act(() => {
        result.current.clearError();
      });

      // Проверяем что ошибка сброшена
      expect(result.current.error).toBe(null);
    });

    it('должен сбрасывать токен и botUsername при новой генерации', async () => {
      vi.spyOn(telegramAPI, 'getMyTelegramLink').mockResolvedValue({ linked: false });
      vi.spyOn(telegramAPI, 'generateLinkToken')
        .mockResolvedValueOnce({ token: 'token1', bot_username: 'Bot1' })
        .mockResolvedValueOnce({ token: 'token2', bot_username: 'Bot2' });

      const { result } = renderHook(() => useTelegram());

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Первая генерация
      await act(async () => {
        await result.current.generateToken();
      });
      expect(result.current.token).toBe('token1');
      expect(result.current.botUsername).toBe('Bot1');

      // Вторая генерация
      await act(async () => {
        await result.current.generateToken();
      });
      expect(result.current.token).toBe('token2');
      expect(result.current.botUsername).toBe('Bot2');
    });
  });
});
