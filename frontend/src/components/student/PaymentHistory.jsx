import { usePayments } from '../../hooks/usePayments.js';
import Spinner from '../common/Spinner.jsx';
import { formatDate, formatTime } from '../../utils/dateFormat.js';
import './PaymentHistory.css';

/**
 * История платежей пользователя
 */
export const PaymentHistory = () => {
  const { payments, isLoadingHistory, historyError } = usePayments();

  const getStatusLabel = (status) => {
    const labels = {
      pending: 'Ожидание',
      succeeded: 'Успешно',
      cancelled: 'Отменен',
      failed: 'Ошибка',
    };
    return labels[status] || status;
  };

  const getStatusClass = (status) => {
    return `status-${status}`;
  };

  if (isLoadingHistory) {
    return (
      <div className="payment-history-loading">
        <Spinner />
      </div>
    );
  }

  if (historyError) {
    return (
      <div className="payment-history-error" role="alert">
        <p>Ошибка загрузки истории платежей</p>
      </div>
    );
  }

  if (!payments || payments.length === 0) {
    return (
      <div className="payment-history-empty">
        <p>История платежей пуста</p>
      </div>
    );
  }

  return (
    <div className="payment-history" data-testid="payment-history">
      <h3 className="payment-history-title">История платежей</h3>

      <div className="payment-history-table-wrapper">
        <table className="payment-history-table">
          <thead>
            <tr>
              <th>Дата</th>
              <th>Сумма</th>
              <th>Кредиты</th>
              <th>Статус</th>
            </tr>
          </thead>
          <tbody>
            {payments.map((payment) => (
              <tr key={payment.id} data-testid="payment-row">
                <td>
                  <div className="payment-date">
                    {formatDate(payment.created_at)}
                  </div>
                  <div className="payment-time">
                    {formatTime(payment.created_at)}
                  </div>
                </td>
                <td className="payment-amount">
                  {payment.amount.toLocaleString('ru-RU')} ₽
                </td>
                <td className="payment-credits">
                  {payment.credits} {payment.credits === 1 ? 'кредит' : 'кредитов'}
                </td>
                <td>
                  <span className={`payment-status ${getStatusClass(payment.status)}`}>
                    {getStatusLabel(payment.status)}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default PaymentHistory;
