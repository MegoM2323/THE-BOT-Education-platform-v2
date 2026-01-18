import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import './PaymentSuccessPage.css';

/**
 * Страница успешной оплаты
 * Показывается после возврата с YooKassa
 */
export const PaymentSuccessPage = () => {
  const navigate = useNavigate();

  useEffect(() => {
    // Через 3 секунды редирект на страницу оплаты
    const timer = setTimeout(() => {
      navigate('/dashboard/student/payment');
    }, 3000);

    return () => clearTimeout(timer);
  }, [navigate]);

  return (
    <div className="payment-success-page" data-testid="payment-success-page">
      <div className="payment-success-content">
        <div className="success-icon">✓</div>
        <h1 className="success-title">Платеж обрабатывается</h1>
        <p className="success-message">
          Ваш платеж успешно создан и обрабатывается.
          <br />
          Кредиты будут начислены на ваш счет в течение нескольких минут.
        </p>
        <p className="redirect-message">
          Через несколько секунд вы будете перенаправлены на страницу оплаты...
        </p>
      </div>
    </div>
  );
};

export default PaymentSuccessPage;
