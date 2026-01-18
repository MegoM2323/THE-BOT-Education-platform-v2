import { useState, useEffect, useMemo } from 'react';
import { logger } from '../../utils/logger.js';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import LessonCard from './LessonCard.jsx';
import Spinner from '../common/Spinner.jsx';
import { useLessons } from '../../hooks/useLessons.js';
import { useSwap } from '../../hooks/useSwap.js';
import ConfirmModal from '../common/ConfirmModal.jsx';
import './SwapModal.css';

export const SwapModal = ({ isOpen, onClose, booking, onSuccess }) => {
  const { lessons, loading: lessonsLoading, fetchLessons } = useLessons();
  const { performSwap, swapping } = useSwap();
  const [selectedLesson, setSelectedLesson] = useState(null);
  const [showConfirm, setShowConfirm] = useState(false);
  const [swapError, setSwapError] = useState(null);

  useEffect(() => {
    if (isOpen) {
      fetchLessons();
      setSwapError(null);
    }
  }, [isOpen, fetchLessons]);

  const availableLessons = useMemo(() => {
    if (!booking) return [];

    // Получить ID текущего занятия - может быть в booking.lesson_id или booking.lesson.id
    const currentLessonId = booking.lesson_id || booking.lesson?.id;

    return lessons.filter((lesson) => {
      // Исключить текущее занятие
      if (lesson.id === currentLessonId) return false;

      // Только будущие занятия
      const isFuture = new Date(lesson.start_time) > new Date();
      if (!isFuture) return false;

      // Должны быть доступные места
      const spotsLeft = lesson.max_students - (lesson.current_students || 0);
      if (spotsLeft <= 0) return false;

      // Должно оставаться не менее 24 часов
      const hoursUntilLesson = (new Date(lesson.start_time) - new Date()) / (1000 * 60 * 60);
      if (hoursUntilLesson < 24) return false;

      return true;
    });
  }, [lessons, booking]);

  const handleSwap = async () => {
    if (!selectedLesson || !booking) return;

    setSwapError(null);
    try {
      // API ожидает old_lesson_id и new_lesson_id, а не booking ID и lesson ID
      const currentLessonId = booking.lesson_id || booking.lesson?.id;
      await performSwap(currentLessonId, selectedLesson.id);
      setShowConfirm(false);
      setSelectedLesson(null);
      onSuccess();
      onClose();
    } catch (error) {
      const errorMessage = error?.message || 'Ошибка при обмене занятия';
      setSwapError(errorMessage);
      console.error('Swap error:', error);
      // Не закрываем модаль, чтобы пользователь видел ошибку
    }
  };

  const handleSelectLesson = (lesson) => {
    setSelectedLesson(lesson);
    setSwapError(null);
    setShowConfirm(true);
  };

  // Получить текущее занятие для отображения
  const currentLesson = booking?.lesson || {
    id: booking?.lesson_id,
    start_time: booking?.start_time,
    end_time: booking?.end_time,
    teacher_name: booking?.teacher_name,
    subject: booking?.subject,
    max_students: booking?.max_students,
  };

  return (
    <>
      <Modal
        isOpen={isOpen}
        onClose={onClose}
        title="Обменять занятие"
        data-testid="swap-modal"
      >
        <div className="swap-modal-content">
          <div className="swap-current">
            <h4 className="swap-section-title">Текущее занятие:</h4>
            {booking && <LessonCard lesson={currentLesson} showActions={false} />}
          </div>

          <div className="swap-divider">→</div>

          <div className="swap-available">
            <h4 className="swap-section-title">Доступные для обмена:</h4>
            {lessonsLoading ? (
              <div className="swap-loading">
                <Spinner />
              </div>
            ) : availableLessons.length === 0 ? (
              <div className="swap-empty">
                <p>Нет доступных занятий для обмена</p>
              </div>
            ) : (
              <div className="swap-lessons-list" data-testid="target-lesson-select">
                {availableLessons.map((lesson) => (
                  <div key={lesson.id} className="swap-lesson-item">
                    <LessonCard lesson={lesson} showActions={false} />
                    <Button
                      onClick={() => handleSelectLesson(lesson)}
                      fullWidth
                      size="small"
                    >
                      Выбрать
                    </Button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </Modal>

      <ConfirmModal
        isOpen={showConfirm}
        onClose={() => {
          setShowConfirm(false);
          setSwapError(null);
        }}
        onConfirm={handleSwap}
        title="Подтверждение обмена"
        message={
          swapError
            ? `Ошибка: ${swapError}`
            : `Вы действительно хотите обменять занятие на ${selectedLesson ? new Date(selectedLesson.start_time).toLocaleDateString('ru-RU') : ''}?`
        }
        confirmText={swapError ? 'Попробовать снова' : 'Обменять'}
        loading={swapping}
        isDanger={!!swapError}
        data-testid="confirm-swap"
      />
    </>
  );
};

export default SwapModal;
