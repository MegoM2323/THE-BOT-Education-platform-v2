/**
 * Custom hooks для работы с рассылками уроков
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as lessonBroadcastsAPI from '../api/lessonBroadcasts.js';

/**
 * Hook для получения истории рассылок урока
 * @param {string} lessonId - UUID урока
 * @returns {Object} Query result с broadcasts data
 */
export const useLessonBroadcasts = (lessonId) => {
  return useQuery({
    queryKey: ['lessonBroadcasts', lessonId],
    queryFn: async ({ signal }) => {
      // Передаем signal правильно - API функция ожидает его как 3-й аргумент
      return await lessonBroadcastsAPI.getLessonBroadcasts(lessonId, {}, signal);
    },
    enabled: !!lessonId,
    staleTime: 30000, // 30 секунд
  });
};

/**
 * Hook для отправки рассылки
 * @returns {Object} Mutation object
 */
export const useSendLessonBroadcast = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ lessonId, message, files, signal }) =>
      lessonBroadcastsAPI.createLessonBroadcast(lessonId, message, files, signal),
    onSuccess: (data, variables) => {
      // Инвалидировать кэш для этого урока
      queryClient.invalidateQueries({ queryKey: ['lessonBroadcasts', variables.lessonId] });
    },
  });
};

/**
 * Hook для получения деталей рассылки
 * @param {string} lessonId - UUID урока
 * @param {string} broadcastId - UUID рассылки
 * @returns {Object} Query result с broadcast details
 */
export const useBroadcastDetails = (lessonId, broadcastId) => {
  return useQuery({
    queryKey: ['broadcastDetails', lessonId, broadcastId],
    queryFn: async ({ signal }) => {
      // Передаем signal правильно - API функция ожидает его как 3-й аргумент
      return await lessonBroadcastsAPI.getLessonBroadcast(lessonId, broadcastId, signal);
    },
    enabled: !!lessonId && !!broadcastId,
  });
};

export default {
  useLessonBroadcasts,
  useSendLessonBroadcast,
  useBroadcastDetails,
};
