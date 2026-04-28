-- バス会社オペレーター向け招待コードテーブル
CREATE TABLE invite_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  company_id UUID NOT NULL REFERENCES bus_companies(id),
  used_by UUID REFERENCES users(id),
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invite_codes_code ON invite_codes(code);

-- 初期招待コード（各バス会社に1件ずつ）
INSERT INTO invite_codes (code, company_id) VALUES
  ('RYUKYU-2024', 'aaaaaaaa-0000-0000-0000-000000000001'),
  ('NAHA-2024',   'aaaaaaaa-0000-0000-0000-000000000002'),
  ('OKINAWA-2024','aaaaaaaa-0000-0000-0000-000000000003'),
  ('TOYO-2024',   'aaaaaaaa-0000-0000-0000-000000000004');
