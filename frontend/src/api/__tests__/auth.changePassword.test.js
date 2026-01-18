import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { changePassword } from '../auth.js';
import apiClient from '../client.js';

vi.mock('../client.js');

describe('auth.js - changePassword function', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    console.error.mockRestore();
  });

  describe('Scenario 1: API request - correct endpoint and method', () => {
    it('should send POST request to /auth/change-password endpoint', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('oldPassword', 'newPassword');

      expect(apiClient.post).toHaveBeenCalledWith(
        '/auth/change-password',
        expect.any(Object),
        expect.any(Object)
      );
    });

    it('should use correct endpoint path', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('oldPassword', 'newPassword');

      const calls = apiClient.post.mock.calls;
      expect(calls[0][0]).toBe('/auth/change-password');
    });
  });

  describe('Scenario 2: API request - correct payload structure', () => {
    it('should transmit old_password field', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('oldPassword123', 'newPassword123');

      const payload = apiClient.post.mock.calls[0][1];
      expect(payload).toHaveProperty('old_password');
      expect(payload.old_password).toBe('oldPassword123');
    });

    it('should transmit new_password field', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('oldPassword123', 'newPassword123');

      const payload = apiClient.post.mock.calls[0][1];
      expect(payload).toHaveProperty('new_password');
      expect(payload.new_password).toBe('newPassword123');
    });

    it('should use snake_case field names for API payload', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('oldPass', 'newPass');

      const payload = apiClient.post.mock.calls[0][1];
      expect(payload).toEqual({
        old_password: 'oldPass',
        new_password: 'newPass',
      });
    });

    it('should not include any other fields in payload', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('oldPassword', 'newPassword');

      const payload = apiClient.post.mock.calls[0][1];
      const keys = Object.keys(payload);
      expect(keys.length).toBe(2);
      expect(keys).toContain('old_password');
      expect(keys).toContain('new_password');
    });
  });

  describe('Scenario 3: Error handling - 401 Unauthorized', () => {
    it('should handle 401 error (incorrect current password)', async () => {
      const error = new Error('Unauthorized');
      error.status = 401;
      apiClient.post.mockRejectedValue(error);

      await expect(changePassword('wrongPassword', 'newPassword')).rejects.toThrow();
      expect(console.error).toHaveBeenCalled();
    });

    it('should log specific message for 401 error', async () => {
      const error = new Error('Unauthorized');
      error.status = 401;
      error.message = 'Current password is incorrect';
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('wrongPassword', 'newPassword');
      } catch {
        // Expected to throw
      }

      expect(console.error).toHaveBeenCalled();
      const errorLog = console.error.mock.calls[0];
      // Should log something about password, not expose actual password
      expect(typeof errorLog[0]).toBe('string');
    });
  });

  describe('Scenario 4: Error handling - 400 Bad Request', () => {
    it('should handle 400 error (weak password or same as old)', async () => {
      const error = new Error('Bad Request');
      error.status = 400;
      apiClient.post.mockRejectedValue(error);

      await expect(changePassword('password', 'password')).rejects.toThrow();
      expect(console.error).toHaveBeenCalled();
    });

    it('should log specific message for 400 error', async () => {
      const error = new Error('Bad Request');
      error.status = 400;
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('oldPass', 'samePass');
      } catch {
        // Expected to throw
      }

      expect(console.error).toHaveBeenCalled();
      const errorLog = console.error.mock.calls[0];
      expect(errorLog[0]).toContain('password');
    });
  });

  describe('Scenario 5: Error logging - never logs passwords', () => {
    it('should not log old password in error message', async () => {
      const error = new Error('Unauthorized');
      error.status = 401;
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('secretOldPassword', 'newPassword');
      } catch {
        // Expected to throw
      }

      const allErrorLogs = console.error.mock.calls;
      const errorLogString = JSON.stringify(allErrorLogs);
      expect(errorLogString).not.toContain('secretOldPassword');
    });

    it('should not log new password in error message', async () => {
      const error = new Error('Bad Request');
      error.status = 400;
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('oldPassword', 'secretNewPassword');
      } catch {
        // Expected to throw
      }

      const allErrorLogs = console.error.mock.calls;
      const errorLogString = JSON.stringify(allErrorLogs);
      expect(errorLogString).not.toContain('secretNewPassword');
    });

    it('should log descriptive error message without exposing passwords', async () => {
      const error = new Error('Unauthorized');
      error.status = 401;
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('password1', 'password2');
      } catch {
        // Expected to throw
      }

      expect(console.error).toHaveBeenCalled();
      const logs = console.error.mock.calls;
      expect(logs[0][0]).toBeTruthy();
      expect(typeof logs[0][0]).toBe('string');

      // Verify the log doesn't contain obvious passwords
      const logContent = logs[0][0].toLowerCase();
      expect(logContent).not.toContain('password1');
      expect(logContent).not.toContain('password2');
    });

    it('should not include passwords in error object logged', async () => {
      const error = new Error('Server error');
      error.status = 500;
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('oldSecret', 'newSecret');
      } catch {
        // Expected to throw
      }

      const allLogs = JSON.stringify(console.error.mock.calls);
      expect(allLogs).not.toContain('oldSecret');
      expect(allLogs).not.toContain('newSecret');
    });
  });

  describe('Scenario 6: Error handling - various status codes', () => {
    it('should handle 500 server error', async () => {
      const error = new Error('Server error');
      error.status = 500;
      apiClient.post.mockRejectedValue(error);

      await expect(changePassword('old', 'new')).rejects.toThrow();
      expect(console.error).toHaveBeenCalled();
    });

    it('should handle network errors', async () => {
      const error = new Error('Network error');
      apiClient.post.mockRejectedValue(error);

      await expect(changePassword('old', 'new')).rejects.toThrow();
      expect(console.error).toHaveBeenCalled();
    });

    it('should not log AbortError (cancelled requests)', async () => {
      const error = new Error('Aborted');
      error.name = 'AbortError';
      apiClient.post.mockRejectedValue(error);

      try {
        await changePassword('old', 'new');
      } catch {
        // Expected to throw
      }

      // AbortError should not be logged
      expect(console.error).not.toHaveBeenCalled();
    });
  });

  describe('Scenario 7: Options passing', () => {
    it('should pass options object to apiClient.post', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      const options = { timeout: 5000 };
      await changePassword('old', 'new', options);

      expect(apiClient.post).toHaveBeenCalledWith(
        expect.any(String),
        expect.any(Object),
        expect.objectContaining(options)
      );
    });

    it('should merge with default options', async () => {
      apiClient.post.mockResolvedValue({ message: 'success' });

      await changePassword('old', 'new', { customOption: 'value' });

      const options = apiClient.post.mock.calls[0][2];
      expect(options).toHaveProperty('customOption', 'value');
    });

    it('should return successful response', async () => {
      const mockResponse = { message: 'Password changed successfully' };
      apiClient.post.mockResolvedValue(mockResponse);

      const result = await changePassword('old', 'new');

      expect(result).toEqual(mockResponse);
    });
  });
});
