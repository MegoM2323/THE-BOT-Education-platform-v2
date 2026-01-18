import LoginForm from '../components/auth/LoginForm.jsx';
import './Login.css';

export const Login = () => {

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-card">
          <h1 className="login-title">Вход в систему</h1>
          <p className="login-subtitle">Войдите для доступа к платформе</p>
          <LoginForm />
        </div>
      </div>
    </div>
  );
};

export default Login;
