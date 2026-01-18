import { useState, useEffect, useRef } from 'react';
import { useCredits } from '../../hooks/useCredits.js';
import Spinner from '../common/Spinner.jsx';
import { formatDate, formatTime } from '../../utils/dateFormat.js';
import { normalizeOperationType, translateReason, OPERATION_TYPE_LABELS, getOperationTypeLabel, getTransactionSignByType, isPositiveOperation } from '../../utils/creditsFormatting.js';
import './CreditsHistory.css';

export const CreditsHistory = () => {
  const { balance, history, historyLoading, historyError, fetchHistory } = useCredits();
  const [operationTypeFilter, setOperationTypeFilter] = useState('');
  const [isFirstLoad, setIsFirstLoad] = useState(true);

  // Сохраняем предыдущее значение баланса для отслеживания изменений
  const prevBalanceRef = useRef(balance);

  // Загрузка истории при монтировании
  useEffect(() => {
    const loadHistory = async () => {
      await fetchHistory();
      setIsFirstLoad(false);
    };
    loadHistory();
  }, [fetchHistory]);

  // Перезагрузка истории при изменении баланса (после add/deduct операций)
  useEffect(() => {
    // Пропускаем первую загрузку и случаи когда баланс еще не определен
    if (isFirstLoad || balance === null || prevBalanceRef.current === null) {
      prevBalanceRef.current = balance;
      return;
    }

    // Если баланс изменился - перезагружаем историю
    if (prevBalanceRef.current !== balance) {
      fetchHistory();
      prevBalanceRef.current = balance;
    }
  }, [balance, isFirstLoad, fetchHistory]);

  const filteredHistory = operationTypeFilter
    ? history.filter((transaction) => normalizeOperationType(transaction.operation_type) === operationTypeFilter)
    : history;

  if (isFirstLoad && historyLoading) {
    return (
      <div className="credits-history-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="credits-history" data-testid="credits-history">
      <div className="credits-history-header">
        <h2 className="credits-history-title">История кредитов</h2>
        <div className="credits-balance-display">
          <span className="balance-label">Текущий баланс:</span>
          <span className="balance-value">
            {balance === null ? 'Загрузка...' : `${balance} кредитов`}
          </span>
        </div>
      </div>

      {historyError && (
        <div className="credits-history-error" role="alert">
          Ошибка загрузки истории: {historyError.userMessage || historyError.message}
          <button
            className="error-retry-btn"
            onClick={() => fetchHistory()}
            disabled={historyLoading}
          >
            {historyLoading ? 'Загрузка...' : 'Повторить'}
          </button>
        </div>
      )}

      <div className="credits-history-filters">
        <label className="filter-label">Тип операции:</label>
        <select
          className="filter-select"
          value={operationTypeFilter}
          onChange={(e) => setOperationTypeFilter(e.target.value)}
          disabled={historyLoading}
        >
          <option value="">Все</option>
          <option value="add">Начисление</option>
          <option value="deduct">Списание</option>
          <option value="refund">Возврат</option>
        </select>
        {historyLoading && <span className="filter-loading-indicator">Обновление...</span>}
      </div>

      {!historyError && filteredHistory.length === 0 && !historyLoading ? (
        <div className="credits-history-empty">
          <p>
            {history.length === 0
              ? 'История операций пуста'
              : 'Нет операций, соответствующих выбранному фильтру'}
          </p>
        </div>
      ) : filteredHistory.length === 0 ? null : (
        <div className="credits-history-table-wrapper">
          <table className="credits-history-table">
            <thead>
              <tr>
                <th>Дата и время</th>
                <th>Операция</th>
                <th>Количество</th>
                <th>Причина</th>
              </tr>
            </thead>
            <tbody>
              {filteredHistory.map((transaction) => (
                <tr key={transaction.id} data-testid="transaction">
                  <td>
                    <div className="transaction-date">
                      {formatDate(transaction.created_at)}
                    </div>
                    <div className="transaction-time">
                      {formatTime(transaction.created_at)}
                    </div>
                  </td>
                  <td>
                    <span className={`transaction-type type-${normalizeOperationType(transaction.operation_type)}`}>
                      {getOperationTypeLabel(transaction.operation_type)}
                    </span>
                  </td>
                  <td>
                    <span className={`transaction-amount ${isPositiveOperation(transaction.operation_type) ? 'positive' : 'negative'}`}>
                      {getTransactionSignByType(transaction.operation_type)}{Math.abs(transaction.amount)}
                    </span>
                  </td>
                  <td className="transaction-description" title={transaction.reason || ''}>
                    {translateReason(transaction.reason)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default CreditsHistory;
