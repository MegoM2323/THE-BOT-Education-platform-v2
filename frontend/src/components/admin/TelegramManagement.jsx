import { useState, useEffect } from 'react';
import TelegramUsersTable from './TelegramUsersTable.jsx';
import BroadcastListManager from './BroadcastListManager.jsx';
import BroadcastComposer from './BroadcastComposer.jsx';
import BroadcastHistory from './BroadcastHistory.jsx';
import { useBroadcast } from '../../hooks/useBroadcast.js';
import './TelegramManagement.css';

/**
 * Главный компонент управления Telegram рассылками
 * Объединяет все функции управления Telegram в одном интерфейсе с табами
 */
export const TelegramManagement = () => {
  const [activeTab, setActiveTab] = useState('users');
  const { linkedUsers, loading, error, fetchLinkedUsers } = useBroadcast();
  const [loadError, setLoadError] = useState(null);

  // Загрузить пользователей сразу при монтировании компонента
  // Независимо от активного таба - данные всегда готовы
  useEffect(() => {
    const loadUsers = async () => {
      try {
        await fetchLinkedUsers();
      } catch (err) {
        const errorMessage = err.message || 'Ошибка загрузки пользователей';
        setLoadError(errorMessage);
        console.error('Ошибка загрузки Telegram пользователей:', err);
      }
    };

    loadUsers();
  }, [fetchLinkedUsers]);

  const tabs = [
    {
      id: 'users',
      label: 'Пользователи',
      icon: (
        <svg className="tab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
    },
    {
      id: 'lists',
      label: 'Списки рассылки',
      icon: (
        <svg className="tab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
        </svg>
      ),
    },
    {
      id: 'compose',
      label: 'Новая рассылка',
      icon: (
        <svg className="tab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
        </svg>
      ),
    },
    {
      id: 'history',
      label: 'История',
      icon: (
        <svg className="tab-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      ),
    },
  ];

  const handleTabChange = (tabId) => {
    setActiveTab(tabId);
    setLoadError(null);
    // Загрузка данных происходит автоматически через useEffect
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'users':
        return (
          <div className="tab-content">
            <div className="tab-header">
              <h3 className="tab-title">Пользователи с Telegram</h3>
              <p className="tab-description">
                Просмотр всех пользователей, которые привязали свой Telegram аккаунт к платформе
              </p>
            </div>
            {(loadError || error) && (
              <div className="tab-error-message">
                <svg className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path
                    d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
                    fill="currentColor"
                  />
                </svg>
                <span>{loadError || error}</span>
              </div>
            )}
            <TelegramUsersTable
              users={linkedUsers}
              loading={loading}
              showCheckboxes={false}
            />
          </div>
        );

      case 'lists':
        return <BroadcastListManager />;

      case 'compose':
        return <BroadcastComposer />;

      case 'history':
        return <BroadcastHistory />;

      default:
        return null;
    }
  };

  return (
    <div className="telegram-management">
      <div className="telegram-management-header">
        <div className="header-content">
          <div className="header-title-wrapper">
            <svg className="telegram-header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2ZM16.64 8.8C16.49 10.38 15.84 14.22 15.51 15.99C15.37 16.74 15.09 16.99 14.83 17.01C14.25 17.07 13.81 16.64 13.25 16.27C12.37 15.69 11.87 15.33 11.02 14.77C10.03 14.12 10.67 13.76 11.24 13.18C11.39 13.03 13.95 10.7 14 10.49C14.0069 10.4582 14.006 10.4252 13.9973 10.3938C13.9886 10.3624 13.9724 10.3337 13.95 10.31C13.89 10.26 13.81 10.28 13.74 10.29C13.65 10.31 12.25 11.24 9.52 13.08C9.12 13.35 8.76 13.49 8.44 13.48C8.08 13.47 7.4 13.28 6.89 13.11C6.26 12.91 5.77 12.8 5.81 12.45C5.83 12.27 6.08 12.09 6.55 11.9C9.47 10.63 11.41 9.79 12.38 9.39C15.16 8.23 15.73 8.03 16.11 8.03C16.19 8.03 16.38 8.05 16.5 8.15C16.6 8.23 16.63 8.34 16.64 8.42C16.63 8.48 16.65 8.66 16.64 8.8Z"
                fill="currentColor"
              />
            </svg>
            <div>
              <h2 className="page-title">Управление Telegram</h2>
              <p className="page-subtitle">Рассылки и уведомления через Telegram</p>
            </div>
          </div>
        </div>
      </div>

      <div className="telegram-management-tabs">
        <div className="tabs-nav">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => handleTabChange(tab.id)}
              aria-label={tab.label}
            >
              {tab.icon}
              <span className="tab-label">{tab.label}</span>
            </button>
          ))}
        </div>

        <div className="tabs-content">
          {renderTabContent()}
        </div>
      </div>
    </div>
  );
};

export default TelegramManagement;
