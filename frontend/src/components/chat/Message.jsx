import { useState } from 'react';
import { logger } from '../../utils/logger.js';
import { getFileDownloadURL } from '../../api/chat.js';
import { useNotification } from '../../hooks/useNotification';
import './Message.css';

/**
 * Компонент для отображения одного сообщения
 */
const Message = ({ message, currentUserId, roomId }) => {
  const [downloadingFile, setDownloadingFile] = useState(null);
  const notification = useNotification();

  const isOwnMessage = message.sender_id === currentUserId;
  const messageTime = new Date(message.created_at);

  /**
   * Получить иконку для типа файла по MIME type
   */
  const getFileIcon = (mimeType) => {
    if (!mimeType) return '📄';

    if (mimeType.startsWith('image/')) return '🖼️';
    if (mimeType.startsWith('video/')) return '🎥';
    if (mimeType.startsWith('audio/')) return '🎵';
    if (mimeType.includes('pdf')) return '📕';
    if (mimeType.includes('word') || mimeType.includes('document')) return '📝';
    if (mimeType.includes('sheet') || mimeType.includes('excel')) return '📊';
    if (mimeType.includes('zip') || mimeType.includes('rar') || mimeType.includes('archive')) return '📦';

    return '📄';
  };

  /**
   * Форматировать размер файла
   */
  const formatFileSize = (bytes) => {
    if (!bytes || bytes === 0) return '0 B';

    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  /**
   * Скачать файл
   */
  const handleDownloadFile = async (attachment) => {
    try {
      setDownloadingFile(attachment.id);

      const url = getFileDownloadURL(roomId, attachment.id);
      const link = document.createElement('a');
      link.href = url;
      link.download = attachment.file_name || 'download';
      link.target = '_blank';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error('Ошибка скачивания файла:', error);
      notification.error('Не удалось скачать файл');
    } finally {
      setDownloadingFile(null);
    }
  };

  /**
   * Получить текст статуса модерации
   */
  const getModerationStatusText = (status) => {
    switch (status) {
      case 'pending_moderation':
        return '⏳ На модерации';
      case 'delivered':
        return '✓ Доставлено';
      case 'blocked':
        return '🚫 Заблокировано';
      default:
        return '';
    }
  };

  /**
   * Получить класс для статуса модерации
   */
  const getModerationStatusClass = (status) => {
    switch (status) {
      case 'pending_moderation':
        return 'status-pending';
      case 'delivered':
        return 'status-delivered';
      case 'blocked':
        return 'status-blocked';
      default:
        return '';
    }
  };

  return (
    <div className={`message ${isOwnMessage ? 'message-own' : 'message-other'}`}>
      <div className="message-bubble">
        {/* Имя отправителя (только для чужих сообщений) */}
        {!isOwnMessage && (
          <div className="message-sender-name">{message.sender_name || 'Пользователь'}</div>
        )}

        {/* Текст сообщения */}
        {message.message_text && (
          <div className="message-text">{message.message_text}</div>
        )}

        {/* Вложенные файлы */}
        {message.attachments && message.attachments.length > 0 && (
          <div className="message-attachments">
            {message.attachments.map((attachment) => (
              <div key={attachment.id} className="attachment-item">
                <span className="attachment-icon">
                  {getFileIcon(attachment.mime_type)}
                </span>
                <div className="attachment-info">
                  <div className="attachment-filename" title={attachment.file_name}>
                    {attachment.file_name}
                  </div>
                  <div className="attachment-size">
                    {formatFileSize(attachment.file_size)}
                  </div>
                </div>
                <button
                  className="attachment-download-btn"
                  onClick={() => handleDownloadFile(attachment)}
                  disabled={downloadingFile === attachment.id}
                  title="Скачать файл"
                >
                  {downloadingFile === attachment.id ? '⏳' : '⬇️'}
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Время и статус */}
        <div className="message-footer">
          <span className="message-time">
            {messageTime.toLocaleTimeString('ru-RU', {
              hour: '2-digit',
              minute: '2-digit',
            })}
          </span>

          {/* Статус модерации (только для своих сообщений, не показываем для delivered) */}
          {isOwnMessage && message.status && message.status !== 'delivered' && (
            <span
              className={`message-status ${getModerationStatusClass(message.status)}`}
              title={getModerationStatusText(message.status)}
            >
              {getModerationStatusText(message.status)}
            </span>
          )}
        </div>
      </div>
    </div>
  );
};

export default Message;
