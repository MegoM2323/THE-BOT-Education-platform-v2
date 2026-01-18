import { useState, useEffect } from 'react';
import Button from "../common/Button.jsx";
import Modal from "../common/Modal.jsx";
import ConfirmModal from '../common/ConfirmModal.jsx';
import Input from '../common/Input.jsx';
import Spinner from '../common/Spinner.jsx';
import TelegramUsersTable from './TelegramUsersTable.jsx';
import { useBroadcast } from '../../hooks/useBroadcast.js';
import { useNotification } from '../../hooks/useNotification';
import './BroadcastListManager.css';

/**
 * Компонент управления списками рассылки
 */
export const BroadcastListManager = () => {
  const notification = useNotification();
  const {
    linkedUsers,
    broadcastLists,
    loading,
    fetchLinkedUsers,
    fetchBroadcastLists,
    createList,
    updateList,
    deleteList,
  } = useBroadcast();

  const [showListModal, setShowListModal] = useState(false);
  const [editingList, setEditingList] = useState(null);
  const [deleteListId, setDeleteListId] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  });
  const [selectedUserIds, setSelectedUserIds] = useState([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchBroadcastLists();
  }, [fetchBroadcastLists]);

  const handleOpenCreate = async () => {
    setEditingList(null);
    setFormData({ name: '', description: '' });
    setSelectedUserIds([]);

    // Загрузить пользователей для выбора
    try {
      await fetchLinkedUsers();
      setShowListModal(true);
    } catch (error) {
      // Ошибка уже обработана в хуке
    }
  };

  const handleOpenEdit = async (list) => {
    setEditingList(list);
    setFormData({
      name: list.name,
      description: list.description || '',
    });
    setSelectedUserIds(list.user_ids || []);

    // Загрузить пользователей для выбора
    try {
      await fetchLinkedUsers();
      setShowListModal(true);
    } catch (error) {
      // Ошибка уже обработана в хуке
    }
  };

  const handleCloseModal = () => {
    setShowListModal(false);
    setEditingList(null);
    setFormData({ name: '', description: '' });
    setSelectedUserIds([]);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Валидация формы
    if (!formData.name || formData.name.trim().length === 0) {
      notification.warning('Введите название списка');
      return;
    }

    if (selectedUserIds.length === 0) {
      notification.warning('Выберите хотя бы одного пользователя');
      return;
    }

    setSubmitting(true);

    try {
      const data = {
        ...formData,
        name: formData.name.trim(),
        description: formData.description?.trim() || '',
        user_ids: selectedUserIds,
      };

      if (editingList) {
        await updateList(editingList.id, data);
      } else {
        await createList(data);
      }

      handleCloseModal();
      fetchBroadcastLists();
    } catch (error) {
      // Ошибка уже обработана в хуке
      console.error('Ошибка при сохранении списка:', error);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteListId) return;

    try {
      await deleteList(deleteListId);
      setDeleteListId(null);
      fetchBroadcastLists();
    } catch (error) {
      // Ошибка уже обработана в хуке
    }
  };

  if (loading && broadcastLists.length === 0) {
    return (
      <div className="broadcast-list-manager-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="broadcast-list-manager">
      <div className="broadcast-list-header">
        <h3 className="broadcast-list-title">Списки рассылки</h3>
        <Button onClick={handleOpenCreate}>
          Создать список
        </Button>
      </div>

      <div className="broadcast-lists-grid">
        {broadcastLists.length === 0 ? (
          <div className="empty-state">
            <p>Нет созданных списков рассылки</p>
            <Button onClick={handleOpenCreate}>
              Создать первый список
            </Button>
          </div>
        ) : (
          broadcastLists.map((list) => (
            <div key={list.id} className="broadcast-list-card">
              <div className="list-card-header">
                <h4 className="list-card-title">{list.name}</h4>
                <div className="list-card-badge">
                  {list.user_count || list.user_ids?.length || 0} получателей
                </div>
              </div>

              {list.description && (
                <p className="list-card-description">{list.description}</p>
              )}

              <div className="list-card-meta">
                <span className="list-card-date">
                  Создан: {new Date(list.created_at).toLocaleDateString('ru-RU')}
                </span>
              </div>

              <div className="list-card-actions">
                <Button
                  size="small"
                  variant="outline"
                  onClick={() => handleOpenEdit(list)}
                >
                  Редактировать
                </Button>
                <Button
                  size="small"
                  variant="danger"
                  onClick={() => setDeleteListId(list.id)}
                >
                  Удалить
                </Button>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Модальное окно создания/редактирования */}
      <Modal
        isOpen={showListModal}
        onClose={handleCloseModal}
        title={editingList ? 'Редактировать список' : 'Создать список рассылки'}
      >
        <form onSubmit={handleSubmit} className="broadcast-list-form">
          <Input
            name="list-name"
            label="Название списка"
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            placeholder="Например: Все студенты"
          />

          <div className="form-group">
            <label className="form-label">Описание</label>
            <textarea
              className="form-textarea"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Краткое описание списка (необязательно)"
              rows={3}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Выберите получателей</label>
            <TelegramUsersTable
              users={linkedUsers}
              selectedUsers={selectedUserIds}
              onSelectionChange={setSelectedUserIds}
              showCheckboxes={true}
            />
          </div>

          <div className="broadcast-list-form-actions">
            <Button type="button" variant="secondary" onClick={handleCloseModal}>
              Отмена
            </Button>
            <Button type="submit" loading={submitting} disabled={selectedUserIds.length === 0}>
              {editingList ? 'Сохранить' : 'Создать'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Модальное окно подтверждения удаления */}
      <ConfirmModal
        isOpen={!!deleteListId}
        onClose={() => setDeleteListId(null)}
        onConfirm={handleDelete}
        title="Удаление списка рассылки"
        message="Вы действительно хотите удалить этот список? Это действие нельзя отменить."
        confirmText="Удалить"
        confirmVariant="danger"
      />
    </div>
  );
};

export default BroadcastListManager;
