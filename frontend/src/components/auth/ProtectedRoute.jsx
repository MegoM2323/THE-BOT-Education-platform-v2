import { Navigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth.js';
import Spinner from '../common/Spinner.jsx';

/**
 * ProtectedRoute - предотвращает race condition при проверке аутентификации
 *
 * Fixes FB002 (B4) - HIGH Race condition in ProtectedRoute
 *
 * Гарантии:
 * - Показывает loading до завершения проверки аутентификации
 * - Не делает redirect пока loading === true
 * - Предотвращает flash неавторизованного контента
 * - Использует replace для предотвращения добавления в history
 */
export const ProtectedRoute = ({ children, allowedRoles = [] }) => {
  const { user, loading, isAuthenticated } = useAuth();

  // CRITICAL: Показываем loading пока auth state не определён
  // Это предотвращает flash login page или unauthorized content
  if (loading) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: '#fff',
          zIndex: 9999
        }}
        data-testid="protected-route-loading"
        role="status"
        aria-label="Checking authentication"
      >
        <Spinner size="lg" />
      </div>
    );
  }

  // Auth check завершён, теперь безопасно проверять isAuthenticated
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Проверка ролей только после того как user загружен
  if (allowedRoles.length > 0 && !allowedRoles.includes(user?.role)) {
    return <Navigate to="/" replace />;
  }

  return children;
};

export default ProtectedRoute;
