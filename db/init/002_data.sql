INSERT INTO users (email, password_hash, role) VALUES 
('operator@example.com', '$2a$10$gGfd2fx6PFQWlVAuPbk5fuxK6/J1Agg4afxbdsAbJdeez17HTNHqa', 'bus_operator'),
('shipper@example.com', '$2a$10$gGfd2fx6PFQWlVAuPbk5fuxK6/J1Agg4afxbdsAbJdeez17HTNHqa', 'shipper')
ON CONFLICT (email) DO NOTHING;

-- サンプルスケジュール（東京→大阪、東京→名古屋）
INSERT INTO schedules (
  id, operator_id,
  origin_lat, origin_lng, origin_name,
  dest_lat, dest_lng, dest_name,
  depart_at, arrive_at,
  max_weight_kg, max_size_cm, avail_weight_kg,
  status, route_geojson
)
SELECT
  '11111111-1111-1111-1111-111111111111',
  u.id,
  35.6812, 139.7671, 'Tokyo Station',
  34.6937, 135.5023, 'Osaka Station',
  NOW() + INTERVAL '2 days',
  NOW() + INTERVAL '2 days 5 hours',
  50.0, 150.0, 47.0,
  'open', NULL
FROM users u WHERE u.email = 'operator@example.com'
ON CONFLICT (id) DO NOTHING;

INSERT INTO schedules (
  id, operator_id,
  origin_lat, origin_lng, origin_name,
  dest_lat, dest_lng, dest_name,
  depart_at, arrive_at,
  max_weight_kg, max_size_cm, avail_weight_kg,
  status, route_geojson
)
SELECT
  '22222222-2222-2222-2222-222222222222',
  u.id,
  35.6812, 139.7671, 'Tokyo Station',
  35.1815, 136.9066, 'Nagoya Station',
  NOW() + INTERVAL '3 days',
  NOW() + INTERVAL '3 days 2 hours',
  30.0, 100.0, 30.0,
  'open', NULL
FROM users u WHERE u.email = 'operator@example.com'
ON CONFLICT (id) DO NOTHING;

-- サンプル予約（東京→大阪スケジュールに2件）
INSERT INTO bookings (
  id, schedule_id, shipper_id,
  tracking_number, weight_kg, size_cm,
  content_desc, recipient_name, recipient_phone, recipient_addr,
  status
)
SELECT
  'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  '11111111-1111-1111-1111-111111111111',
  u.id,
  'TRK-SAMPLE1', 2.0, 60.0,
  'Clothing', 'Taro Yamada', '090-1234-5678', '1-1-1 Kita-ku, Osaka',
  'accepted'
FROM users u WHERE u.email = 'shipper@example.com'
ON CONFLICT (id) DO NOTHING;

INSERT INTO bookings (
  id, schedule_id, shipper_id,
  tracking_number, weight_kg, size_cm,
  content_desc, recipient_name, recipient_phone, recipient_addr,
  status
)
SELECT
  'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
  '11111111-1111-1111-1111-111111111111',
  u.id,
  'TRK-SAMPLE2', 1.0, 40.0,
  'Books', 'Hanako Suzuki', '080-9876-5432', '2-2-2 Naka-ku, Sakai, Osaka',
  'accepted'
FROM users u WHERE u.email = 'shipper@example.com'
ON CONFLICT (id) DO NOTHING;
