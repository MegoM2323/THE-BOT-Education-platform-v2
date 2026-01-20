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
            <div className="header-logo-icon">
              <svg width="40" height="40" viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
                <defs>
                  <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#8B5CF6" />
                    <stop offset="100%" stopColor="#A78BFA" />
                  </linearGradient>
                </defs>
                <rect width="40" height="40" rx="8" fill="url(#logoGradient)" />
                <path d="M12 14C12 13.4477 12.4477 13 13 13H20C20.5523 13 21 13.4477 21 14V26C21 26.5523 20.5523 27 20 27H13C12.4477 27 12 26.5523 12 26V14Z" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M19 14C19 13.4477 19.4477 13 20 13H27C27.5523 13 28 13.4477 28 14V26C28 26.5523 27.5523 27 27 27H20C19.4477 27 19 26.5523 19 26V14Z" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M20 20L20 27" stroke="white" strokeWidth="2" strokeLinecap="round"/>
                <circle cx="20" cy="24" r="1.5" fill="white"/>
              </svg>
            </div>
            <span className="header-logo-text">THE BOT</span>
          </Link>
        ) : (
          <a
            href={import.meta.env.DEV ? window.location.origin : "/"}
            onClick={handleLogoClick}
            className="header-logo"
          >
            <div className="header-logo-icon">
              <svg width="40" height="40" viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
                <defs>
                  <linearGradient id="logoGradientUnauth" x1="0%" y1="0%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#8B5CF6" />
                    <stop offset="100%" stopColor="#A78BFA" />
                  </linearGradient>
                </defs>
                <rect width="40" height="40" rx="8" fill="url(#logoGradientUnauth)" />
                <path d="M12 14C12 13.4477 12.4477 13 13 13H20C20.5523 13 21 13.4477 21 14V26C21 26.5523 20.5523 27 20 27H13C12.4477 27 12 26.5523 12 26V14Z" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M19 14C19 13.4477 19.4477 13 20 13H27C27.5523 13 28 13.4477 28 14V26C28 26.5523 27.5523 27 27 27H20C19.4477 27 19 26.5523 19 26V14Z" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M20 20L20 27" stroke="white" strokeWidth="2" strokeLinecap="round"/>
                <circle cx="20" cy="24" r="1.5" fill="white"/>
              </svg>
            </div>
            <span className="header-logo-text">THE BOT</span>
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
