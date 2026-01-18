import { useState, useEffect } from 'react';
import { logger } from '../../utils/logger.js';
import { useAuth } from '../../hooks/useAuth.js';
import { useNotification } from '../../hooks/useNotification.js';
import { getMyRooms, getOrCreateRoom } from '../../api/chat.js';
import { getStudentsAll, getTeachersAll } from '../../api/users.js';
import Spinner from '../common/Spinner.jsx';
import './ChatList.css';

/**
 * Компонент списка чатов
 */
const ChatList = ({ selectedRoom, onRoomSelect, urlRoomId }) => {
  const { user } = useAuth();
  const { showNotification } = useNotification();

  const [rooms, setRooms] = useState([]);
  const [availableUsers, setAvailableUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [creatingRoom, setCreatingRoom] = useState(null);

  /**
   * Загрузить комнаты текущего пользователя
   */
  const loadRooms = async () => {
    try {
      const data = await getMyRooms();
      setRooms(data || []);
    } catch (error) {
      console.error('Ошибка загрузки комнат:', error);
      showNotification('Ошибка загрузки чатов', 'error');
    } finally {
      setLoading(false);
    }
  };

  /**
   * Загрузить список доступных пользователей
   * Студенты видят преподавателей, преподаватели видят студентов
   */
  const loadAvailableUsers = async () => {
    try {
      let users = [];

      if (user?.role === 'student') {
        // Студенты видят преподавателей
        users = await getTeachersAll();
      } else if (user?.role === 'teacher') {
        // Преподаватели видят студентов
        users = await getStudentsAll();
      }

      setAvailableUsers(users || []);
    } catch (error) {
      console.error('Ошибка загрузки пользователей:', error);
    }
  };

  /**
   * Инициализация: загрузить комнаты и доступных пользователей
   */
  useEffect(() => {
    if (user?.id) {
      loadRooms();
      loadAvailableUsers();
    }
  }, [user?.id]);

  /**
   * Автоматически выбрать комнату из URL если urlRoomId указан
   */
  useEffect(() => {
    if (urlRoomId && rooms.length > 0) {
      const room = rooms.find(r => r.id === urlRoomId);
      if (room && room.id !== selectedRoom?.id) {
        console.log('[ChatList] Автоматический выбор комнаты из URL:', urlRoomId);
        onRoomSelect(room);
      }
    }
  }, [urlRoomId, rooms, selectedRoom?.id, onRoomSelect]);

  /**
   * Создать или открыть комнату с пользователем
   */
  const handleSelectUser = async (userId) => {
    try {
      setCreatingRoom(userId);

      // Создать или получить существующую комнату
      const room = await getOrCreateRoom(userId);

      // Добавить комнату в список если её ещё нет
      setRooms((prev) => {
        const exists = prev.some((r) => r.id === room.id);
        if (exists) {
          return prev;
        }
        return [room, ...prev];
      });

      // Выбрать комнату
      onRoomSelect(room);
    } catch (error) {
      console.error('Ошибка создания комнаты:', error);
      const errorMsg = error.response?.data?.message || 'Не удалось открыть чат';
      showNotification(errorMsg, 'error');
    } finally {
      setCreatingRoom(null);
    }
  };

  /**
   * Форматировать время последнего сообщения
   */
  const formatLastMessageTime = (timestamp) => {
    if (!timestamp) return '';

    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'только что';
    if (diffMins < 60) return `${diffMins} мин назад`;
    if (diffHours < 24) return `${diffHours} ч назад`;
    if (diffDays === 1) return 'вчера';
    if (diffDays < 7) return `${diffDays} дн назад`;

    return date.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'short',
    });
  };

  /**
   * Получить превью последнего сообщения
   */
  const getLastMessagePreview = (room) => {
    if (!room.last_message) return 'Нет сообщений';

    const msg = room.last_message;

    // Если сообщение заблокировано модерацией
    if (msg.moderation_status === 'blocked') {
      return '🚫 Сообщение заблокировано';
    }

    // Если есть вложения
    if (msg.attachments && msg.attachments.length > 0) {
      const fileCount = msg.attachments.length;
      const fileText = fileCount === 1 ? 'файл' : fileCount < 5 ? 'файла' : 'файлов';
      return `📎 ${fileCount} ${fileText}`;
    }

    // Обрезать длинное сообщение
    const maxLength = 50;
    if (msg.message && msg.message.length > maxLength) {
      return msg.message.substring(0, maxLength) + '...';
    }

    return msg.message || 'Файл';
  };

  if (loading) {
    return (
      <div className="chat-list-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="chat-list">
      {/* Header */}
      <div className="chat-list-header">
        <h2 className="chat-list-title">Чаты</h2>
      </div>

      {/* Список существующих комнат */}
      {rooms.length > 0 && (
        <div className="chat-rooms-section">
          <div className="chat-section-header">Активные чаты</div>
          <div className="chat-rooms-list">
            {rooms.map((room) => (
              <div
                key={room.id}
                className={`chat-room-item ${selectedRoom?.id === room.id ? 'chat-room-selected' : ''}`}
                onClick={() => onRoomSelect(room)}
              >
                <div className="chat-room-avatar">
                  {room.participant_name?.[0]?.toUpperCase() || '?'}
                </div>
                <div className="chat-room-info">
                  <div className="chat-room-header-row">
                    <div className="chat-room-name">{room.participant_name || 'Пользователь'}</div>
                    <div className="chat-room-time">
                      {formatLastMessageTime(room.last_message?.created_at)}
                    </div>
                  </div>
                  <div className="chat-room-last-message">
                    {getLastMessagePreview(room)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Список доступных пользователей для начала чата */}
      {availableUsers.length > 0 && (
        <div className="chat-users-section">
          <div className="chat-section-header">
            {user?.role === 'student' ? 'Преподаватели' : 'Студенты'}
          </div>
          <div className="chat-users-list">
            {availableUsers.map((availableUser) => {
              // Проверить, есть ли уже комната с этим пользователем
              const existingRoom = rooms.find(
                (r) => r.participant_id === availableUser.id
              );

              // Если комната уже существует, не показывать в списке доступных
              if (existingRoom) return null;

              return (
                <div
                  key={availableUser.id}
                  className="chat-user-item"
                  onClick={() => handleSelectUser(availableUser.id)}
                >
                  <div className="chat-user-avatar">
                    {availableUser.full_name?.[0]?.toUpperCase() || '?'}
                  </div>
                  <div className="chat-user-info">
                    <div className="chat-user-name">{availableUser.full_name || 'Пользователь'}</div>
                    <div className="chat-user-role">
                      {availableUser.role === 'teacher' ? 'Преподаватель' : 'Студент'}
                    </div>
                  </div>
                  {creatingRoom === availableUser.id && (
                    <div className="chat-user-creating">
                      <Spinner size="small" />
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Пустое состояние */}
      {rooms.length === 0 && availableUsers.length === 0 && (
        <div className="chat-list-empty">
          <div className="chat-list-empty-icon">💬</div>
          <p>Нет доступных чатов</p>
        </div>
      )}
    </div>
  );
};

export default ChatList;
