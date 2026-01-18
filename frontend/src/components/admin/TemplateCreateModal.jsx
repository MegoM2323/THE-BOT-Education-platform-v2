import { useState } from 'react';
import Button from "../common/Button.jsx";
import Input from '../common/Input.jsx';
import Modal from "../common/Modal.jsx";
import { useTemplates } from '../../hooks/useTemplates.js';
import './TemplateCreateModal.css';

export const TemplateCreateModal = ({ isOpen, onClose, onSuccess }) => {
  const { createTemplate, isCreating } = useTemplates();
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  });
  const [errors, setErrors] = useState({});

  const handleChange = (field, value) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Очистка ошибки при изменении поля
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  const validateForm = () => {
    const newErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Название обязательно';
    } else if (formData.name.trim().length < 3) {
      newErrors.name = 'Название должно содержать минимум 3 символа';
    } else if (formData.name.trim().length > 100) {
      newErrors.name = 'Название не должно превышать 100 символов';
    }

    if (formData.description.trim().length > 500) {
      newErrors.description = 'Описание не должно превышать 500 символов';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      const newTemplate = await createTemplate({
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
      });

      // Сброс формы после успешного создания
      setFormData({ name: '', description: '' });
      setErrors({});

      if (onSuccess) {
        onSuccess(newTemplate);
      }
    } catch (error) {
      // Обработка ошибок в хуке
    }
  };

  const handleClose = () => {
    if (!isCreating) {
      setFormData({ name: '', description: '' });
      setErrors({});
      onClose();
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Создать новый шаблон"
      loading={isCreating}
    >
      <form onSubmit={handleSubmit} className="template-create-form">
        <p className="template-create-description">
          Создайте шаблон для еженедельного расписания занятий. После создания вы сможете
          добавить занятия в шаблон и применять его к разным неделям.
        </p>

        <div className="form-field">
          <Input
            label="Название шаблона *"
            type="text"
            value={formData.name}
            onChange={(e) => handleChange('name', e.target.value)}
            placeholder="Например: Осенний семестр 2024"
            autoFocus
            disabled={isCreating}
            error={errors.name}
            maxLength={100}
          />
          {errors.name && <span className="error-text">{errors.name}</span>}
        </div>

        <div className="form-field">
          <label htmlFor="description" className="input-label">
            Описание (необязательно)
          </label>
          <textarea
            id="description"
            className="template-description-textarea"
            value={formData.description}
            onChange={(e) => handleChange('description', e.target.value)}
            placeholder="Краткое описание шаблона (необязательно)"
            rows={4}
            disabled={isCreating}
            maxLength={500}
          />
          {errors.description && <span className="error-text">{errors.description}</span>}
          <span className="char-count">
            {formData.description.length}/500
          </span>
        </div>

        <div className="template-create-actions">
          <Button
            type="button"
            variant="secondary"
            onClick={handleClose}
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
  );
};

export default TemplateCreateModal;
