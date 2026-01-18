import { useState, useEffect } from "react";
import { Routes, Route, Navigate, useNavigate } from "react-router-dom";
import Sidebar from "../components/common/Sidebar.jsx";
import ErrorBoundary from "../components/common/ErrorBoundary.jsx";
import StudentErrorBoundary from "../components/student/StudentErrorBoundary.jsx";
import StudentSchedule from "./StudentSchedule.jsx";
import StudentHomework from "./StudentHomework.jsx";
import CreditsHistory from "../components/student/CreditsHistory.jsx";
import StudentProfile from "../components/student/StudentProfile.jsx";
import Spinner from "../components/common/Spinner.jsx";
import { useCredits } from "../hooks/useCredits.js";
import { useAuth } from "../hooks/useAuth.js";
import { ROLES } from "../utils/constants.js";
// HIDDEN: Payment and Chat sections temporarily disabled
// import PaymentPage from './PaymentPage.js';
// import PaymentSuccessPage from './PaymentSuccessPage.js';
// import ChatPage from './ChatPage.js';
import "./Dashboard.css";

export const StudentDashboard = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [isAuthorizing, setIsAuthorizing] = useState(true);
  const { balance } = useCredits();
  const { user, loading } = useAuth();
  const navigate = useNavigate();

  // Проверка роли пользователя при монтировании компонента
  useEffect(() => {
    // Ждем завершения загрузки данных пользователя
    if (loading) {
      return;
    }

    // Проверяем роль пользователя
    if (!user) {
      // Пользователь не авторизован - редирект на логин
      navigate("/login", { replace: true });
      return;
    }

    if (user.role !== ROLES.STUDENT) {
      // Пользователь авторизован, но не является студентом
      // Перенаправляем на соответствующий дашборд
      if (user.role === ROLES.ADMIN) {
        navigate("/dashboard/admin", { replace: true });
      } else {
        // Неизвестная роль - на главную
        navigate("/", { replace: true });
      }
      return;
    }

    // Пользователь - студент, можно показывать контент
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

  // Показываем спиннер во время проверки авторизации
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

  // Защита от отображения контента неавторизованным пользователям
  if (!user || user.role !== ROLES.STUDENT) {
    return null;
  }

  const navLinks = [
    {
      path: "/dashboard/student/schedule",
      label: "Расписание",
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
      testId: "nav-schedule",
    },
    {
      path: "/dashboard/student/homework",
      label: "Домашнее задание",
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
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
      ),
      testId: "nav-homework",
    },
    // HIDDEN: Chat section temporarily disabled
    // {
    //   path: '/dashboard/student/chat',
    //   label: 'Чат',
    //   icon: <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>,
    //   testId: 'nav-chat'
    // },
    // HIDDEN: Payment section temporarily disabled
    // {
    //   path: '/dashboard/student/payment',
    //   label: 'Оплата',
    //   icon: <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" /></svg>,
    //   testId: 'nav-payment'
    // },
    {
      path: "/dashboard/student/credits",
      label: "История кредитов",
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
            d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
          />
        </svg>
      ),
      testId: "nav-credits",
    },
    {
      path: "/dashboard/student/profile",
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
  ];

  return (
    <div className="dashboard">
      <Sidebar
        links={navLinks}
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        gridLayout={false}
        credits={balance ?? 0}
      />

      <main
        className={`dashboard-main ${sidebarCollapsed ? "sidebar-collapsed" : ""}`}
      >
        <ErrorBoundary>
          <Routes>
            <Route path="/" element={<Navigate to="schedule" replace />} />
            <Route
              path="schedule"
              element={
                <StudentErrorBoundary>
                  <StudentSchedule />
                </StudentErrorBoundary>
              }
            />
            <Route
              path="homework"
              element={
                <StudentErrorBoundary>
                  <StudentHomework />
                </StudentErrorBoundary>
              }
            />
            {/* HIDDEN: Chat routes temporarily disabled */}
            {/* <Route path="chat" element={<StudentErrorBoundary><ChatPage /></StudentErrorBoundary>} /> */}
            {/* <Route path="chat/:roomId" element={<StudentErrorBoundary><ChatPage /></StudentErrorBoundary>} /> */}
            {/* HIDDEN: Payment routes temporarily disabled */}
            {/* <Route path="payment" element={<StudentErrorBoundary><PaymentPage /></StudentErrorBoundary>} /> */}
            {/* <Route path="payment-success" element={<StudentErrorBoundary><PaymentSuccessPage /></StudentErrorBoundary>} /> */}
            <Route
              path="credits"
              element={
                <StudentErrorBoundary>
                  <CreditsHistory />
                </StudentErrorBoundary>
              }
            />
            <Route
              path="profile"
              element={
                <StudentErrorBoundary>
                  <StudentProfile />
                </StudentErrorBoundary>
              }
            />
          </Routes>
        </ErrorBoundary>
      </main>
    </div>
  );
};

export default StudentDashboard;
