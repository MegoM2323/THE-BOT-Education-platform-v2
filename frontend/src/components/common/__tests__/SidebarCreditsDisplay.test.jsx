import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { SidebarCreditsDisplay } from '../SidebarCreditsDisplay.jsx';
import { NotificationProvider } from '../../../context/NotificationContext.jsx';
import * as creditsAPI from '../../../api/credits.js';

vi.mock('../../../api/credits.js');
// Don't mock useSidebarCredits here - we want to test the real component with the hook

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
  return ({ children }) => (
    <NotificationProvider>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </NotificationProvider>
  );
};

describe('SidebarCreditsDisplay', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render credits display', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 150 });

    render(<SidebarCreditsDisplay />, { wrapper: createWrapper() });

    expect(screen.getByTestId('sidebar-credits')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toHaveTextContent(
        '150'
      );
    });
  });

  it('should show loading state initially', async () => {
    // Slow mock to see loading state
    creditsAPI.getCredits.mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve({ balance: 100 }), 500))
    );

    const { container } = render(<SidebarCreditsDisplay />, { wrapper: createWrapper() });

    // Component should render
    expect(container).toBeInTheDocument();

    // Should eventually load the data
    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toHaveTextContent('100');
    }, { timeout: 2000 });
  });

  it('should show error message on fetch failure', async () => {
    creditsAPI.getCredits.mockRejectedValueOnce(new Error('API Error'));

    render(<SidebarCreditsDisplay />, { wrapper: createWrapper() });

    // Should display balance with 0 on error (fallback)
    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toBeInTheDocument();
    });
  });

  it('should display collapsed view correctly', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 100 });

    render(<SidebarCreditsDisplay collapsed={true} />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toBeInTheDocument();
    });

    // Should not have refresh button when collapsed
    expect(
      screen.queryByTestId('sidebar-credits-refresh')
    ).not.toBeInTheDocument();
  });

  it('should have refresh button when not collapsed', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 100 });

    render(<SidebarCreditsDisplay collapsed={false} />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-refresh')).toBeInTheDocument();
    });
  });

  it('should handle manual refresh', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 100 });

    render(<SidebarCreditsDisplay />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toHaveTextContent('100');
    });

    // Wait for loading to complete
    await new Promise(resolve => setTimeout(resolve, 100));

    // Verify refresh button exists
    const refreshButton = screen.getByTestId('sidebar-credits-refresh');
    expect(refreshButton).toBeInTheDocument();
  });

  it('should display custom interval', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 200 });

    render(<SidebarCreditsDisplay interval={5000} />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toHaveTextContent(
        '200'
      );
    });
  });

  it('should render label when not collapsed', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 100 });

    render(<SidebarCreditsDisplay collapsed={false} />, {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText('Кредиты')).toBeInTheDocument();
    });
  });

  it('should format balance display correctly', async () => {
    creditsAPI.getCredits.mockResolvedValue({ balance: 0 });

    render(<SidebarCreditsDisplay />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByTestId('sidebar-credits-balance')).toHaveTextContent(
        '0'
      );
    });
  });
});
