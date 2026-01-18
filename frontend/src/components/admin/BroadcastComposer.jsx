import { useState, useEffect } from 'react';
import Button from '../common/Button.jsx';
import Modal from '../common/Modal.jsx';
import Spinner from '../common/Spinner.jsx';
import TelegramUsersTable from './TelegramUsersTable.jsx';
import { useBroadcast } from '../../hooks/useBroadcast.js';
import './BroadcastComposer.css';

const MAX_MESSAGE_LENGTH = 4096;
const MIN_MESSAGE_LENGTH = 1;

/**
 * Компонент для создания и отправки рассылок
 */
export const BroadcastComposer = () => {
  const {
    broadcastLists,
    linkedUsers,
    loading,
    error,
    fetchBroadcastLists,
    fetchLinkedUsers,
    sendBroadcast,
  } = useBroadcast();

  const [recipientType, setRecipientType] = useState('list'); // 'list' or 'users'
  const [selectedListId, setSelectedListId] = useState('');
  const [selectedUserIds, setSelectedUserIds] = useState([]);
  const [message, setMessage] = useState('');
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [sending, setSending] = useState(false);
  const [localError, setLocalError] = useState(null);
  const [validationErrors, setValidationErrors] = useState({});

  useEffect(() => {
    fetchBroadcastLists();
    fetchLinkedUsers();
  }, [fetchBroadcastLists, fetchLinkedUsers]);

  // Compare UUIDs as strings, not numbers - parseInt() on UUID returns NaN
  const selectedList = broadcastLists.find(list => list.id === selectedListId);
  const recipientCount = recipientType === 'list'
    ? (selectedList?.user_count || selectedList?.user_ids?.length || 0)
    : selectedUserIds.length;
  const characterCount = message.length;
  const trimmedMessage = message.trim();

  // Валидация формы
  const validateForm = () => {
    const errors = {};

    if (recipientType === 'list' && !selectedListId) {
      errors.recipients = 'Выберите список получателей';
    }

    if (recipientType === 'users' && selectedUserIds.length === 0) {
      errors.recipients = 'Выберите хотя бы одного получателя';
    }

    if (!trimmedMessage) {
      errors.message = 'Сообщение не может быть пустым';
    } else if (trimmedMessage.length < MIN_MESSAGE_LENGTH) {
      errors.message = 'Сообщение слишком короткое';
    } else if (characterCount > MAX_MESSAGE_LENGTH) {
      errors.message = `Сообщение не может быть длиннее ${MAX_MESSAGE_LENGTH} символов`;
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const hasRecipients = recipientType === 'list' ? !!selectedListId : selectedUserIds.length > 0;
  const isValid = hasRecipients && trimmedMessage.length > 0 && characterCount <= MAX_MESSAGE_LENGTH;

  const handleSend = async () => {
    if (!validateForm()) {
      setLocalError('Пожалуйста, исправьте ошибки в форме');
      return;
    }

    setSending(true);
    setLocalError(null);

    try {
      const payload = {
        message: trimmedMessage,
      };

      // Добавляем list_id или user_ids в зависимости от типа получателей
      if (recipientType === 'list') {
        payload.list_id = selectedListId;
      } else {
        payload.user_ids = selectedUserIds;
      }

      await sendBroadcast(payload);

      // Сброс формы после успешной отправки
      setRecipientType('list');
      setSelectedListId('');
      setSelectedUserIds([]);
      setMessage('');
      setShowConfirmModal(false);
      setValidationErrors({});
    } catch (err) {
      setLocalError(err.message || 'Ошибка при отправке рассылки');
    } finally {
      setSending(false);
    }
  };

  const handleOpenConfirm = () => {
    if (validateForm()) {
      setLocalError(null);
      setShowConfirmModal(true);
    } else {
      setLocalError('Пожалуйста, исправьте ошибки в форме');
    }
  };

  const handleMessageChange = (e) => {
    setMessage(e.target.value);
    // Очистить ошибку сообщения при начале печати
    if (validationErrors.message) {
      setValidationErrors(prev => ({ ...prev, message: null }));
    }
  };

  const handleListChange = (e) => {
    setSelectedListId(e.target.value);
    // Очистить ошибку получателей при выборе
    if (validationErrors.recipients) {
      setValidationErrors(prev => ({ ...prev, recipients: null }));
    }
  };

  const handleRecipientTypeChange = (type) => {
    setRecipientType(type);
    // Очистить ошибки при смене типа
    setValidationErrors({});
  };

  const handleUserSelectionChange = (userIds) => {
    setSelectedUserIds(userIds);
    // Очистить ошибку получателей при выборе
    if (validationErrors.recipients) {
      setValidationErrors(prev => ({ ...prev, recipients: null }));
    }
  };

  if (loading && broadcastLists.length === 0) {
    return (
      <div className="broadcast-composer-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="broadcast-composer">
      <div className="broadcast-composer-header">
        <h3 className="broadcast-composer-title">Новая рассылка</h3>
        <p className="broadcast-composer-subtitle">
          Создайте сообщение для отправки выбранной группе получателей
        </p>
      </div>

      {(localError || error) && (
        <div className="broadcast-error-message">
          <svg className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
              fill="currentColor"
            />
          </svg>
          <span>{localError || error}</span>
        </div>
      )}

      <div className="broadcast-composer-form">
        {/* Выбор типа получателей */}
        <div className="form-group">
          <label className="form-label">
            Тип получателей
            <span className="required-mark">*</span>
          </label>
          <div className="recipient-type-tabs">
            <button
              type="button"
              className={`recipient-type-tab ${recipientType === 'list' ? 'active' : ''}`}
              onClick={() => handleRecipientTypeChange('list')}
            >
              <svg className="tab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
              </svg>
              Список рассылки
            </button>
            <button
              type="button"
              className={`recipient-type-tab ${recipientType === 'users' ? 'active' : ''}`}
              onClick={() => handleRecipientTypeChange('users')}
            >
              <svg className="tab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              Выбрать пользователей
            </button>
          </div>
        </div>

        {/* Выбор списка рассылки */}
        {recipientType === 'list' && (
          <div className="form-group">
            <label className="form-label" htmlFor="list-select">
              Список получателей
              <span className="required-mark">*</span>
            </label>
            {broadcastLists.length === 0 ? (
              <div className="no-lists-message">
                <p>У вас нет созданных списков рассылки.</p>
                <p>Перейдите во вкладку &quot;Списки рассылки&quot;, чтобы создать список.</p>
              </div>
            ) : (
              <>
                <select
                  id="list-select"
                  className={`form-select ${validationErrors.recipients ? 'error' : ''}`}
                  value={selectedListId}
                  onChange={handleListChange}
                  required
                >
                  <option value="">Выберите список</option>
                  {broadcastLists.map((list) => (
                    <option key={list.id} value={list.id}>
                      {list.name} ({list.user_count || list.user_ids?.length || 0} получателей)
                    </option>
                  ))}
                </select>
                {validationErrors.recipients && (
                  <div className="field-error">{validationErrors.recipients}</div>
                )}
              </>
            )}

            {selectedList && (
              <div className="list-info">
                <div className="list-info-badge">
                  Получателей: {recipientCount}
                </div>
                {selectedList.description && (
                  <p className="list-info-description">{selectedList.description}</p>
                )}
              </div>
            )}
          </div>
        )}

        {/* Выбор конкретных пользователей */}
        {recipientType === 'users' && (
          <div className="form-group">
            <label className="form-label">
              Получатели
              <span className="required-mark">*</span>
            </label>
            {validationErrors.recipients && (
              <div className="field-error">{validationErrors.recipients}</div>
            )}
            <div className="users-selection-container">
              <TelegramUsersTable
                users={linkedUsers}
                loading={loading}
                selectedUsers={selectedUserIds}
                onSelectionChange={handleUserSelectionChange}
                showCheckboxes={true}
              />
            </div>
          </div>
        )}

        <div className="form-group">
          <label className="form-label" htmlFor="message-textarea">
            Текст сообщения
            <span className="required-mark">*</span>
          </label>
          <textarea
            id="message-textarea"
            className={`form-textarea ${validationErrors.message ? 'error' : ''} ${characterCount > MAX_MESSAGE_LENGTH ? 'warning' : ''}`}
            value={message}
            onChange={handleMessageChange}
            placeholder="Введите текст сообщения для рассылки..."
            rows={10}
            required
          />
          <div className="character-counter">
            <span className={characterCount > MAX_MESSAGE_LENGTH ? 'error' : characterCount > MAX_MESSAGE_LENGTH * 0.9 ? 'warning' : ''}>
              {characterCount} / {MAX_MESSAGE_LENGTH} символов
            </span>
          </div>
          {validationErrors.message && (
            <div className="field-error">{validationErrors.message}</div>
          )}
        </div>

        {message.trim() && (
          <div className="message-preview">
            <h4 className="preview-title">Предпросмотр сообщения:</h4>
            <div className="preview-box">
              <div className="preview-header">
                <svg className="telegram-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path
                    d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2ZM16.64 8.8C16.49 10.38 15.84 14.22 15.51 15.99C15.37 16.74 15.09 16.99 14.83 17.01C14.25 17.07 13.81 16.64 13.25 16.27C12.37 15.69 11.87 15.33 11.02 14.77C10.03 14.12 10.67 13.76 11.24 13.18C11.39 13.03 13.95 10.7 14 10.49C14.0069 10.4582 14.006 10.4252 13.9973 10.3938C13.9886 10.3624 13.9724 10.3337 13.95 10.31C13.89 10.26 13.81 10.28 13.74 10.29C13.65 10.31 12.25 11.24 9.52 13.08C9.12 13.35 8.76 13.49 8.44 13.48C8.08 13.47 7.4 13.28 6.89 13.11C6.26 12.91 5.77 12.8 5.81 12.45C5.83 12.27 6.08 12.09 6.55 11.9C9.47 10.63 11.41 9.79 12.38 9.39C15.16 8.23 15.73 8.03 16.11 8.03C16.19 8.03 16.38 8.05 16.5 8.15C16.6 8.23 16.63 8.34 16.64 8.42C16.63 8.48 16.65 8.66 16.64 8.8Z"
                    fill="currentColor"
                  />
                </svg>
                <span>Telegram</span>
              </div>
              <div className="preview-content">{message}</div>
            </div>
          </div>
        )}

        <div className="broadcast-composer-actions">
          <Button
            onClick={handleOpenConfirm}
            disabled={!isValid}
            size="large"
          >
            Отправить рассылку
          </Button>
        </div>
      </div>

      {/* Модальное окно подтверждения */}
      <Modal
        isOpen={showConfirmModal}
        onClose={() => setShowConfirmModal(false)}
        title="Подтверждение отправки"
      >
        <div className="confirm-content">
          <p className="confirm-message">
            Вы действительно хотите отправить рассылку?
          </p>
          <div className="confirm-details">
            {recipientType === 'list' && (
              <div className="confirm-detail">
                <span className="detail-label">Список:</span>
                <span className="detail-value">{selectedList?.name}</span>
              </div>
            )}
            {recipientType === 'users' && (
              <div className="confirm-detail">
                <span className="detail-label">Тип:</span>
                <span className="detail-value">Выбранные пользователи</span>
              </div>
            )}
            <div className="confirm-detail">
              <span className="detail-label">Получателей:</span>
              <span className="detail-value">{recipientCount}</span>
            </div>
            <div className="confirm-detail">
              <span className="detail-label">Символов:</span>
              <span className="detail-value">{characterCount}</span>
            </div>
          </div>
          <div className="confirm-actions">
            <Button
              variant="secondary"
              onClick={() => setShowConfirmModal(false)}
              disabled={sending}
            >
              Отмена
            </Button>
            <Button
              onClick={handleSend}
              loading={sending}
            >
              Отправить
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default BroadcastComposer;
