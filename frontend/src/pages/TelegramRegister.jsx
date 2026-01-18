import { useState, useRef, useEffect } from 'react';
import { logger } from '../utils/logger.js';
import { useNavigate } from 'react-router-dom';
import Button from '../components/common/Button.jsx';
import { useNotification } from '../hooks/useNotification.js';
import * as authAPI from '../api/auth.js';
import './TelegramRegister.css';

/**
 * Страница регистрации студента через Telegram username
 */
export const TelegramRegister = () => {
  const navigate = useNavigate();
  const { showNotification } = useNotification();

  const [telegramUsername, setTelegramUsername] = useState('');
  const [loading, setLoading] = useState(false);

  // Ref для хранения ID таймера навигации
  const navigationTimerRef = useRef(null);

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Валидация
    if (!telegramUsername.trim()) {
      showNotification('Пожалуйста, введите Telegram username', 'error');
      return;
    }

    // Убираем @ если есть
    const username = telegramUsername.trim().startsWith('@')
      ? telegramUsername.trim().substring(1)
      : telegramUsername.trim();

    try {
      setLoading(true);

      // Отправляем запрос на регистрацию
      const response = await authAPI.registerViaTelegram(username);

      showNotification('Регистрация успешна! Вы можете войти в систему.', 'success');

      // Перенаправляем на страницу входа
      navigationTimerRef.current = setTimeout(() => {
        navigate('/login');
        navigationTimerRef.current = null;
      }, 1500);
    } catch (error) {
      console.error('Ошибка регистрации:', error);
      const errorMsg = error.response?.data?.message || 'Ошибка при регистрации';
      showNotification(errorMsg, 'error');
    } finally {
      setLoading(false);
    }
  };

  // Cleanup таймера при unmount
  useEffect(() => {
    return () => {
      if (navigationTimerRef.current) {
        clearTimeout(navigationTimerRef.current);
      }
    };
  }, []);

  return (
    <div className="telegram-register-page">
      <div className="telegram-register-container">
        <div className="telegram-register-card">
          <h1 className="register-title">Регистрация через Telegram</h1>

          <p className="register-description">
            Введите ваш Telegram username для регистрации в системе.
            Это позволит администратору отправлять вам уведомления в Telegram.
          </p>

          <form onSubmit={handleSubmit} className="register-form">
            <div className="form-group">
              <label htmlFor="telegram-username">Telegram Username:</label>
              <div className="input-wrapper">
                <span className="input-prefix">@</span>
                <input
                  id="telegram-username"
                  type="text"
                  placeholder="example_username"
                  value={telegramUsername}
                  onChange={(e) => setTelegramUsername(e.target.value)}
                  disabled={loading}
                  autoComplete="off"
                  required
                />
              </div>
              <p className="input-hint">
                Введите ваш Telegram username без символа @. Вы можете найти его в настройках профиля Telegram.
              </p>
            </div>

            <Button
              type="submit"
              fullWidth
              loading={loading}
              disabled={!telegramUsername.trim()}
            >
              Зарегистрироваться
            </Button>
          </form>

          <div className="register-footer">
            <p>
              Уже зарегистрированы?{' '}
              <button
                className="link-button"
                onClick={() => navigate('/login')}
                disabled={loading}
              >
                Войдите здесь
              </button>
            </p>
          </div>
        </div>

        <div className="register-info">
          <div className="info-box">
            <h3>Как это работает?</h3>
            <ol>
              <li>Введите ваш Telegram username</li>
              <li>Вам будет создан аккаунт студента</li>
              <li>Используйте ваш email для входа в систему</li>
              <li>Администратор сможет отправлять вам уведомления в Telegram</li>
            </ol>
          </div>

          <div className="info-box">
            <h3>Как найти свой Telegram username?</h3>
            <ol>
              <li>Откройте приложение Telegram</li>
              <li>Перейдите в Параметры (Settings)</li>
              <li>Нажмите на свой профиль</li>
              <li>Ваш username указан в строке Username</li>
              <li>Скопируйте username без символа @</li>
            </ol>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TelegramRegister;
