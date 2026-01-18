/**
 * API эндпоинты для чата
 */

import apiClient from "./client.js";

/**
 * Получить или создать комнату с участником
 * @param {string} participantId - ID второго участника (UUID)
 * @returns {Promise<Object>} Объект комнаты
 * @throws {APIError} Если возникла ошибка при создании комнаты
 */
export const getOrCreateRoom = async (participantId) => {
  try {
    console.log("[chat.js] getOrCreateRoom called:", {
      participantId,
      participantIdType: typeof participantId,
      participantIdLength: participantId?.length,
    });

    if (!participantId) {
      console.error("[chat.js] ERROR: participantId is empty or undefined!");
      throw new Error("participant_id is required");
    }

    const response = await apiClient.post("/chat/rooms", {
      participant_id: participantId,
    });

    console.log("[chat.js] getOrCreateRoom response:", response);
    return response;
  } catch (error) {
    console.error("[chat.js] getOrCreateRoom error:", {
      message: error.message,
      status: error.status,
      data: error.data,
    });
    throw error;
  }
};

/**
 * Получить список комнат текущего пользователя
 * @returns {Promise<Array>} Список комнат с информацией об участниках
 * @throws {APIError} Если возникла ошибка при получении списка комнат
 */
export const getMyRooms = async () => {
  try {
    const response = await apiClient.get("/chat/rooms");
    // Backend возвращает {success: true, data: [...]}
    // apiClient auto-unwraps data.data, so response is already the array
    return Array.isArray(response) ? response : [];
  } catch (error) {
    console.error("Error fetching chat rooms:", error);
    throw error;
  }
};

/**
 * Получить все комнаты чатов (для администратора)
 * @returns {Promise<Array>} Список всех комнат с информацией об участниках
 * @throws {APIError} Если возникла ошибка при получении списка
 */
export const getAllRooms = async () => {
  try {
    const response = await apiClient.get("/admin/chats");
    const chats = response?.chats || response;
    return Array.isArray(chats) ? chats : [];
  } catch (error) {
    console.error("Error fetching all chat rooms (admin):", error);
    throw error;
  }
};

/**
 * Отправить сообщение в комнату
 * @param {string} roomId - ID комнаты (UUID)
 * @param {string} message - Текст сообщения
 * @param {File[]} files - Массив файлов для загрузки (опционально)
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Object>} Отправленное сообщение с полями: id, status, moderation_completed_at, attachments
 * @throws {APIError} Если возникла ошибка при отправке сообщения
 */
export const sendMessage = async (roomId, message, files = [], signal) => {
  try {
    // DEBUG: Логируем параметры
    console.log("[chat.js] sendMessage called:", {
      roomId,
      message,
      messageType: typeof message,
      messageLength: message?.length,
      filesCount: files?.length,
    });

    // Всегда используем FormData (backend всегда ожидает multipart/form-data)
    const formData = new FormData();
    formData.append("message", message);
    if (files && files.length > 0) {
      files.forEach((file) => formData.append("files", file));
    }

    // DEBUG: Проверяем что в FormData
    console.log("[chat.js] FormData содержит:");
    for (let [key, value] of formData.entries()) {
      console.log(`  ${key}:`, value);
    }

    console.log("[chat.js] Sending POST to:", `/chat/rooms/${roomId}/messages`);

    // Используем apiClient для поддержки AbortController
    const data = await apiClient.post(
      `/chat/rooms/${roomId}/messages`,
      formData,
      {
        signal, // Передаем signal для отмены запроса
      },
    );

    console.log("[chat.js] Response data:", data);
    return data;
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== "AbortError") {
      console.error("[chat.js] Error sending message:", error);
    }
    throw error;
  }
};

/**
 * Получить сообщения из комнаты
 * @param {string} roomId - ID комнаты (UUID)
 * @param {number} limit - Количество сообщений (по умолчанию 50)
 * @param {number} offset - Смещение для пагинации (по умолчанию 0)
 * @returns {Promise<Array>} Список сообщений с полями: id, sender_id, message_text, status, attachments, created_at
 * @throws {APIError} Если возникла ошибка при получении сообщений
 */
export const getMessages = async (roomId, limit = 50, offset = 0) => {
  try {
    const params = new URLSearchParams();
    params.append("limit", limit.toString());
    params.append("offset", offset.toString());

    const response = await apiClient.get(
      `/chat/rooms/${roomId}/messages?${params.toString()}`,
    );

    // Backend возвращает {success: true, data: [...]}
    // apiClient auto-unwraps data.data, so response is already the array
    return Array.isArray(response) ? response : [];
  } catch (error) {
    console.error("Error fetching messages:", error);
    throw error;
  }
};

/**
 * Скачать файл из сообщения
 * @param {string} roomId - ID комнаты (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @param {AbortSignal} [signal] - Сигнал для отмены запроса (предотвращает утечки памяти)
 * @returns {Promise<Blob>} Файл для скачивания
 * @throws {Error} Если возникла ошибка при скачивании файла
 */
export const downloadFile = async (roomId, fileId, signal) => {
  try {
    // Для blob ответов используем прямой fetch с сигналом
    const baseURL = import.meta.env.VITE_API_URL || "/api/v1";
    const response = await fetch(
      `${baseURL}/chat/rooms/${roomId}/files/${fileId}`,
      {
        method: "GET",
        credentials: "include",
        signal, // Передаем signal для отмены запроса
      },
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage =
        errorData?.error?.message ||
        errorData?.message ||
        "Failed to download file";
      throw new Error(errorMessage);
    }

    return await response.blob();
  } catch (error) {
    // Не логируем AbortError как ошибку
    if (error.name !== "AbortError") {
      console.error("Error downloading file:", error);
    }
    throw error;
  }
};

/**
 * Получить URL для скачивания файла
 * @param {string} roomId - ID комнаты (UUID)
 * @param {string} fileId - ID файла (UUID)
 * @returns {string} URL для скачивания
 */
export const getFileDownloadURL = (roomId, fileId) => {
  const baseURL = import.meta.env.VITE_API_URL || "/api/v1";
  return `${baseURL}/chat/rooms/${roomId}/files/${fileId}`;
};

export default {
  getOrCreateRoom,
  getMyRooms,
  getAllRooms,
  sendMessage,
  getMessages,
  downloadFile,
  getFileDownloadURL,
};
