import { useState, useEffect, useRef, useCallback } from 'react';
import { logger } from '../../utils/logger.js';
import { useAuth } from '../../hooks/useAuth.js';
import { useNotification } from '../../hooks/useNotification.js';
import { useSSE } from '../../hooks/useSSE.js';
import { sendMessage, getMessages } from '../../api/chat.js';
import Message from './Message.jsx';
import Spinner from '../common/Spinner.jsx';
import './ChatWindow.css';

/**
 * Компонент окна чата с сообщениями
 */
const ChatWindow = ({ room }) => {
  const { user } = useAuth();
  const { showNotification } = useNotification();
  const { lastMessage, lastDeletedMessage, lastStatusUpdate, isConnected } = useSSE();

  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [messageText, setMessageText] = useState('');
  const [selectedFiles, setSelectedFiles] = useState([]);

  const messagesEndRef = useRef(null);
  const fileInputRef = useRef(null);
  const textareaRef = useRef(null);
  const processedMessageIdsRef = useRef(new Set());

  /**
   * Загрузить сообщения из комнаты
   */
  const loadMessages = useCallback(async () => {
    if (!room?.id) return;

    try {
      const data = await getMessages(room.id, 50, 0);
      const sorted = (data || []).sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
      setMessages(sorted);
    } catch (error) {
      console.error('Ошибка загрузки сообщений:', error);
      showNotification('Ошибка загрузки сообщений', 'error');
    } finally {
      setLoading(false);
    }
  }, [room?.id, showNotification]);

  /**
   * Инициализация: загрузить сообщения при смене комнаты
   */
  useEffect(() => {
    if (!room?.id) return;

    setLoading(true);
    processedMessageIdsRef.current.clear();
    loadMessages();
  }, [room?.id, loadMessages]);

  /**
   * SSE: Обработка новых сообщений
   */
  useEffect(() => {
    if (!lastMessage || !room?.id) return;

    if (lastMessage.chat_id !== room.id) return;

    if (processedMessageIdsRef.current.has(lastMessage.id)) return;

    if (lastMessage.sender_id === user?.id) return;

    processedMessageIdsRef.current.add(lastMessage.id);

    setMessages((prev) => {
      const exists = prev.some((msg) => msg.id === lastMessage.id);
      if (exists) return prev;

      return [...prev, lastMessage];
    });

    logger.debug('[ChatWindow] SSE new message added:', lastMessage.id);
  }, [lastMessage, room?.id, user?.id]);

  /**
   * SSE: Обработка удалённых сообщений
   */
  useEffect(() => {
    if (!lastDeletedMessage || !room?.id) return;

    if (lastDeletedMessage.chat_id !== room.id) return;

    setMessages((prev) =>
      prev.filter((msg) => msg.id !== lastDeletedMessage.message_id)
    );

    logger.debug('[ChatWindow] SSE message deleted:', lastDeletedMessage.message_id);
  }, [lastDeletedMessage, room?.id]);

  /**
   * SSE: Обработка обновления статуса сообщения
   */
  useEffect(() => {
    if (!lastStatusUpdate || !room?.id) return;

    if (lastStatusUpdate.chat_id !== room.id) return;

    setMessages((prev) =>
      prev.map((msg) =>
        msg.id === lastStatusUpdate.message_id
          ? { ...msg, status: lastStatusUpdate.status }
          : msg
      )
    );

    logger.debug('[ChatWindow] SSE message status updated:', lastStatusUpdate.message_id, lastStatusUpdate.status);
  }, [lastStatusUpdate, room?.id]);

  /**
   * Автоскролл к последнему сообщению
   */
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  /**
   * Обработчик отправки сообщения
   */
  const handleSendMessage = async (e) => {
    e.preventDefault();

    const trimmedText = messageText.trim();

    // Проверка: должен быть текст или файлы
    if (!trimmedText && selectedFiles.length === 0) {
      showNotification('Введите сообщение или прикрепите файл', 'error');
      return;
    }

    try {
      setSending(true);

      // DEBUG: Логируем что отправляем
      console.log('[ChatWindow] Отправка сообщения:', {
        roomId: room.id,
        messageText: trimmedText,
        messageLength: trimmedText.length,
        filesCount: selectedFiles.length
      });

      // Отправить сообщение
      const newMessage = await sendMessage(room.id, trimmedText, selectedFiles);

      // DEBUG: Логируем что получили
      console.log('[ChatWindow] Сообщение отправлено:', newMessage);

      processedMessageIdsRef.current.add(newMessage.id);

      setMessages((prev) => {
        const updated = [...prev, newMessage];
        return updated.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
      });

      // Очистить форму
      setMessageText('');
      setSelectedFiles([]);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }

      // Вернуть фокус в textarea
      if (textareaRef.current) {
        textareaRef.current.focus();
      }
    } catch (error) {
      console.error('[ChatWindow] Ошибка отправки сообщения:', error);
      const errorMsg = error.data?.error?.message || error.message || 'Не удалось отправить сообщение';
      showNotification(errorMsg, 'error');
    } finally {
      setSending(false);
    }
  };

  /**
   * Обработчик выбора файлов
   */
  const handleFileSelect = (e) => {
    const files = Array.from(e.target.files || []);
    setSelectedFiles(files);
  };

  /**
   * Удалить выбранный файл
   */
  const handleRemoveFile = (index) => {
    setSelectedFiles((prev) => prev.filter((_, i) => i !== index));
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  /**
   * Обработчик нажатия Enter в textarea (отправка сообщения)
   */
  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage(e);
    }
  };

  if (!room) {
    return (
      <div className="chat-window-empty">
        <div className="chat-empty-message">
          <div className="chat-empty-icon">💬</div>
          <h3>Выберите чат</h3>
          <p>Выберите пользователя из списка слева, чтобы начать общение</p>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-window">
      {/* Header с информацией о собеседнике */}
      <div className="chat-window-header">
        <div className="chat-participant-info">
          <div className="chat-participant-avatar">
            {room.participant_name?.[0]?.toUpperCase() || '?'}
          </div>
          <div className="chat-participant-details">
            <div className="chat-participant-name">{room.participant_name || 'Пользователь'}</div>
            <div className="chat-participant-role">
              {room.participant_role === 'methodologist' || room.participant_role === 'teacher' ? 'Преподаватель' : 'Студент'}
            </div>
          </div>
        </div>
      </div>

      {/* Список сообщений */}
      <div className="chat-messages-container">
        {loading ? (
          <div className="chat-loading">
            <Spinner />
          </div>
        ) : messages.length === 0 ? (
          <div className="chat-no-messages">
            <div className="chat-no-messages-icon">✉️</div>
            <p>Сообщений пока нет</p>
            <p className="chat-no-messages-hint">Отправьте первое сообщение!</p>
          </div>
        ) : (
          <div className="chat-messages-list">
            {messages.map((message) => (
              <Message
                key={message.id}
                message={message}
                currentUserId={user?.id}
                roomId={room.id}
              />
            ))}
            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Форма отправки сообщения */}
      <div className="chat-input-container">
        {/* Превью выбранных файлов */}
        {selectedFiles.length > 0 && (
          <div className="chat-selected-files">
            {selectedFiles.map((file, index) => (
              <div key={index} className="chat-selected-file-item">
                <span className="chat-file-name">{file.name}</span>
                <button
                  type="button"
                  className="chat-file-remove"
                  onClick={() => handleRemoveFile(index)}
                  title="Удалить файл"
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        )}

        <form className="chat-input-form" onSubmit={handleSendMessage}>
          <button
            type="button"
            className="chat-attach-btn"
            onClick={() => fileInputRef.current?.click()}
            disabled={sending}
            title="Прикрепить файл"
          >
            📎
          </button>

          <input
            ref={fileInputRef}
            type="file"
            multiple
            onChange={handleFileSelect}
            style={{ display: 'none' }}
            accept="*/*"
          />

          <textarea
            ref={textareaRef}
            className="chat-input-textarea"
            placeholder="Введите сообщение..."
            value={messageText}
            onChange={(e) => setMessageText(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={sending}
            rows={1}
          />

          <button
            type="submit"
            className="chat-send-btn"
            disabled={sending || (!messageText.trim() && selectedFiles.length === 0)}
            title="Отправить"
          >
            {sending ? '⏳' : '➤'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default ChatWindow;
