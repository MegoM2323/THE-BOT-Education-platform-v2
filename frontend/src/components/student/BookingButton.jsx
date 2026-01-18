import { useState } from 'react';
import { logger } from '../../utils/logger.js';
import Button from "../common/Button.jsx";
import ConfirmModal from '../common/ConfirmModal.jsx';
import { useCredits } from '../../hooks/useCredits.js';
import './BookingButton.css';

export const BookingButton = ({ lesson, onBook, disabled = false, isBooked = false }) => {
  const [showConfirm, setShowConfirm] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { balance, loading: creditsLoading } = useCredits();

  const spotsLeft = lesson.max_students - (lesson.current_students || 0);
  const isFull = spotsLeft <= 0;
  const hasNoCredits = (balance ?? 0) <= 0;

  const isDisabled = disabled || isFull || hasNoCredits || isBooked || creditsLoading;

  const getTooltip = () => {
    if (isBooked) return 'Вы уже записаны на это занятие';
    if (isFull) return 'Нет свободных мест';
    if (hasNoCredits) return 'Недостаточно кредитов';
    return '';
  };

  const handleClick = () => {
    if (isDisabled) return;
    setError(null);
    setShowConfirm(true);
  };

  const handleConfirm = async () => {
    setLoading(true);
    setError(null);
    try {
      await onBook(lesson.id);
      setShowConfirm(false);
    } catch (error) {
      // Обработка специфичных ошибок от API
      const errorMessage = error?.response?.data?.detail || error?.message || 'Ошибка при записи на занятие';

      // Специальная обработка для распространенных ошибок
      if (error?.response?.status === 402) {
        setError('Недостаточно кредитов для записи на занятие');
      } else if (error?.response?.status === 409) {
        setError('Вы уже записаны на это занятие');
      } else if (error?.response?.status === 400) {
        setError(errorMessage || 'Невозможно записаться на это занятие');
      } else {
        setError(errorMessage);
      }
      logger.error('Booking error:', error);
      // Не закрываем модаль, чтобы пользователь видел ошибку
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <div className="booking-button-wrapper">
        <Button
          onClick={handleClick}
          disabled={isDisabled}
          fullWidth
          title={getTooltip()}
          data-testid="book-button"
        >
          {isBooked ? 'Записан' : isFull ? 'Нет мест' : 'Записаться'}
        </Button>
        {isDisabled && getTooltip() && (
          <span className="booking-button-tooltip">{getTooltip()}</span>
        )}
      </div>

      <ConfirmModal
        isOpen={showConfirm}
        onClose={() => {
          setShowConfirm(false);
          setError(null);
        }}
        onConfirm={handleConfirm}
        title="Подтверждение записи"
        message={
          error
            ? `Ошибка: ${error instanceof Error ? error.message : (error?.message || String(error))}`
            : `Вы действительно хотите записаться на занятие ${new Date(lesson.start_time).toLocaleDateString('ru-RU')}? С вашего счета будет списан 1 кредит.`
        }
        confirmText={error ? 'Попробовать снова' : 'Записаться'}
        loading={loading}
        isDanger={!!error}
      />
    </>
  );
};

export default BookingButton;
