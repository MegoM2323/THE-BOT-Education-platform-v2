import React from 'react';
import { logger } from '../../utils/logger.js';
import { useNavigate } from 'react-router-dom';
import './StudentErrorBoundary.css';

/**
 * Error Boundary компонент для перехвата и отображения ошибок рендера
 * в компонентах студента
 *
 * Использование:
 * <StudentErrorBoundary>
 *   <MyComponent />
 * </StudentErrorBoundary>
 */
class StudentErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorTimestamp: null,
    };
  }

  /**
   * Вызывается когда был выброшен exception
   * Возвращает новое состояние для запуска fallback UI
   */
  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  /**
   * Логирует детали ошибки после того как она была выброшена
   */
  componentDidCatch(error, errorInfo) {
    const timestamp = new Date().toISOString();

    // Сохраняем состояние для рендера
    this.setState({
      error,
      errorInfo,
      errorTimestamp: timestamp,
    });

    // Логируем полную информацию об ошибке в консоль
    console.error('=== СТУДЕНТ КОМПОНЕНТ ОШИБКА ===');
    console.error('Время ошибки:', timestamp);
    console.error('Тип ошибки:', error.name);
    console.error('Сообщение об ошибке:', error.message);
    console.error('Stack trace:', error.stack);
    console.error('Информация о компоненте:', errorInfo.componentStack);
    console.error('=====================================');

    // Отправляем ошибку на сервер для мониторинга (если необходимо)
    this.logErrorToServer(error, errorInfo, timestamp);
  }

  /**
   * Логирует ошибку на сервер
   * (Может быть интегрирован с сервисом мониторинга ошибок)
   */
  logErrorToServer = (error, errorInfo, timestamp) => {
    try {
      // Здесь можно добавить отправку на сервер
      // Пример для сервиса типа Sentry или собственного эндпоинта:
      /*
      fetch('/api/errors/log', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          timestamp,
          errorName: error.name,
          errorMessage: error.message,
          stack: error.stack,
          componentStack: errorInfo.componentStack,
          userAgent: navigator.userAgent,
          url: window.location.href,
        }),
      }).catch(err => console.error('Ошибка при отправке логов на сервер:', err));
      */
    } catch (err) {
      console.error('Ошибка при логировании на сервер:', err);
    }
  };

  /**
   * Сбрасывает состояние ошибки
   */
  resetError = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorTimestamp: null,
    });
  };

  /**
   * Обновляет страницу
   */
  handleRefresh = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="student-error-boundary-container">
          <div className="student-error-boundary-content">
            <div className="student-error-header">
              <div className="student-error-icon">⚠️</div>
              <h1 className="student-error-title">Произошла ошибка</h1>
              <p className="student-error-subtitle">
                При загрузке компонента произошла непредвиденная ошибка
              </p>
            </div>

            <div className="student-error-details">
              <div className="student-error-section">
                <h2 className="student-error-section-title">Тип ошибки</h2>
                <code className="student-error-code">
                  {this.state.error?.name || 'Error'}
                </code>
              </div>

              <div className="student-error-section">
                <h2 className="student-error-section-title">Сообщение об ошибке</h2>
                <code className="student-error-code">
                  {this.state.error?.message || 'Неизвестная ошибка'}
                </code>
              </div>

              {/* Stack trace видна только в development режиме */}
              {process.env.NODE_ENV === 'development' && (
                <>
                  {this.state.errorInfo && (
                    <div className="student-error-section">
                      <h2 className="student-error-section-title">Stack Trace</h2>
                      <pre className="student-error-stack">
                        {this.state.error?.stack}
                      </pre>
                    </div>
                  )}

                  {this.state.errorInfo?.componentStack && (
                    <div className="student-error-section">
                      <h2 className="student-error-section-title">Component Stack</h2>
                      <pre className="student-error-stack">
                        {this.state.errorInfo.componentStack}
                      </pre>
                    </div>
                  )}

                  {this.state.errorTimestamp && (
                    <div className="student-error-section">
                      <h2 className="student-error-section-title">Время ошибки</h2>
                      <code className="student-error-code">
                        {this.state.errorTimestamp}
                      </code>
                    </div>
                  )}
                </>
              )}
            </div>

            <div className="student-error-actions">
              <button
                onClick={this.handleRefresh}
                className="student-error-button student-error-button-primary"
                type="button"
              >
                🔄 Обновить страницу
              </button>
              <StudentErrorBoundaryNavigateButton
                onNavigate={this.resetError}
              />
            </div>

            {process.env.NODE_ENV === 'development' && (
              <div className="student-error-footer">
                <p className="student-error-footer-text">
                  Эта информация видна только в режиме разработки.
                </p>
              </div>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * Вспомогательный компонент для кнопки навигации
 * Используется внутри Error Boundary для доступа к useNavigate hook
 */
function StudentErrorBoundaryNavigateButton({ onNavigate }) {
  const navigate = useNavigate();

  const handleNavigateHome = () => {
    onNavigate();
    navigate('/');
  };

  return (
    <button
      onClick={handleNavigateHome}
      className="student-error-button student-error-button-secondary"
      type="button"
    >
      🏠 Вернуться на главную
    </button>
  );
}

export default StudentErrorBoundary;
