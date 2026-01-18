/**
 * Утилиты для форматирования дат
 */

import { DAYS_OF_WEEK, DAYS_OF_WEEK_SHORT, MONTHS } from './constants.js';

/**
 * Форматирование даты в DD.MM.YYYY (в локальном часовом поясе)
 * @param {Date|string} date - Дата для форматирования (UTC из API)
 * @returns {string} Отформатированная дата в локальном часовом поясе
 */
export const formatDate = (date) => {
  const d = new Date(date);
  // getDate(), getMonth(), getFullYear() автоматически возвращают локальные значения
  const day = String(d.getDate()).padStart(2, '0');
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const year = d.getFullYear();
  return `${day}.${month}.${year}`;
};

/**
 * Форматирование времени в HH:MM (в локальном часовом поясе)
 * @param {Date|string} date - Дата для форматирования (UTC из API)
 * @returns {string} Отформатированное время в локальном часовом поясе
 */
export const formatTime = (date) => {
  const d = new Date(date);
  // getHours() и getMinutes() автоматически возвращают локальное время
  const hours = String(d.getHours()).padStart(2, '0');
  const minutes = String(d.getMinutes()).padStart(2, '0');
  return `${hours}:${minutes}`;
};

/**
 * Форматирование даты и времени в DD.MM.YYYY HH:MM
 * @param {Date|string} date - Дата для форматирования
 * @returns {string} Отформатированные дата и время
 */
export const formatDateTime = (date) => {
  return `${formatDate(date)} ${formatTime(date)}`;
};

/**
 * Форматирование даты в читаемый формат: "1 января 2024"
 * @param {Date|string} date - Дата для форматирования
 * @returns {string} Отформатированная дата
 */
export const formatDateReadable = (date) => {
  const d = new Date(date);
  const day = d.getDate();
  const month = MONTHS[d.getMonth()];
  const year = d.getFullYear();
  return `${day} ${month} ${year}`;
};

/**
 * Форматирование даты с днем недели: "Понедельник, 1 января"
 * @param {Date|string} date - Дата для форматирования
 * @returns {string} Отформатированная дата с днем недели
 */
export const formatDateWithDay = (date) => {
  const d = new Date(date);
  const dayOfWeek = DAYS_OF_WEEK[d.getDay()];
  const day = d.getDate();
  const month = MONTHS[d.getMonth()];
  return `${dayOfWeek}, ${day} ${month}`;
};

/**
 * Форматирование диапазона дат в формате "29 Декабрь 2025 Г. - 4 Январь 2026 Г."
 * @param {Date|string} startDate - Начальная дата
 * @param {Date|string} endDate - Конечная дата
 * @returns {string} Отформатированный диапазон дат
 */
export const formatDateRange = (startDate, endDate) => {
  const start = new Date(startDate);
  const end = new Date(endDate);
  
  // Месяцы в именительном падеже для отображения диапазона
  const monthsNominative = [
    'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
    'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
  ];
  
  const startDay = start.getDate();
  const startMonth = monthsNominative[start.getMonth()];
  const startYear = start.getFullYear();
  
  const endDay = end.getDate();
  const endMonth = monthsNominative[end.getMonth()];
  const endYear = end.getFullYear();
  
  return `${startDay} ${startMonth} ${startYear} Г. - ${endDay} ${endMonth} ${endYear} Г.`;
};

/**
 * Краткое форматирование даты с днем недели: "Пн, 1 янв"
 * @param {Date|string} date - Дата для форматирования
 * @returns {string} Кратко отформатированная дата с днем недели
 */
export const formatDateShort = (date) => {
  const d = new Date(date);
  const dayOfWeek = DAYS_OF_WEEK_SHORT[d.getDay()];
  const day = d.getDate();
  const month = MONTHS[d.getMonth()].slice(0, 3);
  return `${dayOfWeek}, ${day} ${month}`;
};

/**
 * Получение относительной строки даты (сегодня, завтра, вчера)
 * @param {Date|string} date - Дата для проверки
 * @returns {string} Относительная строка даты или отформатированная дата
 */
export const getRelativeDate = (date) => {
  const d = new Date(date);
  const today = new Date();
  const tomorrow = new Date(today);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  // Сброс времени для сравнения только дат
  d.setHours(0, 0, 0, 0);
  today.setHours(0, 0, 0, 0);
  tomorrow.setHours(0, 0, 0, 0);
  yesterday.setHours(0, 0, 0, 0);

  if (d.getTime() === today.getTime()) return 'Сегодня';
  if (d.getTime() === tomorrow.getTime()) return 'Завтра';
  if (d.getTime() === yesterday.getTime()) return 'Вчера';

  return formatDateWithDay(date);
};

/**
 * Проверка, является ли дата сегодняшней
 * @param {Date|string} date - Дата для проверки
 * @returns {boolean}
 */
export const isToday = (date) => {
  const d = new Date(date);
  const today = new Date();
  return (
    d.getDate() === today.getDate() &&
    d.getMonth() === today.getMonth() &&
    d.getFullYear() === today.getFullYear()
  );
};

/**
 * Проверка, находится ли дата в прошлом
 * @param {Date|string} date - Дата для проверки
 * @returns {boolean}
 */
export const isPast = (date) => {
  return new Date(date) < new Date();
};

/**
 * Проверка, находится ли дата в будущем
 * @param {Date|string} date - Дата для проверки
 * @returns {boolean}
 */
export const isFuture = (date) => {
  return new Date(date) > new Date();
};

/**
 * Получение продолжительности между двумя датами в минутах
 * @param {Date|string} start - Начальная дата
 * @param {Date|string} end - Конечная дата
 * @returns {number} Продолжительность в минутах
 */
export const getDuration = (start, end) => {
  const diff = new Date(end) - new Date(start);
  return Math.floor(diff / 1000 / 60);
};

/**
 * Форматирование продолжительности в читаемый формат
 * @param {number} minutes - Продолжительность в минутах
 * @returns {string} Отформатированная продолжительность
 */
export const formatDuration = (minutes) => {
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;

  if (hours === 0) return `${mins} мин`;
  if (mins === 0) return `${hours} ч`;
  return `${hours} ч ${mins} мин`;
};

/**
 * Преобразование даты в ISO строку для API (UTC)
 * @param {Date|string} date - Дата для преобразования (локальное время)
 * @returns {string} ISO строка в UTC для отправки на backend
 */
export const toISOString = (date) => {
  return new Date(date).toISOString();
};

/**
 * Парсинг ISO строки из API в объект Date
 * @param {string} isoString - ISO строка (UTC от backend)
 * @returns {Date} Объект Date (автоматически конвертирован в локальное время)
 */
export const fromISOString = (isoString) => {
  return new Date(isoString);
};

/**
 * Получение названия текущего часового пояса пользователя
 * @returns {string} Название часового пояса (например, "Europe/Moscow")
 */
export const getUserTimezone = () => {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
};

/**
 * Получение смещения часового пояса в часах
 * @param {Date|string} date - Дата для проверки
 * @returns {string} Смещение в формате "+03:00" или "-05:00"
 */
export const getTimezoneOffset = (date = new Date()) => {
  const d = new Date(date);
  const offset = -d.getTimezoneOffset(); // getTimezoneOffset() возвращает обратное значение
  const hours = Math.floor(Math.abs(offset) / 60);
  const minutes = Math.abs(offset) % 60;
  const sign = offset >= 0 ? '+' : '-';
  return `${sign}${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
};

/**
 * Конвертирует дату в строку YYYY-MM-DD в локальном часовом поясе
 * Используется для отправки дат на backend без timezone смещения
 * (в отличие от toISOString().split("T")[0] который конвертирует в UTC)
 * @param {Date|string} date - Дата для конвертации
 * @returns {string} Дата в формате YYYY-MM-DD в локальном часовом поясе
 */
export const toLocalDateString = (date) => {
  const d = new Date(date);
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

export default {
  formatDate,
  formatTime,
  formatDateTime,
  formatDateReadable,
  formatDateWithDay,
  formatDateShort,
  formatDateRange,
  getRelativeDate,
  isToday,
  isPast,
  isFuture,
  getDuration,
  formatDuration,
  toISOString,
  fromISOString,
  getUserTimezone,
  getTimezoneOffset,
  toLocalDateString,
};
