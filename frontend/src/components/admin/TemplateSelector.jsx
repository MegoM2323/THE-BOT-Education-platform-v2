import { useState } from 'react';
import Button from "../common/Button.jsx";
import Modal from "../common/Modal.jsx";
import ConfirmModal from '../common/ConfirmModal.jsx';
import Input from '../common/Input.jsx';
import { useTemplates } from '../../hooks/useTemplates.js';
import { formatDate } from '../../utils/dateFormat.js';
import './TemplateSelector.css';

export const TemplateSelector = ({ onSelect, selectedTemplate }) => {
  const { templates, loading, deleteTemplate, createTemplate, isDeleting, isCreating } = useTemplates();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [deleteTemplateId, setDeleteTemplateId] = useState(null);
  const [newTemplateName, setNewTemplateName] = useState('');

  const handleCreateTemplate = async (e) => {
    e.preventDefault();
    if (!newTemplateName.trim()) {
      return;
    }

    try {
      const created = await createTemplate({
        name: newTemplateName.trim(),
        description: '',
      });
      setNewTemplateName('');
      setShowCreateModal(false);
      onSelect(created);
    } catch (error) {
      // Error handled by hook
    }
  };

  const handleDeleteTemplate = async () => {
    if (!deleteTemplateId) return;

    try {
      await deleteTemplate(deleteTemplateId);
      if (selectedTemplate?.id === deleteTemplateId) {
        onSelect(null);
      }
      setDeleteTemplateId(null);
    } catch (error) {
      // Error handled by hook
    }
  };

  if (loading) {
    return <div className="template-selector-loading">Загрузка шаблонов...</div>;
  }

  return (
    <div className="template-selector">
      <div className="template-selector-header">
        <label className="template-selector-label">Выберите шаблон</label>
        <Button
          size="small"
          onClick={() => setShowCreateModal(true)}
          data-testid="create-template-button"
        >
          + Новый шаблон
        </Button>
      </div>

      <div className="template-selector-content">
        <select
          className="template-selector-dropdown"
          value={selectedTemplate?.id || ''}
          onChange={(e) => {
            const template = templates.find((t) => t.id === e.target.value);
            onSelect(template || null);
          }}
          data-testid="template-select"
        >
          <option value="">-- Выберите шаблон --</option>
          {templates.map((template) => (
            <option key={template.id} value={template.id}>
              {template.name} ({template.lesson_count || 0} занятий)
            </option>
          ))}
        </select>

        {selectedTemplate && (
          <div className="template-selector-actions">
            <Button
              size="small"
              variant="danger"
              onClick={() => setDeleteTemplateId(selectedTemplate.id)}
              disabled={isDeleting}
              data-testid="delete-template-button"
            >
              Удалить шаблон
            </Button>
          </div>
        )}
      </div>

      {selectedTemplate && (
        <div className="template-info">
          <div className="template-info-item">
            <span className="template-info-label">Создан:</span>
            <span className="template-info-value">
              {formatDate(selectedTemplate.created_at)}
            </span>
          </div>
          {selectedTemplate.description && (
            <div className="template-info-item">
              <span className="template-info-label">Описание:</span>
              <span className="template-info-value">{selectedTemplate.description}</span>
            </div>
          )}
        </div>
      )}

      <Modal
        isOpen={showCreateModal}
        onClose={() => {
          setShowCreateModal(false);
          setNewTemplateName('');
        }}
        title="Создать новый шаблон"
        loading={isCreating}
      >
        <form onSubmit={handleCreateTemplate} className="template-create-form">
          <Input
            label="Название шаблона"
            type="text"
            value={newTemplateName}
            onChange={(e) => setNewTemplateName(e.target.value)}
            placeholder="например: Стандартная неделя, Летнее расписание"
            required
            autoFocus
          />
          <div className="template-create-actions">
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setShowCreateModal(false);
                setNewTemplateName('');
              }}
              disabled={isCreating}
            >
              Отмена
            </Button>
            <Button type="submit" loading={isCreating}>
              Создать шаблон
            </Button>
          </div>
        </form>
      </Modal>

      <ConfirmModal
        isOpen={!!deleteTemplateId}
        onClose={() => setDeleteTemplateId(null)}
        onConfirm={handleDeleteTemplate}
        title="Удалить шаблон"
        message="Вы уверены, что хотите удалить этот шаблон? Это действие нельзя отменить."
        confirmText="Удалить"
        confirmVariant="danger"
      />
    </div>
  );
};

export default TemplateSelector;
