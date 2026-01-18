# Production Chat System Guide

Это руководство описывает как работает система автоматического создания чатов на production и как её мониторить.

## Архитектура системы

Чаты в THE_BOT создаются **автоматически** когда бронь становится активной:

```
Преподаватель создает урок → Студент бронирует урок → Бронь активируется →
→ Триггер БД срабатывает → Чат автоматически создается
```

---

## Когда создаётся чат

### Сценарий 1: Автоматическое создание (Trigger при создании/активации брони)

**Условия:**
- Бронь вставляется с `status = 'active'` ИЛИ обновляется на `status = 'active'`
- Связанный урок существует в таблице `lessons`
- Учитель и студент - разные люди

**Процесс:**
1. Система создает новую бронь с `status = 'active'` (INSERT) или обновляет статус (UPDATE)
2. Триггер `booking_create_chat` или `booking_create_chat_update` срабатывает (BEFORE)
3. Функция `create_chat_on_booking_active()` получает `teacher_id` из урока
4. Чат создаётся между преподавателем и студентом
5. UNIQUE constraint защищает от дублирования: `UNIQUE(teacher_id, student_id)`

**Время создания:** Автоматически при INSERT/UPDATE брони (когда становится активной)

### Сценарий 2: При первом обращении в чат

Если по какой-то причине чат не был создан триггером, то при попытке открыть чат:

```go
// backend/internal/service/chat_service.go
// GetOrCreateRoom() вызывает chatRepo.GetOrCreateRoom()
// Если чата нет → создаёт на лету
```

**Время создания:** Лениво, при первом обращении пользователя

### Сценарий 3: Batch Backfill (при миграции БД)

При деплое новой версии с миграцией 050:

```sql
-- Миграция 050 запускает эту функцию:
-- Находит все активные брони без чатов и создаёт их
```

**Время создания:** Один раз при первом деплое (миграция 050)

---

## База данных: Структура

### Таблица: lessons
```sql
CREATE TABLE lessons (
    id UUID PRIMARY KEY,
    teacher_id UUID NOT NULL REFERENCES users(id),
    student_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMP,
    end_time TIMESTAMP,      -- Ключевое поле для триггера
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP      -- Мягкое удаление
);
```

### Таблица: bookings
```sql
CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    lesson_id UUID NOT NULL REFERENCES lessons(id),
    student_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50),       -- 'active', 'completed', 'cancelled'
    created_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### Таблица: chat_rooms
```sql
CREATE TABLE chat_rooms (
    id UUID PRIMARY KEY,
    teacher_id UUID NOT NULL REFERENCES users(id),
    student_id UUID NOT NULL REFERENCES users(id),
    last_message_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,

    UNIQUE(teacher_id, student_id)  -- Защита от дублирования
);
```

### Таблица: messages
```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    chat_room_id UUID NOT NULL REFERENCES chat_rooms(id),
    sender_id UUID NOT NULL REFERENCES users(id),
    message_text TEXT,
    status VARCHAR(50),       -- 'pending_moderation', 'delivered', 'blocked'
    moderation_completed_at TIMESTAMP,
    created_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

---

## Триггер и функции

### Триггер 1: booking_create_chat (INSERT)

```sql
CREATE TRIGGER booking_create_chat
BEFORE INSERT ON bookings
FOR EACH ROW
WHEN (NEW.status = 'active')
EXECUTE FUNCTION create_chat_on_booking_active();
```

**Условия срабатывания:**
- Событие: `BEFORE INSERT` на таблице `bookings`
- `NEW.status = 'active'` (новая бронь активна)

**Что происходит:**
1. Функция `create_chat_on_booking_active()` вызывается
2. Получает `teacher_id` из связанного урока
3. Создаёт чат между преподавателем и студентом
4. UNIQUE constraint предотвращает дублирование

### Триггер 2: booking_create_chat_update (UPDATE)

```sql
CREATE TRIGGER booking_create_chat_update
BEFORE UPDATE ON bookings
FOR EACH ROW
WHEN (NEW.status = 'active' AND OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION create_chat_on_booking_active();
```

**Условия срабатывания:**
- Событие: `BEFORE UPDATE` на таблице `bookings`
- `NEW.status = 'active'` (статус стал активным)
- `OLD.status IS DISTINCT FROM NEW.status` (статус изменился)

**Что происходит:**
1. Функция `create_chat_on_booking_active()` вызывается
2. Создаёт чат если его ещё нет
3. Срабатывает когда одобренная/отложенная бронь переходит в активное состояние

---

## Проверка на Production

### 1. Проверить что триггеры существуют

```bash
# На production сервере (mg@5.129.249.206):
ssh mg@5.129.249.206

# Подключиться к БД
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U postgres -d thebot_db
```

```sql
-- Проверить триггеры на таблице bookings
SELECT
    trigger_name,
    event_object_table,
    event_manipulation
FROM information_schema.triggers
WHERE event_object_table = 'bookings'
  AND trigger_name LIKE 'booking_create_chat%';

-- Результат должен показать 2 триггера:
-- booking_create_chat (BEFORE INSERT)
-- booking_create_chat_update (BEFORE UPDATE)
```

### 2. Проверить что функция существует

```sql
-- Функция для создания чата при активации брони
SELECT routine_name, routine_type
FROM information_schema.routines
WHERE routine_name = 'create_chat_on_booking_active'
AND routine_type = 'FUNCTION';

-- Должен вернуть одну функцию
```

### 3. Проверить статистику чатов

```sql
-- Статистика по урокам и броням
SELECT
    'Total lessons' as metric,
    COUNT(*) as count
FROM lessons
WHERE deleted_at IS NULL

UNION ALL

SELECT
    'Completed lessons (end_time < NOW)',
    COUNT(*)
FROM lessons
WHERE end_time < CURRENT_TIMESTAMP
  AND deleted_at IS NULL

UNION ALL

SELECT
    'Active bookings',
    COUNT(*)
FROM bookings
WHERE status = 'active'
  AND deleted_at IS NULL

UNION ALL

SELECT
    'Chat rooms created',
    COUNT(*)
FROM chat_rooms
WHERE deleted_at IS NULL

UNION ALL

SELECT
    'Chat rooms with messages',
    COUNT(DISTINCT cr.id)
FROM chat_rooms cr
INNER JOIN messages m ON m.chat_room_id = cr.id
WHERE cr.deleted_at IS NULL
  AND m.deleted_at IS NULL;
```

### 4. Найти кандидатов на создание чатов

```sql
-- Найти завершённые уроки без чатов
SELECT
    l.id as lesson_id,
    l.teacher_id,
    b.student_id,
    COUNT(b.id) as bookings_count,
    COUNT(DISTINCT cr.id) as existing_chats
FROM lessons l
INNER JOIN bookings b ON b.lesson_id = l.id AND b.status = 'active'
LEFT JOIN chat_rooms cr ON cr.teacher_id = l.teacher_id
    AND cr.student_id = b.student_id
    AND cr.deleted_at IS NULL
WHERE l.end_time < CURRENT_TIMESTAMP
  AND l.deleted_at IS NULL
  AND cr.id IS NULL
GROUP BY l.id, l.teacher_id, b.student_id
ORDER BY l.end_time DESC;

-- Результат: список уроков которые должны иметь чаты но их нет
```

---

## Тестовый сценарий: Вручную создать чат

### Вариант 1: Via SQL (локально)

```sql
-- 1. Найти существующих учителя и студента
SELECT id, first_name, role FROM users LIMIT 10;

-- 2. Создать тестовый урок в прошлом
INSERT INTO lessons (id, teacher_id, student_id, start_time, end_time, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    '11111111-1111-1111-1111-111111111111',  -- ID преподавателя
    '22222222-2222-2222-2222-222222222222',  -- ID студента
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour',  -- end_time в прошлом - важно!
    NOW(),
    NOW()
) RETURNING id as lesson_id;

-- 3. Создать активную бронь для этого урока
INSERT INTO bookings (id, lesson_id, student_id, status, created_at)
VALUES (
    gen_random_uuid(),
    '33333333-3333-3333-3333-333333333333',  -- ID урока из шага 2
    '22222222-2222-2222-2222-222222222222',  -- ID студента
    'active',
    NOW()
);

-- 4. Обновить урок чтобы сработал триггер
UPDATE lessons
SET updated_at = NOW()
WHERE id = '33333333-3333-3333-3333-333333333333';

-- 5. Проверить что чат создан
SELECT * FROM chat_rooms
WHERE teacher_id = '11111111-1111-1111-1111-111111111111'
  AND student_id = '22222222-2222-2222-2222-222222222222';

-- Результат: должен показать новый чат с created_at = NOW()
```

### Вариант 2: Via Backend API

```bash
# 1. Получить authentication token
curl -X POST https://the-bot.ru/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "teacher@example.com", "password": "password"}'

# Сохранить TOKEN из ответа

# 2. Получить или создать чат с конкретным студентом
curl -X POST https://the-bot.ru/api/chats/get-or-create \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"other_user_id": "STUDENT_UUID"}'

# Результат: JSON с chat_room object, включая room_id
```

---

## Команды для мониторинга

### Backend сервис

```bash
# На production сервере (mg@5.129.249.206):

# Просмотр статуса backend
systemctl status thebot-backend.service

# Просмотр логов backend (real-time)
journalctl -u thebot-backend.service -f

# Поиск ошибок в логах backend
journalctl -u thebot-backend.service -p err --since "1 hour ago"

# Поиск сообщений о создании чатов
journalctl -u thebot-backend.service | grep -i "chat"

# Статистика логов
journalctl -u thebot-backend.service --since "1 hour ago" | wc -l
```

### Database логирование

```bash
# Включить логирование триггеров (на production)
sudo -u postgres psql -d thebot_db -c "
ALTER DATABASE thebot_db SET log_statement = 'all';
ALTER DATABASE thebot_db SET log_duration = on;
"

# Просмотр логов PostgreSQL
tail -f /var/log/postgresql/postgresql.log | grep -i "trigger\|chat"

# Выключить логирование (чтобы не заполнять диск)
sudo -u postgres psql -d thebot_db -c "
ALTER DATABASE thebot_db SET log_statement = 'none';
"
```

### Chat creation script

На production уже есть скрипт для проверки:

```bash
# На production сервере
cd /home/mg/the-bot

# Проверить статус чат-системы (read-only)
./scripts/deployment/verify-chat-creation.sh --check-only

# Проверить и выполнить backfill (если нужно создать недостающие чаты)
./scripts/deployment/verify-chat-creation.sh --backfill

# Проверить с подробным логированием
./scripts/deployment/verify-chat-creation.sh --check-only --verbose
```

---

## Troubleshooting

### Проблема: Чаты не создаются

**Диагностика:**

```sql
-- 1. Проверить триггеры существуют на таблице bookings
SELECT * FROM information_schema.triggers
WHERE event_object_table = 'bookings'
  AND trigger_name LIKE 'booking_create_chat%';

-- Должен вернуть 2 строки. Если пусто → триггеры не созданы или удалены

-- 2. Проверить функцию существует
SELECT * FROM information_schema.routines
WHERE routine_name = 'create_chat_on_booking_active'
  AND routine_type = 'FUNCTION';

-- Должен вернуть 1 строку

-- 3. Проверить что есть активные брони
SELECT COUNT(*)
FROM bookings b
INNER JOIN lessons l ON b.lesson_id = l.id
WHERE b.status = 'active'
  AND l.deleted_at IS NULL
  AND b.deleted_at IS NULL;

-- Должен вернуть > 0

-- 4. Проверить что активные брони без чатов
SELECT COUNT(DISTINCT b.id) as bookings_without_chats
FROM bookings b
INNER JOIN lessons l ON b.lesson_id = l.id
LEFT JOIN chat_rooms cr ON cr.teacher_id = l.teacher_id
    AND cr.student_id = b.student_id
WHERE b.status = 'active'
  AND l.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND cr.id IS NULL;

-- Если > 0 → есть брони без чатов
```

**Решение:**

1. Проверить логи backend: `journalctl -u thebot-backend.service`
2. Проверить логи PostgreSQL: `tail -f /var/log/postgresql/postgresql.log`
3. Проверить применена ли миграция 050: `./scripts/deployment/verify-chat-creation.sh --check-only`
4. Если хотите создать чаты для существующих активных броней, запустите backfill

### Проблема: Дублирующиеся чаты

**Причина:** Нарушение UNIQUE constraint в чём-то невозможно - защита встроена в БД.

**Диагностика:**

```sql
-- Найти дублирующиеся чаты (не должно быть)
SELECT
    teacher_id,
    student_id,
    COUNT(*) as count
FROM chat_rooms
WHERE deleted_at IS NULL
GROUP BY teacher_id, student_id
HAVING COUNT(*) > 1;

-- Если вернул результаты → консистентность нарушена (критично!)
```

**Решение:**

Если есть дубли → это ошибка в коде приложения. Написать issue в GitHub.

### Проблема: Триггер не срабатывает

**Причина:** Триггер срабатывает ТОЛЬКО когда бронь переходит в статус `'active'`

**Диагностика:**

```sql
-- Проверить последние изменения броней
SELECT id, status, created_at, updated_at
FROM bookings
WHERE deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT 10;

-- Проверить броњи в статусе 'active'
SELECT COUNT(*) as active_bookings
FROM bookings
WHERE status = 'active'
  AND deleted_at IS NULL;

-- Если = 0, то нет активных броней - триггер не сработает
```

**Решение:**

Триггер срабатывает при INSERT/UPDATE брони со статусом `'active'`. Если бронь создана с другим статусом, нужно обновить её статус:

```sql
-- Обновить статус брони на 'active' для срабатывания триггера
UPDATE bookings
SET status = 'active', updated_at = CURRENT_TIMESTAMP
WHERE id = 'YOUR_BOOKING_ID' AND status != 'active';
```

### Проблема: Backend возвращает ошибку при открытии чата

**Симптомы:** API возвращает 500 при `/api/chats/get-or-create`

**Диагностика:**

```bash
# Проверить логи backend
journalctl -u thebot-backend.service -f

# Поискать строку с чатом и ошибкой
journalctl -u thebot-backend.service | grep -i "error" | grep -i "chat"
```

**Решение:**

1. Проверить что оба пользователя (teacher и student) существуют в БД:
   ```sql
   SELECT id, email, role FROM users WHERE id IN ('USER_ID_1', 'USER_ID_2');
   ```

2. Проверить что в таблице `lessons` есть связь:
   ```sql
   SELECT * FROM lessons WHERE teacher_id = 'TEACHER_ID' LIMIT 1;
   ```

3. Проверить что в коде `chat_service.go` нет bagов (смотри код в `backend/internal/service/chat_service.go`)

---

## Логирование и мониторинг

### На Production должны быть настроены:

1. **PostgreSQL logs** для отслеживания триггеров:
   ```bash
   # На сервере mg@5.129.249.206
   tail -f /var/log/postgresql/postgresql.log | grep "create_chat"
   ```

2. **Backend logs** в systemd:
   ```bash
   journalctl -u thebot-backend.service -f
   ```

3. **Metrics** (опционально, в Prometheus если есть):
   - Количество созданных чатов за период
   - Время lag между `end_time` и `created_at` чата
   - Ошибки в chat creation логах

### Метрики для мониторинга

```sql
-- 1. Количество чатов созданных за последний день
SELECT COUNT(*) as chats_created_today
FROM chat_rooms
WHERE created_at >= CURRENT_DATE
  AND deleted_at IS NULL;

-- 2. Средний lag между end_time урока и created_at чата
SELECT
    AVG(EXTRACT(EPOCH FROM (cr.created_at - l.end_time))) as avg_lag_seconds,
    MAX(EXTRACT(EPOCH FROM (cr.created_at - l.end_time))) as max_lag_seconds
FROM chat_rooms cr
INNER JOIN lessons l ON cr.teacher_id = l.teacher_id
WHERE cr.created_at >= CURRENT_DATE - INTERVAL '7 days'
  AND l.end_time < cr.created_at;  -- Только завершённые уроки

-- 3. Количество очень долгих лагов (> 1 часа)
SELECT COUNT(*) as slow_creations
FROM chat_rooms cr
INNER JOIN lessons l ON cr.teacher_id = l.teacher_id
WHERE cr.created_at >= CURRENT_DATE - INTERVAL '7 days'
  AND EXTRACT(EPOCH FROM (cr.created_at - l.end_time)) > 3600;
```

---

## Сценарий: Проверить всё перед production deploy

```bash
#!/bin/bash
# Run on production server before/after deploy

set -euo pipefail

echo "=== Chat System Pre-Deployment Checks ==="

# 1. Check database connection
echo "1. Testing database connection..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db -c "SELECT version();" > /dev/null
echo "   ✓ Database connected"

# 2. Check trigger exists
echo "2. Checking trigger 'lesson_completion_create_chat'..."
TRIGGER_COUNT=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db \
    -t -c "SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_name = 'lesson_completion_create_chat';")
if [ "$TRIGGER_COUNT" -gt 0 ]; then
    echo "   ✓ Trigger exists"
else
    echo "   ✗ Trigger NOT FOUND (critical error)"
    exit 1
fi

# 3. Check functions
echo "3. Checking PL/pgSQL functions..."
FUNC1=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db \
    -t -c "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name = 'create_chat_after_lesson_completion';")
FUNC2=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db \
    -t -c "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name = 'create_chats_for_completed_lessons';")

if [ "$FUNC1" -gt 0 ] && [ "$FUNC2" -gt 0 ]; then
    echo "   ✓ Both functions exist"
else
    echo "   ✗ Functions missing"
    exit 1
fi

# 4. Check for candidates
echo "4. Checking for missing chats..."
MISSING=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db \
    -t -c "SELECT COUNT(DISTINCT b.id)
FROM lessons l
INNER JOIN bookings b ON b.lesson_id = l.id AND b.status = 'active'
LEFT JOIN chat_rooms cr ON cr.teacher_id = l.teacher_id AND cr.student_id = b.student_id
WHERE l.end_time < CURRENT_TIMESTAMP AND l.deleted_at IS NULL AND cr.id IS NULL;")

if [ "$MISSING" -gt 0 ]; then
    echo "   ⚠ Found $MISSING missing chats - running backfill..."
    CREATED=$(PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db \
        -t -c "SELECT create_chats_for_completed_lessons();")
    echo "   ✓ Created $CREATED chats"
else
    echo "   ✓ No missing chats found"
fi

# 5. Check backend service
echo "5. Checking backend service status..."
if systemctl is-active --quiet thebot-backend.service; then
    echo "   ✓ Backend service is running"
else
    echo "   ✗ Backend service is NOT running"
    exit 1
fi

echo ""
echo "=== All checks passed! ==="
```

---

## Быстрые команды для production

```bash
# На сервере mg@5.129.249.206

# Проверить статус всех систем
./scripts/deployment/verify-chat-creation.sh --check-only

# Создать недостающие чаты (backfill)
./scripts/deployment/verify-chat-creation.sh --backfill

# Просмотр логов chat creation
journalctl -u thebot-backend.service | grep -i "chat"

# Подключиться к БД и выполнить SQL
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -d thebot_db

# Перезапустить backend если нужно
sudo systemctl restart thebot-backend.service

# Просмотр статуса backend
systemctl status thebot-backend.service
```

---

## Резюме

### Как это работает:
1. Система создает бронь с `status = 'active'` или обновляет её статус на `'active'`
2. Триггер `booking_create_chat` (INSERT) или `booking_create_chat_update` (UPDATE) срабатывает (BEFORE)
3. Функция `create_chat_on_booking_active()` получает `teacher_id` из урока и создаёт чат
4. UNIQUE constraint защищает от дублирования
5. Если триггер не сработал (редко) → `GetOrCreateRoom()` создаёт чат лениво при обращении

### Что мониторить:
- Количество активных броней vs. созданных чатов
- Ошибки в логах backend при создании чатов
- Наличие триггеров и функции в БД
- Логи БД на предмет ошибок в триггерах

### При проблемах:
1. Проверить наличие триггеров на таблице `bookings`: `SELECT * FROM information_schema.triggers WHERE event_object_table = 'bookings'`
2. Проверить логи: `journalctl -u thebot-backend.service -f`
3. Проверить логи БД: `tail -f /var/log/postgresql/postgresql.log`
4. Если броньки уже существуют без чатов, можно создать их вручную через backend API
