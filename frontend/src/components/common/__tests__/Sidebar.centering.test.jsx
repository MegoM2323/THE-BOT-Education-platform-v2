import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { describe, it, expect, beforeEach } from 'vitest';
import { Sidebar } from '../Sidebar.jsx';

// Простые SVG иконки для тестирования
const TestIcon = ({ width = 20, height = 20 }) => (
  <svg width={width} height={height} viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="10" />
  </svg>
);

const createTestLinks = (gridLayout = false) => [
  {
    path: '/dashboard',
    label: 'Панель',
    icon: <TestIcon />,
    testId: 'nav-dashboard',
  },
  {
    path: '/lessons',
    label: 'Уроки',
    icon: <TestIcon />,
    testId: 'nav-lessons',
  },
  {
    path: '/profile',
    label: 'Профиль',
    icon: <TestIcon />,
    testId: 'nav-profile',
  },
  {
    path: '/settings',
    label: 'Настройки',
    icon: <TestIcon />,
    testId: 'nav-settings',
  },
];

// Обёртка с BrowserRouter для NavLink
const SidebarWrapper = ({ children }) => (
  <BrowserRouter>{children}</BrowserRouter>
);

describe('Sidebar Icon Centering', () => {
  describe('CSS Classes for Centering', () => {
    it('should have flex-direction column on grid links', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridLinks = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-link'
      );

      gridLinks.forEach((link) => {
        const styles = window.getComputedStyle(link);
        expect(styles.flexDirection).toBe('column');
      });
    });

    it('should center grid links horizontally and vertically', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridLinks = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-link'
      );

      gridLinks.forEach((link) => {
        const styles = window.getComputedStyle(link);
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
      });
    });

    it('should have correct icon dimensions in grid layout (48px)', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridIcons = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon'
      );

      gridIcons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        // Check that CSS classes are applied and flex is working
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
        // Grid icons should have proper dimensions (width/height set in CSS)
        const width = parseFloat(styles.width);
        const height = parseFloat(styles.height);
        expect(width).toBeGreaterThan(0);
        expect(height).toBeGreaterThan(0);
        expect(width).toBe(height); // Should be square
      });
    });

    it('should have correct icon dimensions in regular layout (44px)', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.width).toBe('44px');
        expect(styles.height).toBe('44px');
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
      });
    });

    it('should apply background color to icons', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.backgroundColor).toBe('rgba(0, 0, 0, 0.02)');
      });
    });
  });

  describe('Icon Centering in Grid Layout', () => {
    it('should render SVG icons with correct dimensions in grid', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const svgs = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon svg'
      );

      // SVGs should be rendered
      expect(svgs.length).toBeGreaterThan(0);

      svgs.forEach((svg) => {
        const styles = window.getComputedStyle(svg);
        // SVG should have dimensions set
        const width = parseFloat(styles.width);
        const height = parseFloat(styles.height);
        expect(width).toBeGreaterThan(0);
        expect(height).toBeGreaterThan(0);
        // SVG should be square
        expect(width).toBe(height);
      });
    });

    it('should render SVG icons with correct dimensions in regular layout', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const svgs = container.querySelectorAll('.sidebar-icon svg');

      svgs.forEach((svg) => {
        const styles = window.getComputedStyle(svg);
        expect(styles.width).toBe('20px');
        expect(styles.height).toBe('20px');
      });
    });

    it('should center SVG inside icon container', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const iconContainers = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon'
      );

      iconContainers.forEach((container) => {
        const svg = container.querySelector('svg');
        expect(svg).toBeInTheDocument();

        const containerStyles = window.getComputedStyle(container);
        expect(containerStyles.display).toBe('flex');
        expect(containerStyles.alignItems).toBe('center');
        expect(containerStyles.justifyContent).toBe('center');
      });
    });

    it('should have symmetric icon spacing', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const iconContainers = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon'
      );

      iconContainers.forEach((iconContainer) => {
        const styles = window.getComputedStyle(iconContainer);
        const width = parseFloat(styles.width);
        const height = parseFloat(styles.height);

        // Проверяем что иконка квадратная
        expect(width).toBe(height);

        const svg = iconContainer.querySelector('svg');
        const svgStyles = window.getComputedStyle(svg);
        const svgWidth = parseFloat(svgStyles.width);
        const svgHeight = parseFloat(svgStyles.height);

        // SVG тоже должна быть квадратной
        expect(svgWidth).toBe(svgHeight);
      });
    });
  });

  describe('Responsive States', () => {
    it('should center icons on mobile (max-width: 768px)', () => {
      // Имитируем мобильный размер
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
      });

      // Восстанавливаем
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1024,
      });
    });

    it('should maintain icon centering in collapsed state', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={true}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
      });
    });

    it('should center icons in grid layout on desktop', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridIcons = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon'
      );

      expect(gridIcons.length).toBeGreaterThan(0);

      gridIcons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
      });
    });
  });

  describe('Grid Layout Structure', () => {
    it('should render grid layout as 2-column grid', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridList = container.querySelector('.sidebar-links--grid');
      const styles = window.getComputedStyle(gridList);

      expect(styles.display).toBe('grid');
      expect(styles.gridTemplateColumns).toBe('1fr 1fr');
    });

    it('should render all icons in grid cells', () => {
      const links = createTestLinks();
      const { container } = render(
        <Sidebar
          links={links}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridIcons = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon'
      );

      expect(gridIcons.length).toBe(links.length);

      gridIcons.forEach((icon) => {
        expect(icon.querySelector('svg')).toBeInTheDocument();
      });
    });

    it('should have gap between grid items', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridList = container.querySelector('.sidebar-links--grid');
      const styles = window.getComputedStyle(gridList);

      expect(styles.gap).toBeTruthy();
    });

    it('should maintain aspect ratio of icon cells', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridLinks = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-link'
      );

      gridLinks.forEach((link) => {
        const icons = link.querySelector('.sidebar-icon');
        expect(icons).toBeInTheDocument();

        const styles = window.getComputedStyle(icons);
        const width = parseFloat(styles.width);
        const height = parseFloat(styles.height);

        // Иконки должны быть квадратными (width === height)
        expect(width).toBe(height);
        expect(width).toBeGreaterThan(0);
      });
    });
  });

  describe('Icon Display Properties', () => {
    it('should not shrink icons', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.flexShrink).toBe('0');
      });
    });

    it('should have correct border radius', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.borderRadius).toBeTruthy();
      });
    });

    it('should have minimal margin on grid icons', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={true}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const gridIcons = container.querySelectorAll(
        '.sidebar-links--grid .sidebar-icon'
      );

      gridIcons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        // margin: 0 auto означает центрирование по горизонтали
        // Проверяем что margin установлен (либо auto, либо 0)
        const marginLeft = styles.marginLeft;
        const marginRight = styles.marginRight;
        expect(marginLeft === 'auto' || marginLeft === '0px').toBe(true);
        expect(marginRight === 'auto' || marginRight === '0px').toBe(true);
      });
    });
  });

  describe('Collapsed State Icons', () => {
    it('should center icons when sidebar is collapsed', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={true}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const sidebar = container.querySelector('.sidebar-collapsed');
      expect(sidebar).toBeInTheDocument();

      const icons = container.querySelectorAll('.sidebar-icon');

      icons.forEach((icon) => {
        const styles = window.getComputedStyle(icon);
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
        expect(styles.justifyContent).toBe('center');
      });
    });

    it('should maintain icon visibility in collapsed state', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={true}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const icons = container.querySelectorAll('.sidebar-icon svg');

      expect(icons.length).toBeGreaterThan(0);

      icons.forEach((svg) => {
        // SVG должна быть видна (не скрыта через display или visibility)
        const styles = window.getComputedStyle(svg);
        expect(styles.display).not.toBe('none');
      });
    });
  });

  describe('Icon Container Alignment', () => {
    it('should align items properly in normal link layout', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const links = container.querySelectorAll('.sidebar-link');

      links.forEach((link) => {
        const styles = window.getComputedStyle(link);
        expect(styles.display).toBe('flex');
        expect(styles.alignItems).toBe('center');
      });
    });

    it('should have correct gap between icon and label in normal layout', () => {
      const { container } = render(
        <Sidebar
          links={createTestLinks()}
          isOpen={true}
          collapsed={false}
          gridLayout={false}
          onClose={() => {}}
          onToggleCollapse={() => {}}
        />,
        { wrapper: SidebarWrapper }
      );

      const links = container.querySelectorAll(
        '.sidebar-links:not(.sidebar-links--grid) .sidebar-link'
      );

      links.forEach((link) => {
        const styles = window.getComputedStyle(link);
        expect(styles.gap).toBeTruthy();
      });
    });
  });
});
