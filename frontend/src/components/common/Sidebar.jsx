import { useEffect, useState } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { useAuth } from "../../hooks/useAuth.js";
import "./Sidebar.css";

export const Sidebar = ({
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

  // Отслеживать размер окна (уже на русском)
  useEffect(() => {
    const handleResize = () => {
      setIsDesktop(window.innerWidth > 768);
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  // Закрывать sidebar при изменении роута (только на мобильных)
  useEffect(() => {
    if (isOpen && !isDesktop) {
      onClose();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.pathname]);

  return (
    <>
      {isOpen && !isDesktop && (
        <div className="sidebar-overlay" onClick={onClose} />
      )}
      <aside
        className={`sidebar ${isOpen ? "sidebar-open" : ""} ${collapsed ? "sidebar-collapsed" : ""}`}
        data-testid="sidebar"
      >
        {!isDesktop && (
          <div className="sidebar-header">
            <div className="sidebar-header-row">
              {/* Кнопка закрытия - только mobile */}
              <button
                className="sidebar-close"
                onClick={onClose}
                aria-label="Закрыть меню"
                data-testid="close-sidebar"
              >
                &times;
              </button>
            </div>
          </div>
        )}

        <nav className="sidebar-nav">
          <ul
            className={`sidebar-links ${gridLayout ? "sidebar-links--grid" : ""}`}
          >
            {links.map((link) => (
              <li key={link.path}>
                <NavLink
                  to={link.path}
                  className={({ isActive }) =>
                    `sidebar-link ${isActive ? "sidebar-link-active" : ""}`
                  }
                  data-testid={
                    link.testId ||
                    `nav-${link.label.toLowerCase().replace(/\s+/g, "-")}`
                  }
                  title={collapsed ? link.label : ""}
                  onClick={() => {
                    if (!isDesktop && isOpen) {
                      onClose();
                    }
                  }}
                >
                  {link.icon && (
                    <span className="sidebar-icon">{link.icon}</span>
                  )}
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

export default Sidebar;
