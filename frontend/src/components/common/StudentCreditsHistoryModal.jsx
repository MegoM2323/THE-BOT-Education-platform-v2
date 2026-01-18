import { useState, useEffect } from 'react';
import { useStudentCreditsHistory } from '../../hooks/useStudentCreditsHistory.js';
import Modal from './Modal.jsx';
import Spinner from './Spinner.jsx';
import { formatDate, formatTime } from '../../utils/dateFormat.js';
import { normalizeOperationType, translateReason, OPERATION_TYPE_LABELS, getOperationTypeLabel, getTransactionSignByType, isPositiveOperation } from '../../utils/creditsFormatting.js';
import '../student/CreditsHistory.css';

export const StudentCreditsHistoryModal = ({ isOpen, onClose, student }) => {
  const [operationTypeFilter, setOperationTypeFilter] = useState('');
  const { history, historyLoading, historyError, fetchHistory } = useStudentCreditsHistory(
    student?.id,
  );

  // Сбрасываем фильтр при изменении студента
  useEffect(() => {
    setOperationTypeFilter('');
  }, [student?.id]);

  if (!isOpen || !student) {
    return null;
  }

  const filteredHistory = operationTypeFilter
    ? history.filter((transaction) => normalizeOperationType(transaction.operation_type) === operationTypeFilter)
    : history;

  const handleRetry = () => {
    fetchHistory();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`История кредитов: ${student.full_name}`}
      loading={historyLoading}
    >
      <div className="credits-history">
        <div className="credits-balance-display">
          <span className="balance-label">Текущий баланс:</span>
          <span className="balance-value">
            {student.credits !== undefined ? `${student.credits} кредитов` : 'Загрузка...'}
          </span>
        </div>

        {historyError && (
          <div className="credits-history-error" role="alert">
            Ошибка загрузки истории: {historyError.userMessage || historyError.message}
            <button
              className="error-retry-btn"
              onClick={handleRetry}
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

        {historyLoading && !history.length && (
          <div className="credits-history-loading">
            <Spinner />
          </div>
        )}

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
    </Modal>
  );
};

export default StudentCreditsHistoryModal;
