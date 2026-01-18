-- 030_add_payment_idempotency.sql
-- Миграция для добавления поддержки идемпотентности платежей
-- Предотвращает обработку одного платежа дважды

-- Добавляем колонку idempotency_key для предотвращения дубликатов при обработке webhook
ALTER TABLE payments
ADD COLUMN idempotency_key VARCHAR(255) UNIQUE,
ADD COLUMN processed_at TIMESTAMP WITH TIME ZONE;

-- Комментарии к новым колонкам
COMMENT ON COLUMN payments.idempotency_key IS 'Ключ идемпотентности для предотвращения обработки одного платежа дважды (обычно ID платежа)';
COMMENT ON COLUMN payments.processed_at IS 'Время успешной обработки платежа (update credits)';

-- Индекс для быстрого поиска по ключу идемпотентности
CREATE INDEX idx_payments_idempotency_key ON payments(idempotency_key)
WHERE idempotency_key IS NOT NULL;
