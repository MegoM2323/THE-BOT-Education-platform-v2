import PaymentCard from '../components/student/PaymentCard.jsx';
import PaymentHistory from '../components/student/PaymentHistory.jsx';
import './PaymentPage.css';

/**
 * Страница оплаты для студента
 * Позволяет покупать кредиты через YooKassa и просматривать историю платежей
 */
export const PaymentPage = () => {
  return (
    <div className="payment-page" data-testid="payment-page">
      <div className="payment-page-header">
        <h1 className="payment-page-title">Оплата кредитов</h1>
        <p className="payment-page-description">
          Купите кредиты для бронирования занятий
        </p>
      </div>

      <div className="payment-page-content">
        <PaymentCard />
        <PaymentHistory />
      </div>
    </div>
  );
};

export default PaymentPage;
