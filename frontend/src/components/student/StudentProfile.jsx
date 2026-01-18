import { useAuth } from '../../hooks/useAuth.js';
import { useCredits } from '../../hooks/useCredits.js';
import TelegramInlineRow from '../common/TelegramInlineRow.jsx';
import ChangePasswordForm from '../common/ChangePasswordForm.jsx';
import Spinner from '../common/Spinner.jsx';
import './StudentProfile.css';

/**
 * Компонент профиля студента
 */
export const StudentProfile = () => {
  const { user, loading: userLoading } = useAuth();
  const { balance, loading: creditsLoading, error: creditsError } = useCredits();

  return (
    <div className="StudentProfile student-profile">
      <div className="profile-header">
        <h2 className="profile-title">Мой профиль</h2>
      </div>

      <div className="profile-content">
        <div className="profile-section">
          <h3 className="section-title">Личная информация</h3>
          {userLoading && !user ? (
            <div className="profile-loading">
              <Spinner />
            </div>
          ) : (
            <div className="profile-info-card">
              <div className="info-row">
                <span className="info-label">Имя:</span>
                <span className="info-value">{user?.full_name || 'Не указано'}</span>
              </div>
              <div className="info-row">
                <span className="info-label">Email:</span>
                <span className="info-value">{user?.email || 'Не указан'}</span>
              </div>
              <div className="info-row">
                <span className="info-label">Роль:</span>
                <span className="info-value">Студент</span>
              </div>
              <div className="info-row">
                <span className="info-label">Telegram:</span>
                <TelegramInlineRow />
              </div>
            </div>
          )}
        </div>

        <div className="profile-section">
          <h3 className="section-title">Баланс кредитов</h3>
          <div className="profile-info-card">
            <div className="info-row">
              <span className="info-label">Доступные кредиты:</span>
              <span className="info-value" data-testid="profile-credits-balance">
                {creditsLoading && balance === null ? (
                  <Spinner size="sm" />
                ) : creditsError ? (
                  <span className="error-text" title={creditsError}>
                    Ошибка загрузки
                  </span>
                ) : (
                  `${balance ?? 0} кредитов`
                )}
              </span>
            </div>
          </div>
        </div>

        <div className="profile-section">
          <h3 className="section-title">Безопасность</h3>
          <div className="profile-info-card">
            <ChangePasswordForm />
          </div>
        </div>
      </div>
    </div>
  );
};

export default StudentProfile;
