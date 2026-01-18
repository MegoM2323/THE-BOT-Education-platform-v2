import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../hooks/useAuth.js';
import { useMyBookings } from '../hooks/useMyBookings.js';
import { useSlowConnection } from '../hooks/useSlowConnection.js';
import * as homeworkAPI from '../api/homework.js';
import LessonBookingModal from '../components/student/LessonBookingModal.jsx';
import { SkeletonList } from '../components/common/SkeletonLoader.jsx';
import SlowConnectionNotice from '../components/common/SlowConnectionNotice.jsx';
import { formatHomeworkText } from '../utils/formatHomeworkText.js';
import './StudentHomework.css';

const StudentHomework = () => {
  const { user } = useAuth();
  const [homeworkList, setHomeworkList] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedLesson, setSelectedLesson] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [dismissedSlowNotice, setDismissedSlowNotice] = useState(false);

  const { myBookings = [], isLoading: bookingsLoading } = useMyBookings();

  // Detect slow connection (show warning after 3 seconds)
  const isSlow = useSlowConnection(loading || bookingsLoading, 3000);

  const fetchHomework = useCallback(async (signal) => {
    try {
      setLoading(true);

      // Получаем список моих записей - защита от null/undefined
      if (!Array.isArray(myBookings) || myBookings.length === 0) {
        setHomeworkList([]);
        setError(null);
        return;
      }

      // Фильтруем занятия: последние 2 недели + все будущие
      const now = new Date();
      const twoWeeksAgo = new Date(now.getTime() - 14 * 24 * 60 * 60 * 1000);

      const homeworkPromises = myBookings
        .filter(booking => {
          // Защита от null/undefined booking или start_time
          if (!booking?.start_time) return false;
          const lessonDate = new Date(booking.start_time);
          return lessonDate >= twoWeeksAgo; // последние 2 недели + будущие
        })
        .map(async (booking) => {
          try {
            // Используем homeworkAPI с поддержкой AbortController
            const files = await homeworkAPI.getHomework(booking.lesson_id, signal);

            return {
              id: booking.lesson_id,
              lesson_date: booking.start_time,
              subject: booking?.subject || 'Занятие',
              teacher_name: booking?.teacher_name ?? 'Не указано',
              homework_text: booking?.homework_text || '',
              files: Array.isArray(files) ? files : [],
              lessonData: {
                ...booking,
                id: booking.lesson_id,
              },
            };
          } catch (err) {
            // Игнорируем AbortError - это нормальное поведение при unmount
            if (err?.name === 'AbortError') {
              return null;
            }
            console.error(`Error fetching homework for lesson ${booking?.lesson_id}:`, err);
            return null;
          }
        });

      const results = await Promise.all(homeworkPromises);
      const validHomework = results.filter(hw => hw !== null && hw?.lesson_date);

      // Сортируем по дате (сначала ближайшие) - с null-safety
      validHomework.sort((a, b) => {
        const dateA = a?.lesson_date ? new Date(a.lesson_date) : new Date(0);
        const dateB = b?.lesson_date ? new Date(b.lesson_date) : new Date(0);
        return dateA - dateB;
      });

      setHomeworkList(validHomework);
      setError(null);
    } catch (err) {
      // Игнорируем AbortError
      if (err?.name === 'AbortError') {
        return;
      }
      console.error('Error fetching homework:', err);
      setError(err?.message || 'Не удалось загрузить домашние задания');
    } finally {
      setLoading(false);
    }
  }, [myBookings]);

  useEffect(() => {
    // Создаём AbortController для отмены запросов при unmount
    const abortController = new AbortController();

    fetchHomework(abortController.signal);

    // Cleanup: отменяем запросы при unmount
    return () => {
      abortController.abort();
    };
  }, [fetchHomework]);

  /**
   * Форматирует дату занятия в короткий формат DD.MM.YYYY
   * @param {string} dateString - ISO строка даты
   * @returns {string} - Отформатированная дата
   */
  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = date.getFullYear();
    return `${day}.${month}.${year}`;
  };

  const handleTileClick = (hw) => {
    setSelectedLesson(hw.lessonData);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedLesson(null);
  };

  // Show skeleton loader while loading
  if (loading || bookingsLoading) {
    return (
      <div className="student-homework">
        <div className="homework-header">
          <h2>Домашние задания</h2>
          <p className="homework-subtitle">
            Отображаются задания за последние 2 недели и предстоящие
          </p>
        </div>

        <div className="homework-grid" aria-busy="true" aria-live="polite">
          <SkeletonList items={4} />
        </div>

        {/* Show slow connection notice if loading takes too long */}
        {isSlow && !dismissedSlowNotice && (
          <SlowConnectionNotice onDismiss={() => setDismissedSlowNotice(true)} />
        )}
      </div>
    );
  }

  if (error) {
    const errorMessage = error instanceof Error ? error.message : (error?.message || String(error));
    return (
      <div className="student-homework">
        <div className="homework-error">
          <p>{errorMessage}</p>
          <button onClick={fetchHomework} className="retry-button">
            Попробовать снова
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="student-homework">
      <div className="homework-header">
        <h2>Домашние задания</h2>
        <p className="homework-subtitle">
          Отображаются задания за последние 2 недели и предстоящие
        </p>
      </div>

      {homeworkList.length === 0 ? (
        <div className="homework-empty">
          <svg className="empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p>У вас пока нет домашних заданий</p>
        </div>
      ) : (
        <div className="homework-grid">
          {homeworkList.map(hw => (
            <div
              key={hw.id}
              className="homework-tile"
              onClick={() => handleTileClick(hw)}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  handleTileClick(hw);
                }
              }}
            >
              <div className="homework-tile-header">
                <div className="homework-subject">{hw.subject}</div>
                <div className="homework-date">Занятие от {formatDate(hw.lesson_date)}</div>
              </div>

              {hw?.teacher_name && (
                <div className="homework-teacher">
                  <span className="teacher-label">Преподаватель:</span>
                  <span className="teacher-name">{hw.teacher_name}</span>
                </div>
              )}

              <div className="homework-content">
                <p className="homework-text">{formatHomeworkText(hw?.homework_text)}</p>
              </div>

              {Array.isArray(hw?.files) && hw.files.length > 0 && (
                <div className="homework-files">
                  <div className="files-label">Прикрепленные файлы:</div>
                  {hw.files.map(file => (
                    <a
                      key={file?.id || `file-${file?.filename}`}
                      href={`${import.meta.env.VITE_API_URL || '/api'}/lessons/${hw?.id}/homework/${file?.id}/download`}
                      className="file-link"
                      target="_blank"
                      rel="noopener noreferrer"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <svg className="file-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                          d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                      </svg>
                      {file?.file_name || 'Файл'}
                    </a>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Модальное окно деталей занятия */}
      {selectedLesson && (
        <LessonBookingModal
          isOpen={isModalOpen}
          onClose={handleCloseModal}
          lesson={selectedLesson}
          myBookings={Array.isArray(myBookings) ? myBookings : []}
          credits={{ balance: user?.credits ?? 0 }}
        />
      )}
    </div>
  );
};

export default StudentHomework;
