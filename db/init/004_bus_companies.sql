-- bus_companies マスターテーブル
CREATE TABLE bus_companies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  -- 荷物置き場の画像（Base64 または URL）
  storage_image_url TEXT,
  storage_description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 初期データ: 沖縄4社
INSERT INTO bus_companies (id, name) VALUES
  ('aaaaaaaa-0000-0000-0000-000000000001', '琉球バス交通'),
  ('aaaaaaaa-0000-0000-0000-000000000002', '那覇バス'),
  ('aaaaaaaa-0000-0000-0000-000000000003', '沖縄バス'),
  ('aaaaaaaa-0000-0000-0000-000000000004', '東陽バス');

-- users テーブルに company_id を追加（bus_operator のみ使用）
ALTER TABLE users ADD COLUMN company_id UUID REFERENCES bus_companies(id);

-- 既存の operator@example.com を沖縄バス所属に
UPDATE users
SET company_id = 'aaaaaaaa-0000-0000-0000-000000000003'
WHERE email = 'operator@example.com';
