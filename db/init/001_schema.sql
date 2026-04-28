-- UUID拡張
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- users テーブル
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('bus_operator', 'shipper')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users(email);

-- schedules テーブル
CREATE TABLE schedules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  operator_id UUID NOT NULL REFERENCES users(id),
  origin_lat FLOAT8 NOT NULL,
  origin_lng FLOAT8 NOT NULL,
  origin_name TEXT NOT NULL DEFAULT '',
  dest_lat FLOAT8 NOT NULL,
  dest_lng FLOAT8 NOT NULL,
  dest_name TEXT NOT NULL DEFAULT '',
  depart_at TIMESTAMPTZ NOT NULL,
  arrive_at TIMESTAMPTZ NOT NULL,
  max_weight_kg FLOAT8 NOT NULL,
  max_size_cm FLOAT8 NOT NULL,
  avail_weight_kg FLOAT8 NOT NULL,
  status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'full', 'departed')),
  route_geojson JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_schedules_operator_id ON schedules(operator_id);
CREATE INDEX idx_schedules_depart_at ON schedules(depart_at);
CREATE INDEX idx_schedules_status ON schedules(status);

-- bookings テーブル
CREATE TABLE bookings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  schedule_id UUID NOT NULL REFERENCES schedules(id),
  shipper_id UUID NOT NULL REFERENCES users(id),
  tracking_number TEXT NOT NULL UNIQUE,
  weight_kg FLOAT8 NOT NULL,
  size_cm FLOAT8 NOT NULL,
  content_desc TEXT NOT NULL DEFAULT '',
  recipient_name TEXT NOT NULL DEFAULT '',
  recipient_phone TEXT NOT NULL DEFAULT '',
  recipient_addr TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'accepted' CHECK (status IN ('accepted', 'loaded', 'in_transit', 'delivered')),
  status_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_bookings_schedule_id ON bookings(schedule_id);
CREATE INDEX idx_bookings_shipper_id ON bookings(shipper_id);
CREATE UNIQUE INDEX idx_bookings_tracking_number ON bookings(tracking_number);

-- booking_status_logs テーブル
CREATE TABLE booking_status_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id UUID NOT NULL REFERENCES bookings(id),
  old_status TEXT NOT NULL,
  new_status TEXT NOT NULL,
  changed_by UUID NOT NULL REFERENCES users(id),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_booking_status_logs_booking_id ON booking_status_logs(booking_id);
