/**
 * Security tests for 401 Unauthorized handling
 * Verifies protection against Open Redirect vulnerabilities (FB003)
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { apiClient, clearCSRFToken } from '../client.js';

describe('API Client 401 Security - Open Redirect Protection', () => {
  let dispatchEventSpy;
  let consoleErrorSpy;

  beforeEach(() => {
    // Mock fetch
    global.fetch = vi.fn();

    // Spy on window.dispatchEvent
    dispatchEventSpy = vi.spyOn(window, 'dispatchEvent');

    // Spy on console.error to verify no errors thrown
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    // Clear CSRF token
    clearCSRFToken();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should NOT use window.location.href for redirects', async () => {
    const originalHref = window.location.href;

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

    // Verify window.location.href was NOT modified
    expect(window.location.href).toBe(originalHref);
  });

  it('should NOT accept redirect URL from response headers', async () => {
    // Mock 401 response with malicious redirect header
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({
        'content-type': 'application/json',
        'Location': 'https://evil.com/phishing',
        'X-Redirect-To': 'https://evil.com/phishing',
      }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    const originalHref = window.location.href;

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow('Unauthorized');

    // Verify window.location.href was NOT modified
    expect(window.location.href).toBe(originalHref);

    // Verify event was dispatched (safe approach)
    expect(dispatchEventSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'auth:unauthorized',
      })
    );
  });

  it('should NOT accept redirect URL from response body', async () => {
    // Mock 401 response with malicious redirect in body
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({
        error: { message: 'Unauthorized' },
        redirect_url: 'https://evil.com/phishing',
        location: 'javascript:alert(document.cookie)',
      }),
    });

    const originalHref = window.location.href;

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow('Unauthorized');

    // Verify window.location.href was NOT modified
    expect(window.location.href).toBe(originalHref);
  });

  it('should only dispatch event with controlled data', async () => {
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

    // Verify event detail contains only safe, controlled data
    const event = dispatchEventSpy.mock.calls[0][0];
    expect(event.type).toBe('auth:unauthorized');
    expect(event.detail.status).toBe(401);
    expect(event.detail.method).toBe(testMethod);
    expect(event.detail.url).toContain(testUrl);

    // Verify NO redirect URL in event detail
    expect(event.detail.redirect).toBeUndefined();
    expect(event.detail.location).toBeUndefined();
    expect(event.detail.redirect_url).toBeUndefined();
  });

  it('should NOT use any URL parameters in redirect', async () => {
    // Mock 401 response
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({ 'content-type': 'application/json' }),
      json: async () => ({ error: { message: 'Unauthorized' } }),
    });

    const originalHref = window.location.href;

    // Attempt API call with malicious redirect in URL
    await expect(
      apiClient.get('/api/test?redirect=https://evil.com')
    ).rejects.toThrow();

    // Verify window.location.href was NOT modified
    expect(window.location.href).toBe(originalHref);
  });

  it('should handle 401 consistently across all HTTP methods', async () => {
    const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];
    const originalHref = window.location.href;

    for (const method of methods) {
      // Mock 401 response
      global.fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({ error: { message: 'Unauthorized' } }),
      });

      // Call appropriate method
      const promise =
        method === 'GET' || method === 'DELETE'
          ? apiClient[method.toLowerCase()]('/api/test')
          : apiClient[method.toLowerCase()]('/api/test', { data: 'test' });

      await expect(promise).rejects.toThrow('Unauthorized');

      // Verify window.location.href was NOT modified
      expect(window.location.href).toBe(originalHref);
    }
  });

  it('should use event-based navigation instead of direct redirect', async () => {
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

    // Verify event-based approach is used (CustomEvent)
    const event = dispatchEventSpy.mock.calls[0][0];
    expect(event).toBeInstanceOf(CustomEvent);
    expect(event.type).toBe('auth:unauthorized');

    // Verify this allows AuthContext to handle navigation safely
    expect(event.detail).toBeDefined();
    expect(event.detail.status).toBe(401);
  });

  it('should NOT expose sensitive data in event detail', async () => {
    // Mock 401 response with sensitive headers
    global.fetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      headers: new Headers({
        'content-type': 'application/json',
        'Set-Cookie': 'session=abc123; HttpOnly',
        'X-CSRF-Token': 'secret-token',
      }),
      json: async () => ({
        error: { message: 'Unauthorized' },
        session_id: 'abc123',
        user_data: { email: 'user@example.com' },
      }),
    });

    // Attempt API call
    await expect(apiClient.get('/api/test')).rejects.toThrow();

    // Verify event detail contains only non-sensitive data
    const event = dispatchEventSpy.mock.calls[0][0];
    expect(event.detail).toBeDefined();
    expect(event.detail.session_id).toBeUndefined();
    expect(event.detail.user_data).toBeUndefined();
    expect(event.detail.headers).toBeUndefined();
    expect(event.detail.csrf_token).toBeUndefined();
  });
});
