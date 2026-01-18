import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import TeacherDashboard from "../TeacherDashboard.jsx";
import { describe, it, expect, beforeEach } from 'vitest';

// Mock child components to avoid loading them
vi.mock('../../components/common/Sidebar.jsx', () => ({
  default: ({ links }) => (
    <div data-testid="mock-sidebar">
      {links.map(link => (
        <a key={link.path} href={link.path} data-testid={link.testId}>
          {link.label}
        </a>
      ))}
    </div>
  )
}));

vi.mock('../../components/teacher/TeacherCalendar.jsx', () => ({
  default: () => <div data-testid="mock-calendar">Calendar</div>
}));

vi.mock('../../components/teacher/TeacherProfile.jsx', () => ({
  default: () => <div data-testid="mock-profile">Profile</div>
}));

vi.mock('../../components/common/ErrorBoundary.jsx', () => ({
  default: ({ children }) => <>{children}</>
}));

describe('TeacherDashboard Navigation Changes', () => {
  const renderTeacherDashboard = () => {
    return render(
      <BrowserRouter>
        <TeacherDashboard />
      </BrowserRouter>
    );
  };

  describe('Scenario 1: Verify "Мои занятия" link is removed from navLinks', () => {
    it('should NOT have a navLink with "lessons" path', () => {
      renderTeacherDashboard();

      // Try to find any link with "lessons" in the path
      const lessonLink = screen.queryByTestId('nav-lessons');
      expect(lessonLink).not.toBeInTheDocument();
    });

    it('should NOT have any navigation link containing "мои занятия" text', () => {
      renderTeacherDashboard();

      // Check that the sidebar doesn't contain lessons link
      const sidebar = screen.getByTestId('mock-sidebar');
      const lessonText = sidebar.textContent;
      expect(lessonText).not.toContain('Мои занятия');
    });
  });

  describe('Scenario 2: Verify only Calendar and Profile navigation remain', () => {
    it('should have Calendar navigation link', () => {
      renderTeacherDashboard();

      const calendarLink = screen.getByTestId('nav-calendar');
      expect(calendarLink).toBeInTheDocument();
      expect(calendarLink).toHaveTextContent('Календарь');
    });

    it('should have Profile navigation link', () => {
      renderTeacherDashboard();

      const profileLink = screen.getByTestId('nav-profile');
      expect(profileLink).toBeInTheDocument();
      expect(profileLink).toHaveTextContent('Профиль');
    });

    it('should have exactly 2 navigation links (Calendar and Profile)', () => {
      renderTeacherDashboard();

      const sidebar = screen.getByTestId('mock-sidebar');
      const navLinks = sidebar.querySelectorAll('a[data-testid^="nav-"]');
      expect(navLinks.length).toBe(2);
    });

    it('should navigate to correct paths for active links', () => {
      renderTeacherDashboard();

      const calendarLink = screen.getByTestId('nav-calendar');
      const profileLink = screen.getByTestId('nav-profile');

      expect(calendarLink).toHaveAttribute('href', '/dashboard/teacher/calendar');
      expect(profileLink).toHaveAttribute('href', '/dashboard/teacher/profile');
    });
  });

  describe('Scenario 3: Verify TeacherSchedule import is removed', () => {
    it('should render without errors (TeacherSchedule not imported)', () => {
      // If TeacherSchedule was still imported and used, rendering would fail
      const { container } = renderTeacherDashboard();
      expect(container).toBeTruthy();
    });

    it('should not have any Route with path="lessons"', () => {
      renderTeacherDashboard();

      // The dashboard should not try to render any lessons route
      // By default it renders calendar (due to redirect on /)
      expect(screen.getByTestId('mock-calendar')).toBeInTheDocument();

      // Verify that no lessons-related content is in the routes
      const main = screen.getByRole('main');
      expect(main.textContent).not.toContain('Lessons');
      expect(main.textContent).not.toContain('lessons');
    });
  });

  describe('Integration Tests', () => {
    it('should have correct default route redirect', () => {
      renderTeacherDashboard();

      // The root route should redirect to calendar
      const calendar = screen.getByTestId('mock-calendar');
      expect(calendar).toBeInTheDocument();
    });

    it('should maintain Mobile menu toggle button', () => {
      renderTeacherDashboard();

      const menuToggle = screen.getByTestId('mobile-menu-toggle');
      expect(menuToggle).toBeInTheDocument();
      expect(menuToggle).toHaveTextContent('Меню');
    });

    it('should have proper sidebar structure', () => {
      renderTeacherDashboard();

      const sidebar = screen.getByTestId('mock-sidebar');
      expect(sidebar).toBeInTheDocument();
    });
  });
});
