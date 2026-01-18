import { Link, useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "../../hooks/useAuth.js";
import { ROLES } from "../../utils/constants.js";
import Button from "../common/Button.jsx";
import "./Header.css";

export const Header = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const getDashboardUrl = () => {
    if (!user) return "/";
    switch (user.role) {
      case "admin":
        return "/dashboard/admin";
      case "methodologist":
        return "/dashboard/methodologist";
      case "teacher":
        return "/dashboard/teacher";
      case "student":
        return "/dashboard/student";
      default:
        return "/";
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  const handleToggleSidebar = () => {
    const isDesktop = window.innerWidth > 768;
    const eventType = isDesktop ? "sidebar:toggleCollapse" : "sidebar:toggle";
    const event = new CustomEvent(eventType);
    window.dispatchEvent(event);
  };

  const isInDashboard = location.pathname.startsWith("/dashboard/");

  const handleLogoClick = (e) => {
    if (isAuthenticated) {
      e.preventDefault();
      if (isInDashboard) {
        // Если в ЛК - переходим на лендинг
        navigate("/");
      } else {
        // Если не в ЛК - переходим в ЛК
        navigate(getDashboardUrl());
      }
    }
    // Если не авторизован, оставляем стандартное поведение ссылки
  };

  const getLogoUrl = () => {
    if (!isAuthenticated) {
      return import.meta.env.DEV ? window.location.origin : "/";
    }
    // Для авторизованных: если в ЛК - на лендинг, иначе - в ЛК
    return isInDashboard ? "/" : getDashboardUrl();
  };

  return (
    <header className="header" data-testid="header">
      <div className="header-left">
        {isInDashboard && (
          <button
            className="header-toggle-btn"
            onClick={handleToggleSidebar}
            aria-label="Переключить боковую панель"
          >
            ☰
          </button>
        )}
      </div>

      <div className="header-center">
        {isAuthenticated ? (
          <Link
            to={getLogoUrl()}
            onClick={handleLogoClick}
            className="header-logo"
          >
            <span className="header-logo-text"><strong>THE BOT</strong></span>
          </Link>
        ) : (
          <a
            href={import.meta.env.DEV ? window.location.origin : "/"}
            onClick={handleLogoClick}
            className="header-logo"
          >
            <span className="header-logo-text"><strong>THE BOT</strong></span>
          </a>
        )}
      </div>

      <div className="header-right">
        {isAuthenticated ? (
          <>
            <span className="header-user-name" data-testid="user-name">
              {user?.fullName || user?.full_name || user?.email}
            </span>
            <button
              className="header-logout-btn"
              onClick={handleLogout}
              data-testid="logout-button"
              aria-label="Выход"
            >
              <svg
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                className="header-logout-icon"
              >
                {/* Дверь с закругленными углами */}
                <rect
                  x="4"
                  y="4"
                  width="9"
                  height="16"
                  rx="1.5"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  fill="none"
                />
                {/* Дверная ручка */}
                <circle
                  cx="9"
                  cy="12"
                  r="1"
                  fill="currentColor"
                />
                {/* Стрелка выхода */}
                <path
                  d="M13 12h5M16 9l3 3-3 3"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </button>
          </>
        ) : (
          <Link to="/login">
            <Button variant="primary" size="sm">
              Войти
            </Button>
          </Link>
        )}
      </div>
    </header>
  );
};

export default Header;
