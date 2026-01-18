/**
 * Custom hooks для работы с домашними заданиями
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as homeworkAPI from '../api/homework.js';

/**
 * Hook для получения списка домашних заданий урока
 * @param {string} lessonId - UUID урока
 * @returns {Object} Query result с homework data
 */
export const useHomework = (lessonId) => {
  return useQuery({
    queryKey: ['homework', lessonId],
    queryFn: async ({ signal }) => {
      // Передаем signal правильно - API функция ожидает его как 2-й аргумент
      return await homeworkAPI.getHomework(lessonId, signal);
    },
    enabled: !!lessonId,
    staleTime: 30000, // 30 секунд
    // Сохраняем предыдущие данные при ошибке, чтобы пользователь видел уже загруженные файлы
    placeholderData: (previousData) => previousData,
  });
};

/**
 * Hook для загрузки домашнего задания
 * @returns {Object} Mutation object
 */
export const useUploadHomework = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ lessonId, file, signal }) => homeworkAPI.uploadHomework(lessonId, file, signal),
    onSuccess: (data, variables) => {
      // Инвалидировать кэш для этого урока
      // В случае ошибки при обновлении списка, она будет обработана в компоненте
      // но успех загрузки файла не будет нарушен
      queryClient.invalidateQueries({ queryKey: ['homework', variables.lessonId] });
    },
  });
};

/**
 * Hook для удаления домашнего задания
 * @returns {Object} Mutation object
 */
export const useDeleteHomework = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ lessonId, fileId, signal }) => homeworkAPI.deleteHomework(lessonId, fileId, signal),
    onSuccess: (data, variables) => {
      // Инвалидировать кэш для этого урока
      queryClient.invalidateQueries({ queryKey: ['homework', variables.lessonId] });
    },
  });
};

export default {
  useHomework,
  useUploadHomework,
  useDeleteHomework,
};
