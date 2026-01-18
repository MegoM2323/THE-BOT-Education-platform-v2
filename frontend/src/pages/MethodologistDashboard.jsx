import { useState, useEffect } from "react";
import { Routes, Route, Navigate, useNavigate } from "react-router-dom";
import Sidebar from "../components/common/Sidebar.jsx";
import ErrorBoundary from "../components/common/ErrorBoundary.jsx";
import Calendar from "../components/admin/Calendar.jsx";
import TelegramManagement from "../components/admin/TelegramManagement.jsx";
import TemplateManagement from "../components/admin/TemplateManagement.jsx";
import TemplateList from "../components/admin/TemplateList.jsx";
import Spinner from "../components/common/Spinner.jsx";
import MethodologistCreditsView from "../components/methodologist/MethodologistCreditsView.jsx";
import { useAuth } from "../hooks/useAuth.js";
import { ROLES } from "../utils/constants.js";
import "./Dashboard.css";

export const MethodologistDashboard = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState(null);
  const [isAuthorizing, setIsAuthorizing] = useState(true);
  const { user, loading } = useAuth();
  const navigate = useNavigate();

  // Проверка роли пользователя при монтировании компонента
  useEffect(() => {
    if (loading) {
      return;
    }

    if (!user) {
      navigate("/login", { replace: true });
      return;
    }

    if (user.role !== ROLES.METHODOLOGIST) {
      if (user.role === ROLES.STUDENT) {
        navigate("/dashboard/student", { replace: true });
      } else if (user.role === ROLES.TEACHER) {
        navigate("/dashboard/teacher", { replace: true });
      } else if (user.role === ROLES.ADMIN) {
        navigate("/dashboard/admin", { replace: true });
      } else {
        navigate("/", { replace: true });
      }
      return;
    }

    setIsAuthorizing(false);
  }, [user, loading, navigate]);

  // Слушатель события toggle сайдбара из Header
  useEffect(() => {
    const handleToggle = () => {
      setSidebarOpen((prev) => !prev);
    };

    const handleToggleCollapse = () => {
      setSidebarCollapsed((prev) => !prev);
    };

    window.addEventListener("sidebar:toggle", handleToggle);
    window.addEventListener("sidebar:toggleCollapse", handleToggleCollapse);
    return () => {
      window.removeEventListener("sidebar:toggle", handleToggle);
      window.removeEventListener(
        "sidebar:toggleCollapse",
        handleToggleCollapse,
      );
    };
  }, []);

  const navLinks = [
    {
      path: "/dashboard/methodologist/calendar",
      label: "Календарь",
      icon: (
        <svg
          className="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
      ),
      testId: "nav-calendar",
    },
    {
      path: "/dashboard/methodologist/templates",
      label: "Шаблоны",
      icon: (
        <svg
          className="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
          />
        </svg>
      ),
      testId: "nav-templates",
    },
    {
      path: "/dashboard/methodologist/telegram",
      label: "Telegram",
      icon: (
        <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2ZM16.64 8.8C16.49 10.38 15.84 14.22 15.51 15.99C15.37 16.74 15.09 16.99 14.83 17.01C14.25 17.07 13.81 16.64 13.25 16.27C12.37 15.69 11.87 15.33 11.02 14.77C10.03 14.12 10.67 13.76 11.24 13.18C11.39 13.03 13.95 10.7 14 10.49C14.0069 10.4582 14.006 10.4252 13.9973 10.3938C13.9886 10.3624 13.9724 10.3337 13.95 10.31C13.89 10.26 13.81 10.28 13.74 10.29C13.65 10.31 12.25 11.24 9.52 13.08C9.12 13.35 8.76 13.49 8.44 13.48C8.08 13.47 7.4 13.28 6.89 13.11C6.26 12.91 5.77 12.8 5.81 12.45C5.83 12.27 6.08 12.09 6.55 11.9C9.47 10.63 11.41 9.79 12.38 9.39C15.16 8.23 15.73 8.03 16.11 8.03C16.19 8.03 16.38 8.05 16.5 8.15C16.6 8.23 16.63 8.34 16.64 8.42C16.63 8.48 16.65 8.66 16.64 8.8Z" />
        </svg>
      ),
      testId: "nav-telegram",
    },
    {
      path: "/dashboard/methodologist/credits",
      label: "Кредиты",
      icon: (
        <svg
          className="w-5 h-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
          <line x1="1" y1="10" x2="23" y2="10" />
        </svg>
      ),
      testId: "nav-credits",
    },
  ];

  if (loading || isAuthorizing) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
        }}
      >
        <Spinner size="lg" />
      </div>
    );
  }

  if (!user || user.role !== ROLES.METHODOLOGIST) {
    return null;
  }

  return (
    <div className="dashboard">
      {!sidebarOpen && (
        <button
          className="dashboard-menu-toggle"
          onClick={() => setSidebarOpen(true)}
          data-testid="mobile-menu-toggle"
          aria-label="Открыть меню"
        >
          Меню
        </button>
      )}

      <Sidebar
        links={navLinks}
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
      />

      <main
        className={`dashboard-main ${sidebarCollapsed ? "sidebar-collapsed" : ""}`}
      >
        <ErrorBoundary>
          <Routes>
            <Route path="/" element={<Navigate to="calendar" replace />} />
            <Route
              path="calendar"
              element={
                <ErrorBoundary>
                  <Calendar />
                </ErrorBoundary>
              }
            />
            <Route
              path="templates"
              element={
                <ErrorBoundary>
                  <div className="templates-page">
                    <TemplateList
                      onSelectTemplate={setSelectedTemplateId}
                      selectedTemplateId={selectedTemplateId}
                    />
                    <TemplateManagement templateId={selectedTemplateId} />
                  </div>
                </ErrorBoundary>
              }
            />
            <Route
              path="telegram"
              element={
                <ErrorBoundary>
                  <TelegramManagement />
                </ErrorBoundary>
              }
            />
            <Route
              path="credits"
              element={
                <ErrorBoundary>
                  <MethodologistCreditsView />
                </ErrorBoundary>
              }
            />
          </Routes>
        </ErrorBoundary>
      </main>
    </div>
  );
};

export default MethodologistDashboard;
