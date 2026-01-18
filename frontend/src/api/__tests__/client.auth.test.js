import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getStoredToken, saveToken, clearToken } from '../client.js';

/**
 * ПРИМЕЧАНИЕ: Эти функции теперь являются no-op (ничего не делают)
 * Аутентификация теперь использует httpOnly cookies, управляемые backend
 * Тесты обновлены для отражения нового поведения
 */
describe('Authentication Token Management (Deprecated - uses httpOnly cookies now)', () => {
  beforeEach(() => {
    // Очищаем localStorage перед каждым тестом
    localStorage.clear();
    vi.clearAllMocks();
    // Mock console.debug для подавления логов
    vi.spyOn(console, 'debug').mockImplementation(() => {});
  });

  afterEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  describe('saveToken (deprecated - no-op)', () => {
    it('should be a no-op and not store anything', () => {
      const testToken = 'test-jwt-token-12345';
      saveToken(testToken);

      // saveToken теперь no-op - ничего не должно быть сохранено
      const stored = localStorage.getItem('auth_token');
      expect(stored).toBeNull();
    });

    it('should log debug message', () => {
      saveToken('test-token');
      expect(console.debug).toHaveBeenCalledWith(
        'saveToken called but ignored - using httpOnly cookies'
      );
    });
  });

  describe('getStoredToken (deprecated - returns null)', () => {
    it('should always return null (token is in httpOnly cookie)', () => {
      // Даже если что-то есть в localStorage, getStoredToken вернет null
      localStorage.setItem('auth_token', 'old-token');
      const retrieved = getStoredToken();
      expect(retrieved).toBeNull();
    });

    it('should return null when no token exists', () => {
      const retrieved = getStoredToken();
      expect(retrieved).toBeNull();
    });
  });

  describe('clearToken (deprecated - no-op)', () => {
    it('should be a no-op (backend handles cookie clearing)', () => {
      localStorage.setItem('auth_token', 'test-token');
      clearToken();

      // clearToken теперь no-op, но localStorage не очищается этой функцией
      // (т.к. мы больше не используем localStorage для токенов)
      expect(console.debug).toHaveBeenCalledWith(
        'clearToken called but ignored - using httpOnly cookies'
      );
    });
  });

  describe('Token lifecycle (deprecated)', () => {
    it('getStoredToken always returns null with httpOnly cookies', () => {
      // С httpOnly cookies токен не доступен через JS
      expect(getStoredToken()).toBeNull();

      // saveToken - no-op
      saveToken('new-jwt-token-xyz');
      expect(getStoredToken()).toBeNull();

      // clearToken - no-op
      clearToken();
      expect(getStoredToken()).toBeNull();
    });
  });
});
