import { useState, useEffect } from 'react';
import Input from './Input.jsx';
import Button from './Button.jsx';
import './TelegramUsernameInput.css';

/**
 * Компонент для ввода и сохранения Telegram username
 * Используется в профилях преподавателей и студентов
 */
export const TelegramUsernameInput = ({
  initialValue = '',
  onSave,
  loading = false,
  error = null,
}) => {
  const [username, setUsername] = useState(initialValue);
  const [validationError, setValidationError] = useState(null);
  const [saveStatus, setSaveStatus] = useState(null); // 'success', 'error', null
  const [hasChanged, setHasChanged] = useState(false);

  // Очистить статус успеха через 3 секунды
  useEffect(() => {
    if (saveStatus === 'success') {
      const timer = setTimeout(() => {
        setSaveStatus(null);
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [saveStatus]);

  // Валидация Telegram username
  const validateUsername = (value) => {
    // Пустое значение допустимо - просто нет username
    if (!value.trim()) {
      setValidationError(null);
      return true;
    }

    // Telegram username: буквы, цифры, подчёркивание, 5-32 символа
    const usernameRegex = /^[a-zA-Z][a-zA-Z0-9_]{4,31}$/;
    if (!usernameRegex.test(value)) {
      setValidationError(
        'Telegram логин: 5-32 символа, начинается с буквы (буквы, цифры, подчёркивание)'
      );
      return false;
    }

    setValidationError(null);
    return true;
  };

  const handleChange = (e) => {
    const value = e.target.value;
    setUsername(value);
    setHasChanged(value !== initialValue);
    // Очистить статусы при изменении значения
    setSaveStatus(null);
    setValidationError(null);
  };

  const handleSave = async () => {
    if (!validateUsername(username)) {
      return;
    }

    try {
      setSaveStatus(null);
      await onSave(username);
      setSaveStatus('success');
      setHasChanged(false);
    } catch (err) {
      setSaveStatus('error');
    }
  };

  const handleCancel = () => {
    setUsername(initialValue);
    setHasChanged(false);
    setSaveStatus(null);
    setValidationError(null);
  };

  return (
    <div className="TelegramUsernameInput telegram-username-input">
      <div className="telegram-username-field">
        <Input
          type="text"
          label="Telegram логин"
          name="telegram_username"
          value={username}
          onChange={handleChange}
          placeholder="Например: software_studio"
          error={validationError || error}
          disabled={loading}
          maxLength={32}
        />
      </div>

      <div className="telegram-username-actions">
        {hasChanged && (
          <>
            <Button
              variant="secondary"
              size="small"
              onClick={handleCancel}
              disabled={loading}
              data-testid="cancel-telegram-button"
            >
              Отменить
            </Button>
            <Button
              variant="primary"
              size="small"
              onClick={handleSave}
              loading={loading}
              disabled={loading || validationError !== null}
              data-testid="save-telegram-button"
            >
              Сохранить
            </Button>
          </>
        )}
      </div>

      {saveStatus === 'success' && (
        <div className="telegram-username-status telegram-username-status-success">
          <svg
            className="status-icon"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"
              fill="currentColor"
            />
          </svg>
          <span>Сохранено успешно</span>
        </div>
      )}

      {saveStatus === 'error' && (
        <div className="telegram-username-status telegram-username-status-error">
          <svg
            className="status-icon"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
              fill="currentColor"
            />
          </svg>
          <span>Ошибка при сохранении</span>
        </div>
      )}
    </div>
  );
};

export default TelegramUsernameInput;
