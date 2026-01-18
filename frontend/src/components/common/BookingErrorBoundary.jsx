import React from 'react';
import { logger } from '../../utils/logger.js';
import './BookingErrorBoundary.css';

/**
 * Error Boundary компонент для обработки ошибок в компонентах данных бронирований
 * Предотвращает полный крах приложения при ошибках загрузки данных
 */
class BookingErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      retryCount: 0,
    };
  }

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(caughtError, errorInfo) {
    // Логируем ошибку для отладки
    console.error('BookingErrorBoundary caught an error:', caughtError);
    console.error('Error info:', errorInfo);

    this.setState({
      error: caughtError,
      errorInfo,
    });
  }

  handleRetry = () => {
    this.setState((prevState) => ({
      hasError: false,
      error: null,
      errorInfo: null,
      retryCount: prevState.retryCount + 1,
    }));

    // Вызовем callback если он предоставлен
    if (this.props.onRetry) {
      this.props.onRetry();
    }
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="booking-error-boundary">
          <div className="booking-error-container">
            <div className="booking-error-icon">⚠️</div>
            <h2 className="booking-error-title">Ошибка загрузки данных</h2>

            <p className="booking-error-message">
              При загрузке информации о бронированиях возникла ошибка. Пожалуйста, попробуйте еще раз.
            </p>

            {this.state.error && (
              <details className="booking-error-details">
                <summary>Подробности ошибки</summary>
                <pre className="booking-error-stack">
                  {this.state.error.toString()}
                  {this.state.errorInfo && this.state.errorInfo.componentStack}
                </pre>
              </details>
            )}

            <div className="booking-error-actions">
              <button
                className="booking-error-retry-button"
                onClick={this.handleRetry}
                aria-label="Попробовать еще раз"
              >
                Попробовать еще раз
              </button>

              <a
                href="/"
                className="booking-error-home-button"
                aria-label="Вернуться на главную"
              >
                На главную
              </a>
            </div>

            {this.state.retryCount > 2 && (
              <p className="booking-error-contact">
                Если проблема продолжается, пожалуйста, свяжитесь с поддержкой.
              </p>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

BookingErrorBoundary.defaultProps = {
  onRetry: null,
};

export default BookingErrorBoundary;
