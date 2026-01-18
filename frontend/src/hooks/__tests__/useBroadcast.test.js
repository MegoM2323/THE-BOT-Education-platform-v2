import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useBroadcast } from '../useBroadcast';
import * as broadcastsAPI from '../../api/broadcasts';
import * as useNotificationModule from '../useNotification';

// Мокаем API и notification hook
vi.mock('../../api/broadcasts');
vi.mock('../useNotification');

describe('useBroadcast', () => {
  const mockNotification = {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(useNotificationModule, 'useNotification').mockReturnValue(mockNotification);
  });

  describe('fetchLinkedUsers', () => {
    it('должен успешно загружать пользователей с Telegram', async () => {
      const mockUsers = [
        { user_id: '1', telegram_id: 111, username: 'user1', role: 'student' },
        { user_id: '2', telegram_id: 222, username: 'user2', role: 'student' },
      ];

      vi.spyOn(broadcastsAPI, 'getLinkedUsers').mockResolvedValue(mockUsers);

      const { result } = renderHook(() => useBroadcast());

      expect(result.current.linkedUsers).toEqual([]);

      await act(async () => {
        await result.current.fetchLinkedUsers('student');
      });

      expect(result.current.linkedUsers).toEqual(mockUsers);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(null);
      expect(broadcastsAPI.getLinkedUsers).toHaveBeenCalledWith('student', expect.objectContaining({ signal: expect.any(AbortSignal) }));
    });

    it('должен обрабатывать ошибки загрузки пользователей', async () => {
      const errorMessage = 'Failed to fetch users';
      vi.spyOn(broadcastsAPI, 'getLinkedUsers').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.fetchLinkedUsers();
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockNotification.error).toHaveBeenCalledWith('Ошибка загрузки пользователей');
    });

    it('должен фильтровать пользователей по роли', async () => {
      const mockStudents = [
        { user_id: '1', telegram_id: 111, username: 'student1', role: 'student' },
      ];

      vi.spyOn(broadcastsAPI, 'getLinkedUsers').mockResolvedValue(mockStudents);

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        await result.current.fetchLinkedUsers('student');
      });

      expect(broadcastsAPI.getLinkedUsers).toHaveBeenCalledWith('student', expect.objectContaining({ signal: expect.any(AbortSignal) }));
      expect(result.current.linkedUsers).toEqual(mockStudents);
    });
  });

  describe('fetchBroadcastLists', () => {
    it('должен успешно загружать списки рассылки', async () => {
      const mockLists = [
        { id: '1', name: 'Students', user_ids: ['1', '2'] },
        { id: '2', name: 'Teachers', user_ids: ['3'] },
      ];

      vi.spyOn(broadcastsAPI, 'getBroadcastLists').mockResolvedValue(mockLists);

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        await result.current.fetchBroadcastLists();
      });

      expect(result.current.broadcastLists).toEqual(mockLists);
      expect(result.current.loading).toBe(false);
      expect(broadcastsAPI.getBroadcastLists).toHaveBeenCalledTimes(1);
    });

    it('должен обрабатывать ошибки загрузки списков', async () => {
      const errorMessage = 'Failed to load lists';
      vi.spyOn(broadcastsAPI, 'getBroadcastLists').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.fetchBroadcastLists();
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockNotification.error).toHaveBeenCalledWith('Ошибка загрузки списков рассылки');
    });
  });

  describe('createList', () => {
    it('должен успешно создавать список рассылки', async () => {
      const newListData = {
        name: 'New List',
        description: 'Description',
        user_ids: ['1', '2'],
      };

      const createdList = {
        id: '123',
        ...newListData,
      };

      vi.spyOn(broadcastsAPI, 'createBroadcastList').mockResolvedValue(createdList);

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        await result.current.createList(newListData);
      });

      expect(result.current.broadcastLists).toContainEqual(createdList);
      expect(mockNotification.success).toHaveBeenCalledWith('Список рассылки создан');
      expect(broadcastsAPI.createBroadcastList).toHaveBeenCalledWith(newListData, expect.objectContaining({ signal: expect.any(AbortSignal) }));
    });

    it('должен обрабатывать ошибки создания списка', async () => {
      const errorMessage = 'Validation error';
      vi.spyOn(broadcastsAPI, 'createBroadcastList').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.createList({ name: 'AB' }); // Короткое имя
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });
  });

  describe('updateList', () => {
    it('должен успешно обновлять список рассылки', async () => {
      const existingList = { id: '123', name: 'Old Name', user_ids: ['1'] };
      const updatedData = { name: 'New Name', user_ids: ['1', '2'] };
      const updatedList = { ...existingList, ...updatedData };

      vi.spyOn(broadcastsAPI, 'updateBroadcastList').mockResolvedValue(updatedList);

      const { result } = renderHook(() => useBroadcast());

      // Устанавливаем начальный список
      act(() => {
        result.current.broadcastLists.push(existingList);
      });

      await act(async () => {
        await result.current.updateList('123', updatedData);
      });

      const found = result.current.broadcastLists.find((l) => l.id === '123');
      expect(found).toEqual(updatedList);
      expect(mockNotification.success).toHaveBeenCalledWith('Список обновлен');
    });

    it('должен обрабатывать ошибки обновления списка', async () => {
      const errorMessage = 'List not found';
      vi.spyOn(broadcastsAPI, 'updateBroadcastList').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.updateList('invalid-id', {});
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });
  });

  describe('deleteList', () => {
    it('должен успешно удалять список рассылки', async () => {
      vi.spyOn(broadcastsAPI, 'deleteBroadcastList').mockResolvedValue();

      const { result } = renderHook(() => useBroadcast());

      // Создаем тестовый список
      act(() => {
        result.current.broadcastLists.push({ id: '123', name: 'Test List' });
      });

      await act(async () => {
        await result.current.deleteList('123');
      });

      expect(result.current.broadcastLists).not.toContainEqual(
        expect.objectContaining({ id: '123' })
      );
      expect(mockNotification.success).toHaveBeenCalledWith('Список удален');
    });

    it('должен обрабатывать ошибки удаления списка', async () => {
      const errorMessage = 'Cannot delete list';
      vi.spyOn(broadcastsAPI, 'deleteBroadcastList').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.deleteList('123');
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });
  });

  describe('sendBroadcast', () => {
    it('должен успешно отправлять рассылку', async () => {
      const broadcastData = {
        list_id: '123',
        message: 'Test message',
      };

      const result_broadcast = {
        id: 'broadcast-1',
        status: 'pending',
        ...broadcastData,
      };

      vi.spyOn(broadcastsAPI, 'sendBroadcast').mockResolvedValue(result_broadcast);

      const { result } = renderHook(() => useBroadcast());

      let response;
      await act(async () => {
        response = await result.current.sendBroadcast(broadcastData);
      });

      expect(response).toEqual(result_broadcast);
      expect(mockNotification.success).toHaveBeenCalledWith('Рассылка отправлена');
      expect(broadcastsAPI.sendBroadcast).toHaveBeenCalledWith(broadcastData, expect.objectContaining({ signal: expect.any(AbortSignal) }));
    });

    it('должен обрабатывать ошибки отправки рассылки', async () => {
      const errorMessage = 'Message too long';
      vi.spyOn(broadcastsAPI, 'sendBroadcast').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.sendBroadcast({ message: 'x'.repeat(5000) });
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });
  });

  describe('fetchBroadcasts', () => {
    it('должен успешно загружать историю рассылок', async () => {
      const mockBroadcasts = [
        { id: '1', message: 'Message 1', status: 'completed' },
        { id: '2', message: 'Message 2', status: 'in_progress' },
      ];

      const mockResponse = {
        broadcasts: mockBroadcasts,
        total: 2,
      };

      vi.spyOn(broadcastsAPI, 'getBroadcasts').mockResolvedValue(mockResponse);

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        await result.current.fetchBroadcasts(20, 0);
      });

      expect(result.current.broadcasts).toEqual(mockBroadcasts);
      expect(result.current.totalBroadcasts).toBe(2);
      expect(broadcastsAPI.getBroadcasts).toHaveBeenCalledWith({ limit: 20, offset: 0 }, expect.objectContaining({ signal: expect.any(AbortSignal) }));
    });

    it('должен обрабатывать пагинацию', async () => {
      const mockResponse = {
        broadcasts: [{ id: '3', message: 'Message 3' }],
        total: 10,
      };

      vi.spyOn(broadcastsAPI, 'getBroadcasts').mockResolvedValue(mockResponse);

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        await result.current.fetchBroadcasts(20, 40);
      });

      expect(broadcastsAPI.getBroadcasts).toHaveBeenCalledWith({ limit: 20, offset: 40 }, expect.objectContaining({ signal: expect.any(AbortSignal) }));
    });

    it('должен обрабатывать ошибки загрузки истории', async () => {
      const errorMessage = 'Server error';
      vi.spyOn(broadcastsAPI, 'getBroadcasts').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.fetchBroadcasts();
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(mockNotification.error).toHaveBeenCalledWith('Ошибка загрузки истории рассылок');
    });
  });

  describe('getBroadcastDetailsById', () => {
    it('должен успешно загружать детали рассылки', async () => {
      const mockDetails = {
        broadcast: { id: '1', message: 'Test', status: 'completed' },
        logs: [
          { user_id: '1', status: 'success', telegram_id: 111 },
          { user_id: '2', status: 'failed', telegram_id: 222, error: 'Bot blocked' },
        ],
      };

      vi.spyOn(broadcastsAPI, 'getBroadcastDetails').mockResolvedValue(mockDetails);

      const { result } = renderHook(() => useBroadcast());

      let details;
      await act(async () => {
        details = await result.current.getBroadcastDetailsById('1');
      });

      expect(details).toEqual(mockDetails);
      expect(broadcastsAPI.getBroadcastDetails).toHaveBeenCalledWith('1', expect.objectContaining({ signal: expect.any(AbortSignal) }));
    });

    it('должен обрабатывать ошибки загрузки деталей', async () => {
      const errorMessage = 'Broadcast not found';
      vi.spyOn(broadcastsAPI, 'getBroadcastDetails').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.getBroadcastDetailsById('invalid-id');
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(mockNotification.error).toHaveBeenCalledWith('Ошибка загрузки деталей рассылки');
    });
  });

  describe('cancelBroadcastById', () => {
    it('должен успешно отменять рассылку', async () => {
      vi.spyOn(broadcastsAPI, 'cancelBroadcast').mockResolvedValue();

      const { result } = renderHook(() => useBroadcast());

      // Устанавливаем начальный список рассылок
      act(() => {
        result.current.broadcasts.push({
          id: '123',
          message: 'Test',
          status: 'in_progress',
        });
      });

      await act(async () => {
        await result.current.cancelBroadcastById('123');
      });

      const broadcast = result.current.broadcasts.find((b) => b.id === '123');
      expect(broadcast.status).toBe('cancelled');
      expect(mockNotification.success).toHaveBeenCalledWith('Рассылка отменена');
    });

    it('должен обрабатывать ошибки отмены рассылки', async () => {
      const errorMessage = 'Cannot cancel completed broadcast';
      vi.spyOn(broadcastsAPI, 'cancelBroadcast').mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useBroadcast());

      await act(async () => {
        try {
          await result.current.cancelBroadcastById('123');
        } catch (err) {
          expect(err.message).toBe(errorMessage);
        }
      });

      expect(mockNotification.error).toHaveBeenCalledWith(errorMessage);
    });
  });

  describe('loading state', () => {
    it('должен устанавливать loading во время операций', async () => {
      vi.spyOn(broadcastsAPI, 'getLinkedUsers').mockImplementation(
        () => new Promise((resolve) => setTimeout(() => resolve([]), 100))
      );

      const { result } = renderHook(() => useBroadcast());

      act(() => {
        result.current.fetchLinkedUsers();
      });

      expect(result.current.loading).toBe(true);

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
    });

    it('должен сбрасывать error перед новой операцией', async () => {
      vi.spyOn(broadcastsAPI, 'getLinkedUsers')
        .mockRejectedValueOnce(new Error('First error'))
        .mockResolvedValueOnce([]);

      const { result } = renderHook(() => useBroadcast());

      // Первый вызов с ошибкой
      await act(async () => {
        try {
          await result.current.fetchLinkedUsers();
        } catch (err) {
          // ignore
        }
      });
      expect(result.current.error).toBe('First error');

      // Второй вызов успешный
      await act(async () => {
        await result.current.fetchLinkedUsers();
      });

      // Ошибка должна быть сброшена
      expect(result.current.error).toBe(null);
    });
  });
});
