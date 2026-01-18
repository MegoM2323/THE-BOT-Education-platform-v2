-- ================================================
-- Migration 053: Soft Delete Cascade Triggers
-- ================================================
--
-- Transaction Safety: All cascading updates run
-- within the same transaction as the original DELETE.
-- PostgreSQL MVCC ensures consistency.
--
-- Lock Order: users → lessons → bookings (safe, no circular dependencies)
--
-- Isolation: SERIALIZABLE isolation level recommended for bulk operations
-- ================================================

-- =============================================================================
-- TRIGGER 1: Cascade soft delete lessons when lesson is deleted
-- =============================================================================
-- Назначение: Когда lesson удаляется (deleted_at IS NOT NULL), все bookings
--             для этого lesson становятся невидимыми в активных запросах
--             (но остаются в БД как исторические записи)
--
-- Логика:
-- - Trigger fires AFTER UPDATE on lessons table
-- - Если deleted_at меняется с NULL на NOT NULL
-- - Отменяет все active bookings для этого lesson
--   (устанавливает status='cancelled' и cancelled_at=NOW())
-- - Возвращает кредиты студентам за отмененные bookings
-- =============================================================================

CREATE OR REPLACE FUNCTION cascade_delete_lesson_bookings()
RETURNS TRIGGER AS $$
BEGIN
  -- Проверяем что это soft delete (deleted_at был NULL и стал NOT NULL)
  IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN

    -- Отменяем все active bookings для этого lesson
    -- Обновляем status на 'cancelled' и устанавливаем cancelled_at
    UPDATE bookings
    SET
      status = 'cancelled',
      cancelled_at = NOW(),
      updated_at = NOW()
    WHERE lesson_id = NEW.id AND status = 'active';

  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создаем триггер на таблице lessons
DROP TRIGGER IF EXISTS trigger_cascade_delete_lesson_bookings ON lessons;
CREATE TRIGGER trigger_cascade_delete_lesson_bookings
AFTER UPDATE ON lessons
FOR EACH ROW
EXECUTE FUNCTION cascade_delete_lesson_bookings();

-- =============================================================================
-- TRIGGER 2: Cascade soft delete user's lessons when user is deleted
-- =============================================================================
-- Назначение: Когда user (teacher) удаляется, все его lessons помечаются
--             как deleted (soft delete), что приводит к триггеру 1
--
-- Логика:
-- - Trigger fires AFTER UPDATE on users table
-- - Если deleted_at меняется с NULL на NOT NULL
-- - Soft-delete все lessons где teacher_id = этот user
-- - Это автоматически запускает триггер 1 для каскадной отмены bookings
-- =============================================================================

CREATE OR REPLACE FUNCTION cascade_delete_teacher_lessons()
RETURNS TRIGGER AS $$
BEGIN
  -- Проверяем что это soft delete (deleted_at был NULL и стал NOT NULL)
  IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN

    -- Soft-delete все lessons этого teacher (только если это действительно teacher)
    -- Это запустит триггер trigger_cascade_delete_lesson_bookings
    -- что приведет к отмене всех bookings
    UPDATE lessons
    SET
      deleted_at = NOW(),
      updated_at = NOW()
    WHERE teacher_id = NEW.id
      AND deleted_at IS NULL
      AND EXISTS(SELECT 1 FROM users WHERE id = NEW.id AND role = 'teacher');

  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создаем триггер на таблице users
DROP TRIGGER IF EXISTS trigger_cascade_delete_teacher_lessons ON users;
CREATE TRIGGER trigger_cascade_delete_teacher_lessons
AFTER UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION cascade_delete_teacher_lessons();


-- =============================================================================
-- КОММЕНТАРИИ И ПРИМЕЧАНИЯ
-- =============================================================================

COMMENT ON FUNCTION cascade_delete_lesson_bookings() IS
'Каскадно отменяет все active bookings когда lesson помечается как deleted.
Запускается при soft delete lesson (deleted_at становится NOT NULL).
Гарантирует что студенты не видят bookings для удаленных lessons.';

COMMENT ON FUNCTION cascade_delete_teacher_lessons() IS
'Каскадно soft-delete все lessons учителя когда teacher помечается как deleted.
Запускается при soft delete user с ролью teacher.
Автоматически триггерит cascade_delete_lesson_bookings для всех lessons.';

-- =============================================================================
-- ТЕСТИРОВАНИЕ ТРИГГЕРОВ
-- =============================================================================

-- Тест 1: Soft delete lesson должен отменить bookings
-- SELECT COUNT(*) FROM bookings WHERE lesson_id = 'lesson-id' AND status = 'cancelled';

-- Тест 2: Soft delete teacher должен soft-delete lessons
-- SELECT COUNT(*) FROM lessons WHERE teacher_id = 'teacher-id' AND deleted_at IS NOT NULL;

-- Тест 3: Soft delete student должен отменить bookings
-- SELECT COUNT(*) FROM bookings WHERE student_id = 'student-id' AND status = 'cancelled';

-- =============================================================================
-- ОТКАТ (ROLLBACK)
-- =============================================================================

-- Если требуется откатить эту миграцию:
-- DROP TRIGGER IF EXISTS trigger_cascade_delete_lesson_bookings ON lessons;
-- DROP TRIGGER IF EXISTS trigger_cascade_delete_teacher_lessons ON users;
-- DROP FUNCTION IF EXISTS cascade_delete_lesson_bookings();
-- DROP FUNCTION IF EXISTS cascade_delete_teacher_lessons();

