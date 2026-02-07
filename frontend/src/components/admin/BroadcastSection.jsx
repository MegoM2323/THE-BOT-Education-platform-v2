import { useState, useRef } from 'react';
import { useLessonBroadcasts, useSendLessonBroadcast } from '../../hooks/useLessonBroadcasts.js';
import { useNotification } from '../../hooks/useNotification.js';
import { useAuth } from '../../hooks/useAuth.js';
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import Modal from "../common/Modal.jsx";
import { formatTime } from '../../utils/dateFormat.js';
import './BroadcastSection.css';

/**
 * Форматирует размер файла в читаемый вид
 */
const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
};

/**
 * Компонент секции рассылок урока
 * @param {boolean} [readOnly] - режим только для чтения (блокирует отправку)
 */
export const BroadcastSection = ({ lessonId, lesson, readOnly = false }) => {
  const { user } = useAuth();
  const { showNotification } = useNotification();
  const fileInputRef = useRef(null);

  // Состояния формы
  const [message, setMessage] = useState('');
  const [files, setFiles] = useState([]);
  const [showPreview, setShowPreview] = useState(false);
  const [selectedBroadcast, setSelectedBroadcast] = useState(null);

  // Загрузка данных рассылок
  const { data: broadcasts = [], isLoading, error } = useLessonBroadcasts(lessonId);

  // Mutation для отправки
  const sendMutation = useSendLessonBroadcast();

  // Проверка прав доступа (блокируется в readOnly режиме)
  const canSend = !readOnly && (user?.role === 'admin' || user?.role === 'teacher' || user?.role === 'teacher');

  /**
   * Обработка выбора файлов
   */
  const handleFileSelect = (e) => {
    const selectedFiles = Array.from(e.target.files);

    // Валидация количества файлов (макс 10)
    if (files.length + selectedFiles.length > 10) {
      showNotification('Максимум 10 файлов на рассылку', 'error');
      return;
    }

    // Валидация размера каждого файла (макс 10MB)
    for (const file of selectedFiles) {
      if (file.size > 10 * 1024 * 1024) {
        showNotification(`Файл ${file.name} превышает 10MB`, 'error');
        return;
      }
    }

    setFiles((prev) => [...prev, ...selectedFiles]);
    // Очистить input для возможности повторной загрузки
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  /**
   * Удалить файл из списка
   */
  const handleRemoveFile = (index) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  /**
   * Показать preview
   */
  const handlePreview = () => {
    if (!message.trim()) {
      showNotification('Введите текст сообщения', 'error');
      return;
    }
    setShowPreview(true);
  };

  /**
   * Отправить рассылку
   */
  const handleSend = async () => {
    if (!message.trim()) {
      showNotification('Введите текст сообщения', 'error');
      return;
    }

    if (message.trim().length > 4096) {
      showNotification('Сообщение не должно превышать 4096 символов', 'error');
      return;
    }

    try {
      await sendMutation.mutateAsync({
        lessonId,
        message: message.trim(),
        files,
      });

      showNotification('Рассылка успешно отправлена', 'success');
      // Очистить форму
      setMessage('');
      setFiles([]);
      setShowPreview(false);
    } catch (error) {
      console.error('Ошибка отправки рассылки:', error);
      let errorMsg = 'Не удалось отправить рассылку';
      if (error.data?.error?.message) {
        errorMsg = error.data.error.message;
      } else if (error.message && error.message !== `HTTP ${error.status}`) {
        errorMsg = error.message;
      }
      showNotification(errorMsg, 'error');
    }
  };

  /**
   * Показать детали рассылки
   */
  const handleBroadcastClick = (broadcast) => {
    setSelectedBroadcast(broadcast);
  };

  /**
   * Получить статус badge
   */
  const getStatusBadge = (status) => {
    const statusMap = {
      pending: { label: 'Ожидает', className: 'status-pending' },
      sending: { label: 'Отправка', className: 'status-sending' },
      completed: { label: 'Завершено', className: 'status-completed' },
      failed: { label: 'Ошибка', className: 'status-failed' },
    };

    const config = statusMap[status] || statusMap.pending;

    return <span className={`broadcast-status-badge ${config.className}`}>{config.label}</span>;
  };

  if (isLoading) {
    return (
      <div className="broadcast-section">
        <Spinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="broadcast-section">
        <div className="broadcast-error">Ошибка загрузки рассылок</div>
      </div>
    );
  }

  return (
    <>
      <div className="broadcast-section">
        {/* Compose Section */}
        {canSend && (
          <div className="broadcast-compose">
            <h4 className="broadcast-subsection-title">Новая рассылка</h4>

            {/* Textarea */}
            <div className="broadcast-compose-textarea">
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="Введите сообщение для студентов урока..."
                rows={5}
                maxLength={4096}
                disabled={sendMutation.isPending}
              />
              <div className="broadcast-char-counter">
                {message.length} / 4096
              </div>
            </div>

            {/* File Upload */}
            <div className="broadcast-files-section">
              <input
                ref={fileInputRef}
                type="file"
                onChange={handleFileSelect}
                style={{ display: 'none' }}
                accept="*/*"
                multiple
              />
              <Button
                variant="secondary"
                size="sm"
                onClick={() => fileInputRef.current?.click()}
                disabled={files.length >= 10 || sendMutation.isPending}
              >
                📎 Прикрепить файлы ({files.length}/10)
              </Button>

              {files.length > 0 && (
                <div className="broadcast-files-list">
                  {files.map((file, index) => (
                    <div key={index} className="broadcast-file-item">
                      <span className="broadcast-file-name" title={file.name}>
                        {file.name} ({formatFileSize(file.size)})
                      </span>
                      <button
                        className="broadcast-file-remove"
                        onClick={() => handleRemoveFile(index)}
                        disabled={sendMutation.isPending}
                        aria-label="Удалить файл"
                      >
                        ×
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="broadcast-compose-actions">
              <Button
                variant="secondary"
                size="sm"
                onClick={handlePreview}
                disabled={!message.trim() || sendMutation.isPending}
              >
                👁️ Предпросмотр
              </Button>
              <Button
                variant="primary"
                onClick={handleSend}
                loading={sendMutation.isPending}
                disabled={!message.trim()}
              >
                📤 Отправить
              </Button>
            </div>
          </div>
        )}

        {/* History Section */}
        <div className="broadcast-history">
          <h4 className="broadcast-subsection-title">История рассылок</h4>

          {broadcasts.length === 0 ? (
            <div className="broadcast-empty">
              <p>Рассылок пока нет</p>
            </div>
          ) : (
            <div className="broadcast-list">
              {broadcasts.map((broadcast) => (
                <div
                  key={broadcast.id}
                  className="broadcast-item"
                  onClick={() => handleBroadcastClick(broadcast)}
                >
                  <div className="broadcast-item-header">
                    <div className="broadcast-sender">
                      {broadcast.sender_name || 'Неизвестный отправитель'}
                    </div>
                    {getStatusBadge(broadcast.status)}
                  </div>
                  <div className="broadcast-item-preview">
                    {broadcast.message.length > 100
                      ? `${broadcast.message.substring(0, 100)}...`
                      : broadcast.message}
                  </div>
                  <div className="broadcast-item-footer">
                    <span className="broadcast-date">
                      {new Date(broadcast.created_at).toLocaleString('ru-RU')}
                    </span>
                    {broadcast.files && broadcast.files.length > 0 && (
                      <span className="broadcast-files-count">
                        📎 {broadcast.files.length} файл(а)
                      </span>
                    )}
                    {broadcast.status === 'completed' && (
                      <span className="broadcast-sent-count">
                        ✓ {broadcast.sent_count} получателей
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Preview Modal */}
      {showPreview && (
        <Modal
          isOpen={showPreview}
          onClose={() => setShowPreview(false)}
          title="Предпросмотр рассылки"
          footer={
            <div className="broadcast-preview-footer">
              <Button variant="secondary" onClick={() => setShowPreview(false)}>
                Закрыть
              </Button>
              <Button variant="primary" onClick={handleSend} loading={sendMutation.isPending}>
                📤 Отправить
              </Button>
            </div>
          }
        >
          <div className="broadcast-preview-content">
            <div className="broadcast-preview-lesson">
              <h4>Урок</h4>
              <p>
                {new Date(lesson.start_time).toLocaleDateString('ru-RU', {
                  weekday: 'long',
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}{' '}
                {formatTime(lesson.start_time)} - {formatTime(lesson.end_time)}
              </p>
              {lesson.subject && <p>Предмет: {lesson.subject}</p>}
            </div>
            <div className="broadcast-preview-message">
              <h4>Сообщение</h4>
              <p className="broadcast-preview-text">{message}</p>
            </div>
            {files.length > 0 && (
              <div className="broadcast-preview-files">
                <h4>Файлы ({files.length})</h4>
                <ul>
                  {files.map((file, index) => (
                    <li key={index}>
                      {file.name} ({formatFileSize(file.size)})
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </Modal>
      )}

      {/* Broadcast Details Modal */}
      {selectedBroadcast && (
        <Modal
          isOpen={!!selectedBroadcast}
          onClose={() => setSelectedBroadcast(null)}
          title="Детали рассылки"
        >
          <div className="broadcast-details-content">
            <div className="broadcast-details-header">
              <div>
                <strong>Отправитель:</strong> {selectedBroadcast.sender_name}
              </div>
              <div>
                <strong>Дата:</strong> {new Date(selectedBroadcast.created_at).toLocaleString('ru-RU')}
              </div>
              <div>
                <strong>Статус:</strong> {getStatusBadge(selectedBroadcast.status)}
              </div>
              {selectedBroadcast.status === 'completed' && (
                <div>
                  <strong>Доставлено:</strong> {selectedBroadcast.sent_count} из{' '}
                  {selectedBroadcast.sent_count + (selectedBroadcast.failed_count || 0)}
                </div>
              )}
            </div>
            <div className="broadcast-details-message">
              <h4>Сообщение</h4>
              <p className="broadcast-details-text">{selectedBroadcast.message}</p>
            </div>
            {selectedBroadcast.files && selectedBroadcast.files.length > 0 && (
              <div className="broadcast-details-files">
                <h4>Файлы ({selectedBroadcast.files.length})</h4>
                <ul>
                  {selectedBroadcast.files.map((file) => (
                    <li key={file.id}>
                      <a
                        href={`${import.meta.env.VITE_API_URL || '/api'}/v1/lessons/${lessonId}/broadcasts/${selectedBroadcast.id}/files/${file.id}/download`}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {file.file_name} ({formatFileSize(file.file_size)})
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        </Modal>
      )}
    </>
  );
};

export default BroadcastSection;
