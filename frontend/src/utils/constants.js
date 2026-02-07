/**
 * Константы приложения
 */

// Роли пользователей
export const ROLES = {
  ADMIN: 'admin',
  TEACHER: 'teacher',
  STUDENT: 'student',
};

// Статусы записей
export const BOOKING_STATUS = {
  ACTIVE: 'active',
  CANCELLED: 'cancelled',
  COMPLETED: 'completed',
};

// Статусы занятий
export const LESSON_STATUS = {
  SCHEDULED: 'scheduled',
  IN_PROGRESS: 'in_progress',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled',
};

// Типы транзакций кредитов
export const CREDIT_TRANSACTION_TYPES = {
  BOOKING: 'booking',
  CANCELLATION: 'cancellation',
  SWAP: 'swap',
  ADMIN_ADD: 'admin_add',
  ADMIN_DEDUCT: 'admin_deduct',
};

// Предметы
export const SUBJECTS = [
  'Математика',
  'Физика',
  'Химия',
  'Информатика',
  'Английский язык',
  'Русский язык',
  'Литература',
  'История',
  'Обществознание',
  'Биология',
  'География',
];

// Дни недели
export const DAYS_OF_WEEK = [
  'Воскресенье',
  'Понедельник',
  'Вторник',
  'Среда',
  'Четверг',
  'Пятница',
  'Суббота',
];

export const DAYS_OF_WEEK_SHORT = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

// Месяцы
export const MONTHS = [
  'Января',
  'Февраля',
  'Марта',
  'Апреля',
  'Мая',
  'Июня',
  'Июля',
  'Августа',
  'Сентября',
  'Октября',
  'Ноября',
  'Декабря',
];

// Временные слоты (часы)
export const TIME_SLOTS = [
  '08:00',
  '09:00',
  '10:00',
  '11:00',
  '12:00',
  '13:00',
  '14:00',
  '15:00',
  '16:00',
  '17:00',
  '18:00',
  '19:00',
  '20:00',
];

// Типы уведомлений
export const NOTIFICATION_TYPES = {
  SUCCESS: 'success',
  ERROR: 'error',
  WARNING: 'warning',
  INFO: 'info',
};

// Базовый URL для API
export const API_URL = '/api';

// Ключи локального хранилища
export const STORAGE_KEYS = {
  USER: 'user',
  AUTH_TOKEN: 'auth_token',
  THEME: 'theme',
};

export default {
  ROLES,
  BOOKING_STATUS,
  LESSON_STATUS,
  CREDIT_TRANSACTION_TYPES,
  SUBJECTS,
  DAYS_OF_WEEK,
  DAYS_OF_WEEK_SHORT,
  MONTHS,
  TIME_SLOTS,
  NOTIFICATION_TYPES,
  API_URL,
  STORAGE_KEYS,
};
