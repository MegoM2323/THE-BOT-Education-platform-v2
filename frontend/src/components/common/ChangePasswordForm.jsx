import { useState } from 'react';
import { changePassword } from '../../api/auth.js';
import Input from './Input.jsx';
import Button from './Button.jsx';
import './ChangePasswordForm.css';

export const ChangePasswordForm = () => {
  const [formData, setFormData] = useState({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  });

  const [errors, setErrors] = useState({});
  const [successMessage, setSuccessMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const validateForm = () => {
    const newErrors = {};

    if (!formData.currentPassword.trim()) {
      newErrors.currentPassword = 'Текущий пароль обязателен';
    }

    if (!formData.newPassword.trim()) {
      newErrors.newPassword = 'Новый пароль обязателен';
    } else if (formData.newPassword.length < 8) {
      newErrors.newPassword = 'Пароль должен содержать минимум 8 символов';
    }

    if (formData.newPassword === formData.currentPassword) {
      newErrors.newPassword = 'Новый пароль не должен совпадать с текущим';
    }

    if (!formData.confirmPassword.trim()) {
      newErrors.confirmPassword = 'Подтверждение пароля обязательно';
    } else if (formData.newPassword !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Пароли не совпадают';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    if (errors[name]) {
      setErrors((prev) => ({
        ...prev,
        [name]: '',
      }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSuccessMessage('');

    if (!validateForm()) {
      return;
    }

    setLoading(true);

    try {
      await changePassword(formData.currentPassword, formData.newPassword);
      setSuccessMessage('Пароль успешно изменен');
      setFormData({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
      });
      setErrors({});
    } catch (error) {
      const err = error;
      if (err.status === 401) {
        setErrors((prev) => ({
          ...prev,
          currentPassword: 'Неверный текущий пароль',
        }));
      } else if (err.status === 400) {
        setErrors((prev) => ({
          ...prev,
          newPassword: 'Пароль слишком слабый или совпадает с текущим',
        }));
      } else {
        setErrors((prev) => ({
          ...prev,
          submit: 'Ошибка при изменении пароля. Попробуйте позже',
        }));
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form className="change-password-form" onSubmit={handleSubmit}>
      <div className="form-fields">
        <Input
          type="password"
          label="Текущий пароль"
          name="currentPassword"
          value={formData.currentPassword}
          onChange={handleChange}
          error={errors.currentPassword}
          placeholder="Введите текущий пароль"
          required
        />

        <Input
          type="password"
          label="Новый пароль"
          name="newPassword"
          value={formData.newPassword}
          onChange={handleChange}
          error={errors.newPassword}
          placeholder="Введите новый пароль (минимум 8 символов)"
          required
        />

        <Input
          type="password"
          label="Подтверждение пароля"
          name="confirmPassword"
          value={formData.confirmPassword}
          onChange={handleChange}
          error={errors.confirmPassword}
          placeholder="Подтвердите новый пароль"
          required
        />
      </div>

      {successMessage && (
        <div className="form-message form-success">{successMessage}</div>
      )}

      {errors.submit && (
        <div className="form-message form-error">{errors.submit}</div>
      )}

      <Button
        type="submit"
        variant="primary"
        size="md"
        loading={loading}
        fullWidth={false}
      >
        Изменить пароль
      </Button>
    </form>
  );
};

export default ChangePasswordForm;
