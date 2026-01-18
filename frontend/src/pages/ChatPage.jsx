import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ChatList from '../components/chat/ChatList.jsx';
import ChatWindow from '../components/chat/ChatWindow.jsx';
import './ChatPage.css';

/**
 * Страница чата с адаптивным layout
 */
const ChatPage = () => {
  const { roomId } = useParams();
  const navigate = useNavigate();

  const [selectedRoom, setSelectedRoom] = useState(null);
  const [showMobileChat, setShowMobileChat] = useState(false);

  /**
   * Обработчик выбора комнаты
   */
  const handleRoomSelect = (room) => {
    setSelectedRoom(room);
    setShowMobileChat(true);

    // Обновить URL
    if (room?.id) {
      navigate(`/chat/${room.id}`, { replace: true });
    }
  };

  /**
   * Вернуться к списку чатов (на mobile)
   */
  const handleBackToList = () => {
    setShowMobileChat(false);
    setSelectedRoom(null);
    navigate('/chat', { replace: true });
  };

  /**
   * При монтировании: если есть roomId в URL, попытаться выбрать комнату
   */
  useEffect(() => {
    if (roomId) {
      // Комната будет выбрана через ChatList после загрузки
      setShowMobileChat(true);
    }
  }, [roomId]);

  return (
    <div className="chat-page">
      <div className="chat-page-container">
        {/* Chat List (слева, или на весь экран на mobile если нет выбранной комнаты) */}
        <div className={`chat-page-list ${showMobileChat ? 'chat-page-list-hidden' : ''}`}>
          <ChatList selectedRoom={selectedRoom} onRoomSelect={handleRoomSelect} urlRoomId={roomId} />
        </div>

        {/* Chat Window (справа, или на весь экран на mobile если выбрана комната) */}
        <div className={`chat-page-window ${showMobileChat ? 'chat-page-window-visible' : ''}`}>
          {selectedRoom && (
            <div className="chat-page-mobile-back">
              <button className="chat-back-button" onClick={handleBackToList}>
                ← Назад к чатам
              </button>
            </div>
          )}
          <ChatWindow room={selectedRoom} />
        </div>
      </div>
    </div>
  );
};

export default ChatPage;
