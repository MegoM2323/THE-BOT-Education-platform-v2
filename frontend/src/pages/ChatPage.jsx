import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import ChatList from '../components/chat/ChatList.jsx';
import ChatWindow from '../components/chat/ChatWindow.jsx';
import './ChatPage.css';

/**
 * Страница чата с адаптивным layout
 */
const ChatPage = () => {
  const { roomId } = useParams();

  const [selectedRoom, setSelectedRoom] = useState(null);
  const [showMobileChat, setShowMobileChat] = useState(false);

  /**
   * Обработчик выбора комнаты
   */
  const handleRoomSelect = (room) => {
    console.log('[ChatPage] handleRoomSelect called with room:', room);
    setSelectedRoom(room);
    setShowMobileChat(true);
    console.log('[ChatPage] selectedRoom set, showMobileChat set to true');
  };

  /**
   * Вернуться к списку чатов (на mobile)
   */
  const handleBackToList = () => {
    setShowMobileChat(false);
    setSelectedRoom(null);
    // URL не обновляем - работаем через state для совместимости со всеми ролями
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

  console.log('[ChatPage] Render - selectedRoom:', selectedRoom, 'showMobileChat:', showMobileChat);

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
