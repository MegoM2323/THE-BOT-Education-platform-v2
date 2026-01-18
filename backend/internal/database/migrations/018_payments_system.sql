-- 014_payments_system.sql
-- Миграция для интеграции платежной системы YooKassa
-- Таблица payments хранит историю платежей (1 кредит = 2800₽)

-- Создание таблицы payments
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK(amount > 0),
    credits INT NOT NULL CHECK(credits > 0),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'succeeded', 'cancelled')),
    yookassa_payment_id VARCHAR(255) UNIQUE,
    confirmation_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Комментарии к таблице и столбцам
COMMENT ON TABLE payments IS 'История платежей через YooKassa (1 кредит = 2800₽)';
COMMENT ON COLUMN payments.user_id IS 'Пользователь, совершивший платеж';
COMMENT ON COLUMN payments.amount IS 'Сумма платежа в рублях';
COMMENT ON COLUMN payments.credits IS 'Количество приобретаемых кредитов';
COMMENT ON COLUMN payments.status IS 'pending: ожидает оплаты, succeeded: успешно оплачен, cancelled: отменен';
COMMENT ON COLUMN payments.yookassa_payment_id IS 'ID платежа в системе YooKassa для webhook и отслеживания';
COMMENT ON COLUMN payments.confirmation_url IS 'URL для подтверждения платежа (redirect на страницу оплаты)';

-- Индексы для производительности
-- Основной индекс для истории платежей пользователя
CREATE INDEX idx_payments_user ON payments(user_id, created_at DESC);

-- Индекс для webhook-обработки (поиск по yookassa_payment_id)
CREATE INDEX idx_payments_yookassa_id ON payments(yookassa_payment_id)
WHERE yookassa_payment_id IS NOT NULL;

-- Частичный индекс для поиска незавершенных платежей (для фоновых задач очистки)
CREATE INDEX idx_payments_pending_status ON payments(status, created_at)
WHERE status = 'pending';

-- Trigger для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_payment_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER payments_update_timestamp
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION update_payment_timestamp();

-- Комментарии к индексам для понимания их назначения
COMMENT ON INDEX idx_payments_user IS 'Оптимизация запросов истории платежей пользователя';
COMMENT ON INDEX idx_payments_yookassa_id IS 'Быстрый поиск платежа по ID YooKassa для webhook';
COMMENT ON INDEX idx_payments_pending_status IS 'Поиск незавершенных платежей для очистки/напоминаний';
