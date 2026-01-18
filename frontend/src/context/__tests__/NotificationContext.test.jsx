import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { NotificationProvider, NotificationContext } from '../NotificationContext.jsx';
import { useContext } from 'react';

function TestComponent() {
  const ctx = useContext(NotificationContext);
  return (
    <div>
      <div data-testid="count">{ctx?.notifications?.length ?? 0}</div>
      <div data-testid="has-add">{ctx?.addNotification ? 'yes' : 'no'}</div>
      <div data-testid="has-remove">{ctx?.removeNotification ? 'yes' : 'no'}</div>
      <div data-testid="has-clear">{ctx?.clearAll ? 'yes' : 'no'}</div>
    </div>
  );
}

describe('NotificationContext', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should provide notification context with all required methods', () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    );

    expect(screen.getByTestId('count')).toHaveTextContent('0');
    expect(screen.getByTestId('has-add')).toHaveTextContent('yes');
    expect(screen.getByTestId('has-remove')).toHaveTextContent('yes');
    expect(screen.getByTestId('has-clear')).toHaveTextContent('yes');
  });

  it('should initialize with empty notifications array', () => {
    render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    );

    expect(screen.getByTestId('count')).toHaveTextContent('0');
  });

  it('should clear timers on unmount without errors', () => {
    const clearTimeoutSpy = vi.spyOn(global, 'clearTimeout');

    const { unmount } = render(
      <NotificationProvider>
        <TestComponent />
      </NotificationProvider>
    );

    unmount();

    expect(clearTimeoutSpy).toBeDefined();
    clearTimeoutSpy.mockRestore();
  });
});
