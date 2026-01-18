import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import TelegramLinkCard from '../common/TelegramLinkCard.jsx';
import { useTelegram } from '../../hooks/useTelegram.js';

vi.mock('../../hooks/useTelegram.js');

describe('TelegramLinkCard', () => {
  const mockUseTelegram = {
    linkStatus: null,
    token: null,
    botUsername: 'TestBot',
    loading: false,
    error: null,
    generateToken: vi.fn(),
    unlinkAccount: vi.fn(),
    refetchLinkStatus: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useTelegram).mockReturnValue(mockUseTelegram);
  });

  describe('Рендер состояния "не привязан"', () => {
    it('должен отображать кнопку привязки когда Telegram не привязан', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/Привязать Telegram/i)).toBeInTheDocument();
    });

    it('должен отображать информацию о преимуществах привязки', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/уведомлений/i)).toBeInTheDocument();
    });
  });

  describe('Рендер состояния "привязан"', () => {
    it('должен отображать информацию о привязанном аккаунте', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: {
          linked: true,
          telegram: {
            telegram_id: 123456789,
            username: 'testuser',
          },
        },
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/@testuser/)).toBeInTheDocument();
      expect(screen.getByText(/Отвязать Telegram/i)).toBeInTheDocument();
    });

    it('должен показывать статус подписки на уведомления', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: {
          linked: true,
          telegram: {
            telegram_id: 123456789,
            username: 'testuser',
          },
        },
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/Привязан/)).toBeInTheDocument();
    });

    it('должен показывать когда пользователь отписался от уведомлений', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: {
          linked: false,
        },
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/Не привязан/)).toBeInTheDocument();
    });
  });

  describe('Генерация токена привязки', () => {
    it('должен открывать модальное окно при клике на кнопку привязки', async () => {
      const generateTokenMock = vi.fn().mockResolvedValue(undefined);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
        generateToken: generateTokenMock,
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByText(/Привязать Telegram/i);
      fireEvent.click(linkButton);

      await waitFor(() => {
        expect(generateTokenMock).toHaveBeenCalledTimes(1);
      });
    });

    it('должен отображать токен и инструкции после генерации', async () => {
      const mockToken = 'test-token-12345';
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
        token: mockToken,
        botUsername: 'TestBot',
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/Telegram Уведомления/i)).toBeInTheDocument();
    });

    it('должен показывать ссылку на бота Telegram', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
        token: 'test-token',
        botUsername: 'TestBot',
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(/Привязать Telegram/i)).toBeInTheDocument();
    });
  });

  describe('Отвязка аккаунта', () => {
    it('должен показывать подтверждение перед отвязкой', async () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: {
          linked: true,
          telegram: {
            telegram_id: 123456789,
            username: 'testuser',
          },
        },
      });

      render(<TelegramLinkCard />);

      const unlinkButton = screen.getByText(/Отвязать Telegram/i);
      fireEvent.click(unlinkButton);

      await waitFor(() => {
        expect(screen.getByText(/Вы действительно хотите отвязать/i)).toBeInTheDocument();
      });
    });

    it('должен вызывать unlinkAccount при подтверждении', async () => {
      const unlinkAccountMock = vi.fn().mockResolvedValue(undefined);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: {
          linked: true,
          telegram: {
            telegram_id: 123456789,
            username: 'testuser',
          },
        },
        unlinkAccount: unlinkAccountMock,
      });

      render(<TelegramLinkCard />);

      const unlinkButton = screen.getByText(/Отвязать Telegram/i);
      fireEvent.click(unlinkButton);

      const confirmButton = await screen.findByRole('button', { name: /^Отвязать$/i });
      fireEvent.click(confirmButton);

      await waitFor(() => {
        expect(unlinkAccountMock).toHaveBeenCalledTimes(1);
      });
    });

    it('не должен отвязывать при отмене', async () => {
      const unlinkAccountMock = vi.fn().mockResolvedValue(undefined);
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: {
          linked: true,
          telegram: {
            telegram_id: 123456789,
            username: 'testuser',
          },
        },
        unlinkAccount: unlinkAccountMock,
      });

      render(<TelegramLinkCard />);

      const unlinkButton = screen.getByText(/Отвязать Telegram/i);
      fireEvent.click(unlinkButton);

      const cancelButton = await screen.findByRole('button', { name: /Отмена/i });
      fireEvent.click(cancelButton);

      expect(unlinkAccountMock).not.toHaveBeenCalled();
    });
  });

  describe('Состояние загрузки', () => {
    it('должен показывать индикатор загрузки', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        loading: true,
        linkStatus: null,
      });

      render(<TelegramLinkCard />);

      expect(screen.getByRole('status')).toBeInTheDocument();
    });

    it('должен блокировать кнопки во время загрузки', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
        loading: true,
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      expect(linkButton).toBeDisabled();
    });
  });

  describe('Обработка ошибок', () => {
    it('должен отображать ошибки', () => {
      const errorMessage = 'Ошибка при загрузке';
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        error: errorMessage,
        linkStatus: { linked: false },
      });

      render(<TelegramLinkCard />);

      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('должен иметь правильные aria-атрибуты для кнопок', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      expect(linkButton).toBeInTheDocument();
    });

    it('должен быть доступен с клавиатуры', () => {
      vi.mocked(useTelegram).mockReturnValue({
        ...mockUseTelegram,
        linkStatus: { linked: false },
      });

      render(<TelegramLinkCard />);

      const linkButton = screen.getByRole('button', { name: /Привязать Telegram/i });
      linkButton.focus();
      expect(linkButton).toHaveFocus();
    });
  });
});
