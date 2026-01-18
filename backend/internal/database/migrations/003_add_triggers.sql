-- 003_add_triggers.sql
-- Database triggers for automation and validation

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers to all relevant tables
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credits_updated_at
    BEFORE UPDATE ON credits
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lessons_updated_at
    BEFORE UPDATE ON lessons
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically create credits account for new students
CREATE OR REPLACE FUNCTION create_credits_for_student()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.role = 'student' THEN
        INSERT INTO credits (user_id, balance)
        VALUES (NEW.id, 0);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_credits_on_student_insert
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_credits_for_student();

-- Function to validate lesson capacity
CREATE OR REPLACE FUNCTION check_lesson_capacity()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.current_students > NEW.max_students THEN
        RAISE EXCEPTION 'Lesson capacity exceeded: current_students (%) cannot exceed max_students (%)',
            NEW.current_students, NEW.max_students;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_lesson_capacity_before_update
    BEFORE UPDATE ON lessons
    FOR EACH ROW
    WHEN (NEW.current_students IS DISTINCT FROM OLD.current_students)
    EXECUTE FUNCTION check_lesson_capacity();

-- Function to prevent negative credit balance
CREATE OR REPLACE FUNCTION check_credit_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.balance < 0 THEN
        RAISE EXCEPTION 'Credit balance cannot be negative: attempted to set balance to %', NEW.balance;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_credit_balance_before_update
    BEFORE UPDATE ON credits
    FOR EACH ROW
    WHEN (NEW.balance IS DISTINCT FROM OLD.balance)
    EXECUTE FUNCTION check_credit_balance();

-- Function to automatically cleanup expired sessions (called periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

-- Comments for triggers
COMMENT ON FUNCTION update_updated_at_column() IS 'Automatically updates updated_at timestamp on row modification';
COMMENT ON FUNCTION create_credits_for_student() IS 'Automatically creates credits account when a new student is registered';
COMMENT ON FUNCTION check_lesson_capacity() IS 'Validates that current_students does not exceed max_students';
COMMENT ON FUNCTION check_credit_balance() IS 'Prevents credit balance from becoming negative';
COMMENT ON FUNCTION cleanup_expired_sessions() IS 'Removes expired sessions (should be called by a cron job or periodically)';
