import { useState, useEffect } from "react";
import { Routes, Route, Navigate, useNavigate } from "react-router-dom";
import Sidebar from "../components/common/Sidebar.jsx";
import ErrorBoundary from "../components/common/ErrorBoundary.jsx";
import MethodologistCalendar from "../components/methodologist/MethodologistCalendar.jsx";
import MethodologistProfile from "../components/methodologist/MethodologistProfile.jsx";
import TemplateManagement from "../components/admin/TemplateManagement.jsx";
import TemplateList from "../components/admin/TemplateList.jsx";
import Spinner from "../components/common/Spinner.jsx";
import MethodologistCreditsView from "../components/methodologist/MethodologistCreditsView.jsx";
import ChatPage from "./ChatPage.jsx";
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
      path: "/dashboard/methodologist/profile",
      label: "Профиль",
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
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
          />
        </svg>
      ),
      testId: "nav-profile",
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
    {
      path: "/dashboard/methodologist/chat",
      label: "Чат",
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
            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
          />
        </svg>
      ),
      testId: "nav-methodologist-chat",
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
                  <MethodologistCalendar />
                </ErrorBoundary>
              }
            />
            <Route
              path="profile"
              element={
                <ErrorBoundary>
                  <MethodologistProfile />
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
              path="credits"
              element={
                <ErrorBoundary>
                  <MethodologistCreditsView />
                </ErrorBoundary>
              }
            />
            <Route
              path="chat"
              element={
                <ErrorBoundary>
                  <ChatPage />
                </ErrorBoundary>
              }
            />
            <Route
              path="chat/:roomId"
              element={
                <ErrorBoundary>
                  <ChatPage />
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
