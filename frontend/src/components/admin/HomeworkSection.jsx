import { useState, useRef, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useHomework, useUploadHomework, useDeleteHomework } from '../../hooks/useHomework.js';
import { useNotification } from '../../hooks/useNotification.js';
import { useAuth } from '../../hooks/useAuth.js';
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import ConfirmModal from '../common/ConfirmModal.jsx';
import * as homeworkAPI from '../../api/homework.js';
import * as lessonAPI from '../../api/lessons.js';
import { formatHomeworkText } from '../../utils/formatHomeworkText.js';
import './HomeworkSection.css';

/**
 * Форматирует размер файла в читаемый вид
 * @param {number} bytes - Размер файла в байтах
 * @returns {string} Отформатированный размер
 */
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
};

/**
 * Компонент секции домашних заданий
 * @param {function} [onHomeworkSaved] - колбэк после успешного сохранения текста
 * @param {boolean} [readOnly] - режим только для чтения (блокирует редактирование)
 */
export const HomeworkSection = ({ lessonId, lesson, onHomeworkCountChange, onHomeworkSaved, readOnly = false }) => {
  const { user } = useAuth();
  const { showNotification } = useNotification();
  const queryClient = useQueryClient();
  const fileInputRef = useRef(null);
  const saveTimeoutRef = useRef(null);
  const pendingTextRef = useRef(null);

  const [isDragging, setIsDragging] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [fileToDelete, setFileToDelete] = useState(null);
  const [homeworkText, setHomeworkText] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState(null);
  // Отслеживание ID урока для предотвращения race condition
  const [currentLessonId, setCurrentLessonId] = useState(lessonId);

  // Загрузка данных домашних заданий
  // React Query автоматически отменяет предыдущие запросы при изменении queryKey
  // Используем гарантированный массив, чтобы избежать ошибок при валидации
  const { data: homeworkData, isLoading, error, isFetching } = useHomework(lessonId);
  const homeworkList = Array.isArray(homeworkData) ? homeworkData : [];

  // Mutations
  const uploadMutation = useUploadHomework();
  const deleteMutation = useDeleteHomework();

  // Проверка прав доступа
  // Admin может редактировать все уроки
  // Teacher может редактировать только свои уроки (где teacher_id === user.id)
  // Methodologist может редактировать все уроки
  // Student может только просматривать (canEdit = false)
  // Примечание: Нормализуем UUID к lowercase для надежного сравнения
  const normalizeId = (id) => String(id || '').toLowerCase().trim();
  const isTeacherOwnLesson = user?.role === 'teacher' &&
    lesson?.teacher_id &&
    normalizeId(lesson.teacher_id) === normalizeId(user?.id);

  const canEdit = !readOnly && (
    user?.role === 'admin' ||
    isTeacherOwnLesson ||
    user?.role === 'methodologist'
  );

  /**
   * Локально обновляем кэш уроков, чтобы сразу видеть новое описание
   * даже до завершения повторной загрузки данных
   */
  const updateLessonCaches = (lessonIdToUpdate, text) => {
    const updateValue = (cached) => {
      if (!cached) return cached;

      const updateItem = (item) =>
        item?.id === lessonIdToUpdate ? { ...item, homework_text: text } : item;

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
        return { ...cached, lesson: { ...cached.lesson, homework_text: text } };
      }

      if (cached?.id === lessonIdToUpdate) {
        return { ...cached, homework_text: text };
      }

      return cached;
    };

    ['lessons', 'myLessons', 'studentLessons', 'teacher-schedule'].forEach((key) => {
      queryClient.setQueriesData({ queryKey: [key], exact: false }, updateValue);
    });
  };

  // Обновить счетчик домашних заданий в родительском компоненте
  // ВАЖНО: Обновляем только если данные соответствуют текущему уроку (защита от race condition)
  useEffect(() => {
    if (onHomeworkCountChange && homeworkList.length !== undefined && lessonId === currentLessonId) {
      onHomeworkCountChange(homeworkList.length);
    }
  }, [homeworkList.length, onHomeworkCountChange, lessonId, currentLessonId]);

  /**
   * Сохранение текста домашнего задания
   * ВАЖНО: Сохраняем только если урок не изменился (защита от race condition)
   */
  const saveHomeworkText = async (text, targetLessonId) => {
    // Проверяем, что урок не изменился с момента начала редактирования
    if (targetLessonId !== currentLessonId) {
      console.warn('Homework text save cancelled: lesson changed');
      return;
    }

    setIsSaving(true);
    try {
      await lessonAPI.updateLesson(targetLessonId, { homework_text: text });

      // Проверяем снова после сохранения
      if (targetLessonId === currentLessonId) {
        updateLessonCaches(targetLessonId, text);
        if (onHomeworkSaved) {
          onHomeworkSaved(text);
        }
        setLastSaved(new Date());

        // Инвалидировать React Query кеши для уроков чтобы переload обновленные данные
        // Это гарантирует что если пользователь откроет урок заново, будут новые данные
        queryClient.invalidateQueries({ queryKey: ['lessons'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['myLessons'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['studentLessons'], exact: false });
        queryClient.invalidateQueries({ queryKey: ['teacher-schedule'], exact: false });
      }
    } catch (error) {
      console.error('Error saving homework text:', error);
      // Показываем уведомление только если урок всё ещё актуален
      if (targetLessonId === currentLessonId) {
        // Извлекаем сообщение об ошибке из ответа API
        const errorMessage = error.response?.data?.message || 
                            error.response?.data?.error?.message ||
                            error.message ||
                            'Не удалось сохранить описание';
        showNotification(errorMessage, 'error');
      }
    } finally {
      setIsSaving(false);
    }
  };

  /**
   * Обработка изменения текста с debounce
   */
  const handleTextChange = (e) => {
    const newText = e.target.value;
    setHomeworkText(newText);
     // Сохраняем последнее введённое значение для возможного сохранения при размонтировании
    pendingTextRef.current = newText;

    // Очистить предыдущий таймер
    if (saveTimeoutRef.current) {
      clearTimeout(saveTimeoutRef.current);
    }

    // Сохраняем ID урока в момент создания таймера
    const targetLessonId = lessonId;

    // Установить новый таймер на автосохранение (500ms)
    saveTimeoutRef.current = setTimeout(() => {
      if (canEdit) {
        const textToSave = pendingTextRef.current;
        // Передаём сохранённый ID урока для проверки в saveHomeworkText
        if (typeof textToSave === 'string') {
          saveHomeworkText(textToSave, targetLessonId).then(() => {
            // Очищаем pendingTextRef только если мы всё ещё сохраняем тот же урок
            if (targetLessonId === currentLessonId && pendingTextRef.current === textToSave) {
              pendingTextRef.current = null;
            }
          });
        }
      }
    }, 500);
  };

  /**
   * Отслеживание смены урока для предотвращения race condition
   * При смене lessonId сбрасываем состояние и отменяем pending операции
   */
  useEffect(() => {
    if (lessonId !== currentLessonId) {
      // Отменить pending сохранение текста
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
        saveTimeoutRef.current = null;
      }

      // Сбросить состояние сохранения
      setIsSaving(false);
      setLastSaved(null);

      // Обновить текущий ID урока
      setCurrentLessonId(lessonId);

      // React Query автоматически отменит предыдущий запрос благодаря AbortSignal
      // и запустит новый с обновленным lessonId в queryKey
    }
  }, [lessonId, currentLessonId]);

  /**
   * Загрузка текста домашнего задания при монтировании или смене урока
   */
  useEffect(() => {
    if (lesson?.homework_text !== undefined) {
      setHomeworkText(lesson.homework_text || '');
    }
  }, [lesson]);

  /**
   * Очистка таймера при размонтировании
   * ВАЖНО: Если был несохранённый текст, при размонтировании попытаемся его сохранить
   */
  useEffect(() => {
    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
        saveTimeoutRef.current = null;
      }

      // Если есть несохранённый текст и пользователь имеет право редактировать,
      // пытаемся сохранить его "на выходе", чтобы не терять изменения
      if (canEdit && pendingTextRef.current !== null && currentLessonId) {
        const textToSave = pendingTextRef.current;
        pendingTextRef.current = null;
        // Запускаем сохранение асинхронно, не дожидаясь завершения
        // (компонент уже размонтируется, но API-запрос выполнится)
        saveHomeworkText(textToSave, currentLessonId);
      }
    };
  }, [canEdit, currentLessonId]);

  /**
   * Обработка выбора файла
   */
  const handleFileSelect = async (files) => {
    if (!files || files.length === 0) return;

    const file = files[0];

    // Валидация размера (макс 10MB)
    if (file.size > 10 * 1024 * 1024) {
      showNotification('Размер файла не должен превышать 10MB', 'error');
      return;
    }

    // Валидация количества файлов (макс 10)
    // Проверяем только если список успешно загружен (не при ошибке)
    if (!error && homeworkList.length >= 10) {
      showNotification('Максимум 10 файлов на урок', 'error');
      return;
    }

    try {
      await uploadMutation.mutateAsync({ lessonId, file });
      showNotification('Файл успешно загружен', 'success');
      // Очистить input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch (error) {
      console.error('Ошибка загрузки файла:', error);
      // Извлекаем сообщение об ошибке из ответа API
      const errorMsg = error?.response?.data?.error?.message || 
                      error?.response?.data?.message || 
                      error?.message || 
                      'Не удалось загрузить файл';
      showNotification(errorMsg, 'error');
    }
  };

  /**
   * Обработка drag and drop
   */
  const handleDragOver = (e) => {
    e.preventDefault();
    if (!canEdit) return;
    setIsDragging(true);
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    if (!canEdit) return;

    const files = e.dataTransfer.files;
    handleFileSelect(files);
  };

  /**
   * Обработка клика на input
   */
  const handleUploadClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  /**
   * Обработка удаления файла
   */
  const handleDeleteClick = (file) => {
    setFileToDelete(file);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!fileToDelete) return;

    try {
      await deleteMutation.mutateAsync({ lessonId, fileId: fileToDelete.id });
      showNotification('Файл успешно удалён', 'success');
      setShowDeleteConfirm(false);
      setFileToDelete(null);
    } catch (error) {
      console.error('Ошибка удаления файла:', error);
      const errorMsg = error.response?.data?.message || 'Не удалось удалить файл';
      showNotification(errorMsg, 'error');
    }
  };

  /**
   * Обработка скачивания файла
   */
  const handleDownload = (file) => {
    homeworkAPI.downloadHomework(lessonId, file.id);
  };

  // Показываем полный спиннер только при первой загрузке
  if (isLoading) {
    return (
      <div className="homework-section">
        <Spinner />
      </div>
    );
  }

  return (
    <>
      <div className="homework-section" style={{ position: 'relative' }}>
        {/* Показываем ошибку как предупреждение, но не скрываем компонент */}
        {error && (
          <div className="homework-error-message">
            ⚠️ Ошибка загрузки списка домашних заданий. Файлы могут быть не синхронизированы.
          </div>
        )}
        
        {/* Индикатор загрузки при переключении между уроками (isFetching без isLoading) */}
        {isFetching && !isLoading && (
          <div className="homework-transition-loading">
            <Spinner size="sm" />
            <span>Загрузка...</span>
          </div>
        )}
        {/* Текстовый блок для описания домашнего задания */}
        <div className="homework-text-section">
          <div className="homework-text-header">
            <label htmlFor={`homework-text-${lessonId}`}>
              Описание домашнего задания
            </label>
            {canEdit && (
              <span className="homework-text-status">
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

          {/* Для редактирования (admin/teacher) показываем textarea */}
          {canEdit ? (
            <textarea
              id={`homework-text-${lessonId}`}
              className="homework-text-input"
              value={homeworkText}
              onChange={handleTextChange}
              placeholder="Введите текст домашнего задания..."
              rows={4}
            />
          ) : (
            /* Для просмотра (student) показываем div с кликабельными ссылками */
            <div className="homework-text-display">
              {homeworkText ? formatHomeworkText(homeworkText) : (
                <span className="homework-text-empty">Описание не добавлено</span>
              )}
            </div>
          )}
        </div>

        {/* Drag and drop зона */}
        {canEdit && (
          <div
            className={`homework-upload-zone ${isDragging ? 'dragging' : ''}`}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={handleUploadClick}
          >
            <input
              ref={fileInputRef}
              type="file"
              onChange={(e) => handleFileSelect(e.target.files)}
              style={{ display: 'none' }}
              accept="*/*"
            />
            <div className="upload-icon">📎</div>
            <p className="upload-text">
              {uploadMutation.isPending
                ? 'Загрузка...'
                : 'Перетащите файл сюда или кликните для выбора'}
            </p>
            <p className="upload-hint">Максимум 10MB, до 10 файлов на урок</p>
          </div>
        )}

        {/* Список файлов */}
        {homeworkList.length === 0 ? (
          <div className="homework-empty">
            <p>Домашние задания не добавлены</p>
          </div>
        ) : (
          <div className="homework-list">
            {homeworkList.map((file) => (
              <div
                key={file.id}
                className="homework-item homework-item-clickable"
                onClick={() => handleDownload(file)}
              >
                <div className="homework-file-info">
                  <div className="homework-file-icon">📄</div>
                  <div className="homework-file-details">
                    <div className="homework-file-name" title={file.file_name}>
                      {file.file_name}
                    </div>
                    <div className="homework-file-meta">
                      {formatFileSize(file.file_size)} •
                      {new Date(file.created_at).toLocaleDateString('ru-RU')}
                      {file.created_by_name && ` • ${file.created_by_name}`}
                    </div>
                  </div>
                </div>
                {canEdit && (
                  <div className="homework-file-actions">
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteClick(file);
                      }}
                      loading={deleteMutation.isPending && fileToDelete?.id === file.id}
                      title="Удалить"
                    >
                      🗑️
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Подтверждение удаления */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => {
          setShowDeleteConfirm(false);
          setFileToDelete(null);
        }}
        onConfirm={confirmDelete}
        title="Удаление файла"
        message={`Вы уверены, что хотите удалить файл "${fileToDelete?.file_name}"?`}
        confirmText="Удалить"
        confirmVariant="danger"
        loading={deleteMutation.isPending}
      />
    </>
  );
};

export default HomeworkSection;
