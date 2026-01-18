import { useReducer, useCallback, useEffect, useRef } from 'react';
import * as broadcastsAPI from '../api/broadcasts.js';
import { useNotification } from './useNotification.js';

// Reducer для управления состоянием (избегаем race conditions)
const broadcastReducer = (state, action) => {
  switch (action.type) {
    case 'FETCH_START':
      return { ...state, loading: true, error: null };
    case 'FETCH_LINKED_USERS_SUCCESS':
      return { ...state, loading: false, linkedUsers: action.payload };
    case 'FETCH_BROADCAST_LISTS_SUCCESS':
      return { ...state, loading: false, broadcastLists: action.payload };
    case 'FETCH_BROADCASTS_SUCCESS':
      return {
        ...state,
        loading: false,
        broadcasts: action.payload.broadcasts,
        totalBroadcasts: action.payload.total
      };
    case 'CREATE_LIST_SUCCESS':
      return {
        ...state,
        loading: false,
        broadcastLists: [...state.broadcastLists, action.payload]
      };
    case 'UPDATE_LIST_SUCCESS':
      return {
        ...state,
        loading: false,
        broadcastLists: state.broadcastLists.map(list =>
          list.id === action.payload.id ? action.payload : list
        )
      };
    case 'DELETE_LIST_SUCCESS':
      return {
        ...state,
        loading: false,
        broadcastLists: state.broadcastLists.filter(list => list.id !== action.payload)
      };
    case 'CANCEL_BROADCAST_SUCCESS':
      return {
        ...state,
        loading: false,
        broadcasts: state.broadcasts.map(broadcast =>
          broadcast.id === action.payload ? { ...broadcast, status: 'cancelled' } : broadcast
        )
      };
    case 'CLEAR_LOADING':
      return { ...state, loading: false };
    case 'FETCH_ERROR':
      return { ...state, loading: false, error: action.payload };
    default:
      return state;
  }
};

const initialState = {
  linkedUsers: [],
  broadcastLists: [],
  broadcasts: [],
  totalBroadcasts: 0,
  loading: false,
  error: null,
};

/**
 * Кастомный хук для управления рассылками в Telegram
 * FIX T176: Используем useReducer для предотвращения race conditions при concurrent setState
 * FIX T175: useCallback с exhaustive-deps для предотвращения stale closures
 *           Используем ref для notification методов, чтобы callbacks были стабильны
 * FIX T173: Добавлена поддержка AbortController для отмены запросов при unmount
 *
 * ВАЖНО: notification методы хранятся в ref для стабильности callbacks
 * Все callbacks зависят только от dispatch (stable) и refs (stable)
 */
export const useBroadcast = () => {
  const [state, dispatch] = useReducer(broadcastReducer, initialState);
  const notification = useNotification();

  // FIX T175: Храним notification методы в ref для предотвращения stale closures
  // Ref всегда содержит актуальные методы, но не триггерит ре-рендер callbacks
  const notificationRef = useRef(notification);

  // Обновляем ref при изменении notification (хотя методы должны быть stable)
  useEffect(() => {
    notificationRef.current = notification;
  }, [notification]);

  // FIX T173: Ref для хранения активных AbortController
  const abortControllersRef = useRef([]);

  // Cleanup на unmount - отменяем все активные запросы
  useEffect(() => {
    // FIX T175: Сохраняем ref в локальную переменную для cleanup
    const controllers = abortControllersRef.current;
    return () => {
      controllers.forEach(controller => {
        controller.abort();
      });
      controllers.length = 0;
    };
  }, []);

  // Утилита для создания AbortController с автоматической очисткой
  const createRequest = useCallback(() => {
    const controller = new AbortController();
    abortControllersRef.current.push(controller);

    const cleanup = () => {
      const index = abortControllersRef.current.indexOf(controller);
      if (index > -1) {
        abortControllersRef.current.splice(index, 1);
      }
    };

    return { signal: controller.signal, cleanup };
  }, []);

  /**
   * Загрузить пользователей с привязанным Telegram
   */
  const fetchLinkedUsers = useCallback(
    async (role) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        const data = await broadcastsAPI.getLinkedUsers(role, { signal });
        // API client already unwraps {success: true, data: {users: [...], count: N}}
        // So data is already the inner object, or could be array directly
        const users = Array.isArray(data) ? data : (data?.users || []);
        dispatch({ type: 'FETCH_LINKED_USERS_SUCCESS', payload: users });
        return users;
      } catch (err) {
        // Игнорируем AbortError - это нормальное поведение при unmount
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error('Ошибка загрузки пользователей');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Загрузить списки рассылки
   */
  const fetchBroadcastLists = useCallback(async () => {
    const { signal, cleanup } = createRequest();
    try {
      dispatch({ type: 'FETCH_START' });
      const data = await broadcastsAPI.getBroadcastLists({ signal });
      // API client already unwraps {success: true, data: {lists: [...], count: N}}
      // Could be array directly or object with lists property
      const lists = Array.isArray(data) ? data : (data?.lists || []);
      dispatch({ type: 'FETCH_BROADCAST_LISTS_SUCCESS', payload: lists });
      return lists;
    } catch (err) {
      // Игнорируем AbortError
      if (err.name === 'AbortError') {
        return;
      }
      dispatch({ type: 'FETCH_ERROR', payload: err.message });
      notificationRef.current.error('Ошибка загрузки списков рассылки');
      throw err;
    } finally {
      cleanup();
    }
  }, [createRequest]);

  /**
   * Создать список рассылки
   */
  const createList = useCallback(
    async (data) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        const response = await broadcastsAPI.createBroadcastList(data, { signal });
        // API returns {list: {...}} or just the list object
        const newList = response?.list || response;
        dispatch({ type: 'CREATE_LIST_SUCCESS', payload: newList });
        notificationRef.current.success('Список рассылки создан');
        return newList;
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error(err.message || 'Ошибка создания списка');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Обновить список рассылки
   */
  const updateList = useCallback(
    async (id, data) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        const response = await broadcastsAPI.updateBroadcastList(id, data, { signal });
        // API returns {list: {...}} or just the list object
        const updatedList = response?.list || response;
        dispatch({ type: 'UPDATE_LIST_SUCCESS', payload: updatedList });
        notificationRef.current.success('Список обновлен');
        return updatedList;
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error(err.message || 'Ошибка обновления списка');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Удалить список рассылки
   */
  const deleteList = useCallback(
    async (id) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        await broadcastsAPI.deleteBroadcastList(id, { signal });
        dispatch({ type: 'DELETE_LIST_SUCCESS', payload: id });
        notificationRef.current.success('Список удален');
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error(err.message || 'Ошибка удаления списка');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Отправить рассылку
   */
  const sendBroadcast = useCallback(
    async (data) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        const response = await broadcastsAPI.sendBroadcast(data, { signal });
        // API returns {broadcast: {...}, message: "..."}
        const broadcast = response?.broadcast || response;
        dispatch({ type: 'CLEAR_LOADING' });
        notificationRef.current.success('Рассылка отправлена');
        return broadcast;
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error(err.message || 'Ошибка отправки рассылки');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Загрузить историю рассылок
   */
  const fetchBroadcasts = useCallback(
    async (limit = 20, offset = 0) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        const data = await broadcastsAPI.getBroadcasts({ limit, offset }, { signal });
        // API returns {broadcasts: [...], count: N, total: N, limit: N, offset: N}
        // OR could be array directly for backward compatibility
        const broadcasts = Array.isArray(data) ? data : (data?.broadcasts || []);
        const total = Array.isArray(data) ? data.length : (data?.total || broadcasts.length);
        dispatch({
          type: 'FETCH_BROADCASTS_SUCCESS',
          payload: { broadcasts, total }
        });
        return data;
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error('Ошибка загрузки истории рассылок');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Получить детали рассылки
   */
  const getBroadcastDetailsById = useCallback(
    async (id) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        const data = await broadcastsAPI.getBroadcastDetails(id, { signal });
        // API returns {broadcast: {...}, logs: [...]} or the raw data structure
        dispatch({ type: 'CLEAR_LOADING' });
        return data;
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error('Ошибка загрузки деталей рассылки');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  /**
   * Отменить рассылку
   */
  const cancelBroadcastById = useCallback(
    async (id) => {
      const { signal, cleanup } = createRequest();
      try {
        dispatch({ type: 'FETCH_START' });
        await broadcastsAPI.cancelBroadcast(id, { signal });
        dispatch({ type: 'CANCEL_BROADCAST_SUCCESS', payload: id });
        notificationRef.current.success('Рассылка отменена');
      } catch (err) {
        // Игнорируем AbortError
        if (err.name === 'AbortError') {
          return;
        }
        dispatch({ type: 'FETCH_ERROR', payload: err.message });
        notificationRef.current.error(err.message || 'Ошибка отмены рассылки');
        throw err;
      } finally {
        cleanup();
      }
    },
    [createRequest]
  );

  return {
    linkedUsers: state.linkedUsers,
    broadcastLists: state.broadcastLists,
    broadcasts: state.broadcasts,
    totalBroadcasts: state.totalBroadcasts,
    loading: state.loading,
    error: state.error,
    fetchLinkedUsers,
    fetchBroadcastLists,
    createList,
    updateList,
    deleteList,
    sendBroadcast,
    fetchBroadcasts,
    getBroadcastDetailsById,
    cancelBroadcastById,
  };
};

export default useBroadcast;
