import { useState, useEffect, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import Button from "../common/Button.jsx";
import Input from '../common/Input.jsx';
import Modal from "../common/Modal.jsx";
import ConfirmModal from '../common/ConfirmModal.jsx';
import { SkeletonTable } from '../common/SkeletonLoader.jsx';
import SlowConnectionNotice from '../common/SlowConnectionNotice.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import { useSlowConnection } from '../../hooks/useSlowConnection.js';
import { useAuth } from '../../hooks/useAuth.js';
import * as usersAPI from '../../api/users.js';
import './UserManagement.css';

export const UserManagement = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [roleFilter, setRoleFilter] = useState('');
  const [showUserModal, setShowUserModal] = useState(false);
  const [editingUser, setEditingUser] = useState(null);
  const [deleteUserId, setDeleteUserId] = useState(null);
  const [formData, setFormData] = useState({
    email: '',
    full_name: '',
    password: '',
    role: 'student',
  });
  const [submitting, setSubmitting] = useState(false);
  const [dismissedSlowNotice, setDismissedSlowNotice] = useState(false);
  const notification = useNotification();
  const queryClient = useQueryClient();
  const { user: currentUser } = useAuth();

  // Detect slow connection
  const isSlow = useSlowConnection(loading, 3000);

  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true);
      const filters = roleFilter ? { role: roleFilter } : {};
      const data = await usersAPI.getUsersAll(filters);
      setUsers(data);
    } catch (error) {
      console.error('Failed to load users - Status:', error?.response?.status, 'Message:', error?.message);

      if (error?.response?.status === 403) {
        notification.error('Доступ запрещен. Требуются права администратора.');
      } else if (error?.response?.status === 401) {
        notification.error('Сеанс истекшей. Пожалуйста, авторизуйтесь заново.');
      } else {
        notification.error('Ошибка загрузки пользователей');
      }
    } finally {
      setLoading(false);
    }
  }, [roleFilter, notification]);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const handleOpenCreate = () => {
    setEditingUser(null);
    setFormData({
      email: '',
      full_name: '',
      password: '',
      role: 'student',
    });
    setShowUserModal(true);
  };

  const handleOpenEdit = (user) => {
    setEditingUser(user);
    setFormData({
      email: user.email,
      full_name: user.full_name,
      password: '',
      role: user.role,
    });
    setShowUserModal(true);
  };

  const handleCloseModal = () => {
    setShowUserModal(false);
    setEditingUser(null);
    setFormData({
      email: '',
      full_name: '',
      password: '',
      role: 'student',
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      // Валидация обязательных полей
      if (!formData.email || !formData.email.trim()) {
        throw new Error('Email обязателен');
      }

      // Валидация email формата
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(formData.email.trim())) {
        throw new Error('Некорректный формат email');
      }

      if (!formData.full_name || !formData.full_name.trim()) {
        throw new Error('Полное имя обязательно');
      }

      // Валидация минимальной длины имени (как на backend)
      if (formData.full_name.trim().length < 2) {
        throw new Error('Полное имя должно содержать минимум 2 символа');
      }

      if (!editingUser && (!formData.password || !formData.password.trim())) {
        throw new Error('Пароль обязателен при создании');
      }

      if (formData.password && formData.password.length < 8) {
        throw new Error('Пароль должен содержать минимум 8 символов');
      }

      if (!['student', 'methodologist', 'admin'].includes(formData.role)) {
        throw new Error('Выберите валидную роль');
      }

      if (editingUser) {
        // Проверка: админ не может понизить себе роль
        if (currentUser && currentUser.id === editingUser.id && currentUser.role === 'admin') {
          if (formData.role !== 'admin') {
            throw new Error('Вы не можете понизить свою роль');
          }
        }

        // Обновить пользователя
        const updates = {
          email: formData.email.trim(),
          full_name: formData.full_name.trim(),
          role: formData.role,
        };
        if (formData.password && formData.password.trim()) {
          updates.password = formData.password; // Только если заполнено
        }
        const response = await usersAPI.updateUser(editingUser.id, updates);

        // Немедленно обновить пользователя в локальном состоянии для отзывчивости UI
        // API может вернуть { user: {...} } или напрямую объект пользователя
        const updatedUser = response?.user || response;
        if (updatedUser && updatedUser.id) {
          setUsers(prevUsers =>
            prevUsers.map(u => u.id === editingUser.id ? updatedUser : u)
          );
        }

        notification.success('Пользователь обновлен');
      } else {
        // Создать пользователя с корректными данными
        const userData = {
          email: formData.email.trim(),
          full_name: formData.full_name.trim(),
          password: formData.password,
          role: formData.role,
        };
        const response = await usersAPI.createUser(userData);
        // API может вернуть { user: {...} } или напрямую объект пользователя
        const newUser = response?.user || response;
        if (newUser && newUser.id) {
          // Добавить нового пользователя в локальное состояние
          setUsers(prevUsers => [...prevUsers, newUser]);
        }
        notification.success('Пользователь создан');
      }
      handleCloseModal();

      // Инвалидировать кэш React Query для обновления других компонентов
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['students'] });
      queryClient.invalidateQueries({ queryKey: ['teachers'] });

      // Перезагрузить данные для синхронизации с сервером
      await fetchUsers();
    } catch (error) {
      // Обработка ошибок с правильным методом notification
      const errorMessage = error?.message || 'Ошибка сохранения пользователя';
      console.error('Ошибка при сохранении пользователя:', error);
      notification.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteUserId) return;

    // Проверка: админ не может удалить себя
    if (currentUser && currentUser.id === deleteUserId) {
      notification.error('Вы не можете удалить себя');
      setDeleteUserId(null);
      return;
    }

    try {
      await usersAPI.deleteUser(deleteUserId);

      // Инвалидировать кэш React Query
      queryClient.invalidateQueries({ queryKey: ['users'] });
      queryClient.invalidateQueries({ queryKey: ['students'] });
      queryClient.invalidateQueries({ queryKey: ['teachers'] });
      queryClient.invalidateQueries({ queryKey: ['credits'] });

      notification.success('Пользователь удален');
      setDeleteUserId(null);
      fetchUsers();
    } catch (error) {
      notification.error(error.message || 'Ошибка удаления пользователя');
    }
  };

  const getRoleLabel = (role) => {
    const labels = {
      student: 'Студент',
      methodologist: 'Учитель',
      admin: 'Администратор',
    };
    return labels[role] || role;
  };

  if (loading && users.length === 0) {
    return (
      <div className="user-management" data-testid="user-management">
        <div className="user-management-header">
          <h2 className="user-management-title">Управление пользователями</h2>
          <Button onClick={handleOpenCreate} disabled data-testid="create-user-button">Добавить пользователя</Button>
        </div>

        <div className="user-management-filters">
          <label className="filter-label">Роль:</label>
          <select
            className="filter-select"
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            data-testid="role-filter"
            disabled
          >
            <option value="">Все</option>
            <option value="student">Студенты</option>
            <option value="methodologist">Учителя</option>
            <option value="admin">Администраторы</option>
          </select>
        </div>

        <div className="user-management-table-wrapper" aria-busy="true" aria-live="polite">
          <SkeletonTable rows={5} columns={4} />
        </div>

        {/* Show slow connection notice if loading takes too long */}
        {isSlow && !dismissedSlowNotice && (
          <SlowConnectionNotice onDismiss={() => setDismissedSlowNotice(true)} />
        )}
      </div>
    );
  }

  return (
    <div className="user-management" data-testid="user-management">
      <div className="user-management-header">
        <h2 className="user-management-title">Управление пользователями</h2>
        <Button onClick={handleOpenCreate} data-testid="create-user-button">Добавить пользователя</Button>
      </div>

      <div className="user-management-filters">
        <label className="filter-label">Роль:</label>
        <select
          className="filter-select"
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          data-testid="role-filter"
        >
          <option value="">Все</option>
          <option value="student">Студенты</option>
          <option value="methodologist">Учителя</option>
          <option value="admin">Администраторы</option>
        </select>
      </div>

      <div className="user-management-table-wrapper">
        <table className="user-management-table">
          <thead>
            <tr>
              <th>Имя</th>
              <th>Email</th>
              <th>Роль</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {Array.isArray(users) && users.length > 0 ? (
              users.map((user) => (
                <tr key={user.id} data-testid="user-row">
                  <td>{user.full_name}</td>
                  <td>{user.email}</td>
                  <td>
                    <span className={`role-badge role-${user.role}`}>
                      {getRoleLabel(user.role)}
                    </span>
                  </td>
                  <td>
                    <div className="user-actions">
                      <Button
                        size="small"
                        variant="outline"
                        onClick={() => handleOpenEdit(user)}
                        data-testid="edit-user-button"
                      >
                        Редактировать
                      </Button>
                      {currentUser && currentUser.id !== user.id && (
                        <Button
                          size="small"
                          variant="danger"
                          onClick={() => setDeleteUserId(user.id)}
                          data-testid="delete-user-button"
                        >
                          Удалить
                        </Button>
                      )}
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="4" className="empty-message">
                  {Array.isArray(users) ? 'Пользователи не найдены' : 'Ошибка загрузки данных'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Modal
        isOpen={showUserModal}
        onClose={handleCloseModal}
        title={editingUser ? 'Редактировать пользователя' : 'Создать пользователя'}
      >
        <form onSubmit={handleSubmit} className="user-form" data-testid="user-form">
          <Input
            label="Email"
            name="email"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData({ ...formData, email: e.target.value })}
            required
            data-testid="email-input"
          />
          <Input
            label="Полное имя"
            name="full_name"
            type="text"
            value={formData.full_name}
            onChange={(e) => setFormData({ ...formData, full_name: e.target.value })}
            required
            data-testid="full-name-input"
          />
          <Input
            label={editingUser ? 'Пароль (оставьте пустым, чтобы не менять)' : 'Пароль'}
            name="password"
            type="password"
            value={formData.password}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
            required={!editingUser}
            data-testid="password-input"
          />
          <div className="form-group">
            <label className="form-label" htmlFor="role">Роль</label>
            <select
              id="role"
              name="role"
              className="form-select"
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value })}
              required
              disabled={editingUser && currentUser && currentUser.id === editingUser.id && currentUser.role === 'admin'}
              data-testid="role-select"
            >
              <option value="student">Студент</option>
              <option value="methodologist">Учитель</option>
              <option value="admin">Администратор</option>
            </select>
            {editingUser && currentUser && currentUser.id === editingUser.id && currentUser.role === 'admin' && (
              <small className="form-help-text" style={{ display: 'block', marginTop: '4px', color: '#666' }}>
                Вы не можете изменить свою роль
              </small>
            )}
          </div>
          <div className="user-form-actions">
            <Button type="button" variant="secondary" onClick={handleCloseModal}>
              Отмена
            </Button>
            <Button type="submit" loading={submitting}>
              {editingUser ? 'Сохранить' : 'Создать'}
            </Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        isOpen={!!deleteUserId}
        onClose={() => setDeleteUserId(null)}
        onConfirm={handleDelete}
        title="Удаление пользователя"
        message="Вы действительно хотите удалить этого пользователя? Это действие нельзя отменить."
        confirmText="Удалить"
        confirmVariant="danger"
      />
    </div>
  );
};

export default UserManagement;
