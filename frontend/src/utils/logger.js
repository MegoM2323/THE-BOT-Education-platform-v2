/**
 * Development logging utility
 * Логирование только в режиме разработки
 */

const isDev = import.meta?.env?.DEV || false;

export const logger = {
  /**
   * Debug logs - только для разработки
   */
  log: (...args) => {
    if (isDev) {
      console.log(...args);
    }
  },

  /**
   * Debug logs - только для разработки
   */
  debug: (...args) => {
    if (isDev) {
      console.debug(...args);
    }
  },

  /**
   * Warning logs - только для разработки
   */
  warn: (...args) => {
    if (isDev) {
      console.warn(...args);
    }
  },

  /**
   * Info logs - только для разработки
   */
  info: (...args) => {
    if (isDev) {
      console.info(...args);
    }
  },

  /**
   * Error logs - всегда (включая production)
   * Используй только для критичных ошибок
   */
  error: (...args) => {
    console.error(...args);
  },
};

// Короткие алиасы для удобства
export const log = logger.log;
export const debug = logger.debug;
export const warn = logger.warn;
export const info = logger.info;
export const error = logger.error;

export default logger;
