import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext.jsx';
import { NotificationProvider } from './context/NotificationContext.jsx';
import ErrorBoundary from './components/common/ErrorBoundary.jsx';
import Header from './components/layout/Header.jsx';
import Notification from './components/common/Notification.jsx';
import ProtectedRoute from './components/auth/ProtectedRoute.jsx';
import AuthEventHandler from './components/auth/AuthEventHandler.jsx';
import Landing from './pages/Landing.jsx';
import Login from './pages/Login.jsx';
import TelegramRegister from './pages/TelegramRegister.jsx';
import StudentDashboard from './pages/StudentDashboard.jsx';
import AdminDashboard from './pages/AdminDashboard.jsx';
import MethodologistDashboard from './pages/MethodologistDashboard.jsx';
import ChatPage from './pages/ChatPage.jsx';
import { ROLES } from './utils/constants.js';

function App() {
  return (
    <ErrorBoundary>
      <Router>
        <AuthProvider>
          <NotificationProvider>
            <AuthEventHandler />
            <Header />
            <Notification />
            <ErrorBoundary>
              <Routes>
                <Route path="/" element={<Landing />} />
                <Route path="/login" element={<Login />} />
                <Route path="/telegram-register" element={<TelegramRegister />} />

                <Route
                  path="/dashboard/student/*"
                  element={
                    <ProtectedRoute allowedRoles={[ROLES.STUDENT]}>
                      <StudentDashboard />
                    </ProtectedRoute>
                  }
                />

                <Route
                  path="/dashboard/admin/*"
                  element={
                    <ProtectedRoute allowedRoles={[ROLES.ADMIN]}>
                      <AdminDashboard />
                    </ProtectedRoute>
                  }
                />

                <Route
                  path="/dashboard/methodologist/*"
                  element={
                    <ProtectedRoute allowedRoles={[ROLES.METHODOLOGIST]}>
                      <MethodologistDashboard />
                    </ProtectedRoute>
                  }
                />

                <Route
                  path="/chat"
                  element={
                    <ErrorBoundary>
                      <ProtectedRoute allowedRoles={[ROLES.STUDENT]}>
                        <ChatPage />
                      </ProtectedRoute>
                    </ErrorBoundary>
                  }
                />

                <Route
                  path="/chat/:roomId"
                  element={
                    <ErrorBoundary>
                      <ProtectedRoute allowedRoles={[ROLES.STUDENT]}>
                        <ChatPage />
                      </ProtectedRoute>
                    </ErrorBoundary>
                  }
                />

                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </ErrorBoundary>
          </NotificationProvider>
        </AuthProvider>
      </Router>
    </ErrorBoundary>
  );
}

export default App;
