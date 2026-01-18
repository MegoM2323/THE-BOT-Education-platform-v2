import React from 'react';
import { formatHomeworkText } from '../formatHomeworkText';

describe('formatHomeworkText', () => {
  describe('Simple text without special characters', () => {
    test('should return array with plain text', () => {
      const input = 'Простой текст';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result).toHaveLength(1);
      expect(result[0]).toBe('Простой текст');
    });

    test('should handle English text', () => {
      const input = 'Simple text';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result[0]).toBe('Simple text');
    });

    test('should handle text with punctuation', () => {
      const input = 'Text with punctuation: это текст!';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result[0]).toBe('Text with punctuation: это текст!');
    });
  });

  describe('Text with single newline', () => {
    test('should convert single newline to br element', () => {
      const input = 'Строка 1\nСтрока 2';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result).toHaveLength(3);
      expect(result[0]).toBe('Строка 1');
      expect(result[1].type).toBe('br');
      expect(result[2]).toBe('Строка 2');
    });

    test('should handle newline with spaces', () => {
      const input = 'Line 1  \n  Line 2';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result).toHaveLength(3);
      expect(result[1].type).toBe('br');
    });
  });

  describe('Text with multiple newlines', () => {
    test('should convert multiple newlines to multiple br elements', () => {
      const input = 'Строка 1\nСтрока 2\nСтрока 3';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result).toHaveLength(5);
      expect(result[0]).toBe('Строка 1');
      expect(result[1].type).toBe('br');
      expect(result[2]).toBe('Строка 2');
      expect(result[3].type).toBe('br');
      expect(result[4]).toBe('Строка 3');
    });

    test('should handle four lines', () => {
      const input = 'Line 1\nLine 2\nLine 3\nLine 4';
      const result = formatHomeworkText(input);

      expect(result).toHaveLength(7);
      const brCount = result.filter(el => React.isValidElement(el) && el.type === 'br').length;
      expect(brCount).toBe(3);
    });
  });

  describe('Text with URL without newlines', () => {
    test('should convert URL to anchor tag', () => {
      const input = 'Ссылка: https://example.com';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result).toHaveLength(2);
      expect(result[0]).toBe('Ссылка: ');
      expect(result[1].type).toBe('a');
      expect(result[1].props.href).toBe('https://example.com');
      expect(result[1].props.target).toBe('_blank');
      expect(result[1].props.rel).toBe('noopener noreferrer');
      expect(result[1].props.children).toBe('https://example.com');
    });

    test('should handle URL at the beginning', () => {
      const input = 'https://example.com это ссылка';
      const result = formatHomeworkText(input);

      expect(result[0].type).toBe('a');
      expect(result[1]).toBe(' это ссылка');
    });

    test('should handle URL at the end', () => {
      const input = 'Посетите https://example.com';
      const result = formatHomeworkText(input);

      expect(result[0]).toBe('Посетите ');
      expect(result[1].type).toBe('a');
    });

    test('should handle http URL', () => {
      const input = 'Ссылка: http://example.com';
      const result = formatHomeworkText(input);

      expect(result[1].type).toBe('a');
      expect(result[1].props.href).toBe('http://example.com');
    });
  });

  describe('Text with URL and newlines', () => {
    test('should handle URL with newlines before and after', () => {
      const input = 'Информация\nhttps://example.com\nПримечание';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result[0]).toBe('Информация');
      expect(result[1].type).toBe('br');
      expect(result[2].type).toBe('a');
      expect(result[2].props.href).toBe('https://example.com');
      expect(result[3].type).toBe('br');
      expect(result[4]).toBe('Примечание');
    });

    test('should handle URL with text on same line', () => {
      const input = 'Посетите https://example.com для деталей\nСпасибо';
      const result = formatHomeworkText(input);

      expect(result[0]).toBe('Посетите ');
      expect(result[1].type).toBe('a');
      expect(result[2]).toBe(' для деталей');
      expect(result[3].type).toBe('br');
      expect(result[4]).toBe('Спасибо');
    });
  });

  describe('Multiple URLs in one line', () => {
    test('should convert multiple URLs to multiple anchor tags', () => {
      const input = 'https://example.com и https://another.com';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      const linkCount = result.filter(el => React.isValidElement(el) && el.type === 'a').length;
      expect(linkCount).toBe(2);

      expect(result[0].type).toBe('a');
      expect(result[0].props.href).toBe('https://example.com');
      expect(result[1]).toBe(' и ');
      expect(result[2].type).toBe('a');
      expect(result[2].props.href).toBe('https://another.com');
    });

    test('should handle three URLs in one line', () => {
      const input = 'https://a.com https://b.com https://c.com';
      const result = formatHomeworkText(input);

      const linkCount = result.filter(el => React.isValidElement(el) && el.type === 'a').length;
      expect(linkCount).toBe(3);
    });
  });

  describe('Empty text and null', () => {
    test('should return null for null input', () => {
      const result = formatHomeworkText(null);
      expect(result).toBeNull();
    });

    test('should return null for empty string', () => {
      const result = formatHomeworkText('');
      expect(result).toBeNull();
    });

    test('should return null for whitespace only', () => {
      const result = formatHomeworkText('   ');
      expect(result).toBeNull();
    });

    test('should return null for tab characters only', () => {
      const result = formatHomeworkText('\t\t');
      expect(result).toBeNull();
    });

    test('should return null for undefined', () => {
      const result = formatHomeworkText(undefined);
      expect(result).toBeNull();
    });
  });

  describe('Multiple newlines in a row', () => {
    test('should preserve empty lines between text', () => {
      const input = 'Текст\n\n\nТекст';
      const result = formatHomeworkText(input);

      expect(Array.isArray(result)).toBe(true);
      expect(result[0]).toBe('Текст');
      expect(result[1].type).toBe('br');
      expect(result[2]).toBe('');
      expect(result[3].type).toBe('br');
      expect(result[4]).toBe('');
      expect(result[5].type).toBe('br');
      expect(result[6]).toBe('Текст');
    });

    test('should handle two consecutive newlines', () => {
      const input = 'Line 1\n\nLine 2';
      const result = formatHomeworkText(input);

      expect(result[0]).toBe('Line 1');
      expect(result[1].type).toBe('br');
      expect(result[2]).toBe('');
      expect(result[3].type).toBe('br');
      expect(result[4]).toBe('Line 2');
    });
  });

  describe('Complex scenarios', () => {
    test('should handle URL with query parameters', () => {
      const input = 'Ссылка: https://example.com/path?query=value&other=123';
      const result = formatHomeworkText(input);

      expect(result[1].type).toBe('a');
      expect(result[1].props.href).toBe('https://example.com/path?query=value&other=123');
    });

    test('should handle URL followed by punctuation on next line', () => {
      const input = 'https://example.com\n.';
      const result = formatHomeworkText(input);

      expect(result[0].type).toBe('a');
      expect(result[1].type).toBe('br');
      expect(result[2]).toBe('.');
    });

    test('should handle mixed content with URLs and multiple lines', () => {
      const input = 'Посмотрите: https://example.com\nДля подробности https://another.com\nСпасибо';
      const result = formatHomeworkText(input);

      const linkCount = result.filter(el => React.isValidElement(el) && el.type === 'a').length;
      expect(linkCount).toBe(2);

      const brCount = result.filter(el => React.isValidElement(el) && el.type === 'br').length;
      expect(brCount).toBe(2);
    });

    test('should not match URLs with broken protocol', () => {
      const input = 'Это не ссылка: ht tp://example.com';
      const result = formatHomeworkText(input);

      const linkCount = result.filter(el => React.isValidElement(el) && el.type === 'a').length;
      expect(linkCount).toBe(0);
      expect(result[0]).toContain('ht tp://example.com');
    });
  });

  describe('Edge cases', () => {
    test('should handle only URL without other text', () => {
      const input = 'https://example.com';
      const result = formatHomeworkText(input);

      expect(result).toHaveLength(1);
      expect(result[0].type).toBe('a');
    });

    test('should return null for only newline character (after trim it is empty)', () => {
      const input = '\n';
      const result = formatHomeworkText(input);

      // After trim(), "\n" becomes "", so function returns null
      expect(result).toBeNull();
    });

    test('should handle text with non-string type (number)', () => {
      const result = formatHomeworkText(123);
      expect(result).toBeNull();
    });

    test('should handle text with non-string type (object)', () => {
      const result = formatHomeworkText({ text: 'hello' });
      expect(result).toBeNull();
    });

    test('should generate unique keys for br elements', () => {
      const input = 'Line 1\nLine 2\nLine 3';
      const result = formatHomeworkText(input);

      const brElements = result.filter(el => React.isValidElement(el) && el.type === 'br');
      const brKeys = brElements.map(el => el.key);

      expect(new Set(brKeys).size).toBe(brKeys.length);
    });

    test('should generate unique keys for link elements', () => {
      const input = 'https://a.com https://b.com';
      const result = formatHomeworkText(input);

      const linkElements = result.filter(el => React.isValidElement(el) && el.type === 'a');
      const linkKeys = linkElements.map(el => el.key);

      expect(new Set(linkKeys).size).toBe(linkKeys.length);
    });
  });

  describe('Return type validation', () => {
    test('should return array of React elements and strings', () => {
      const input = 'Text\nhttps://example.com\nMore';
      const result = formatHomeworkText(input);

      result.forEach(item => {
        const isString = typeof item === 'string';
        const isReactElement = React.isValidElement(item);
        expect(isString || isReactElement).toBe(true);
      });
    });

    test('should not return null when text has valid content', () => {
      const inputs = [
        'a',
        ' a ',
        'https://example.com',
        'text\nmore',
      ];

      inputs.forEach(input => {
        const result = formatHomeworkText(input);
        expect(result).not.toBeNull();
        expect(Array.isArray(result)).toBe(true);
      });
    });
  });
});
