import { useQuery } from "@tanstack/react-query";
import * as lessonsAPI from "../api/lessons.js";

/**
 * Хук для загрузки расписания преподавателя
 * Использует эндпоинт /api/v1/teacher/schedule который фильтрует занятия по teacher_id на backend
 *
 * @param {string} startDate - Начало диапазона (ISO format)
 * @param {string} endDate - Конец диапазона (ISO format)
 * @returns {Object} Query result с lessons, isLoading, error
 */
export const useTeacherSchedule = (startDate, endDate) => {
  return useQuery({
    queryKey: ["teacher-schedule", startDate, endDate],
    queryFn: async () => {
      // Backend автоматически фильтрует по teacher_id текущего пользователя
      // и возвращает {lessons: [...], count: N}
      const data = await lessonsAPI.getTeacherSchedule({
        start_date: startDate,
        end_date: endDate,
      });

      // Нормализация данных (backend возвращает {lessons, count})
      const lessonsArray = Array.isArray(data) ? data : data?.lessons || [];

      return lessonsArray;
    },
    enabled: !!startDate && !!endDate,
    staleTime: 30000, // 30 секунд
  });
};

export default useTeacherSchedule;
