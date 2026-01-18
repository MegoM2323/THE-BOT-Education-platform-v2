import { Component } from 'react';
import { logger } from '../../utils/logger.js';

/**
 * Универсальный Error Boundary компонент для перехвата ошибок рендера
 *
 * Features:
 * - Ловит ошибки в дочерних компонентах
 * - Показывает fallback UI с деталями ошибки
 * - Позволяет повторить попытку (retry)
 * - Логирует ошибки в консоль (можно расширить до отправки на сервер)
 * - Подсчитывает количество попыток для предотвращения бесконечных циклов
 *
 * Usage:
 * <ErrorBoundary fallback={<CustomFallback />}>
 *   <YourComponent />
 * </ErrorBoundary>
 */
class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      retryCount: 0,
      errorTimestamp: null
    };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    const timestamp = new Date().toISOString();

    // Логируем детали ошибки
    console.error('=== ERROR BOUNDARY CAUGHT ERROR ===');
    console.error('Timestamp:', timestamp);
    console.error('Error:', error);
    console.error('Error Info:', errorInfo);
    console.error('Component Stack:', errorInfo.componentStack);
    console.error('===================================');

    this.setState({
      error,
      errorInfo,
      errorTimestamp: timestamp
    });

    // Вызываем callback если предоставлен
    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }

    // Опционально: отправка на сервер для мониторинга
    this.logErrorToServer(error, errorInfo, timestamp);
  }

  /**
   * Отправка ошибки на сервер для мониторинга
   * В будущем можно интегрировать с Sentry, LogRocket и т.д.
   */
  logErrorToServer = (error, errorInfo, timestamp) => {
    try {
      // Placeholder для будущей интеграции с сервисом мониторинга
      // fetch('/api/errors/log', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({
      //     timestamp,
      //     errorName: error.name,
      //     errorMessage: error.message,
      //     stack: error.stack,
      //     componentStack: errorInfo.componentStack,
      //     userAgent: navigator.userAgent,
      //     url: window.location.href,
      //   }),
      // });
    } catch (err) {
      console.error('Failed to log error to server:', err);
    }
  };

  /**
   * Сброс состояния ошибки и повторная попытка рендера
   */
  handleRetry = () => {
    this.setState(prevState => ({
      hasError: false,
      error: null,
      errorInfo: null,
      retryCount: prevState.retryCount + 1,
      errorTimestamp: null
    }));

    // Вызываем callback если предоставлен
    if (this.props.onRetry) {
      this.props.onRetry();
    }
  };

  /**
   * Перезагрузка страницы (для критичных ошибок)
   */
  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      // Если предоставлен кастомный fallback, используем его
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Дефолтный fallback UI
      return (
        <div style={{
          padding: '20px',
          backgroundColor: '#fee',
          border: '1px solid #f99',
          borderRadius: '8px',
          margin: '20px',
          maxWidth: '800px',
          fontFamily: 'system-ui, -apple-system, sans-serif'
        }}>
          <div style={{ marginBottom: '16px' }}>
            <div style={{ fontSize: '48px', marginBottom: '8px' }}>⚠️</div>
            <h3 style={{ margin: '0 0 8px 0', color: '#d32f2f' }}>
              Произошла ошибка
            </h3>
            <p style={{ margin: '0', color: '#666' }}>
              При отображении компонента возникла непредвиденная ошибка
            </p>
          </div>

          <div style={{
            backgroundColor: '#fff',
            border: '1px solid #ddd',
            borderRadius: '4px',
            padding: '12px',
            marginBottom: '16px'
          }}>
            <div style={{ marginBottom: '8px' }}>
              <strong>Тип ошибки:</strong>
              <code style={{
                display: 'block',
                marginTop: '4px',
                padding: '8px',
                backgroundColor: '#f5f5f5',
                borderRadius: '4px',
                fontSize: '14px'
              }}>
                {this.state.error?.name || 'Error'}
              </code>
            </div>

            <div>
              <strong>Сообщение:</strong>
              <code style={{
                display: 'block',
                marginTop: '4px',
                padding: '8px',
                backgroundColor: '#f5f5f5',
                borderRadius: '4px',
                fontSize: '14px'
              }}>
                {this.state.error?.message || 'Неизвестная ошибка'}
              </code>
            </div>
          </div>

          {/* Детали стека только в dev режиме */}
          {process.env.NODE_ENV === 'development' && this.state.errorInfo && (
            <details style={{ marginBottom: '16px' }}>
              <summary style={{ cursor: 'pointer', marginBottom: '8px' }}>
                Подробности (только в dev режиме)
              </summary>
              <pre style={{
                backgroundColor: '#f5f5f5',
                padding: '12px',
                borderRadius: '4px',
                overflow: 'auto',
                fontSize: '12px',
                maxHeight: '200px'
              }}>
                {this.state.error?.stack}
                {'\n\nComponent Stack:'}
                {this.state.errorInfo.componentStack}
              </pre>
            </details>
          )}

          <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
            <button
              onClick={this.handleRetry}
              style={{
                padding: '10px 20px',
                backgroundColor: '#007bff',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '14px',
                fontWeight: '500'
              }}
            >
              🔄 Попробовать заново
            </button>

            <button
              onClick={this.handleReload}
              style={{
                padding: '10px 20px',
                backgroundColor: '#6c757d',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '14px',
                fontWeight: '500'
              }}
            >
              ⟳ Перезагрузить страницу
            </button>

            <a
              href="/"
              style={{
                padding: '10px 20px',
                backgroundColor: '#28a745',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                textDecoration: 'none',
                fontSize: '14px',
                fontWeight: '500',
                display: 'inline-block'
              }}
            >
              🏠 На главную
            </a>
          </div>

          {/* Предупреждение после множественных попыток */}
          {this.state.retryCount > 2 && (
            <div style={{
              marginTop: '16px',
              padding: '12px',
              backgroundColor: '#fff3cd',
              border: '1px solid #ffc107',
              borderRadius: '4px',
              fontSize: '14px'
            }}>
              ⚠️ Проблема сохраняется после {this.state.retryCount} попыток.
              Рекомендуем перезагрузить страницу или обратиться в поддержку.
            </div>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}

ErrorBoundary.defaultProps = {
  fallback: null,
  onError: null,
  onRetry: null
};

export default ErrorBoundary;
