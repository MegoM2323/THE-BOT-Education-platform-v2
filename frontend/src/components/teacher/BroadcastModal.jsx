import { useState } from 'react';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import { formatTime } from '../../utils/dateFormat.js';
import './BroadcastModal.css';

/**
 * Модальное окно для отправки рассылки студентам занятия (для преподавателя)
 * @param {boolean} isOpen - Открыто ли модальное окно
 * @param {Function} onClose - Callback при закрытии
 * @param {Object} lesson - Объект занятия
 * @param {Array} students - Массив студентов на занятии
 * @param {Function} onSend - Callback при отправке сообщения (message) => Promise
 */
export const BroadcastModal = ({ isOpen, onClose, lesson, students, onSend }) => {
  const [message, setMessage] = useState('');
  const [sending, setSending] = useState(false);
  const [error, setError] = useState(null);

  const handleSend = async () => {
    if (!message.trim()) {
      setError('Введите текст сообщения');
      return;
    }

    if (message.length > 4096) {
      setError('Сообщение не может быть длиннее 4096 символов');
      return;
    }

    try {
      setSending(true);
      setError(null);
      await onSend(message);
      setMessage('');
      onClose();
    } catch (err) {
      // Обработка ошибок
      if (err.response?.status === 403) {
        setError('Это не ваше занятие');
      } else if (err.response?.status === 400) {
        setError('Сообщение слишком длинное или пустое');
      } else if (err.response?.status === 404) {
        setError('Занятие не найдено');
      } else {
        setError(err.message || 'Ошибка отправки рассылки');
      }
    } finally {
      setSending(false);
    }
  };

  const handleClose = () => {
    if (!sending) {
      setMessage('');
      setError(null);
      onClose();
    }
  };

  if (!lesson) return null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="📢 Отправить рассылку"
      footer={
        <div className="broadcast-modal-footer">
          <Button
            variant="secondary"
            onClick={handleClose}
            disabled={sending}
          >
            Отмена
          </Button>
          <Button
            variant="primary"
            onClick={handleSend}
            loading={sending}
            disabled={!message.trim() || sending}
          >
            Отправить
          </Button>
        </div>
      }
    >
      <div className="broadcast-modal-content">
        {/* Информация о занятии */}
        <div className="broadcast-lesson-info">
          <h4>Занятие</h4>
          <div className="broadcast-info-item">
            <strong>Время:</strong>{' '}
            {new Date(lesson.start_time).toLocaleDateString('ru-RU', {
              weekday: 'long',
              year: 'numeric',
              month: 'long',
              day: 'numeric'
            })}{' '}
            {formatTime(lesson.start_time)} - {formatTime(lesson.end_time)}
          </div>
          {lesson.description && (
            <div className="broadcast-info-item">
              <strong>Описание:</strong> {lesson.description}
            </div>
          )}
        </div>

        {/* Список получателей */}
        <div className="broadcast-recipients">
          <h4>Получатели ({students.length})</h4>
          {students.length === 0 ? (
            <p className="no-students">На занятие пока никто не записан</p>
          ) : (
            <ul className="broadcast-students-list">
              {students.map((student) => (
                <li key={student.id}>
                  {student.full_name || student.student_name || student.email}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Текст сообщения */}
        <div className="broadcast-message-input">
          <label htmlFor="broadcast-message">Текст сообщения</label>
          <textarea
            id="broadcast-message"
            value={message}
            onChange={(e) => {
              setMessage(e.target.value);
              setError(null);
            }}
            placeholder="Введите сообщение для студентов..."
            rows={5}
            maxLength={4096}
            disabled={sending}
          />
          <div className="broadcast-char-counter">
            {message.length} / 4096
          </div>
        </div>

        {/* Ошибка */}
        {error && (
          <div className="broadcast-error">
            {error instanceof Error ? error.message : (error?.message || String(error))}
          </div>
        )}
      </div>
    </Modal>
  );
};

export default BroadcastModal;
