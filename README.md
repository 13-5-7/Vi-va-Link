# バス回送物流プラットフォーム

バスの回送便を活用した荷物輸送マッチングシステムです。バス会社（オペレーター）が運行スケジュールを登録し、荷主がそのスケジュールに荷物を予約できます。QRコードによる非対面受け渡しと、リアルタイムの荷物追跡に対応しています。

## 起動方法

```bash
docker-compose up --build
```

| サービス | URL |
|---|---|
| フロントエンド | http://localhost:5173 |
| バックエンドAPI | http://localhost:8080 |
| ヘルスチェック | http://localhost:8080/health |

初回起動時に `db/init/` 以下のSQLが順番に自動適用されます。

### テストアカウント

| ロール | メールアドレス | パスワード |
|---|---|---|
| バス会社オペレーター | operator@example.com | password |
| 荷主 | shipper@example.com | password |

---

## 技術スタック

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.22 / Echo v4 |
| データベース | PostgreSQL 16 |
| フロントエンド | Vue 3 / Vite / Tailwind CSS / Pinia |
| 地図 | Leaflet + OpenStreetMap (Nominatim) |
| ルーティング | OSRM（利用不可時は直線フォールバック） |
| 認証 | JWT (HS256, 有効期限24時間) |
| コンテナ | Docker / Docker Compose |

---

## アーキテクチャ

```
frontend (Vue3)
    │  HTTP (Vite proxy → /api/v1)
    ▼
backend (Go/Echo)
    ├── handler/      HTTPハンドラー・バリデーション
    ├── service/      ビジネスロジック
    ├── repository/   DBアクセス
    ├── model/        データモデル
    └── middleware/   JWT認証・CORS・エラーハンドリング
    │
    ▼
PostgreSQL (bus_logistics DB)
```

---

## ユーザーロール

| ロール | 説明 |
|---|---|
| `bus_operator` | バス会社のオペレーター。スケジュール管理・QRスキャンによるステータス更新を行う |
| `shipper` | 荷主。スケジュールを検索して荷物を予約し、追跡番号で状況を確認する |

---

## 画面一覧

### オペレーター

| パス | 画面 |
|---|---|
| /operator/login | ログイン |
| /operator/dashboard | ダッシュボード |
| /operator/schedules | スケジュール一覧（予約状況・ステータス変更・削除） |
| /operator/schedules/new | スケジュール登録（地図クリック or 地名検索） |
| /operator/qrscan | QRスキャン（荷物ステータスを自動更新） |
| /operator/mypage | マイページ（所属会社・荷物置き場の画像と説明を管理） |

### 荷主

| パス | 画面 |
|---|---|
| /shipper/login | ログイン |
| /shipper/dashboard | ダッシュボード |
| /shipper/schedules | スケジュール検索（出発地・目的地・日時で絞り込み） |
| /shipper/bookings/new | 予約登録（完了後にQRコード表示・印刷） |
| /shipper/bookings | 予約一覧 |
| /tracking | 荷物追跡（追跡番号入力、30秒ポーリング） |

---

## APIエンドポイント

### 認証

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| POST | /api/v1/auth/register | なし | ユーザー登録 |
| POST | /api/v1/auth/login | なし | ログイン（JWTトークン返却） |

### スケジュール

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| GET | /api/v1/schedules | operator | 自分のスケジュール一覧（予約情報含む） |
| POST | /api/v1/schedules | operator | スケジュール作成 |
| GET | /api/v1/schedules/search | shipper | スケジュール検索（位置・日時フィルター） |
| GET | /api/v1/schedules/:id | JWT | スケジュール詳細 |
| PATCH | /api/v1/schedules/:id/status | operator | スケジュールステータス更新 |
| DELETE | /api/v1/schedules/:id | operator | スケジュール削除 |

### 予約

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| GET | /api/v1/bookings | shipper | 自分の予約一覧 |
| POST | /api/v1/bookings | shipper | 予約作成（追跡番号を自動発行） |
| GET | /api/v1/bookings/:id | JWT | 予約詳細 |
| PATCH | /api/v1/bookings/:id/status | operator | 荷物ステータス更新（ログ記録） |

### 追跡・ルーティング・会社

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| GET | /api/v1/tracking/:tracking_number | なし | 追跡番号で荷物状況を照会 |
| GET | /api/v1/routing | なし | 2点間のルート取得（OSRM） |
| GET | /api/v1/companies | なし | バス会社一覧 |
| GET | /api/v1/companies/me | operator | 自社情報取得 |
| PATCH | /api/v1/companies/me/storage | operator | 荷物置き場の画像・説明を更新 |

---

## ビジネスルール

### 荷物サイズ制限

- 1個あたりの重量: 最大 **10kg**
- 1個あたりのサイズ（3辺合計）: 最大 **140cm**
- スケジュール作成時にも `max_size_cm` の上限は 140cm

### 予約ステータス遷移

前方向への遷移のみ許可。逆戻り不可。

```
accepted（受付済）→ loaded（積載済）→ in_transit（輸送中）→ delivered（配達済）
```

QRスキャン画面でスキャンするたびに次のステータスへ自動遷移します。ステータス変更のたびに `booking_status_logs` にログが記録されます。

### スケジュールステータス遷移

```
open（受付中）→ full（満載）→ departed（出発済）→ arrived（到着済）
```

- 予約が存在する、または `departed` / `arrived` のスケジュールは削除不可
- `departed` / `arrived` への遷移時、紐づく予約のステータスは自動連動しない（QRスキャンで個別更新）

### 積載容量管理

予約作成時にトランザクション内で `schedules.avail_weight_kg` をデクリメントします。同時予約による超過を防ぐため `FOR UPDATE` ロックを使用しています。

---

## DBスキーマ

### テーブル一覧

| テーブル | 説明 |
|---|---|
| users | ユーザー（オペレーター・荷主共通） |
| bus_companies | バス会社マスター（沖縄4社初期登録済み） |
| schedules | 運行スケジュール |
| bookings | 荷物予約 |
| booking_status_logs | 予約ステータス変更ログ |

### users

| カラム | 型 | 説明 |
|---|---|---|
| id | UUID | PK |
| email | TEXT | ユニーク |
| password_hash | TEXT | bcryptハッシュ |
| role | TEXT | `bus_operator` or `shipper` |
| company_id | UUID | バス会社ID（オペレーターのみ） |
| created_at | TIMESTAMPTZ | |

### bus_companies

| カラム | 型 | 説明 |
|---|---|---|
| id | UUID | PK |
| name | TEXT | 会社名（ユニーク） |
| storage_image_url | TEXT | 荷物置き場の画像（Base64 or URL） |
| storage_description | TEXT | 荷物置き場の説明文 |
| created_at | TIMESTAMPTZ | |

初期データ: 琉球バス交通、那覇バス、沖縄バス、東陽バス

### schedules

| カラム | 型 | 説明 |
|---|---|---|
| id | UUID | PK |
| operator_id | UUID | FK → users |
| origin_lat/lng | FLOAT8 | 出発地座標 |
| origin_name | TEXT | 出発地名 |
| dest_lat/lng | FLOAT8 | 目的地座標 |
| dest_name | TEXT | 目的地名 |
| depart_at | TIMESTAMPTZ | 出発日時 |
| arrive_at | TIMESTAMPTZ | 到着予定日時 |
| max_weight_kg | FLOAT8 | 最大積載重量 |
| max_size_cm | FLOAT8 | 最大積載サイズ（3辺合計） |
| avail_weight_kg | FLOAT8 | 残余積載重量 |
| status | TEXT | open / full / departed / arrived |
| route_geojson | JSONB | ルートのGeoJSON |
| created_at | TIMESTAMPTZ | |

### bookings

| カラム | 型 | 説明 |
|---|---|---|
| id | UUID | PK |
| schedule_id | UUID | FK → schedules |
| shipper_id | UUID | FK → users |
| tracking_number | TEXT | 追跡番号（ユニーク、TRK-XXXXXXXX形式） |
| weight_kg | FLOAT8 | 重量 |
| size_cm | FLOAT8 | サイズ（3辺合計） |
| content_desc | TEXT | 内容物の概要 |
| recipient_name | TEXT | 受取人名 |
| recipient_phone | TEXT | 受取人電話番号 |
| recipient_addr | TEXT | 受取人住所 |
| status | TEXT | accepted / loaded / in_transit / delivered |
| status_updated_at | TIMESTAMPTZ | 最終ステータス更新日時 |
| created_at | TIMESTAMPTZ | |

### booking_status_logs

| カラム | 型 | 説明 |
|---|---|---|
| id | UUID | PK |
| booking_id | UUID | FK → bookings |
| old_status | TEXT | 変更前ステータス |
| new_status | TEXT | 変更後ステータス |
| changed_by | UUID | FK → users（操作したオペレーター） |
| changed_at | TIMESTAMPTZ | |

---

## 開発メモ

### DB接続（コンテナ起動中）

```bash
docker compose exec db psql -U app -d bus_logistics
```

### よく使うコマンド

```bash
# 再ビルドして起動
docker compose up -d --build

# DBデータを含めて完全リセット（注意: データ消去）
docker compose down -v

# ビルドキャッシュ削除
docker builder prune -f

# 脆弱性チェック
docker compose run --rm frontend npm audit
```
