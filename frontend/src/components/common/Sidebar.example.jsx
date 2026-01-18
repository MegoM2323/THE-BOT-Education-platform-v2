/**
 * ПРИМЕР: Как интегрировать SidebarCreditsDisplay в существующий Sidebar
 *
 * Это демонстрирует желаемый результат после интеграции
 */

import { useEffect, useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth.js';
import { SidebarCreditsDisplay } from './SidebarCreditsDisplay.jsx';
import './Sidebar.css';

export const SidebarExample = ({
  links,
  isOpen,
  onClose,
  collapsed,
  onToggleCollapse,
  gridLayout = false,
}) => {
  const { user } = useAuth();
  const location = useLocation();
  const [isDesktop, setIsDesktop] = useState(window.innerWidth > 768);

  // Отслеживать размер окна
  useEffect(() => {
    const handleResize = () => {
      setIsDesktop(window.innerWidth > 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Закрывать sidebar при изменении роута (только на мобильных)
  useEffect(() => {
    if (isOpen && !isDesktop) {
      onClose();
    }
  }, [location.pathname, isOpen, isDesktop, onClose]);

  return (
    <>
      {isOpen && !isDesktop && (
        <div className="sidebar-overlay" onClick={onClose} />
      )}
      <aside
        className={`sidebar ${isOpen ? 'sidebar-open' : ''} ${
          collapsed ? 'sidebar-collapsed' : ''
        }`}
        data-testid="sidebar"
      >
        <div className="sidebar-header">
          <div className="sidebar-header-row">
            <h2 className="sidebar-title">
              {!collapsed && (
                <>
                  {user?.role === 'student' && 'Студент'}
                  {user?.role === 'teacher' && 'Преподаватель'}
                  {user?.role === 'admin' && 'Администратор'}
                </>
              )}
            </h2>

            {/* Кнопка переключения - только desktop */}
            {isDesktop && onToggleCollapse && (
              <button
                className="sidebar-toggle"
                onClick={onToggleCollapse}
                aria-label={collapsed ? 'Развернуть меню' : 'Свернуть меню'}
                data-testid="toggle-sidebar"
                title={collapsed ? 'Развернуть меню' : 'Свернуть меню'}
              >
                <svg
                  width="20"
                  height="20"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path d="M15 18l-6-6 6-6" />
                </svg>
              </button>
            )}

            {/* Кнопка закрытия - только mobile */}
            {!isDesktop && (
              <button
                className="sidebar-close"
                onClick={onClose}
                aria-label="Закрыть меню"
                data-testid="close-sidebar"
              >
                &times;
              </button>
            )}
          </div>
        </div>

        {/* НОВОЕ: Добавляем SidebarCreditsDisplay для студентов */}
        {user?.role === 'student' && (
          <SidebarCreditsDisplay collapsed={collapsed} interval={10000} />
        )}

        <nav className="sidebar-nav">
          <ul className={`sidebar-links ${gridLayout ? 'sidebar-links--grid' : ''}`}>
            {links.map((link) => (
              <li key={link.path}>
                <NavLink
                  to={link.path}
                  className={({ isActive }) =>
                    `sidebar-link ${isActive ? 'sidebar-link-active' : ''}`
                  }
                  data-testid={link.testId || `nav-${link.label.toLowerCase().replace(/\s+/g, '-')}`}
                  title={collapsed ? link.label : ''}
                >
                  {link.icon && <span className="sidebar-icon">{link.icon}</span>}
                  <span className="sidebar-label">{link.label}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </aside>
    </>
  );
};

export default SidebarExample;
