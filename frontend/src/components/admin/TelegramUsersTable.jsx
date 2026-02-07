import { useState, useEffect } from 'react';
import Spinner from '../common/Spinner.jsx';
import './TelegramUsersTable.css';

/**
 * Таблица пользователей с привязанным Telegram
 * Используется для выбора получателей рассылки
 */
export const TelegramUsersTable = ({
  users = [],
  loading = false,
  selectedUsers = [],
  onSelectionChange,
  showCheckboxes = true
}) => {
  const [roleFilter, setRoleFilter] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [filteredUsers, setFilteredUsers] = useState([]);
  const [selectAll, setSelectAll] = useState(false);

  // Фильтрация пользователей
  useEffect(() => {
    let result = users;

    // Фильтр по роли
    if (roleFilter) {
      result = result.filter(user => user.role === roleFilter);
    }

    // Поиск по имени или email
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter(user =>
        user.full_name?.toLowerCase().includes(query) ||
        user.email?.toLowerCase().includes(query) ||
        user.telegram?.username?.toLowerCase().includes(query)
      );
    }

    setFilteredUsers(result);
  }, [users, roleFilter, searchQuery]);

  // Обновление состояния "выбрать все" когда меняется список выбранных
  useEffect(() => {
    if (filteredUsers.length > 0) {
      const allSelected = filteredUsers.every(user =>
        selectedUsers.includes(user.id)
      );
      setSelectAll(allSelected);
    }
  }, [selectedUsers, filteredUsers]);

  const handleSelectAll = (e) => {
    const checked = e.target.checked;
    setSelectAll(checked);

    if (checked) {
      // Добавить всех отфильтрованных пользователей
      const allIds = [...new Set([...selectedUsers, ...filteredUsers.map(u => u.id)])];
      onSelectionChange(allIds);
    } else {
      // Убрать всех отфильтрованных пользователей
      const filteredIds = filteredUsers.map(u => u.id);
      const remaining = selectedUsers.filter(id => !filteredIds.includes(id));
      onSelectionChange(remaining);
    }
  };

  const handleSelectUser = (userId) => {
    if (selectedUsers.includes(userId)) {
      onSelectionChange(selectedUsers.filter(id => id !== userId));
    } else {
      onSelectionChange([...selectedUsers, userId]);
    }
  };

  const getRoleLabel = (role) => {
    const labels = {
      student: 'Студент',
      teacher: 'Преподаватель',
      admin: 'Администратор',
    };
    return labels[role] || role;
  };

  const formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="telegram-users-table-loading">
        <Spinner />
      </div>
    );
  }

  if (!users || users.length === 0) {
    return (
      <div className="telegram-users-table-container">
        <div className="telegram-users-empty-state">
          <svg className="empty-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
              fill="currentColor"
            />
          </svg>
          <h3 className="empty-title">Нет пользователей с Telegram</h3>
          <p className="empty-description">
            Пока нет пользователей, которые привязали свой Telegram аккаунт.
            Пользователи могут привязать Telegram через свой профиль.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="telegram-users-table-container">
      <div className="telegram-users-filters">
        <div className="filter-group">
          <label className="filter-label" htmlFor="role-filter">Роль:</label>
          <select
            id="role-filter"
            className="filter-select"
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
          >
            <option value="">Все</option>
            <option value="student">Студенты</option>
            <option value="teacher">Преподаватели</option>
            <option value="admin">Администраторы</option>
          </select>
        </div>

        <div className="filter-group">
          <label className="filter-label" htmlFor="search-filter">Поиск:</label>
          <input
            id="search-filter"
            type="text"
            className="filter-input"
            placeholder="Имя, email или username..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {showCheckboxes && (
          <div className="selected-count">
            Выбрано: {selectedUsers.length}
          </div>
        )}
      </div>

      <div className="telegram-users-table-wrapper">
        <table className="telegram-users-table">
          <thead>
            <tr>
              {showCheckboxes && (
                <th className="checkbox-column">
                  <input
                    type="checkbox"
                    checked={selectAll}
                    onChange={handleSelectAll}
                    disabled={filteredUsers.length === 0}
                    aria-label="Выбрать всех"
                  />
                </th>
              )}
              <th>ФИО</th>
              <th>Email</th>
              <th>Роль</th>
              <th>Telegram Username</th>
              <th>Telegram ID</th>
              <th>Дата привязки</th>
            </tr>
          </thead>
          <tbody>
            {filteredUsers.length === 0 ? (
              <tr>
                <td colSpan={showCheckboxes ? 7 : 6} className="empty-row">
                  {searchQuery || roleFilter
                    ? 'Пользователи не найдены'
                    : 'Нет пользователей с привязанным Telegram'}
                </td>
              </tr>
            ) : (
              filteredUsers.map((user) => (
                <tr
                  key={user.id}
                  className={selectedUsers.includes(user.id) ? 'selected-row' : ''}
                  onClick={() => showCheckboxes && handleSelectUser(user.id)}
                  style={{ cursor: showCheckboxes ? 'pointer' : 'default' }}
                >
                  {showCheckboxes && (
                    <td className="checkbox-column" onClick={(e) => e.stopPropagation()}>
                      <input
                        type="checkbox"
                        checked={selectedUsers.includes(user.id)}
                        onChange={() => handleSelectUser(user.id)}
                        aria-label={`Выбрать ${user.full_name}`}
                      />
                    </td>
                  )}
                  <td>{user.full_name}</td>
                  <td>{user.email}</td>
                  <td>
                    <span className={`role-badge role-${user.role}`}>
                      {getRoleLabel(user.role)}
                    </span>
                  </td>
                  <td>
                    {user.telegram?.username ? (
                      <span className="telegram-username">
                        @{user.telegram.username}
                      </span>
                    ) : (
                      <span className="no-data">Не привязан</span>
                    )}
                  </td>
                  <td>
                    <code className="telegram-id">
                      {user.telegram?.telegram_id && String(user.telegram.telegram_id).trim()
                        ? user.telegram.telegram_id
                        : '-'}
                    </code>
                  </td>
                  <td>{user.telegram?.linked_at ? formatDate(user.telegram.linked_at) : '-'}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default TelegramUsersTable;
