-- bookings.status に 'cancelled' を追加
ALTER TABLE bookings
  DROP CONSTRAINT IF EXISTS bookings_status_check;

ALTER TABLE bookings
  ADD CONSTRAINT bookings_status_check
  CHECK (status IN ('accepted', 'loaded', 'in_transit', 'delivered', 'cancelled'));
