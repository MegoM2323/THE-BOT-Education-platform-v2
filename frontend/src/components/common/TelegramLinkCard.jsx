import { useState, useEffect, useRef } from 'react';
import Button from './Button.jsx';
import Modal from './Modal.jsx';
import ConfirmModal from './ConfirmModal.jsx';
import Spinner from './Spinner.jsx';
import { useTelegram } from '../../hooks/useTelegram.js';
import { isSafariOnIOS } from '../../utils/browserDetection.js';
import './TelegramLinkCard.css';

/**
 * Компонент для привязки Telegram аккаунта
 */
export const TelegramLinkCard = () => {
  const {
    linkStatus,
    token,
    botUsername,
    loading,
    error,
    generateToken,
    unlinkAccount,
    refetchLinkStatus,
  } = useTelegram();

  const [showLinkModal, setShowLinkModal] = useState(false);
  const [showUnlinkConfirm, setShowUnlinkConfirm] = useState(false);
  const [linkUrl, setLinkUrl] = useState('');
  const [copySuccess, setCopySuccess] = useState(false);
  const [localError, setLocalError] = useState(null);

  // Ref для хранения ID таймера копирования
  const copyTimerRef = useRef(null);

  // Генерировать ссылку на бота при открытии модального окна
  useEffect(() => {
    if (showLinkModal && token && botUsername) {
      const url = `https://t.me/${botUsername}?start=${token}`;
      setLinkUrl(url);
    }
  }, [showLinkModal, token, botUsername]);

  // Очистить сообщение об ошибке при закрытии модального окна
  useEffect(() => {
    if (!showLinkModal && !showUnlinkConfirm) {
      setLocalError(null);
    }
  }, [showLinkModal, showUnlinkConfirm]);

  const handleOpenLinkModal = async () => {
    try {
      setLocalError(null);
      await generateToken();
      setShowLinkModal(true);
    } catch (err) {
      setLocalError(err.message || 'Не удалось генерировать токен');
    }
  };

  const handleCloseLinkModal = () => {
    setShowLinkModal(false);
    setLocalError(null);
    // Проверяем статус привязки после закрытия модального окна
    refetchLinkStatus();
  };

  const handleUnlink = async () => {
    try {
      setLocalError(null);
      await unlinkAccount();
      setShowUnlinkConfirm(false);
    } catch (err) {
      setLocalError(err.message || 'Не удалось отвязать Telegram');
    }
  };

  const handleCopyLink = () => {
    if (linkUrl) {
      // Очистить предыдущий таймер если есть
      if (copyTimerRef.current) {
        clearTimeout(copyTimerRef.current);
      }

      navigator.clipboard.writeText(linkUrl);
      setCopySuccess(true);

      copyTimerRef.current = setTimeout(() => {
        setCopySuccess(false);
        copyTimerRef.current = null;
      }, 2000);
    }
  };

  // Cleanup таймера при unmount
  useEffect(() => {
    return () => {
      if (copyTimerRef.current) {
        clearTimeout(copyTimerRef.current);
      }
    };
  }, []);

  if (loading && !linkStatus) {
    return (
      <div className="telegram-link-card">
        <div className="telegram-link-loading">
          <Spinner />
        </div>
      </div>
    );
  }

  const isLinked = linkStatus?.linked;

  return (
    <div className="telegram-link-card">
      <div className="telegram-link-header">
        <div className="telegram-link-title-wrapper">
          <svg className="telegram-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2ZM16.64 8.8C16.49 10.38 15.84 14.22 15.51 15.99C15.37 16.74 15.09 16.99 14.83 17.01C14.25 17.07 13.81 16.64 13.25 16.27C12.37 15.69 11.87 15.33 11.02 14.77C10.03 14.12 10.67 13.76 11.24 13.18C11.39 13.03 13.95 10.7 14 10.49C14.0069 10.4582 14.006 10.4252 13.9973 10.3938C13.9886 10.3624 13.9724 10.3337 13.95 10.31C13.89 10.26 13.81 10.28 13.74 10.29C13.65 10.31 12.25 11.24 9.52 13.08C9.12 13.35 8.76 13.49 8.44 13.48C8.08 13.47 7.4 13.28 6.89 13.11C6.26 12.91 5.77 12.8 5.81 12.45C5.83 12.27 6.08 12.09 6.55 11.9C9.47 10.63 11.41 9.79 12.38 9.39C15.16 8.23 15.73 8.03 16.11 8.03C16.19 8.03 16.38 8.05 16.5 8.15C16.6 8.23 16.63 8.34 16.64 8.42C16.63 8.48 16.65 8.66 16.64 8.8Z"
              fill="currentColor"
            />
          </svg>
          <h3 className="telegram-link-title">Telegram Уведомления</h3>
        </div>
        <div className={`telegram-status ${isLinked ? 'telegram-status-linked' : 'telegram-status-not-linked'}`}>
          {isLinked ? 'Привязан' : 'Не привязан'}
        </div>
      </div>

      {(error || localError) && (
        <div className="telegram-error-message">
          <svg className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
              fill="currentColor"
            />
          </svg>
          <span>{error || localError}</span>
        </div>
      )}

      <div className="telegram-link-body">
        {isLinked ? (
          <div className="telegram-linked-info">
            <p className="telegram-info-text">
              Ваш Telegram аккаунт привязан. Вы будете получать уведомления о занятиях и других важных событиях.
            </p>
            {linkStatus.telegram && (
              <div className="telegram-user-info">
                <div className="telegram-user-detail">
                  <span className="telegram-detail-label">Telegram ID:</span>
                  <span className="telegram-detail-value">{linkStatus.telegram.telegram_id}</span>
                </div>
                {linkStatus.telegram.username && (
                  <div className="telegram-user-detail">
                    <span className="telegram-detail-label">Username:</span>
                    <span className="telegram-detail-value">@{linkStatus.telegram.username}</span>
                  </div>
                )}
                {(linkStatus.telegram.first_name || linkStatus.telegram.last_name) && (
                  <div className="telegram-user-detail">
                    <span className="telegram-detail-label">Имя:</span>
                    <span className="telegram-detail-value">
                      {linkStatus.telegram.first_name} {linkStatus.telegram.last_name}
                    </span>
                  </div>
                )}
              </div>
            )}
            <Button variant="danger" size="small" onClick={() => setShowUnlinkConfirm(true)} disabled={loading}>
              Отвязать Telegram
            </Button>
          </div>
        ) : (
          <div className="telegram-not-linked-info">
            <p className="telegram-info-text">
              Привяжите ваш Telegram аккаунт для получения уведомлений о занятиях, изменениях расписания и других важных событиях.
            </p>
            <Button onClick={handleOpenLinkModal} loading={loading} disabled={loading}>
              Привязать Telegram
            </Button>
          </div>
        )}
      </div>

      {/* Модальное окно привязки */}
      <Modal
        isOpen={showLinkModal}
        onClose={handleCloseLinkModal}
        title="Привязка Telegram аккаунта"
      >
        <div className="telegram-link-instructions">
          <div className="telegram-step">
            <div className="telegram-step-number">1</div>
            <div className="telegram-step-content">
              <h4 className="telegram-step-title">Откройте бота в Telegram</h4>
              <p className="telegram-step-description">
                {isSafariOnIOS()
                  ? 'Нажмите на кнопку ниже, чтобы открыть бот в Telegram. Вы будете перенаправлены автоматически.'
                  : 'Нажмите на кнопку ниже или скопируйте ссылку и откройте её в Telegram'
                }
              </p>
              {localError && (
                <div className="telegram-modal-error">
                  {localError}
                </div>
              )}
              {linkUrl && (
                <div className="telegram-link-url-wrapper">
                  <input
                    type="text"
                    className="telegram-link-url"
                    value={linkUrl}
                    readOnly
                    aria-label="Ссылка на бота"
                  />
                  <Button size="small" variant="outline" onClick={handleCopyLink} title={copySuccess ? 'Скопировано!' : 'Копировать'}>
                    {copySuccess ? 'Скопировано!' : 'Копировать'}
                  </Button>
                </div>
              )}
              {linkUrl && (
                isSafariOnIOS() ? (
                  <Button
                    variant="primary"
                    fullWidth
                    onClick={() => {
                      window.location.href = linkUrl;
                    }}
                  >
                    Открыть бота в Telegram
                  </Button>
                ) : (
                  <a
                    href={linkUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="telegram-open-bot-link"
                  >
                    <Button variant="primary" fullWidth>
                      Открыть бота в Telegram
                    </Button>
                  </a>
                )
              )}
            </div>
          </div>

          <div className="telegram-step">
            <div className="telegram-step-number">2</div>
            <div className="telegram-step-content">
              <h4 className="telegram-step-title">Нажмите кнопку &quot;Старт&quot;</h4>
              <p className="telegram-step-description">
                В открывшемся чате с ботом нажмите кнопку &quot;Старт&quot; или отправьте команду /start
              </p>
            </div>
          </div>

          <div className="telegram-step">
            <div className="telegram-step-number">3</div>
            <div className="telegram-step-content">
              <h4 className="telegram-step-title">Подтверждение привязки</h4>
              <p className="telegram-step-description">
                Бот автоматически привяжет ваш аккаунт. После этого вы можете закрыть это окно.
              </p>
            </div>
          </div>

          <div className="telegram-link-footer">
            <Button variant="secondary" onClick={handleCloseLinkModal}>
              Готово
            </Button>
          </div>
        </div>
      </Modal>

      {/* Модальное окно подтверждения отвязки */}
      <ConfirmModal
        isOpen={showUnlinkConfirm}
        onClose={() => setShowUnlinkConfirm(false)}
        onConfirm={handleUnlink}
        title="Отвязка Telegram"
        message="Вы действительно хотите отвязать Telegram? Вы перестанете получать уведомления."
        confirmText="Отвязать"
        confirmVariant="danger"
      />
    </div>
  );
};

export default TelegramLinkCard;
