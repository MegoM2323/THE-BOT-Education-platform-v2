import { useState } from 'react';
import { logger } from '../../utils/logger.js';
import { validatePaymentRedirectUrl } from '../../utils/validation.js';
import { usePayments } from '../../hooks/usePayments.js';
import { useStudentCredits } from '../../hooks/useStudentCredits.js';
import Spinner from '../common/Spinner.jsx';
import './PaymentCard.css';

/**
 * Карточка покупки кредитов
 */
export const PaymentCard = () => {
  const { createPayment, isCreatingPayment } = usePayments();
  const { credits } = useStudentCredits();
  const [isProcessing, setIsProcessing] = useState(false);
  const [creditQuantity, setCreditQuantity] = useState(1);
  const [validationError, setValidationError] = useState('');

  const PRICE_PER_CREDIT = 2800;
  const MIN_CREDITS = 1;
  const MAX_CREDITS = 100;

  // Рассчитать общую стоимость (обработка пустого значения)
  const totalPrice = (creditQuantity === '' || isNaN(creditQuantity)) ? 0 : creditQuantity * PRICE_PER_CREDIT;

  // Валидация количества кредитов
  const handleQuantityChange = (e) => {
    const value = e.target.value;

    // Разрешить пустое поле для удобства ввода
    if (value === '') {
      setCreditQuantity('');
      setValidationError('');
      return;
    }

    const numValue = parseInt(value, 10);

    // Проверка на корректное число
    if (isNaN(numValue)) {
      setValidationError('Введите корректное число');
      return;
    }

    // Проверка минимума
    if (numValue < MIN_CREDITS) {
      setCreditQuantity(numValue);
      setValidationError(`Минимум ${MIN_CREDITS} кредит`);
      return;
    }

    // Проверка максимума
    if (numValue > MAX_CREDITS) {
      setCreditQuantity(MAX_CREDITS);
      setValidationError(`Максимум ${MAX_CREDITS} кредитов`);
      return;
    }

    // Валидное значение
    setCreditQuantity(numValue);
    setValidationError('');
  };

  // Обработка потери фокуса - установить минимум если пусто
  const handleBlur = () => {
    if (creditQuantity === '' || creditQuantity < MIN_CREDITS) {
      setCreditQuantity(MIN_CREDITS);
      setValidationError('');
    }
  };

  const handleBuyCredits = async () => {
    if (isProcessing || isCreatingPayment) return;

    // Финальная валидация перед покупкой
    const quantity = parseInt(creditQuantity, 10);
    if (isNaN(quantity) || quantity < MIN_CREDITS || quantity > MAX_CREDITS) {
      setValidationError(`Введите число от ${MIN_CREDITS} до ${MAX_CREDITS}`);
      return;
    }

    try {
      setIsProcessing(true);
      setValidationError('');
      const response = await createPayment(quantity);

      // Валидация URL редиректа перед переходом
      if (response.confirmation_url) {
        const validation = validatePaymentRedirectUrl(response.confirmation_url);

        if (validation.isValid) {
          // URL безопасный - выполняем редирект
          logger.info('Redirecting to payment gateway', { url: validation.sanitizedUrl });
          window.location.href = validation.sanitizedUrl;
        } else {
          // URL не прошел валидацию - логируем и показываем ошибку
          logger.error('Invalid payment redirect URL', {
            url: response.confirmation_url,
            error: validation.error,
          });
          console.error('[Security] Payment redirect blocked:', validation.error);
          setValidationError('Ошибка безопасности: недопустимый URL платежной системы');
          setIsProcessing(false);
        }
      } else {
        console.error('No confirmation_url in response:', response);
        setValidationError('Ошибка создания платежа');
        setIsProcessing(false);
      }
    } catch (error) {
      console.error('Payment creation failed:', error);
      setValidationError('Ошибка создания платежа. Попробуйте еще раз.');
      setIsProcessing(false);
    }
  };

  return (
    <div className="payment-card" data-testid="payment-card">
      <h3 className="payment-card-title">Покупка кредитов</h3>

      <div className="credit-balance-section">
        <span className="balance-label">Текущий баланс:</span>
        <span className="balance-value" data-testid="current-balance">
          {credits?.balance ?? 0} {credits?.balance === 1 ? 'кредит' : 'кредитов'}
        </span>
      </div>

      <div className="quantity-section">
        <label htmlFor="credit-quantity" className="quantity-label">
          Количество кредитов:
        </label>
        <input
          id="credit-quantity"
          type="number"
          min={MIN_CREDITS}
          max={MAX_CREDITS}
          value={creditQuantity}
          onChange={handleQuantityChange}
          onBlur={handleBlur}
          className={`quantity-input ${validationError ? 'quantity-input-error' : ''}`}
          disabled={isProcessing || isCreatingPayment}
          data-testid="credit-quantity-input"
        />
        {validationError && (
          <div className="validation-error" data-testid="validation-error">
            {validationError}
          </div>
        )}
      </div>

      <div className="price-info">
        <div className="price-row">
          <span className="price-label">Цена за кредит:</span>
          <span className="price-value">{PRICE_PER_CREDIT.toLocaleString('ru-RU')}₽</span>
        </div>
        <div className="price-row total-row">
          <span className="price-label">Итого:</span>
          <span className="price-value total-price" data-testid="total-price">
            {totalPrice.toLocaleString('ru-RU')}₽
          </span>
        </div>
      </div>

      <button
        className="buy-credits-btn"
        onClick={handleBuyCredits}
        disabled={isProcessing || isCreatingPayment || !!validationError || creditQuantity === ''}
        data-testid="buy-credits-btn"
      >
        {isProcessing || isCreatingPayment ? (
          <>
            <Spinner size="small" />
            <span>Создание платежа...</span>
          </>
        ) : (
          `Купить ${creditQuantity || 0} ${creditQuantity === 1 ? 'кредит' : creditQuantity > 1 && creditQuantity < 5 ? 'кредита' : 'кредитов'}`
        )}
      </button>

      <div className="payment-info">
        <p className="payment-info-text">
          После нажатия кнопки вы будете перенаправлены на страницу оплаты YooKassa
        </p>
      </div>
    </div>
  );
};

export default PaymentCard;
