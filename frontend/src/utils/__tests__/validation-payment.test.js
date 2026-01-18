/**
 * Тесты для валидации URL редиректов платежной системы
 */

import { validatePaymentRedirectUrl } from '../validation.js';

describe('validatePaymentRedirectUrl', () => {
  describe('Valid URLs', () => {
    it('should accept valid YooKassa URL', () => {
      const url = 'https://yookassa.ru/checkout/payments/1234567890';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
      expect(result.error).toBeNull();
      expect(result.sanitizedUrl).toBe(url);
    });

    it('should accept valid YooMoney URL', () => {
      const url = 'https://yoomoney.ru/checkout/confirm?order=12345';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
      expect(result.error).toBeNull();
      expect(result.sanitizedUrl).toBe(url);
    });

    it('should accept valid Yandex Money URL', () => {
      const url = 'https://money.yandex.ru/payment/confirm';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
      expect(result.error).toBeNull();
      expect(result.sanitizedUrl).toBe(url);
    });

    it('should accept YooKassa subdomain', () => {
      const url = 'https://checkout.yookassa.ru/payments/1234';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
      expect(result.error).toBeNull();
    });
  });

  describe('Invalid URLs - Security Threats', () => {
    it('should reject HTTP (non-HTTPS) URLs', () => {
      const url = 'http://yookassa.ru/checkout/payments/1234';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('HTTPS');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject unauthorized domain', () => {
      const url = 'https://evil.com/fake-payment';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не разрешен');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject phishing domain mimicking YooKassa', () => {
      const url = 'https://yookassa.ru.evil.com/checkout';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не разрешен');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject similar-looking domain', () => {
      const url = 'https://yookassa-ru.com/checkout';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не разрешен');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject localhost URL', () => {
      const url = 'https://localhost:8080/fake-payment';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не разрешен');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject IP address URL', () => {
      const url = 'https://192.168.1.100/payment';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не разрешен');
      expect(result.sanitizedUrl).toBeNull();
    });
  });

  describe('Invalid URL Formats', () => {
    it('should reject empty string', () => {
      const result = validatePaymentRedirectUrl('');

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не указан');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject null', () => {
      const result = validatePaymentRedirectUrl(null);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не указан');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject undefined', () => {
      const result = validatePaymentRedirectUrl(undefined);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не указан');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject malformed URL', () => {
      const result = validatePaymentRedirectUrl('not-a-valid-url');

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('Неверный формат URL');
      expect(result.sanitizedUrl).toBeNull();
    });

    it('should reject non-string input', () => {
      const result = validatePaymentRedirectUrl(12345);

      expect(result.isValid).toBe(false);
      expect(result.error).toContain('не указан');
      expect(result.sanitizedUrl).toBeNull();
    });
  });

  describe('Edge Cases', () => {
    it('should normalize URL and return sanitized version', () => {
      const url = 'https://yookassa.ru/checkout?param1=value1&param2=value2';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
      expect(result.sanitizedUrl).toBeTruthy();
    });

    it('should handle URL with fragment', () => {
      const url = 'https://yookassa.ru/checkout#section';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
      expect(result.sanitizedUrl).toBe(url);
    });

    it('should be case-insensitive for domain matching', () => {
      const url = 'https://YooKassa.RU/checkout';
      const result = validatePaymentRedirectUrl(url);

      expect(result.isValid).toBe(true);
    });
  });
});
