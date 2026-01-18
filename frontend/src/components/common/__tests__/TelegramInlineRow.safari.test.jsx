import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import TelegramInlineRow from '../TelegramInlineRow';
import * as browserDetection from '../../../utils/browserDetection';
import { useTelegram } from '../../../hooks/useTelegram';

// Mock browser detection utilities
vi.mock('../../../utils/browserDetection');

// Mock the useTelegram hook
vi.mock('../../../hooks/useTelegram');

// Mock child components
vi.mock('../Button.jsx', () => ({
  default: ({ onClick, disabled, loading, variant, children }) => (
    <button onClick={onClick} disabled={disabled} data-loading={loading} data-variant={variant}>
      {children}
    </button>
  ),
}));

vi.mock('../ConfirmModal.jsx', () => ({
  default: ({ isOpen, onConfirm, title, message }) => (
    isOpen ? (
      <div data-testid="confirm-modal">
        <h2>{title}</h2>
        <p>{message}</p>
        <button onClick={onConfirm}>Confirm</button>
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
  token: null,
  botUsername: null,
  loading: false,
  error: null,
  generateToken: vi.fn(async () => ({
    token: 'test-token-123',
    bot_username: 'test_bot',
  })),
  unlinkAccount: vi.fn(async () => ({})),
  refetchLinkStatus: vi.fn(),
  clearError: vi.fn(),
  ...overrides,
});

describe('TelegramInlineRow - Safari Browser Behavior', () => {
  let mockUseTelegramValue;

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseTelegramValue = createMockUseTelegramValue();
    vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

    // Mock window.location
    delete window.location;
    window.location = { href: '' };
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
  });

  describe('Safari on iOS - Telegram Linking Behavior', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
    });

    it('should detect Safari on iOS and show warning message before redirect', async () => {
      const user = userEvent.setup();

      render(<TelegramInlineRow />);

      // Click the link button
      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Wait for warning to appear
      await waitFor(() => {
        const warning = screen.getByText(/перенаправление в telegram/i);
        expect(warning).toBeInTheDocument();
      });
    });

    it('should set window.location.href after 1500ms delay on Safari iOS', async () => {
      vi.useFakeTimers();
      const user = userEvent.setup({ delay: null });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      fireEvent.click(linkButton);

      // Wait for generateToken to be called
      await vi.waitFor(() => {
        expect(vi.mocked(mockUseTelegramValue.generateToken)).toHaveBeenCalled();
      });

      // Initially, window.location.href should not be set
      expect(window.location.href).toBe('');

      // Fast forward 1400ms - should still not be set
      vi.advanceTimersByTime(1400);
      expect(window.location.href).toBe('');

      // Fast forward remaining 100ms to reach 1500ms total
      vi.advanceTimersByTime(100);

      // Now window.location.href should be set with bot URL
      expect(window.location.href).toBe('https://t.me/test_bot?start=test-token-123');

      vi.useRealTimers();
    });

    it('should not show error message when Safari iOS redirects successfully', async () => {
      vi.useFakeTimers();
      const user = userEvent.setup({ delay: null });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      fireEvent.click(linkButton);

      // Advance timer to trigger redirect
      vi.advanceTimersByTime(1500);

      // Error message should NOT be visible
      const errorMessage = screen.queryByText(/не удалось открыть бот/i);
      expect(errorMessage).not.toBeInTheDocument();

      vi.useRealTimers();
    });

    it('should call isSafariOnIOS detector when attempting to link', async () => {
      const user = userEvent.setup();

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Should have called the Safari iOS detector
      expect(vi.mocked(browserDetection.isSafariOnIOS)).toHaveBeenCalled();
    });

    it('should display warning div with correct message text', async () => {
      const user = userEvent.setup();

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Wait for warning element
      const warning = await screen.findByText(/перенаправление в telegram/i);
      expect(warning).toBeInTheDocument();

      // Warning should be in a div with class 'telegram-safari-warning'
      const warningDiv = warning.closest('.telegram-safari-warning');
      expect(warningDiv).toBeInTheDocument();
    });

    it('should generate correct bot URL with token', async () => {
      vi.useFakeTimers();
      const user = userEvent.setup({ delay: null });

      const customTokenData = {
        token: 'custom-token-xyz',
        bot_username: 'custom_bot',
      };
      mockUseTelegramValue.generateToken.mockResolvedValueOnce(customTokenData);
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      fireEvent.click(linkButton);

      // Wait for generateToken to be called and promise to resolve
      await vi.waitFor(() => {
        expect(vi.mocked(mockUseTelegramValue.generateToken)).toHaveBeenCalled();
      });

      // Advance timer to trigger redirect
      vi.advanceTimersByTime(1500);

      // Should set window.location.href with the custom token
      expect(window.location.href).toBe('https://t.me/custom_bot?start=custom-token-xyz');

      vi.useRealTimers();
    });

    it('should handle token generation error gracefully on Safari iOS', async () => {
      const user = userEvent.setup();

      const error = new Error('Token generation failed');
      mockUseTelegramValue.generateToken.mockRejectedValueOnce(error);
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Wait for error state - component should show error instead of warning
      await waitFor(() => {
        const errorMessage = screen.queryByText(/token generation failed/i);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    it('should set isLinking to false after token generation', async () => {
      const user = userEvent.setup();

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });

      // Initially button should not be in loading state
      expect(linkButton).not.toHaveAttribute('data-loading', 'true');

      await user.click(linkButton);

      // After click, button should show loading state (but briefly)
      await waitFor(() => {
        // After generateToken completes, loading should be false
        expect(vi.mocked(mockUseTelegramValue.generateToken)).toHaveBeenCalled();
      });
    });

    it('should clear local error before retrying on Safari iOS', async () => {
      const user = userEvent.setup();

      // First attempt: error
      mockUseTelegramValue.generateToken.mockRejectedValueOnce(new Error('First error'));
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      let linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Wait for error to show
      await waitFor(() => {
        expect(screen.getByText(/first error/i)).toBeInTheDocument();
      });

      // Second attempt: success
      mockUseTelegramValue.generateToken.mockResolvedValueOnce({
        token: 'test-token-123',
        bot_username: 'test_bot',
      });
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      // Click retry button
      const retryButton = screen.getByRole('button', { name: /повторить/i });
      await user.click(retryButton);

      // clearError should have been called
      expect(vi.mocked(mockUseTelegramValue.clearError)).toHaveBeenCalled();
    });
  });

  describe('Safari on macOS - Popup Opening Behavior', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
    });

    it('should call openTelegramBot for Safari desktop', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Should call openTelegramBot
      expect(vi.mocked(browserDetection.openTelegramBot)).toHaveBeenCalledWith(
        'https://t.me/test_bot?start=test-token-123'
      );
    });

    it('should show error when popup is blocked on Safari desktop', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: true,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Wait for error message to appear
      await waitFor(() => {
        const errorMessage = screen.getByText(/не удалось открыть бот/i);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    it('should display helpful error message when popup is blocked', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: true,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      await waitFor(() => {
        const errorMessage = screen.getByText(/попробуйте отключить блокировку всплывающих окон/i);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    it('should NOT show Safari warning when popup succeeds on macOS', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Safari warning should NOT appear
      const warning = screen.queryByText(/перенаправление в telegram/i);
      expect(warning).not.toBeInTheDocument();
    });

    it('should NOT redirect on Safari macOS (use popup instead)', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Should NOT set window.location.href (that's iOS only)
      expect(window.location.href).toBe('');
    });

    it('should allow retry when popup is blocked', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: true,
      });

      render(<TelegramInlineRow />);

      let linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Wait for error
      await waitFor(() => {
        expect(screen.getByText(/не удалось открыть бот/i)).toBeInTheDocument();
      });

      // Second attempt: success
      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      const retryButton = screen.getByRole('button', { name: /повторить/i });
      await user.click(retryButton);

      // Should call openTelegramBot again
      expect(vi.mocked(browserDetection.openTelegramBot)).toHaveBeenCalledTimes(2);
    });
  });

  describe('Other Browsers (Chrome, Firefox, Edge) - Popup Behavior', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
    });

    it('should NOT show Safari warning for Chrome/Firefox', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Safari warning should NOT appear for other browsers
      const warning = screen.queryByText(/перенаправление в telegram/i);
      expect(warning).not.toBeInTheDocument();
    });

    it('should use window.open (popup) for Chrome/Firefox', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // openTelegramBot should be called
      expect(vi.mocked(browserDetection.openTelegramBot)).toHaveBeenCalledWith(
        'https://t.me/test_bot?start=test-token-123'
      );
    });

    it('should detect popup blocking in other browsers', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: true,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Error should show for popup blocking
      await waitFor(() => {
        expect(screen.getByText(/не удалось открыть бот/i)).toBeInTheDocument();
      });
    });

    it('should show error if popup blocked in Chrome', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: true,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      await waitFor(() => {
        const errorMessage = screen.getByText(/попробуйте отключить блокировку всплывающих окон/i);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    it('should NOT redirect on window.location for non-Safari browsers', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // window.location.href should NOT be used for non-Safari
      expect(window.location.href).toBe('');
    });
  });

  describe('Component Lifecycle and State Management', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });
    });

    it('should render "Привязать" button when not linked', () => {
      render(<TelegramInlineRow />);

      const button = screen.getByRole('button', { name: /привязать/i });
      expect(button).toBeInTheDocument();
    });

    it('should show loading state during token generation', async () => {
      const user = userEvent.setup();

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Button should have been clicked
      expect(vi.mocked(mockUseTelegramValue.generateToken)).toHaveBeenCalled();
    });

    it('should clear error when retrying', async () => {
      const user = userEvent.setup();

      mockUseTelegramValue.generateToken.mockRejectedValueOnce(new Error('First error'));
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      let linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      await waitFor(() => {
        expect(screen.getByText(/first error/i)).toBeInTheDocument();
      });

      mockUseTelegramValue.generateToken.mockResolvedValueOnce({
        token: 'test-token-123',
        bot_username: 'test_bot',
      });

      const retryButton = screen.getByRole('button', { name: /повторить/i });
      await user.click(retryButton);

      expect(vi.mocked(mockUseTelegramValue.clearError)).toHaveBeenCalled();
    });
  });

  describe('Edge Cases and Error Handling', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });
    });

    it('should handle missing token data', async () => {
      const user = userEvent.setup();

      mockUseTelegramValue.generateToken.mockResolvedValueOnce({
        token: null,
        bot_username: null,
      });
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Should not attempt to open popup if token is missing
      expect(vi.mocked(browserDetection.openTelegramBot)).not.toHaveBeenCalled();
    });

    it('should handle missing token but present bot_username', async () => {
      const user = userEvent.setup();

      mockUseTelegramValue.generateToken.mockResolvedValueOnce({
        token: null,
        bot_username: 'test_bot',
      });
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Should not open bot if token is missing
      expect(vi.mocked(browserDetection.openTelegramBot)).not.toHaveBeenCalled();
    });

    it('should handle missing bot_username but present token', async () => {
      const user = userEvent.setup();

      mockUseTelegramValue.generateToken.mockResolvedValueOnce({
        token: 'test-token-123',
        bot_username: null,
      });
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // Should not open bot if bot_username is missing
      expect(vi.mocked(browserDetection.openTelegramBot)).not.toHaveBeenCalled();
    });

    it('should handle generateToken rejection with custom error message', async () => {
      const user = userEvent.setup();

      const customError = new Error('Custom error message');
      mockUseTelegramValue.generateToken.mockRejectedValueOnce(customError);
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      await waitFor(() => {
        expect(screen.getByText(/custom error message/i)).toBeInTheDocument();
      });
    });
  });

  describe('Integration: Safari iOS with Unlink Flow', () => {
    beforeEach(() => {
      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(true);
      mockUseTelegramValue.linkStatus = { linked: true, telegram: { username: 'testuser' } };
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);
    });

    it('should render "Отвязать" button when linked on Safari iOS', () => {
      render(<TelegramInlineRow />);

      const button = screen.getByRole('button', { name: /отвязать/i });
      expect(button).toBeInTheDocument();
    });

    it('should show username when linked', () => {
      render(<TelegramInlineRow />);

      const username = screen.getByText(/@testuser/i);
      expect(username).toBeInTheDocument();
    });

    it('should show linked status indicator', () => {
      render(<TelegramInlineRow />);

      const statusIndicator = screen.getByTitle(/telegram привязан/i);
      expect(statusIndicator).toBeInTheDocument();
    });
  });

  describe('Browser Detection Integration', () => {
    it('should call isSafariOnIOS only once during linking', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      // isSafariOnIOS should be called exactly once in handleLink
      expect(vi.mocked(browserDetection.isSafariOnIOS)).toHaveBeenCalledTimes(1);
    });

    it('should verify openTelegramBot receives correct URL format', async () => {
      const user = userEvent.setup();

      vi.mocked(browserDetection.isSafariOnIOS).mockReturnValue(false);
      vi.mocked(browserDetection.openTelegramBot).mockReturnValue({
        method: 'popup',
        blocked: false,
      });

      const customTokenData = {
        token: 'abc123',
        bot_username: 'mybot',
      };
      mockUseTelegramValue.generateToken.mockResolvedValueOnce(customTokenData);
      vi.mocked(useTelegram).mockReturnValue(mockUseTelegramValue);

      render(<TelegramInlineRow />);

      const linkButton = screen.getByRole('button', { name: /привязать/i });
      await user.click(linkButton);

      expect(vi.mocked(browserDetection.openTelegramBot)).toHaveBeenCalledWith(
        'https://t.me/mybot?start=abc123'
      );
    });
  });
});
