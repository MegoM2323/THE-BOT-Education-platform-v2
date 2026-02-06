import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth.js';
import { useNotification } from '../../hooks/useNotification.js';
import { validateLoginForm } from '../../utils/validation.js';
import Input from '../common/Input.jsx';
import Button from "../common/Button.jsx";
import { ROLES } from '../../utils/constants.js';
import './LoginForm.css';

export const LoginForm = () => {
  const [formData, setFormData] = useState({ email: '', password: '' });
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [passwordFocused, setPasswordFocused] = useState(false);
  const { login } = useAuth();
  const notification = useNotification();
  const navigate = useNavigate();

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: null }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    const validation = validateLoginForm(formData);
    if (!validation.isValid) {
      setErrors(validation.errors);
      return;
    }

    try {
      setLoading(true);
      const user = await login(formData.email, formData.password);

      // Микротаск для завершения React батчинга перед навигацией.
      // Это гарантирует, что Context.isAuthenticated будет обновлён
      // ДО того, как ProtectedRoute проверит аутентификацию.
      await new Promise(resolve => setTimeout(resolve, 0));

      notification.success('Добро пожаловать!');

      // Перенаправление в зависимости от роли
      switch (user.role) {
        case ROLES.ADMIN:
          navigate('/dashboard/admin/users');
          break;
        case ROLES.STUDENT:
          navigate('/dashboard/student/schedule');
          break;
        default:
          navigate('/');
      }
    } catch (error) {
      // Log retry attempts if they happened (for debugging)
      console.debug('[LoginForm] Login failed with error:', error.message);
      notification.error(error.message || 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="login-form" data-testid="login-form">
      {errors.general && (
        <div className="login-error" data-testid="error-message">
          {errors.general}
        </div>
      )}
      <div data-testid="email-input">
        <Input
          label="Email"
          type="email"
          name="email"
          value={formData.email}
          onChange={handleChange}
          error={errors.email}
          required
        />
      </div>
      <div className="password-input-wrapper" data-testid="password-input">
        <div className="input-container">
          <label htmlFor="password" className="input-label">
            Пароль
            <span className="input-required">*</span>
          </label>
          <div style={{ position: 'relative', width: '100%' }}>
            <input
              id="password"
              type={showPassword ? "text" : "password"}
              name="password"
              value={formData.password}
              onChange={handleChange}
              onFocus={() => setPasswordFocused(true)}
              onBlur={() => setPasswordFocused(false)}
              placeholder=""
              required
              className={`input-field ${errors.password ? 'input-error' : ''} ${passwordFocused ? 'input-focused' : ''} input-with-password-toggle`}
            />
            <button
              type="button"
              className="toggle-password"
              onClick={() => setShowPassword(!showPassword)}
              data-testid="toggle-password"
              aria-label={showPassword ? "Скрыть пароль" : "Показать пароль"}
            >
              {showPassword ? (
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"></path>
                  <path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"></path>
                  <path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"></path>
                  <line x1="2" y1="2" x2="22" y2="22"></line>
                </svg>
              ) : (
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"></path>
                  <circle cx="12" cy="12" r="3"></circle>
                </svg>
              )}
            </button>
          </div>
          {errors.password && <span className="input-error-message">{errors.password}</span>}
        </div>
      </div>
      {loading && <div data-testid="loading-spinner">Загрузка...</div>}
      <Button type="submit" loading={loading} fullWidth data-testid="login-button">
        Войти
      </Button>
      <div className="login-footer">
        <Link to="/" className="back-to-home-link">
          ← Вернуться на главную
        </Link>
      </div>
    </form>
  );
};

export default LoginForm;
