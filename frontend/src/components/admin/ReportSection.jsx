import { useState, useRef, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useNotification } from '../../hooks/useNotification.js';
import { useAuth } from '../../hooks/useAuth.js';
import * as lessonAPI from '../../api/lessons.js';
import Button from '../common/Button.jsx';
import './ReportSection.css';

/**
 * Компонент секции отчета о занятии
 * Позволяет преподавателю и админу добавлять текстовый отчет о проведенном занятии
 */
export const ReportSection = ({ lessonId, lesson, students = [] }) => {
  const { user } = useAuth();
  const { showNotification } = useNotification();
  const queryClient = useQueryClient();
  const saveTimeoutRef = useRef(null);
  const pendingTextRef = useRef(null);

  const [reportText, setReportText] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState(null);
  const [sendingReport, setSendingReport] = useState(false);

  // Проверка прав доступа
  // Admin может редактировать отчеты всех уроков
  // Teacher может редактировать отчеты только своих уроков (где teacher_id === user.id)
  // Teacher может редактировать отчеты всех уроков
  // Student может только просматривать (canEdit = false)
  const normalizeId = (id) => String(id || '').toLowerCase().trim();
  const isTeacherOwnLesson = user?.role === 'teacher' &&
    lesson?.teacher_id &&
    normalizeId(lesson.teacher_id) === normalizeId(user?.id);

  const canEdit = user?.role === 'admin' ||
    isTeacherOwnLesson ||
    user?.role === 'teacher';

  // Проверка: можно ли редактировать отчет (только после начала занятия)
  const canEditReport = () => {
    if (!canEdit) return false;
    if (user?.role === 'admin') return true;
    return new Date(lesson?.start_time || 0) <= new Date();
  };

  /**
   * Локально обновляем кэш уроков, чтобы сразу видеть новое описание
   */
  const updateLessonCaches = (lessonIdToUpdate, text) => {
    const updateValue = (cached) => {
      if (!cached) return cached;

      const updateItem = (item) =>
        item?.id === lessonIdToUpdate ? { ...item, report_text: text } : item;

      if (Array.isArray(cached)) {
        return cached.map(updateItem);
      }

      if (Array.isArray(cached?.lessons)) {
        return { ...cached, lessons: cached.lessons.map(updateItem) };
      }

      if (Array.isArray(cached?.data)) {
        return { ...cached, data: cached.data.map(updateItem) };
      }

      if (cached?.lesson?.id === lessonIdToUpdate) {
        return { ...cached, lesson: { ...cached.lesson, report_text: text } };
      }

      if (cached?.id === lessonIdToUpdate) {
        return { ...cached, report_text: text };
      }

      return cached;
    };

    ['lessons', 'myLessons', 'studentLessons', 'teacher-schedule'].forEach((key) => {
      queryClient.setQueriesData({ queryKey: [key], exact: false }, updateValue);
    });
  };

  /**
   * Сохранение текста отчета
   */
  const saveReportText = async (text) => {
    if (!canEditReport()) {
      showNotification('Отчет можно редактировать только после начала занятия', 'error');
      return;
    }

    setIsSaving(true);
    try {
      await lessonAPI.updateLesson(lessonId, { report_text: text });

      updateLessonCaches(lessonId, text);
      setLastSaved(new Date());

      // Инвалидировать React Query кеши для уроков
      queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
      queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false });
    } catch (error) {
      console.error('Error saving report text:', error);
      const errorMessage = error.response?.data?.message ||
                          error.response?.data?.error?.message ||
                          error.message ||
                          'Не удалось сохранить отчет';
      showNotification(errorMessage, 'error');
    } finally {
      setIsSaving(false);
    }
  };

  /**
   * Обработка изменения текста с debounce
   */
  const handleTextChange = (e) => {
    const newText = e.target.value;
    setReportText(newText);
    pendingTextRef.current = newText;

    // Очистить предыдущий таймер
    if (saveTimeoutRef.current) {
      clearTimeout(saveTimeoutRef.current);
    }

    // Установить новый таймер на автосохранение (500ms)
    saveTimeoutRef.current = setTimeout(() => {
      if (canEditReport() && typeof newText === 'string') {
        const textToSave = pendingTextRef.current;
        saveReportText(textToSave).then(() => {
          if (pendingTextRef.current === textToSave) {
            pendingTextRef.current = null;
          }
        });
      }
    }, 500);
  };

  /**
   * Загрузка текста отчета при монтировании или смене урока
   */
  useEffect(() => {
    if (lesson?.report_text !== undefined) {
      setReportText(lesson.report_text || '');
    }
  }, [lesson, lessonId]);

  /**
   * Отправка отчета родителям
   */
  const handleSendReportToParents = async () => {
    if (!reportText) {
      showNotification('Отчет о занятии пустой', 'error');
      return;
    }

    setSendingReport(true);
    try {
      const result = await lessonAPI.sendReportToParents(lessonId);

      if (result.sent > 0) {
        showNotification(`Отчет отправлен ${result.sent} родителям`, 'success');
      }
      if (result.failed > 0) {
        showNotification(`Не удалось отправить ${result.failed}`, 'warning');
        if (result.errors?.length > 0) {
          console.error('Failed to send reports:', result.errors);
        }
      }
    } catch (error) {
      console.error('Error sending report to parents:', error);
      const errorMessage = error.response?.data?.message ||
                          error.response?.data?.error?.message ||
                          error.message ||
                          'Не удалось отправить отчет';
      showNotification(errorMessage, 'error');
    } finally {
      setSendingReport(false);
    }
  };

  /**
   * Очистка таймера при размонтировании
   */
  useEffect(() => {
    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
        saveTimeoutRef.current = null;
      }

      // Если есть несохранённый текст и пользователь имеет право редактировать,
      // пытаемся сохранить его "на выходе"
      if (canEditReport() && pendingTextRef.current !== null) {
        const textToSave = pendingTextRef.current;
        pendingTextRef.current = null;
        saveReportText(textToSave);
      }
    };
  }, [canEditReport]);

  const isEditable = canEditReport();

  return (
    <div className="report-section">
      {/* Предупреждение если занятие еще не началось */}
      {!isEditable && canEdit && (
        <div className="report-warning">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 1a7 7 0 100 14A7 7 0 008 1zM7 4h2v4H7V4zm0 5h2v2H7V9z"/>
          </svg>
          <span>Отчет можно редактировать только после начала занятия</span>
        </div>
      )}

      <div className="report-text-section">
        <div className="report-text-header">
          <label htmlFor={`report-text-${lessonId}`}>
            Отчет о занятии
          </label>
          <div className="report-text-actions">
            {isEditable && (
              <span className="report-text-status">
                {isSaving ? (
                  <span className="saving-indicator">Сохранение...</span>
                ) : lastSaved ? (
                  <span className="saved-indicator">
                    Сохранено {new Date(lastSaved).toLocaleTimeString('ru-RU')}
                  </span>
                ) : null}
              </span>
            )}
          </div>
        </div>

        {/* Для редактирования (admin/teacher после начала занятия) показываем textarea */}
        {isEditable ? (
          <textarea
            id={`report-text-${lessonId}`}
            className="report-text-input"
            value={reportText}
            onChange={handleTextChange}
            placeholder="Опишите, что было на занятии..."
            rows={8}
          />
        ) : (
          /* Для просмотра (student или до начала занятия) показываем div */
          <div className="report-text-display">
            {reportText ? (
              <div className="report-text-content">
                {reportText.split('\n').map((line, idx) => (
                  <p key={idx}>{line || '\u00A0'}</p>
                ))}
              </div>
            ) : (
              <span className="report-text-empty">Отчет не добавлен</span>
            )}
          </div>
        )}

        {/* Кнопка отправки отчета родителям (для admin и teacher) */}
        {reportText && students.length > 0 && (user?.role === 'admin' || user?.role === 'teacher') && (
          <div className="report-parent-actions">
            <Button
              onClick={handleSendReportToParents}
              disabled={sendingReport}
              variant="secondary"
            >
              {sendingReport ? 'Отправка...' : '📤 Отправить отчет родителям'}
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};

export default ReportSection;
