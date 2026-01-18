import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach, beforeEach, vi } from 'vitest';
import { QueryClient } from '@tanstack/react-query';
import React from 'react';

// =============================================================================
// NOTES ON API MOCKING
// =============================================================================
// API client is NOT globally mocked to allow individual tests to set up their
// own mocks with specific behavior. Tests that need API mocking should:
// 1. Use vi.mock('../src/api/client.js', ...) in their test file
// 2. Or mock specific hooks like useBroadcast, useLessons etc.
// =============================================================================

// =============================================================================
// GLOBAL NOTIFICATION MOCK
// =============================================================================
// Mock NotificationContext FIRST to prevent import errors when useNotification
// is loaded via require() in test files
const mockNotificationValue = {
  showNotification: vi.fn(),
  addNotification: vi.fn(),
  removeNotification: vi.fn(),
  clearAll: vi.fn(),
  error: vi.fn(),
  success: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  notifications: [],
};

vi.mock('./src/context/NotificationContext', () => ({
  NotificationContext: React.createContext(mockNotificationValue),
  NotificationProvider: ({ children }) => children,
  default: React.createContext(mockNotificationValue),
}));

vi.mock('./src/context/NotificationContext.jsx', () => ({
  NotificationContext: React.createContext(mockNotificationValue),
  NotificationProvider: ({ children }) => children,
  default: React.createContext(mockNotificationValue),
}));

// Mock useNotification hook
vi.mock('./src/hooks/useNotification', () => ({
  useNotification: vi.fn(() => mockNotificationValue),
  default: vi.fn(() => mockNotificationValue),
}));

vi.mock('./src/hooks/useNotification.js', () => ({
  useNotification: vi.fn(() => mockNotificationValue),
  default: vi.fn(() => mockNotificationValue),
}));

// =============================================================================
// GLOBAL AUTH MOCK
// =============================================================================
const mockAuthValue = {
  user: { id: 1, email: 'test@example.com', role: 'teacher' },
  isAuthenticated: true,
  loading: false,
  login: vi.fn(),
  logout: vi.fn(),
  register: vi.fn(),
  checkAuth: vi.fn(),
  refreshUser: vi.fn(),
  balance: 0,
};

vi.mock('./src/hooks/useAuth', () => ({
  useAuth: vi.fn(() => mockAuthValue),
  default: vi.fn(() => mockAuthValue),
}));

vi.mock('./src/hooks/useAuth.js', () => ({
  useAuth: vi.fn(() => mockAuthValue),
  default: vi.fn(() => mockAuthValue),
}));

// Global mock для AuthContext
vi.mock('./src/context/AuthContext', () => ({
  AuthContext: React.createContext(mockAuthValue),
  AuthProvider: ({ children }) => children,
  useAuth: vi.fn(() => mockAuthValue),
}));

vi.mock('./src/context/AuthContext.jsx', () => ({
  AuthContext: React.createContext(mockAuthValue),
  AuthProvider: ({ children }) => children,
  useAuth: vi.fn(() => mockAuthValue),
}));

// Очистка после каждого теста
afterEach(() => {
  cleanup();
  vi.clearAllTimers(); // Очищаем все таймеры
  vi.useRealTimers(); // Возвращаем реальные таймеры
});

// Сброс моков перед каждым тестом
beforeEach(() => {
  vi.clearAllMocks();
});

// Mock глобальных объектов браузера
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock IntersectionObserver (для lazy loading, виртуализации)
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock window.matchMedia (для responsive тестов)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock window.scrollTo (для scroll тестов)
Object.defineProperty(window, 'scrollTo', {
  writable: true,
  value: vi.fn(),
});

// Создание тестового QueryClient с отключенными retry и оптимизациями
export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false, // Отключаем retry в тестах
        gcTime: Infinity, // Кэш не очищается автоматически
        staleTime: 0, // Данные всегда считаются устаревшими
      },
      mutations: {
        retry: false, // Отключаем retry для мутаций
      },
    },
    logger: {
      log: console.log,
      warn: console.warn,
      error: () => {}, // Подавляем ошибки в логах тестов
    },
  });
}
