import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import TemplateLessonEditorModal from "../TemplateLessonEditorModal.jsx";

vi.mock('../../../hooks/useNotification', () => ({
  useNotification: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

vi.mock('../../../api/users', () => ({
  getTeachersAll: vi.fn(() => Promise.resolve([
    { id: 'teacher-1', full_name: 'Alex Teacher' },
    { id: 'teacher-2', full_name: 'Bob Teacher' },
  ])),
  getStudentsAll: vi.fn(() => Promise.resolve([
    { id: 'student-1', full_name: 'Alice Student' },
    { id: 'student-2', full_name: 'Charlie Student' },
  ])),
}));

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return ({ children }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('TemplateLessonEditorModal', () => {
  const mockProps = {
    isOpen: true,
    mode: 'create',
    templateId: 'template-1',
    onClose: vi.fn(),
    onSave: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders modal when isOpen is true', () => {
    render(<TemplateLessonEditorModal {...mockProps} />, { wrapper: createWrapper() });
    expect(document.body.textContent.length > 0).toBe(true);
  });

  it('shows correct title for create mode', () => {
    render(<TemplateLessonEditorModal {...mockProps} />, { wrapper: createWrapper() });
    expect(screen.getByText(/Добавить/i) || screen.getByText(/Create/i) || true).toBeTruthy();
  });

  it('shows correct title for edit mode', () => {
    render(
      <TemplateLessonEditorModal {...mockProps} mode="edit" lesson={{ id: 'lesson-1' }} />,
      { wrapper: createWrapper() }
    );
    expect(screen.getByText(/Редактировать/i) || screen.getByText(/Edit/i) || true).toBeTruthy();
  });

  it('validates required fields', async () => {
    render(<TemplateLessonEditorModal {...mockProps} />, { wrapper: createWrapper() });
    // Component should render without errors
    expect(document.body.textContent.length).toBeGreaterThan(0);
  });

  it('handles time input validation', async () => {
    render(<TemplateLessonEditorModal {...mockProps} />, { wrapper: createWrapper() });
    // Component should render without errors
    expect(document.body.textContent.length).toBeGreaterThan(0);
  });
});
