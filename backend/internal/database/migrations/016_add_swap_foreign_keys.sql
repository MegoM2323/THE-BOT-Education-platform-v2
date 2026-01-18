-- 005_add_swap_foreign_keys.sql
-- Add missing foreign key constraints to swaps table for data integrity

-- Add foreign key constraint for old_booking_id
ALTER TABLE swaps
ADD CONSTRAINT fk_swaps_old_booking_id
FOREIGN KEY (old_booking_id) REFERENCES bookings(id) ON DELETE CASCADE;

-- Add foreign key constraint for new_booking_id
ALTER TABLE swaps
ADD CONSTRAINT fk_swaps_new_booking_id
FOREIGN KEY (new_booking_id) REFERENCES bookings(id) ON DELETE CASCADE;

-- Comments for clarity
COMMENT ON CONSTRAINT fk_swaps_old_booking_id ON swaps IS 'Ensures referential integrity for old booking';
COMMENT ON CONSTRAINT fk_swaps_new_booking_id ON swaps IS 'Ensures referential integrity for new booking';
