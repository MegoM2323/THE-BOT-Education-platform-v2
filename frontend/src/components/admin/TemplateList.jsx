import { useState } from 'react';
import Button from "../common/Button.jsx";
import Modal from "../common/Modal.jsx";
import { useTemplates } from '../../hooks/useTemplates.js';
import TemplateCreateModal from './TemplateCreateModal.jsx';
import './TemplateList.css';

export const TemplateList = ({ onSelectTemplate, selectedTemplateId }) => {
  const { templates, loading, deleteTemplate, isDeleting } = useTemplates();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);

  const handleDeleteClick = (e, templateId) => {
    e.stopPropagation(); // Предотвратить выбор шаблона при клике на кнопку удаления
    setDeleteConfirmId(templateId);
  };

  const handleConfirmDelete = async () => {
    if (!deleteConfirmId) return;

    try {
      await deleteTemplate(deleteConfirmId);
      setDeleteConfirmId(null);
      // Если удалён выбранный шаблон, сброс выбора
      if (deleteConfirmId === selectedTemplateId) {
        onSelectTemplate(null);
      }
    } catch (error) {
      // Обработка ошибок в хуке
    }
  };

  const getDeletingTemplateName = () => {
    const template = templates.find(t => t.id === deleteConfirmId);
    return template?.name || '';
  };

  if (loading) {
    return (
      <div className="template-list-container">
        <div className="template-list-loading">Загрузка шаблонов...</div>
      </div>
    );
  }

  return (
    <div className="template-list-container">
      <div className="template-list-header">
        <h2 className="template-list-title">Шаблоны расписаний</h2>
        <Button onClick={() => setShowCreateModal(true)}>
          Создать новый шаблон
        </Button>
      </div>

      {templates.length === 0 ? (
        <div className="template-list-empty">
          <p>Нет созданных шаблонов</p>
          <p className="template-list-empty-hint">
            Создайте шаблон для начала работы с еженедельным расписанием
          </p>
        </div>
      ) : (
        <div className="template-list-grid">
          {templates.map((template) => (
            <div
              key={template.id}
              className={`template-card ${
                selectedTemplateId === template.id ? 'template-card-selected' : ''
              }`}
              onClick={() => onSelectTemplate(template.id)}
              data-testid={`template-card-${template.id}`}
            >
              <div className="template-card-header">
                <h3 className="template-card-name">{template.name}</h3>
                <Button
                  size="small"
                  variant="danger"
                  onClick={(e) => handleDeleteClick(e, template.id)}
                  disabled={isDeleting}
                  data-testid={`delete-template-${template.id}`}
                >
                  Удалить
                </Button>
              </div>
              {template.description && (
                <p className="template-card-description">{template.description}</p>
              )}
              <div className="template-card-footer">
                <span className="template-card-lessons">
                  {template.lesson_count || 0} занятий
                </span>
                <span className="template-card-date">
                  Создан: {new Date(template.created_at).toLocaleDateString('ru-RU')}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      <TemplateCreateModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={(newTemplate) => {
          setShowCreateModal(false);
          onSelectTemplate(newTemplate.id);
        }}
      />

      <Modal
        isOpen={deleteConfirmId !== null}
        onClose={() => setDeleteConfirmId(null)}
        title="Подтверждение удаления"
        loading={isDeleting}
      >
        <div className="template-delete-confirmation">
          <p>
            Вы уверены, что хотите удалить шаблон <strong>{getDeletingTemplateName()}</strong>?
          </p>
          <p className="template-delete-warning">
            Это действие нельзя отменить. Все занятия в шаблоне будут удалены.
          </p>
          <div className="template-delete-actions">
            <Button
              variant="secondary"
              onClick={() => setDeleteConfirmId(null)}
              disabled={isDeleting}
            >
              Отмена
            </Button>
            <Button
              variant="danger"
              onClick={handleConfirmDelete}
              loading={isDeleting}
            >
              Удалить шаблон
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default TemplateList;
