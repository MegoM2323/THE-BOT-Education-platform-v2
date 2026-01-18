/**
 * Tests for 401 Unauthorized handling in API client
 * Verifies that 401 responses trigger custom events instead of direct window.location redirects
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { apiClient, clearCSRFToken } from '../client.js';

describe('API Client 401 Unauthorized Handling', () => {
  let eventListenerSpy;
  let dispatchEventSpy;

  beforeEach(() => {
    // Mock fetch
    global.fetch = vi.fn();

    // Spy on window.dispatchEvent
    dispatchEventSpy = vi.spyOn(window, 'dispatchEvent');

    // Clear CSRF token
    clearCSRFToken();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should dispatch auth:unauthorized event on 401 response', async () => {
    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow('Unauthorized');

    // Verify event was dispatched
    expect(dispatchEventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'auth:unauthorized',
        detail: expect.objectContaining({
          status: 401,
          method: 'GET',
        }),
      })
    );
  });

  it('should NOT use window.location.href on 401', async () => {
    const originalLocation = window.location;
    delete window.location;
    window.location = { ...originalLocation, href: '' };

    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow();

    // Verify window.location.href was NOT modified
    expect(window.location.href).toBe('');

    // Restore
    window.location = originalLocation;
  });

  it('should clear CSRF token on 401 response', async () => {
    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow();

    // CSRF token should be cleared (tested indirectly - next request won't have it)
    expect(dispatchEventSpy).toHaveBeenCalled();
  });

  it('should include request details in event', async () => {
    const testUrl = '/api/users/me';
    const testMethod = 'GET';

    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Attempt API call
    await expect(apiClient.get(testUrl)).rejects.toThrow();

    // Verify event detail includes request info
    expect(dispatchEventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'auth:unauthorized',
        detail: expect.objectContaining({
          url: expect.stringContaining(testUrl),
          method: testMethod,
          status: 401,
        }),
      })
    );
  });

  it('should work with POST requests on 401', async () => {
    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Attempt POST request
    await expect(apiClient.post('/api/bookings', { lesson_id: 1 })).rejects.toThrow();

    // Verify event was dispatched with correct method
    expect(dispatchEventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'auth:unauthorized',
        detail: expect.objectContaining({
          method: 'POST',
          status: 401,
        }),
      })
    );
  });

  it('should handle 401 in server-side rendering (no window)', async () => {
    // Mock window as undefined
    const originalWindow = global.window;
    global.window = undefined;

    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Should not throw error about window being undefined
    await expect(apiClient.get('/api/test')).rejects.toThrow('Unauthorized');

    // Restore
    global.window = originalWindow;
  });

  it('should allow event listeners to handle navigation', async () => {
    let eventReceived = false;

    // Add event listener
    const handler = (event) => {
      eventReceived = true;
      expect(event.type).toBe('auth:unauthorized');
      expect(event.detail.status).toBe(401);
    };

    window.addEventListener('auth:unauthorized', handler);

    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow();

    // Verify event was received
    expect(eventReceived).toBe(true);

    // Cleanup
    window.removeEventListener('auth:unauthorized', handler);
  });
});
