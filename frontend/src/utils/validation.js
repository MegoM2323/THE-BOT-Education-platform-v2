/**
 * Утилиты для валидации форм
 */

/**
 * Валидация формата email
 * @param {string} email - Email для валидации
 * @returns {boolean}
 */
export const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

/**
 * Валидация надежности пароля
 * @param {string} password - Пароль для валидации
 * @returns {Object} Результат валидации с полями isValid и errors
 */
export const validatePassword = (password) => {
  const errors = [];

  if (password.length < 8) {
    errors.push('Пароль должен содержать минимум 8 символов');
  }

  if (!/[A-Z]/.test(password)) {
    errors.push('Пароль должен содержать хотя бы одну заглавную букву');
  }

  if (!/[a-z]/.test(password)) {
    errors.push('Пароль должен содержать хотя бы одну строчную букву');
  }

  if (!/[0-9]/.test(password)) {
    errors.push('Пароль должен содержать хотя бы одну цифру');
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

/**
 * Валидация обязательного поля
 * @param {any} value - Значение для валидации
 * @returns {boolean}
 */
export const isRequired = (value) => {
  if (typeof value === 'string') {
    return value.trim().length > 0;
  }
  return value !== null && value !== undefined;
};

/**
 * Валидация минимальной длины
 * @param {string} value - Значение для валидации
 * @param {number} min - Минимальная длина
 * @returns {boolean}
 */
export const minLength = (value, min) => {
  return value.length >= min;
};

/**
 * Валидация максимальной длины
 * @param {string} value - Значение для валидации
 * @param {number} max - Максимальная длина
 * @returns {boolean}
 */
export const maxLength = (value, max) => {
  return value.length <= max;
};

/**
 * Валидация диапазона чисел
 * @param {number} value - Значение для валидации
 * @param {number} min - Минимальное значение
 * @param {number} max - Максимальное значение
 * @returns {boolean}
 */
export const inRange = (value, min, max) => {
  const num = Number(value);
  return num >= min && num <= max;
};

/**
 * Валидация номера телефона (Российский формат)
 * @param {string} phone - Номер телефона для валидации
 * @returns {boolean}
 */
export const isValidPhone = (phone) => {
  const phoneRegex = /^(\+7|8)?[\s-]?\(?[489][0-9]{2}\)?[\s-]?[0-9]{3}[\s-]?[0-9]{2}[\s-]?[0-9]{2}$/;
  return phoneRegex.test(phone);
};

/**
 * Валидация, что дата не в прошлом
 * @param {Date|string} date - Дата для валидации
 * @returns {boolean}
 */
export const isNotPast = (date) => {
  return new Date(date) >= new Date();
};

/**
 * Валидация, что дата находится в диапазоне
 * @param {Date|string} date - Дата для валидации
 * @param {Date|string} start - Начальная дата
 * @param {Date|string} end - Конечная дата
 * @returns {boolean}
 */
export const isDateInRange = (date, start, end) => {
  const d = new Date(date);
  return d >= new Date(start) && d <= new Date(end);
};

/**
 * Валидация формы входа
 * @param {Object} data - Данные формы
 * @returns {Object} Результат валидации с полями isValid и errors
 */
export const validateLoginForm = (data) => {
  const errors = {};

  if (!isRequired(data.email)) {
    errors.email = 'Email обязателен';
  } else if (!isValidEmail(data.email)) {
    errors.email = 'Неверный формат email';
  }

  if (!isRequired(data.password)) {
    errors.password = 'Пароль обязателен';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
};

/**
 * Валидация формы пользователя
 * @param {Object} data - Данные формы
 * @returns {Object} Результат валидации с полями isValid и errors
 */
export const validateUserForm = (data) => {
  const errors = {};

  if (!isRequired(data.name)) {
    errors.name = 'Имя обязательно';
  } else if (!minLength(data.name, 2)) {
    errors.name = 'Имя должно содержать минимум 2 символа';
  }

  if (!isRequired(data.email)) {
    errors.email = 'Email обязателен';
  } else if (!isValidEmail(data.email)) {
    errors.email = 'Неверный формат email';
  }

  if (data.password && data.password.length > 0) {
    const passwordValidation = validatePassword(data.password);
    if (!passwordValidation.isValid) {
      errors.password = passwordValidation.errors[0];
    }
  }

  if (data.phone && !isValidPhone(data.phone)) {
    errors.phone = 'Неверный формат телефона';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
};

/**
 * Валидация формы занятия
 * @param {Object} data - Данные формы
 * @returns {Object} Результат валидации с полями isValid и errors
 */
export const validateLessonForm = (data) => {
  const errors = {};

  if (!isRequired(data.subject)) {
    errors.subject = 'Предмет обязателен';
  }

  if (!isRequired(data.start_time)) {
    errors.start_time = 'Время начала обязательно';
  } else if (!isNotPast(data.start_time)) {
    errors.start_time = 'Время начала не может быть в прошлом';
  }

  if (!isRequired(data.end_time)) {
    errors.end_time = 'Время окончания обязательно';
  } else if (new Date(data.end_time) <= new Date(data.start_time)) {
    errors.end_time = 'Время окончания должно быть позже времени начала';
  }

  if (!isRequired(data.max_students)) {
    errors.max_students = 'Максимальное количество учеников обязательно';
  } else if (!inRange(data.max_students, 1, 20)) {
    errors.max_students = 'Количество учеников должно быть от 1 до 20';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
};

/**
 * Валидация формы кредитов
 * @param {Object} data - Данные формы
 * @returns {Object} Результат валидации с полями isValid и errors
 */
export const validateCreditsForm = (data) => {
  const errors = {};

  if (!isRequired(data.amount)) {
    errors.amount = 'Количество кредитов обязательно';
  } else if (!inRange(data.amount, 1, 1000)) {
    errors.amount = 'Количество кредитов должно быть от 1 до 1000';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
  };
};

/**
 * Whitelist разрешенных доменов для редиректа платежной системы
 * Только YooKassa домены разрешены
 */
const ALLOWED_PAYMENT_DOMAINS = [
  'yookassa.ru',
  'yoomoney.ru',
  'money.yandex.ru',
];

/**
 * Валидация URL для редиректа на платежную систему
 * Проверяет, что URL принадлежит разрешенному домену платежной системы
 * @param {string} url - URL для валидации
 * @returns {Object} Результат валидации с полями isValid, error и sanitizedUrl
 */
export const validatePaymentRedirectUrl = (url) => {
  // Проверка наличия URL
  if (!url || typeof url !== 'string' || url.trim().length === 0) {
    return {
      isValid: false,
      error: 'URL не указан',
      sanitizedUrl: null,
    };
  }

  try {
    // Парсинг URL для валидации структуры
    const parsedUrl = new URL(url);

    // Проверка протокола - только HTTPS разрешен
    if (parsedUrl.protocol !== 'https:') {
      console.warn('[Security] Payment redirect rejected: non-HTTPS protocol', { url });
      return {
        isValid: false,
        error: 'Только HTTPS протокол разрешен для платежных редиректов',
        sanitizedUrl: null,
      };
    }

    // Извлечение hostname и проверка по whitelist
    const hostname = parsedUrl.hostname.toLowerCase();

    // Проверка, что домен соответствует одному из разрешенных
    const isAllowedDomain = ALLOWED_PAYMENT_DOMAINS.some(allowedDomain => {
      // Проверка точного совпадения или поддомена
      return hostname === allowedDomain || hostname.endsWith(`.${allowedDomain}`);
    });

    if (!isAllowedDomain) {
      console.warn('[Security] Payment redirect rejected: unauthorized domain', {
        url,
        hostname,
        allowedDomains: ALLOWED_PAYMENT_DOMAINS,
      });
      return {
        isValid: false,
        error: `Домен ${hostname} не разрешен для платежных редиректов`,
        sanitizedUrl: null,
      };
    }

    // URL валидный и безопасный
    console.log('[Security] Payment redirect validated', { hostname });
    return {
      isValid: true,
      error: null,
      sanitizedUrl: parsedUrl.href, // Используем нормализованный URL
    };
  } catch (error) {
    console.warn('[Security] Payment redirect rejected: invalid URL format', { url, error: error.message });
    return {
      isValid: false,
      error: 'Неверный формат URL',
      sanitizedUrl: null,
    };
  }
};

export default {
  isValidEmail,
  validatePassword,
  isRequired,
  minLength,
  maxLength,
  inRange,
  isValidPhone,
  isNotPast,
  isDateInRange,
  validateLoginForm,
  validateUserForm,
  validateLessonForm,
  validateCreditsForm,
  validatePaymentRedirectUrl,
};
