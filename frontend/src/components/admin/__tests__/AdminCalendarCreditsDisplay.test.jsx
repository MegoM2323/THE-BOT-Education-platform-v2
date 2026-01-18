import { render, screen, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AdminCalendarCreditsDisplay from '../AdminCalendarCreditsDisplay.jsx';
import * as creditsAPI from '../../../api/credits.js';

// Mock the API
vi.mock('../../../api/credits.js');

// Helper to render with React Query provider
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });

const renderWithQueryClient = (component) => {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      {component}
    </QueryClientProvider>
  );
};

describe('AdminCalendarCreditsDisplay', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should not render when selectedStudentId is not provided', () => {
    const { container } = renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId={null} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('should fetch and display student credits when selectedStudentId is provided', async () => {
    creditsAPI.getUserCredits.mockResolvedValue({ balance: 150 });

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId="student-123" />
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin-calendar-credits-balance')).toHaveTextContent('150');
    });

    expect(creditsAPI.getUserCredits).toHaveBeenCalledWith('student-123');
  });

  it('should display loading spinner while fetching', () => {
    creditsAPI.getUserCredits.mockImplementation(() => new Promise(() => {})); // Never resolves

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId="student-123" />
    );

    // Should show Spinner component
    expect(screen.getByTestId('admin-calendar-credits')).toBeInTheDocument();
  });

  it('should display error state when API call fails', async () => {
    creditsAPI.getUserCredits.mockRejectedValue(new Error('API Error'));

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId="student-123" />
    );

    await waitFor(() => {
      expect(screen.getByText('Ошибка')).toBeInTheDocument();
    });
  });

  it('should call onCreditsLoaded callback with balance when credits are loaded', async () => {
    const onCreditsLoaded = vi.fn();
    creditsAPI.getUserCredits.mockResolvedValue({ balance: 200 });

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay
        selectedStudentId="student-123"
        onCreditsLoaded={onCreditsLoaded}
      />
    );

    await waitFor(() => {
      expect(onCreditsLoaded).toHaveBeenCalledWith(200);
    });
  });

  it('should refetch when selectedStudentId changes', async () => {
    creditsAPI.getUserCredits.mockResolvedValue({ balance: 100 });

    const queryClient = createTestQueryClient();
    const { rerender } = render(
      <QueryClientProvider client={queryClient}>
        <AdminCalendarCreditsDisplay selectedStudentId="student-1" />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(creditsAPI.getUserCredits).toHaveBeenCalledWith('student-1');
    });

    creditsAPI.getUserCredits.mockResolvedValue({ balance: 200 });

    rerender(
      <QueryClientProvider client={queryClient}>
        <AdminCalendarCreditsDisplay selectedStudentId="student-2" />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(creditsAPI.getUserCredits).toHaveBeenCalledWith('student-2');
      expect(creditsAPI.getUserCredits).toHaveBeenCalledTimes(2);
    });
  });

  it('should have correct data-testid attribute', async () => {
    creditsAPI.getUserCredits.mockResolvedValue({ balance: 100 });

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId="student-123" />
    );

    expect(screen.getByTestId('admin-calendar-credits')).toBeInTheDocument();
  });

  it('should display 0 when balance is 0', async () => {
    creditsAPI.getUserCredits.mockResolvedValue({ balance: 0 });

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId="student-123" />
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin-calendar-credits-balance')).toHaveTextContent('0');
    });
  });

  it('should default to 0 when balance is null', async () => {
    creditsAPI.getUserCredits.mockResolvedValue({ balance: null });

    renderWithQueryClient(
      <AdminCalendarCreditsDisplay selectedStudentId="student-123" />
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin-calendar-credits-balance')).toHaveTextContent('0');
    });
  });
});
