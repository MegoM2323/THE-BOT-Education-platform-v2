import { useState, useMemo } from 'react';
import { usePaymentSettings } from '../../hooks/usePaymentSettings.js';
import Spinner from '../common/Spinner.jsx';
import Input from '../common/Input.jsx';
import './PaymentSettings.css';

/**
 * Форматирование даты для отображения
 */
const formatDate = (dateString) => {
  if (!dateString) return 'Неизвестно';
  const date = new Date(dateString);
  return date.toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

/**
 * Toggle Switch Component
 */
const ToggleSwitch = ({ enabled, onChange, disabled }) => {
  return (
    <button
      className={`toggle-switch ${enabled ? 'enabled' : 'disabled'}`}
      onClick={() => onChange(!enabled)}
      disabled={disabled}
      aria-label={enabled ? 'Отключить платежи' : 'Включить платежи'}
      data-testid="payment-toggle"
    >
      <span className="toggle-slider"></span>
    </button>
  );
};

/**
 * PaymentSettings - Компонент для управления платежами студентов (admin only)
 */
export const PaymentSettings = () => {
  const { students, isLoading, error, updatePaymentStatus, isUpdating } = usePaymentSettings();
  const [searchText, setSearchText] = useState('');
  const [sortBy, setSortBy] = useState('name'); // 'name' | 'email' | 'status' | 'updated'

  // Filter and sort students
  const filteredStudents = useMemo(() => {
    if (!Array.isArray(students)) return [];

    let filtered = [...students];

    // Apply search filter
    if (searchText.trim()) {
      const search = searchText.toLowerCase().trim();
      filtered = filtered.filter(
        (student) =>
          student.full_name?.toLowerCase().includes(search) ||
          student.email?.toLowerCase().includes(search)
      );
    }

    // Apply sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return (a.full_name || '').localeCompare(b.full_name || '', 'ru-RU');
        case 'email':
          return (a.email || '').localeCompare(b.email || '', 'ru-RU');
        case 'status':
          // Sort by payment_enabled: enabled first
          return (b.payment_enabled === true ? 1 : 0) - (a.payment_enabled === true ? 1 : 0);
        case 'updated':
          // Sort by updated_at: newest first
          const dateA = new Date(a.updated_at || 0);
          const dateB = new Date(b.updated_at || 0);
          return dateB - dateA;
        default:
          return 0;
      }
    });

    return filtered;
  }, [students, searchText, sortBy]);

  const handleToggle = async (userId, currentStatus) => {
    try {
      await updatePaymentStatus(userId, !currentStatus);
    } catch (error) {
      console.error('Toggle payment status failed:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="payment-settings-loading">
        <Spinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="payment-settings-error" role="alert">
        <p>Ошибка загрузки настроек платежей: {error.message || error}</p>
      </div>
    );
  }

  return (
    <div className="payment-settings" data-testid="payment-settings">
      <div className="payment-settings-header">
        <h2 className="payment-settings-title">Управление платежами</h2>
        <p className="payment-settings-description">
          Включите или отключите возможность совершать платежи для каждого студента
        </p>
      </div>

      {/* Search and Sort Controls */}
      <div className="payment-settings-controls">
        <div className="search-box">
          <Input
            type="text"
            placeholder="Поиск по имени или email"
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            data-testid="search-input"
          />
        </div>

        <div className="sort-box">
          <label htmlFor="sort-select" className="sort-label">
            Сортировка:
          </label>
          <select
            id="sort-select"
            className="sort-select"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            data-testid="sort-select"
          >
            <option value="name">По имени</option>
            <option value="email">По email</option>
            <option value="status">По статусу</option>
            <option value="updated">По дате изменения</option>
          </select>
        </div>
      </div>

      {/* Students Table */}
      {filteredStudents.length === 0 ? (
        <div className="empty-state">
          <p>Студенты не найдены</p>
        </div>
      ) : (
        <div className="payment-settings-table-wrapper">
          <table className="payment-settings-table" data-testid="students-table">
            <thead>
              <tr>
                <th>Студент</th>
                <th>Email</th>
                <th>Платежи</th>
                <th>Изменено</th>
              </tr>
            </thead>
            <tbody>
              {filteredStudents.map((student) => (
                <tr key={student.id} data-testid={`student-row-${student.id}`}>
                  <td className="student-name">{student.full_name || 'Неизвестно'}</td>
                  <td className="student-email">{student.email || 'Неизвестно'}</td>
                  <td className="student-payment-status">
                    <div className="payment-toggle-wrapper">
                      <ToggleSwitch
                        enabled={student.payment_enabled === true}
                        onChange={(newStatus) => handleToggle(student.id, student.payment_enabled)}
                        disabled={isUpdating}
                      />
                      <span
                        className={`status-label ${student.payment_enabled ? 'enabled' : 'disabled'}`}
                      >
                        {student.payment_enabled ? 'Включены' : 'Отключены'}
                      </span>
                    </div>
                  </td>
                  <td className="student-updated">{formatDate(student.updated_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Footer info */}
      <div className="payment-settings-footer">
        <p className="total-count">
          Всего студентов: {filteredStudents.length}
          {searchText && ` (из ${students.length})`}
        </p>
      </div>
    </div>
  );
};

export default PaymentSettings;
