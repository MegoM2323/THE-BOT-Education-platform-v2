import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { CreditsHistory } from '../CreditsHistory.jsx';

// Mock useCredits hook
vi.mock('../../../hooks/useCredits.js', () => ({
  useCredits: vi.fn(),
}));

// Mock Spinner component
vi.mock('../../common/Spinner.jsx', () => ({
  default: () => <div data-testid="spinner">Loading...</div>
}));

// Import the mocked hook
import { useCredits } from '../../../hooks/useCredits.js';

describe('CreditsHistory - Error Message Display', () => {
  const mockFetchHistory = vi.fn();
  const defaultMockValue = {
    balance: 100,
    history: [],
    historyPagination: null,
    loading: false,
    error: null,
    historyLoading: false,
    historyError: null,
    fetchHistory: mockFetchHistory,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useCredits.mockReturnValue(defaultMockValue);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  // Test 1: No error - error container not rendered
  it('should not render error message when historyError is null', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: null,
    });

    render(<CreditsHistory />);

    const errorContainer = screen.queryByRole('alert');
    expect(errorContainer).not.toBeInTheDocument();
  });

  // Test 2: historyError with userMessage property
  it('should display userMessage when historyError has userMessage property', () => {
    const userMessage = 'Ваша сессия истекла. Пожалуйста, авторизуйтесь заново.';
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage,
        message: 'Unauthorized',
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement).toBeInTheDocument();
    expect(alertElement.textContent).toContain(userMessage);
  });

  // Test 3: historyError without userMessage - fallback to message
  it('should display message when historyError lacks userMessage property', () => {
    const fallbackMessage = 'Network error occurred';
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        message: fallbackMessage,
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement).toBeInTheDocument();
    expect(alertElement.textContent).toContain(fallbackMessage);
  });

  // Test 4: Error message has role="alert" attribute
  it('should have role="alert" attribute on error container', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: 'Error message text',
        message: 'fallback',
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement).toHaveAttribute('role', 'alert');
  });

  // Test 5: Retry button is present when error is shown
  it('should render retry button when error is displayed', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: 'Error loading history',
        message: 'error',
      },
    });

    render(<CreditsHistory />);

    const retryButton = screen.getByRole('button', { name: /повторить/i });
    expect(retryButton).toBeInTheDocument();
  });

  // Test 6: Retry button calls fetchHistory when clicked
  it('should call fetchHistory when retry button is clicked', () => {
    const newMockFetch = vi.fn();
    useCredits.mockReturnValue({
      ...defaultMockValue,
      fetchHistory: newMockFetch,
      historyError: {
        userMessage: 'Error loading history',
        message: 'error',
      },
    });

    render(<CreditsHistory />);

    // Clear previous calls from useEffect
    newMockFetch.mockClear();

    const retryButton = screen.getByRole('button', { name: /повторить/i });
    fireEvent.click(retryButton);

    expect(newMockFetch).toHaveBeenCalledTimes(1);
  });

  // Test 7: Error container is visible and accessible
  it('should render error container as an accessible alert region', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: 'Error loading history',
        message: 'error',
      },
    });

    render(<CreditsHistory />);

    const alertContainer = screen.getByRole('alert');
    expect(alertContainer).toBeVisible();
  });

  // Test 8: Balance displays correctly even with error
  it('should display current balance even when error occurs', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      balance: 250,
      historyError: {
        userMessage: 'Error loading history',
        message: 'error',
      },
    });

    render(<CreditsHistory />);

    expect(screen.getByText('250 кредитов')).toBeInTheDocument();
  });

  // Test 9: Error message prefix is correct
  it('should include error prefix text in the message', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: 'Сервер временно недоступен',
        message: 'server error',
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement.textContent).toContain('Ошибка загрузки истории:');
    expect(alertElement.textContent).toContain('Сервер временно недоступен');
  });

  // Test 10: Error message with both userMessage and message properties
  it('should prioritize userMessage over message when both are present', () => {
    const userMessage = 'User-friendly error message';
    const technicalMessage = 'Technical error details';

    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage,
        message: technicalMessage,
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement.textContent).toContain(userMessage);
  });

  // Test 11: Error with empty string userMessage falls back to message
  it('should fall back to message when userMessage is empty string', () => {
    const fallbackMessage = 'Fallback message text';

    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: '',
        message: fallbackMessage,
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement.textContent).toContain(fallbackMessage);
  });

  // Test 12: Error with undefined userMessage falls back to message
  it('should fall back to message when userMessage is undefined', () => {
    const fallbackMessage = 'Fallback error message';

    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: undefined,
        message: fallbackMessage,
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement.textContent).toContain(fallbackMessage);
  });

  // Test 13: Error component structure and classes
  it('should have correct CSS class on error container', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: 'Error message',
        message: 'error',
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement).toHaveClass('credits-history-error');
  });

  // Test 14: Retry button has correct CSS class
  it('should have error-retry-btn class on retry button', () => {
    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: 'Error message',
        message: 'error',
      },
    });

    render(<CreditsHistory />);

    const retryButton = screen.getByRole('button');
    expect(retryButton).toHaveClass('error-retry-btn');
  });

  // Test 15: Error with null userMessage falls back to message
  it('should fall back to message when userMessage is null', () => {
    const fallbackMessage = 'Null fallback error';

    useCredits.mockReturnValue({
      ...defaultMockValue,
      historyError: {
        userMessage: null,
        message: fallbackMessage,
      },
    });

    render(<CreditsHistory />);

    const alertElement = screen.getByRole('alert');
    expect(alertElement.textContent).toContain(fallbackMessage);
  });
});
