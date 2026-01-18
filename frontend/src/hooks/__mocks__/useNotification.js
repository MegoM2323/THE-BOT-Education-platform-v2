// Manual mock for useNotification hook
// This avoids importing NotificationContext.jsx which causes issues in tests
import { vi } from 'vitest';

const mockNotification = {
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

export const useNotification = vi.fn(() => mockNotification);

export default useNotification;
