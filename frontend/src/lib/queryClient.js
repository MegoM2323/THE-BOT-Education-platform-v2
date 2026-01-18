import { QueryClient } from '@tanstack/react-query';

/**
 * Глобальный QueryClient для React Query
 *
 * MEMORY LEAK PREVENTION (T173):
 * React Query автоматически отменяет все queries и mutations при размонтировании компонента.
 * Это предотвращает утечки памяти и "setState on unmounted component" warnings.
 *
 * Подробнее: frontend/docs/FETCH_CANCELLATION.md
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 минут
      gcTime: 10 * 60 * 1000, // 10 минут (старое название: cacheTime)
      refetchOnWindowFocus: true, // Автоматическое обновление при возврате в окно для синхронизации с сервером
      refetchOnReconnect: true, // Автоматическое обновление при восстановлении сети
      retry: 1,
      retryDelay: 1000,
      // ВАЖНО: React Query автоматически отменяет запросы при unmount
      // Нет необходимости в ручной отмене запросов в хуках
    },
    mutations: {
      retry: 1,
      retryDelay: 1000,
      // ВАЖНО: Mutations также автоматически отменяются при unmount
    },
  },
});
