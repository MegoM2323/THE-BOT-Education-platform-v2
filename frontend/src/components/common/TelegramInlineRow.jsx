import { useState } from 'react';
import Button from './Button.jsx';
import ConfirmModal from './ConfirmModal.jsx';
import Spinner from './Spinner.jsx';
import { useTelegram } from '../../hooks/useTelegram.js';
import { openTelegramBot, isSafariOnIOS } from '../../utils/browserDetection.js';
import './TelegramInlineRow.css';

/**
 * Компактный inline компонент для отображения статуса привязки Telegram в профиле.
 * Показывает статус (привязан/не привязан), username и кнопку действия.
 */
export const TelegramInlineRow = () => {
  const {
    linkStatus,
    token,
    botUsername,
    loading,
    error,
    generateToken,
    unlinkAccount,
    clearError,
  } = useTelegram();

  const [showUnlinkConfirm, setShowUnlinkConfirm] = useState(false);
  const [isLinking, setIsLinking] = useState(false);
  const [localError, setLocalError] = useState(null);
  const [showSafariWarning, setShowSafariWarning] = useState(false);

  const isLinked = linkStatus?.linked;

  /**
   * Обработчик привязки Telegram.
   * Генерирует токен и открывает бота в новой вкладке.
   * На Safari iOS: показывает предупреждение перед перенаправлением.
   * На других браузерах: использует оптимальный метод открытия окна.
   */
  const handleLink = async () => {
    try {
      setLocalError(null);
      setIsLinking(true);
      const data = await generateToken();

      if (data.token && data.bot_username) {
        const botUrl = `https://t.me/${data.bot_username}?start=${data.token}`;

        // Safari on iOS: show warning before redirect
        if (isSafariOnIOS()) {
          setShowSafariWarning(true);
          // Delay redirect to let user see the warning
          setTimeout(() => {
            window.location.href = botUrl;
          }, 1500);
        } else {
          // Other browsers: use intelligent popup opening
          const result = openTelegramBot(botUrl);
          if (result.blocked) {
            setLocalError('Не удалось открыть бот. Попробуйте отключить блокировку всплывающих окон.');
          }
        }
      }
    } catch (err) {
      // Сохраняем локальную ошибку для отображения
      setLocalError(err.message || 'Не удалось сгенерировать токен');
      // Примечание: useTelegram hook автоматически обновит linkStatus при ошибке
    } finally {
      setIsLinking(false);
    }
  };

  /**
   * Обработчик повторной попытки привязки.
   * Сбрасывает все ошибки и статус, затем пытается снова привязать.
   */
  const handleRetry = async () => {
    // Сбрасываем локальную ошибку и ошибку из хука
    setLocalError(null);
    clearError();
    // Вызываем привязку заново
    await handleLink();
  };

  /**
   * Обработчик отвязки Telegram.
   */
  const handleUnlink = async () => {
    try {
      setLocalError(null);
      await unlinkAccount();
      setShowUnlinkConfirm(false);
    } catch (err) {
      setLocalError(err.message || 'Не удалось отвязать Telegram');
    }
  };

  // Состояние загрузки при первичной загрузке статуса
  if (loading && !linkStatus) {
    return (
      <div className="telegram-inline-row">
        <div className="telegram-loading">
          <Spinner size="sm" className="telegram-loading-spinner" />
          <span>Загрузка...</span>
        </div>
      </div>
    );
  }

  // Отображение ошибки (но только если не привязан)
  // Если ошибка и аккаунт всё ещё привязан - показываем обычный интерфейс
  if ((error || localError) && !isLinked) {
    return (
      <div className="telegram-inline-row">
        <div className="telegram-error">
          <svg className="telegram-error-icon" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
          </svg>
          <span>{error || localError}</span>
        </div>
        <div className="telegram-action-btn">
          <Button
            size="small"
            variant="outline"
            onClick={handleRetry}
            loading={isLinking}
            disabled={isLinking}
          >
            Повторить
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="telegram-inline-row">
      <div className="telegram-status-indicator">
        {isLinked ? (
          <>
            <span className="telegram-status-icon linked" title="Telegram привязан">
              &#10003;
            </span>
            <span className="telegram-username">
              {linkStatus.telegram?.username
                ? `@${linkStatus.telegram.username}`
                : 'Привязан'}
            </span>
          </>
        ) : (
          <>
            <span className="telegram-status-icon not-linked" title="Telegram не привязан">
              &#9675;
            </span>
            <span className="telegram-not-linked-text">Не привязан</span>
          </>
        )}
      </div>

      <div className="telegram-action-btn">
        {isLinked ? (
          <Button
            size="small"
            variant="danger"
            onClick={() => setShowUnlinkConfirm(true)}
            disabled={loading}
          >
            Отвязать
          </Button>
        ) : (
          <Button
            size="small"
            variant="primary"
            onClick={handleLink}
            loading={isLinking}
            disabled={isLinking || loading}
          >
            Привязать
          </Button>
        )}
      </div>

      {/* Модальное окно подтверждения отвязки */}
      <ConfirmModal
        isOpen={showUnlinkConfirm}
        onClose={() => setShowUnlinkConfirm(false)}
        onConfirm={handleUnlink}
        title="Отвязка Telegram"
        message="Вы действительно хотите отвязать Telegram? Вы перестанете получать уведомления о занятиях."
        confirmText="Отвязать"
        confirmVariant="danger"
      />

      {/* Safari iOS warning before redirect */}
      {showSafariWarning && (
        <div className="telegram-safari-warning">
          <span>Перенаправление в Telegram...</span>
        </div>
      )}
    </div>
  );
};

export default TelegramInlineRow;
