import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import TelegramLinkCard from '../TelegramLinkCard.jsx';
import * as browserDetection from '../../../utils/browserDetection.js';
import { useTelegram } from '../../../hooks/useTelegram.js';

// Mock browser detection utilities
vi.mock('../../../utils/browserDetection.js');

// Mock the useTelegram hook
vi.mock('../../../hooks/useTelegram.js');

// Mock child components
vi.mock('../Button.jsx', () => ({
  default: ({ onClick, disabled, loading, variant, children, fullWidth }) => (
    <button
      onClick={onClick}
      disabled={disabled}
      data-loading={loading}
      data-variant={variant}
      data-fullwidth={fullWidth}
    >
      {children}
    </button>
  ),
}));

vi.mock('../Modal.jsx', () => ({
  default: ({ isOpen, onClose, title, children }) => (
    isOpen ? (
      <div data-testid="modal" role="dialog">
        <h2>{title}</h2>
        {children}
        <button onClick={onClose}>Закрыть</button>
      </div>
    ) : null
  ),
}));

vi.mock('../ConfirmModal.jsx', () => ({
  default: ({ isOpen, onConfirm, title, message }) => (
    isOpen ? (
      <div data-testid="confirm-modal" role="dialog">
        <h2>{title}</h2>
        <p>{message}</p>
        <button onClick={onConfirm}>Отвязать</button>
        <button>Отмена</button>
      </div>
    ) : null
  ),
}));

vi.mock('../Spinner.jsx', () => ({
  default: ({ size }) => <div data-testid="spinner" data-size={size}>Loading...</div>,
}));

// Default mock values for useTelegram hook
const createMockUseTelegramValue = (overrides = {}) => ({
  linkStatus: { linked: false },
  token: 'test-token-12345',
  botUsername: 'test_bot',
  loading: false,
  error: null,
  generateToken: vi.fn(async () => ({
    token: 'test-token-12345',
    bot_username: 'test_bot',
  })),
  unlinkAccount: vi.fn(async () => ({})),
  refetchLinkStatus: vi.fn(),
  ...overrides,
});

describe('TelegramLinkCard - Safari Browser Behavior', () => {
  let mockUseTelegramValue;

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseTelegramValue = createMockUseTelegramValue();
    vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

    // Mock window.location
    delete window.location;
    window.location = { href: '' };

    // Mock navigator.clipboard
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn(async (text) => text),
      },
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  describe('Safari iOS - Поведение привязки (Suite 1)', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
    });

    it('Safari iOS показывает "Вы будете перенаправлены автоматически" в инструкциях', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const instructions = screen.getByText(/Вы будете перенаправлены автоматически/i);
        expect(instructions).toBeInTheDocument();
      });
    });

    it('Кнопка "Открыть бота в Telegram" имеет onClick handler (НЕ является anchor tag)', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton).toBeInTheDocument();
        expect(botButton.closest('a')).not.toBeInTheDocument();
      });
    });

    it('onClick устанавливает window.location.href с URL бота', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton).toBeInTheDocument();
      });

      const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
      fireEvent.click(botButton);

      expect(window.location.href).toBe('https://t.me/test_bot?start=test-token-12345');
    });

    it('Нет элемента <a> для Safari iOS (прямая кнопка)', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton).toBeInTheDocument();
      });

      const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
      expect(botButton.tagName).toBe('BUTTON');
      expect(botButton.tagName).not.toBe('A');
    });

    it('Текст кнопки одинаковый для всех браузеров', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton.textContent).toBe('Открыть бота в Telegram');
      });
    });
  });

  describe('Другие браузеры (Chrome, Firefox) - Поведение (Suite 2)', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
    });

    it('НЕ показывает "Вы будете перенаправлены автоматически"', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const instructions = screen.queryByText(/Вы будете перенаправлены автоматически/i);
        expect(instructions).not.toBeInTheDocument();
      });
    });

    it('Показывает стандартное сообщение "Нажмите на кнопку ниже или скопируйте ссылку"', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const instructions = screen.getByText(/Нажмите на кнопку ниже или скопируйте ссылку/i);
        expect(instructions).toBeInTheDocument();
      });
    });

    it('Кнопка обернута в <a> tag с href', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton).toBeInTheDocument();
      });

      const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
      const anchorTag = botButton.closest('a');
      expect(anchorTag).toBeInTheDocument();
      expect(anchorTag).toHaveAttribute('href', 'https://t.me/test_bot?start=test-token-12345');
    });

    it('Anchor tag имеет target="_blank"', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton).toBeInTheDocument();
      });

      const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
      const anchorTag = botButton.closest('a');
      expect(anchorTag).toHaveAttribute('target', '_blank');
    });

    it('Anchor tag имеет rel="noopener noreferrer"', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
        expect(botButton).toBeInTheDocument();
      });

      const botButton = screen.getByRole('button', { name: /Открыть бота в Telegram/i });
      const anchorTag = botButton.closest('a');
      expect(anchorTag).toHaveAttribute('rel', 'noopener noreferrer');
    });
  });

  describe('Прочее функционирование (Suite 3)', () => {
    it('Функция копирования ссылки работает на всех браузерах (Safari)', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const copyButton = screen.getByRole('button', { name: /Копировать/i });
        expect(copyButton).toBeInTheDocument();
      });

      const copyButton = screen.getByRole('button', { name: /Копировать/i });
      await userEvent.click(copyButton);

      await waitFor(() => {
        expect(navigator.clipboard.writeText).toHaveBeenCalledWith('https://t.me/test_bot?start=test-token-12345');
      });
    });

    it('Функция копирования ссылки работает на всех браузерах (Chrome)', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const copyButton = screen.getByRole('button', { name: /Копировать/i });
        expect(copyButton).toBeInTheDocument();
      });

      const copyButton = screen.getByRole('button', { name: /Копировать/i });
      await userEvent.click(copyButton);

      await waitFor(() => {
        expect(navigator.clipboard.writeText).toHaveBeenCalledWith('https://t.me/test_bot?start=test-token-12345');
      });
    });

    it('Обработка ошибок при отсутствии linkUrl', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: null,
        botUsername: null,
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      expect(linkButton).toBeInTheDocument();

      // Копирование не должно работать, если нет URL
      const urlInput = screen.queryByDisplayValue(/https:\/\/t\.me/i);
      expect(urlInput).not.toBeInTheDocument();
    });

    it('Модальное окно закрывается правильно', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const modal = screen.getByTestId('modal');
        expect(modal).toBeInTheDocument();
      });

      const closeButton = screen.getByRole('button', { name: /Закрыть/i });
      await userEvent.click(closeButton);

      await waitFor(() => {
        const modal = screen.queryByTestId('modal');
        expect(modal).not.toBeInTheDocument();
      });
    });
  });

  describe('Проверка браузер-детектора', () => {
    it('isSafariOnIOS вызывается при открытии модального окна (Safari)', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        expect(vi.mocked(browserDetection.isSafariOnIOS)).toHaveBeenCalled();
      });
    });

    it('isSafariOnIOS вызывается при открытии модального окна (Chrome)', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'test-token-12345',
        botUsername: 'test_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        expect(vi.mocked(browserDetection.isSafariOnIOS)).toHaveBeenCalled();
      });
    });
  });

  describe('URL ссылки на бота', () => {
    it('Safari iOS: URL содержит правильный формат t.me/bot?start=token', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'custom-token-abc123',
        botUsername: 'my_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const urlInput = screen.getByDisplayValue(/https:\/\/t\.me\/my_bot\?start=custom-token-abc123/i);
        expect(urlInput).toBeInTheDocument();
      });
    });

    it('Chrome: URL содержит правильный формат t.me/bot?start=token', async () => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegramValue,
        linkStatus: { linked: false },
        token: 'custom-token-xyz789',
        botUsername: 'another_bot',
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      await userEvent.click(linkButton);

      await waitFor(() => {
        const urlInput = screen.getByDisplayValue(/https:\/\/t\.me\/another_bot\?start=custom-token-xyz789/i);
        expect(urlInput).toBeInTheDocument();
      });
    });
  });
});
