/**
 * Comprehensive test suite for browser detection utility
 * Tests isSafari(), isIOS(), isSafariOnIOS(), and openTelegramBot()
 *
 * This is RED phase - tests are written before implementation
 * Expected: All tests will FAIL initially (that's correct!)
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

/**
 * User agent strings for different browsers and platforms
 * Used to mock navigator.userAgent in tests
 */
const USER_AGENTS = {
  // Safari macOS
  SAFARI_MACOS: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Version/14.1.1 Safari/537.36',

  // Safari iOS - iPhone
  SAFARI_IPHONE: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1',

  // Safari iOS - iPad
  SAFARI_IPAD: 'Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1',

  // Chrome macOS
  CHROME_MACOS: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36',

  // Chrome iOS (CriOS)
  CHROME_IOS: 'Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/96.0.4664.45 Mobile/15E148 Safari/604.1',

  // Firefox Desktop
  FIREFOX_DESKTOP: 'Mozilla/5.0 (X11; Linux x86_64; rv:95.0) Gecko/20100101 Firefox/95.0',

  // Android Chrome
  ANDROID_CHROME: 'Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Mobile Safari/537.36',
};

/**
 * Store original navigator properties for restoration
 */
let originalNavigator;

/**
 * Helper: Create a mock navigator object with custom properties
 */
function createMockNavigator(userAgent, vendor = '', platform = 'MacIntel', maxTouchPoints = 0) {
  return {
    userAgent,
    vendor,
    platform,
    maxTouchPoints,
  };
}

/**
 * Helper: Mock navigator and window.open
 */
function mockBrowserEnvironment(userAgent, vendor = '', platform = 'MacIntel', maxTouchPoints = 0) {
  // Save original navigator for restoration
  originalNavigator = global.navigator;

  // Create mock navigator
  const mockNav = createMockNavigator(userAgent, vendor, platform, maxTouchPoints);
  Object.defineProperty(global, 'navigator', {
    value: mockNav,
    writable: true,
    configurable: true,
  });

  // Mock window.open
  global.window.open = vi.fn();

  // Mock window.location.href
  global.window.location = {
    href: '',
  };
}

/**
 * Helper: Restore original navigator after test
 */
function restoreBrowserEnvironment() {
  Object.defineProperty(global, 'navigator', {
    value: originalNavigator,
    configurable: true,
  });
  vi.clearAllMocks();
}

describe('isSafari()', () => {
  afterEach(() => {
    restoreBrowserEnvironment();
  });

  it('should return true for Safari on macOS', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_MACOS, 'Apple');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(true);
  });

  it('should return true for Safari on iOS (iPhone)', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPHONE, 'Apple');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(true);
  });

  it('should return true for Safari on iOS (iPad)', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPAD, 'Apple');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(true);
  });

  it('should return false for Chrome on macOS', () => {
    mockBrowserEnvironment(USER_AGENTS.CHROME_MACOS, 'Google');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(false);
  });

  it('should return false for Chrome on iOS (CriOS)', () => {
    mockBrowserEnvironment(USER_AGENTS.CHROME_IOS, '');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(false);
  });

  it('should return false for Firefox on desktop', () => {
    mockBrowserEnvironment(USER_AGENTS.FIREFOX_DESKTOP, '');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(false);
  });

  it('should return false for Android Chrome', () => {
    mockBrowserEnvironment(USER_AGENTS.ANDROID_CHROME, '');
    const { isSafari } = require('../browserDetection');
    expect(isSafari()).toBe(false);
  });
});

describe('isIOS()', () => {
  afterEach(() => {
    restoreBrowserEnvironment();
  });

  it('should return true for Safari on iPhone', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPHONE, 'Apple');
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(true);
  });

  it('should return true for Safari on iPad', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPAD, 'Apple');
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(true);
  });

  it('should return true for Chrome on iPhone (CriOS)', () => {
    mockBrowserEnvironment(USER_AGENTS.CHROME_IOS, '');
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(true);
  });

  it('should return true for iPad iOS 13+ (MacIntel + touch capability)', () => {
    // iPad OS 13+ reports as MacIntel with touch capability
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPAD, 'Apple', 'MacIntel', 5);
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(true);
  });

  it('should return false for Safari on macOS', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_MACOS, 'Apple', 'MacIntel', 0);
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(false);
  });

  it('should return false for Chrome on macOS', () => {
    mockBrowserEnvironment(USER_AGENTS.CHROME_MACOS, 'Google', 'MacIntel', 0);
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(false);
  });

  it('should return false for Windows desktop', () => {
    mockBrowserEnvironment(USER_AGENTS.FIREFOX_DESKTOP, '', 'Win32', 0);
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(false);
  });

  it('should return false for Android Chrome', () => {
    mockBrowserEnvironment(USER_AGENTS.ANDROID_CHROME, '', 'Linux armv7l', 10);
    const { isIOS } = require('../browserDetection');
    expect(isIOS()).toBe(false);
  });
});

describe('isSafariOnIOS()', () => {
  afterEach(() => {
    restoreBrowserEnvironment();
  });

  it('should return true for Safari on iPhone', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPHONE, 'Apple');
    const { isSafariOnIOS } = require('../browserDetection');
    expect(isSafariOnIOS()).toBe(true);
  });

  it('should return true for Safari on iPad', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_IPAD, 'Apple');
    const { isSafariOnIOS } = require('../browserDetection');
    expect(isSafariOnIOS()).toBe(true);
  });

  it('should return false for Safari on macOS', () => {
    mockBrowserEnvironment(USER_AGENTS.SAFARI_MACOS, 'Apple', 'MacIntel', 0);
    const { isSafariOnIOS } = require('../browserDetection');
    expect(isSafariOnIOS()).toBe(false);
  });

  it('should return false for Chrome on iOS (CriOS)', () => {
    mockBrowserEnvironment(USER_AGENTS.CHROME_IOS, '');
    const { isSafariOnIOS } = require('../browserDetection');
    expect(isSafariOnIOS()).toBe(false);
  });

  it('should return false for Chrome on macOS', () => {
    mockBrowserEnvironment(USER_AGENTS.CHROME_MACOS, 'Google');
    const { isSafariOnIOS } = require('../browserDetection');
    expect(isSafariOnIOS()).toBe(false);
  });

  it('should return false for Firefox', () => {
    mockBrowserEnvironment(USER_AGENTS.FIREFOX_DESKTOP, '');
    const { isSafariOnIOS } = require('../browserDetection');
    expect(isSafariOnIOS()).toBe(false);
  });
});

describe('openTelegramBot(botUrl)', () => {
  afterEach(() => {
    restoreBrowserEnvironment();
  });

  describe('Safari on iOS - redirect method', () => {
    beforeEach(() => {
      mockBrowserEnvironment(USER_AGENTS.SAFARI_IPHONE, 'Apple');
    });

    it('should use window.location.href for Safari on iOS', () => {
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      openTelegramBot(botUrl);

      expect(global.window.location.href).toBe(botUrl);
    });

    it('should return redirect object for Safari on iOS', () => {
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'redirect',
        redirected: true,
      });
    });

    it('should set location.href for iPad Safari too', () => {
      mockBrowserEnvironment(USER_AGENTS.SAFARI_IPAD, 'Apple');
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/another_bot';

      openTelegramBot(botUrl);

      expect(global.window.location.href).toBe(botUrl);
    });
  });

  describe('Safari on macOS - popup method', () => {
    beforeEach(() => {
      mockBrowserEnvironment(USER_AGENTS.SAFARI_MACOS, 'Apple', 'MacIntel', 0);
    });

    it('should use window.open for Safari on macOS', () => {
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      openTelegramBot(botUrl);

      expect(global.window.open).toHaveBeenCalledWith(botUrl, '_blank', 'noopener,noreferrer');
    });

    it('should return popup object for Safari on macOS (popup not blocked)', () => {
      global.window.open = vi.fn(() => ({})); // Mock successful popup
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'popup',
        blocked: false,
      });
    });

    it('should return blocked: true when popup is blocked (null)', () => {
      global.window.open = vi.fn(() => null); // Popup blocked
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'popup',
        blocked: true,
      });
    });

    it('should return blocked: true when popup is blocked (undefined)', () => {
      global.window.open = vi.fn(() => undefined); // Popup blocked
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'popup',
        blocked: true,
      });
    });
  });

  describe('Chrome - popup method', () => {
    beforeEach(() => {
      mockBrowserEnvironment(USER_AGENTS.CHROME_MACOS, 'Google');
    });

    it('should use window.open for Chrome on macOS', () => {
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      openTelegramBot(botUrl);

      expect(global.window.open).toHaveBeenCalledWith(botUrl, '_blank', 'noopener,noreferrer');
    });

    it('should return popup object for Chrome (popup not blocked)', () => {
      global.window.open = vi.fn(() => ({})); // Mock successful popup
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'popup',
        blocked: false,
      });
    });

    it('should return blocked: true when popup is blocked', () => {
      global.window.open = vi.fn(() => null);
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'popup',
        blocked: true,
      });
    });
  });

  describe('Chrome iOS - popup method', () => {
    beforeEach(() => {
      mockBrowserEnvironment(USER_AGENTS.CHROME_IOS, '');
    });

    it('should use window.open for Chrome on iOS', () => {
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      openTelegramBot(botUrl);

      expect(global.window.open).toHaveBeenCalledWith(botUrl, '_blank', 'noopener,noreferrer');
    });

    it('should return popup object for Chrome iOS (popup not blocked)', () => {
      global.window.open = vi.fn(() => ({}));
      const { openTelegramBot } = require('../browserDetection');
      const botUrl = 'https://t.me/test_bot';

      const result = openTelegramBot(botUrl);

      expect(result).toEqual({
        method: 'popup',
        blocked: false,
      });
    });
  });

  describe('Edge cases and error handling', () => {
    it('should handle empty URL', () => {
      mockBrowserEnvironment(USER_AGENTS.SAFARI_IPHONE, 'Apple');
      const { openTelegramBot } = require('../browserDetection');

      const result = openTelegramBot('');

      expect(result).toHaveProperty('method');
      expect(result).toHaveProperty('redirected', true);
    });

    it('should handle URL without protocol', () => {
      mockBrowserEnvironment(USER_AGENTS.CHROME_MACOS, 'Google');
      global.window.open = vi.fn(() => ({}));
      const { openTelegramBot } = require('../browserDetection');

      const result = openTelegramBot('t.me/test_bot');

      expect(global.window.open).toHaveBeenCalled();
      expect(result).toHaveProperty('method');
    });

    it('should handle very long URL', () => {
      mockBrowserEnvironment(USER_AGENTS.SAFARI_IPAD, 'Apple');
      const longUrl = 'https://t.me/test_bot?param=' + 'x'.repeat(1000);

      const { openTelegramBot } = require('../browserDetection');
      const result = openTelegramBot(longUrl);

      expect(global.window.location.href).toBe(longUrl);
      expect(result).toHaveProperty('redirected', true);
    });

    it('should handle rapid consecutive calls (Safari iOS)', () => {
      mockBrowserEnvironment(USER_AGENTS.SAFARI_IPHONE, 'Apple');
      const { openTelegramBot } = require('../browserDetection');

      const result1 = openTelegramBot('https://t.me/bot1');
      const result2 = openTelegramBot('https://t.me/bot2');

      expect(result1).toEqual({ method: 'redirect', redirected: true });
      expect(result2).toEqual({ method: 'redirect', redirected: true });
    });

    it('should handle rapid consecutive calls (Chrome)', () => {
      mockBrowserEnvironment(USER_AGENTS.CHROME_MACOS, 'Google');
      global.window.open = vi.fn(() => ({}));
      const { openTelegramBot } = require('../browserDetection');

      const result1 = openTelegramBot('https://t.me/bot1');
      const result2 = openTelegramBot('https://t.me/bot2');

      expect(global.window.open).toHaveBeenCalledTimes(2);
      expect(result1).toHaveProperty('method', 'popup');
      expect(result2).toHaveProperty('method', 'popup');
    });
  });
});
