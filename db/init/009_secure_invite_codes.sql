-- 既存の推測可能な招待コードを無効化し、セキュアなコードに置き換える
-- 注意: 本番環境では ADMIN_KEY を使って /api/v1/admin/invite-codes から発行すること

-- 古い未使用コードを削除（使用済みは履歴として保持）
DELETE FROM invite_codes
WHERE used_by IS NULL
  AND code IN ('RYUKYU-2024', 'NAHA-2024', 'OKINAWA-2024', 'TOYO-2024');

-- 新しいセキュアな招待コードを挿入（gen_random_bytes で生成した16進数32文字）
-- 実際の運用では管理者APIから動的に発行すること
INSERT INTO invite_codes (code, company_id) VALUES
  (encode(gen_random_bytes(16), 'hex'), 'aaaaaaaa-0000-0000-0000-000000000001'),
  (encode(gen_random_bytes(16), 'hex'), 'aaaaaaaa-0000-0000-0000-000000000002'),
  (encode(gen_random_bytes(16), 'hex'), 'aaaaaaaa-0000-0000-0000-000000000003'),
  (encode(gen_random_bytes(16), 'hex'), 'aaaaaaaa-0000-0000-0000-000000000004');
