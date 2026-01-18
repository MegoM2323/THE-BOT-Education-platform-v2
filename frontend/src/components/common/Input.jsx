import { useState, useRef } from 'react';
import './Input.css';

/**
 * Компонент поля ввода с валидацией
 * Поддерживает типы: text, email, password, date, time, number и другие
 */
export const Input = ({
  type = 'text',
  label,
  name,
  value,
  onChange,
  onBlur,
  onClick,
  placeholder,
  error,
  required = false,
  disabled = false,
  className = '',
  min,
  max,
  step,
  ...props
}) => {
  const [focused, setFocused] = useState(false);
  const inputRef = useRef(null);

  const handleFocus = () => setFocused(true);
  const handleBlur = (e) => {
    setFocused(false);
    if (onBlur) onBlur(e);
  };

  // Обработчик клика для открытия календаря (для date/time типов)
  const handleClick = (e) => {
    // Вызываем переданный onClick если есть
    if (onClick) {
      onClick(e);
    }
    
    // Для date/time типов пытаемся открыть нативный календарь
    if ((type === 'date' || type === 'time' || type === 'datetime-local') && inputRef.current) {
      // Используем showPicker() если доступен (современные браузеры)
      if (typeof inputRef.current.showPicker === 'function') {
        try {
          inputRef.current.showPicker();
        } catch (err) {
          // Если showPicker() не поддерживается или не может быть вызван, 
          // браузер откроет календарь автоматически при фокусе
        }
      }
    }
  };

  // Добавляем специфичный класс для date/time/number типов
  const typeClass = ['date', 'time', 'datetime-local', 'number'].includes(type)
    ? `input-field-${type}`
    : '';

  const inputClasses = [
    'input-field',
    typeClass,
    error && 'input-error',
    focused && 'input-focused',
    disabled && 'input-disabled',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  // Подготовка атрибутов для number/date/time
  const inputProps = {
    id: name,
    type,
    name,
    value,
    onChange,
    onFocus: handleFocus,
    onBlur: handleBlur,
    onClick: handleClick,
    placeholder,
    disabled,
    required,
    className: inputClasses,
    ...props,
  };

  // Добавляем min/max для number и date типов
  if (type === 'number') {
    if (min !== undefined) inputProps.min = min;
    if (max !== undefined) inputProps.max = max;
    if (step !== undefined) inputProps.step = step;
  }
  
  if (type === 'date' || type === 'datetime-local' || type === 'time') {
    if (min !== undefined) inputProps.min = min;
    if (max !== undefined) inputProps.max = max;
  }

  return (
    <div className="input-container">
      {label && (
        <label htmlFor={name} className="input-label">
          {label}
          {required && <span className="input-required">*</span>}
        </label>
      )}
      <input {...inputProps} ref={inputRef} />
      {error && <span className="input-error-message">{error instanceof Error ? error.message : (error?.message || String(error))}</span>}
    </div>
  );
};

export default Input;
