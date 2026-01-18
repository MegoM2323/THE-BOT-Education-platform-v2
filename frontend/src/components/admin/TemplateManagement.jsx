import { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import Button from "../common/Button.jsx";
import Input from '../common/Input.jsx';
import TemplateCalendarView from './TemplateCalendarView.jsx';
import TemplateApplySection from './TemplateApplySection.jsx';
import { useTemplate, useTemplates } from '../../hooks/useTemplates.js';
import { useNotification } from '../../hooks/useNotification.js';
import './TemplateManagement.css';

export const TemplateManagement = ({ templateId }) => {
  const queryClient = useQueryClient();
  const { template, loading, refetch } = useTemplate(templateId);
  const { updateTemplate, isUpdating } = useTemplates();

  const notification = useNotification();
  const [isEditingName, setIsEditingName] = useState(false);
  const [editedName, setEditedName] = useState('');
  const [editedDescription, setEditedDescription] = useState('');

  // Initialize edited values when template loads
  useEffect(() => {
    if (template) {
      setEditedName(template.name || '');
      setEditedDescription(template.description || '');
    }
  }, [template]);

  const handleSaveTemplateInfo = async () => {
    if (!editedName.trim()) {
      notification.error('Название шаблона не может быть пустым');
      return;
    }

    try {
      await updateTemplate(template.id, {
        name: editedName.trim(),
        description: editedDescription.trim(),
      });
      setIsEditingName(false);
      // Кэш уже инвалидирован в useTemplates hook
    } catch (error) {
      // Error handled by hook
    }
  };

  // Если не выбран шаблон, показываем placeholder
  if (!templateId) {
    return (
      <div className="template-management" data-testid="template-management">
        <div className="template-management-placeholder">
          <p>Выберите шаблон из списка или создайте новый</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="template-management" data-testid="template-management">
        <div className="template-management-loading">
          Загрузка шаблона...
        </div>
      </div>
    );
  }

  if (!template) {
    return (
      <div className="template-management" data-testid="template-management">
        <div className="template-management-error">
          Не удалось загрузить шаблон
        </div>
      </div>
    );
  }

  return (
    <div className="template-management" data-testid="template-management">
      <div className="template-management-header">
        <div>
          <div className="template-title-section">
            {isEditingName ? (
              <div className="template-title-edit">
                <Input
                  type="text"
                  value={editedName}
                  onChange={(e) => setEditedName(e.target.value)}
                  placeholder="Название шаблона"
                  autoFocus
                />
                <Input
                  type="text"
                  value={editedDescription}
                  onChange={(e) => setEditedDescription(e.target.value)}
                  placeholder="Описание (необязательно)"
                />
                <div className="template-title-actions">
                  <Button
                    size="small"
                    onClick={handleSaveTemplateInfo}
                    loading={isUpdating}
                  >
                    Сохранить
                  </Button>
                  <Button
                    size="small"
                    variant="secondary"
                    onClick={() => {
                      setEditedName(template.name || '');
                      setEditedDescription(template.description || '');
                      setIsEditingName(false);
                    }}
                    disabled={isUpdating}
                  >
                    Отмена
                  </Button>
                </div>
              </div>
            ) : (
              <div className="template-title-display">
                <h2 className="template-management-title">{template.name}</h2>
                <Button
                  size="small"
                  variant="outline"
                  onClick={() => setIsEditingName(true)}
                >
                  Изменить
                </Button>
              </div>
            )}
          </div>
          <p className="template-management-subtitle">
            {template.description || 'Создайте и управляйте еженедельным расписанием занятий'}
          </p>
        </div>
      </div>

      <TemplateCalendarView
        templateId={template.id}
      />

      {/* Секция применения шаблона */}
      <TemplateApplySection
        templateId={template.id}
        onApplied={(result) => {
          notification.success(
            `Шаблон применен: создано занятий - ${result.created_lessons_count || 0}`
          );
          // Кэши уже инвалидированы в useTemplates hook
        }}
      />
    </div>
  );
};

export default TemplateManagement;
