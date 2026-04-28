# 技術設計ドキュメント: バス回送物流プラットフォーム

## 概要

バスの回送区間（空車走行）を活用して小荷物を配送するプラットフォーム。
バス会社（Bus_Operator）が回送スケジュールを登録し、荷主（Shipper）がそのスケジュールに荷物を予約できる。
QRコードによる非対面受け渡しと、リアルタイムに近い荷物追跡に対応する。

### 技術スタック

| 区分 | 技術 |
|------|------|
| コンテナ | Docker / Docker Compose |
| バックエンド | Go 1.22 / Echo v4 |
| ホットデプロイ | Air |
| フロントエンド | Vue.js 3 / Vite / Tailwind CSS / Pinia |
| データベース | PostgreSQL 16 |
| 地図 | Leaflet.js + OpenStreetMap |
| ルーティング | OSRM（パブリックAPI、フォールバック: 直線距離） |
| ジオコーディング | Nominatim（OpenStreetMap、無料・認証不要） |
| 認証 | JWT (HS256, 有効期限24時間) + bcrypt |
| QRコード | qrcode.vue（フロントエンド生成） / html5-qrcode（スキャン） |
| API形式 | RESTful JSON |
| セキュリティ | CSP / X-Frame-Options / X-Content-Type-Options 等のセキュリティヘッダー |
| 環境変数管理 | .env ファイル（.gitignore 済み）/ .env.example（テンプレート） |

---

## アーキテクチャ

### システム構成図

```
┌─────────────────────────────────────────────────────────────┐
│                        Docker Compose                        │
│                                                             │
│  ┌──────────────┐    ┌──────────────────────────────────┐  │
│  │   frontend   │    │            backend               │  │
│  │  Vue.js 3    │───▶│  Go (Echo framework)             │  │
│  │  Vite        │    │  ┌──────────────────────────┐    │  │
│  │  Leaflet.js  │    │  │  Router / Middleware      │    │  │
│  │  port: 5173  │    │  │  (JWT Auth, CORS, Error)  │    │  │
│  └──────────────┘    │  └──────────────────────────┘    │  │
│                      │  ┌──────┐ ┌────────┐ ┌────────┐  │  │
│                      │  │Auth  │ │Schedule│ │Booking │  │  │
│                      │  │Hdlr  │ │Handler │ │Handler │  │  │
│                      │  └──────┘ └────────┘ └────────┘  │  │
│                      │  ┌──────────┐ ┌───────────────┐  │  │
│                      │  │Tracking  │ │Routing Handler│  │  │
│                      │  │Handler   │ │(OSRM proxy)   │  │  │
│                      │  └──────────┘ └───────────────┘  │  │
│                      │  ┌──────────────────────────┐    │  │
│                      │  │  Company Handler          │    │  │
│                      │  └──────────────────────────┘    │  │
│                      │  port: 8080                       │  │
│                      └──────────────────────────────────┘  │
│                                    │                        │
│                      ┌─────────────▼──────────┐            │
│                      │      PostgreSQL 16      │            │
│                      │      port: 5432         │            │
│                      └────────────────────────┘            │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  外部サービス                                          │  │
│  │  OSRM Public API: router.project-osrm.org            │  │
│  │  フォールバック: 直線距離GeoJSON                        │  │
│  │  Nominatim API: nominatim.openstreetmap.org          │  │
│  │  (地名→緯度経度変換、フロントエンドから直接呼び出し)    │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### モノリシック構成の方針

- 単一の Go バイナリにすべてのハンドラーを含める
- パッケージ分割でドメイン境界を明確にする（handler / service / repository / model）
- ハンドラーはサービスをインターフェース経由で依存し、テスト時にモックに差し替え可能にする

---

## コンポーネントとインターフェース

### バックエンド パッケージ構成

```
backend/
├── main.go
├── config/
│   └── config.go              # 環境変数読み込み（DATABASE_URL, JWT_SECRET, OSRM_BASE_URL）
├── db/
│   └── db.go                  # pgx 接続プール
├── middleware/
│   ├── auth.go                # JWT検証・RequireRole ミドルウェア
│   ├── cors.go                # CORS設定
│   ├── error.go               # カスタムエラーハンドラー
│   └── security.go            # セキュリティヘッダー（CSP / X-Frame-Options 等）
├── handler/
│   ├── interfaces.go          # サービスインターフェース定義・scheduleToMap ヘルパー
│   ├── auth.go                # POST /auth/register（招待コード検証付き）, POST /auth/login
│   ├── schedule.go            # GET/POST /schedules, GET /schedules/search,
│   │                          # GET /schedules/:id, PATCH /schedules/:id/status,
│   │                          # DELETE /schedules/:id, POST /schedules/:id/cancel
│   ├── booking.go             # GET/POST /bookings, GET /bookings/:id,
│   │                          # DELETE /bookings/:id（キャンセル）
│   ├── tracking.go            # GET /tracking/:tracking_number,
│   │                          # PATCH /bookings/:id/status
│   ├── routing.go             # GET /routing (OSRMプロキシ、フォールバックあり)
│   ├── company.go             # GET /companies, GET /companies/me,
│   │                          # PATCH /companies/me/storage
│   └── test/
│       ├── mock_services_test.go
│       ├── auth_handler_test.go
│       ├── booking_handler_test.go
│       ├── schedule_handler_test.go
│       ├── tracking_handler_test.go
│       └── routing_handler_test.go
├── service/
│   ├── auth_service.go        # Register（招待コード検証・company_id 自動設定） / Login（bcrypt + JWT発行）
│   ├── schedule_service.go    # Create / ListByOperator / Search / GetByID /
│   │                          # UpdateScheduleStatus / Delete（cancelled 予約除外） /
│   │                          # Cancel（accepted 予約連動キャンセル・avail_weight_kg 回復）
│   ├── booking_service.go     # Create（トランザクション・0値/負値チェック） / ListByShipper / GetByID /
│   │                          # Cancel（トランザクション、avail_weight_kg 回復）
│   ├── tracking_service.go    # GetByTrackingNumber / UpdateStatus（ログ記録）
│   └── test/
│       ├── mock_repos_test.go
│       ├── auth_service_test.go
│       ├── booking_service_test.go
│       ├── schedule_service_test.go
│       └── tracking_service_test.go
├── repository/
│   ├── user_repo.go           # FindByEmail / FindByID / Create / SetCompanyID / Delete
│   ├── schedule_repo.go       # Create / FindByID / ListByOperator / Search /
│   │                          # UpdateStatus / Delete / AddAvailWeight
│   ├── booking_repo.go        # Create / FindByID / FindByTrackingNumber /
│   │                          # ListByShipper / UpdateStatus / UpdateStatusDirect
│   ├── tracking_repo.go       # InsertStatusLog
│   ├── company_repo.go        # List / FindByID / UpdateStorage
│   ├── invite_repo.go         # FindByCode / MarkUsed（招待コード管理）
│   └── test/
│       ├── mock_repos_test.go
│       ├── user_repo_test.go
│       ├── booking_repo_test.go
│       ├── schedule_repo_test.go
│       ├── schedule_filter_test.go
│       └── tracking_repo_test.go
└── model/
    ├── user.go                # User 構造体・Role 定数
    ├── schedule.go            # Schedule 構造体・ScheduleStatus 型
    ├── booking.go             # Booking 構造体・BookingStatus 型・CanTransitionTo
    │                          # BookingStatusCancelled 定数追加
    ├── company.go             # BusCompany 構造体
    └── test/
        ├── booking_model_test.go
        ├── schedule_model_test.go
        └── user_model_test.go
```

### フロントエンド コンポーネント構成

```
frontend/
├── src/
│   ├── main.js
│   ├── router/
│   │   └── index.js               # Vue Router（ロールガード付き）
│   ├── stores/
│   │   └── auth.js                # Pinia: 認証状態（token, role, userId）
│   ├── views/
│   │   ├── LoginOperator.vue      # Bus_Operator ログイン（登録画面へのリンク付き）
│   │   ├── RegisterOperator.vue   # Bus_Operator 新規登録（招待コード必須）
│   │   ├── LoginShipper.vue       # Shipper ログイン
│   │   ├── OperatorDashboard.vue  # スケジュール登録・一覧・QRスキャン・マイページへのナビ
│   │   ├── OperatorMyPage.vue     # 所属会社情報・荷物置き場画像/説明の管理
│   │   ├── ScheduleCreate.vue     # 地名入力→緯度経度自動セット + 地図クリック登録
│   │   ├── ScheduleList.vue       # Operator用: 一覧クリックで地図自動移動・ステータス変更・削除・運行中止
│   │   │                          # キャンセル済み予約を表示（ステータス変更ボタン非表示）
│   │   ├── ShipperDashboard.vue   # スケジュール検索・予約一覧・荷物追跡・荷物置き場確認へのナビ
│   │   ├── ScheduleSearch.vue     # Shipper用: 一覧クリックで地図自動移動・荷物置き場確認リンク
│   │   ├── BookingCreate.vue      # 予約フォーム・完了後QRコード表示・印刷・荷物置き場確認リンク
│   │   ├── BookingList.vue        # 自分の予約一覧（60秒ポーリング）
│   │   │                          # accepted 予約にキャンセルボタン表示・確認ダイアログ
│   │   ├── CompanyList.vue        # バス会社・荷物置き場一覧（Shipper向け）
│   │   ├── QRScan.vue             # QRスキャン画面（カメラ or 手動入力）
│   │   │                          # cancelled 荷物スキャン時のエラーメッセージ対応
│   │   └── Tracking.vue           # 荷物追跡（認証不要、30秒ポーリング）
│   │                               # cancelled でポーリング停止
│   └── components/
│       ├── MapView.vue            # Leaflet.js ラッパー（center/bounds プロップで地図移動）
│       ├── RouteMap.vue           # 経路表示（MapViewのラッパー、OSRM呼び出し）
│       ├── BookingStatusBadge.vue # ステータスバッジ（cancelled: グレー系バッジ追加）
│       ├── QRCodeDisplay.vue      # QRコード生成・表示（qrcode.vue）
│       └── QRScanner.vue          # QRコードスキャン（html5-qrcode）
```

### 主要インターフェース（Go）

```go
// handler/interfaces.go
type AuthServiceInterface interface {
    Register(ctx context.Context, req service.RegisterRequest) (*model.User, error)
    Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error)
}

type ScheduleServiceInterface interface {
    Create(ctx context.Context, req service.CreateScheduleRequest) (*model.Schedule, error)
    ListByOperator(ctx context.Context, operatorID uuid.UUID) ([]model.Schedule, error)
    Search(ctx context.Context, filter repository.ScheduleFilter) ([]model.Schedule, error)
    GetByID(ctx context.Context, id uuid.UUID) (*model.Schedule, error)
    UpdateScheduleStatus(ctx context.Context, scheduleID uuid.UUID, newStatus model.ScheduleStatus, operatorID uuid.UUID) error
    Delete(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error
    Cancel(ctx context.Context, scheduleID uuid.UUID, operatorID uuid.UUID) error
}

type BookingServiceInterface interface {
    Create(ctx context.Context, req service.CreateBookingRequest) (*model.Booking, error)
    ListByShipper(ctx context.Context, shipperID uuid.UUID) ([]model.Booking, error)
    GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
    Cancel(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error
}

type TrackingServiceInterface interface {
    GetByTrackingNumber(ctx context.Context, trackingNumber string) (*service.TrackingInfo, error)
    UpdateStatus(ctx context.Context, bookingID uuid.UUID, newStatus model.BookingStatus, operatorID uuid.UUID) error
}
```

### APIエンドポイント一覧

| メソッド | パス | 認証 | 説明 |
|---------|------|------|------|
| POST | /api/v1/auth/register | なし | ユーザー登録 |
| POST | /api/v1/auth/login | なし | ログイン・JWT発行 |
| GET | /api/v1/schedules | JWT(Operator) | スケジュール一覧（自社分、予約情報含む） |
| POST | /api/v1/schedules | JWT(Operator) | スケジュール登録 |
| GET | /api/v1/schedules/search | JWT(Shipper) | スケジュール検索（位置・日時フィルター） |
| GET | /api/v1/schedules/:id | JWT | スケジュール詳細 |
| PATCH | /api/v1/schedules/:id/status | JWT(Operator) | スケジュールステータス更新 |
| DELETE | /api/v1/schedules/:id | JWT(Operator) | スケジュール削除（cancelled 予約のみなら削除可） |
| POST | /api/v1/schedules/:id/cancel | JWT(Operator) | スケジュール運行中止（accepted 予約を連動キャンセル） |
| GET | /api/v1/bookings | JWT(Shipper) | 予約一覧（自分の分） |
| POST | /api/v1/bookings | JWT(Shipper) | 荷物予約（追跡番号自動発行） |
| GET | /api/v1/bookings/:id | JWT | 予約詳細 |
| DELETE | /api/v1/bookings/:id | JWT(Shipper) | 予約キャンセル（accepted のみ） |
| PATCH | /api/v1/bookings/:id/status | JWT(Operator) | 荷物ステータス更新（ログ記録） |
| GET | /api/v1/tracking/:tracking_number | なし | 荷物追跡（認証不要） |
| GET | /api/v1/routing | なし | 経路計算（OSRMプロキシ） |
| GET | /api/v1/companies | なし | バス会社一覧（認証不要） |
| GET | /api/v1/companies/me | JWT(Operator) | 自社情報取得 |
| PATCH | /api/v1/companies/me/storage | JWT(Operator) | 荷物置き場の画像・説明を更新 |

---

## データモデル

### ER図

```
┌─────────────────────┐         ┌──────────────────────────┐
│     bus_companies   │         │        users             │
├─────────────────────┤         ├──────────────────────────┤
│ id          UUID PK │◀───┐    │ id          UUID PK      │
│ name        TEXT UQ │    │    │ email       TEXT UQ      │
│ storage_image_url   │    └────│ company_id  UUID FK(null)│
│             TEXT    │         │ password_hash TEXT       │
│ storage_description │         │ role        TEXT         │
│             TEXT    │         │   (bus_operator/shipper) │
│ created_at  TIMESTAMPTZ│      │ created_at  TIMESTAMPTZ  │
└─────────────────────┘         └──────────┬───────────────┘
                                           │
                                           │ operator_id
                                ┌──────────▼───────────────┐
                                │        schedules          │
                                ├──────────────────────────┤
                                │ id              UUID PK  │
                                │ operator_id     UUID FK  │
                                │ origin_lat      FLOAT8   │
                                │ origin_lng      FLOAT8   │
                                │ origin_name     TEXT     │
                                │ dest_lat        FLOAT8   │
                                │ dest_lng        FLOAT8   │
                                │ dest_name       TEXT     │
                                │ depart_at       TIMESTAMPTZ│
                                │ arrive_at       TIMESTAMPTZ│
                                │ max_weight_kg   FLOAT8   │
                                │ max_size_cm     FLOAT8   │
                                │ avail_weight_kg FLOAT8   │
                                │ status          TEXT     │
                                │  (open/full/departed/    │
                                │   arrived)               │
                                │ route_geojson   JSONB    │
                                │ created_at      TIMESTAMPTZ│
                                └──────────┬───────────────┘
                                           │ 1:N
                                ┌──────────▼───────────────┐
                                │        bookings           │
                                ├──────────────────────────┤
                                │ id              UUID PK  │
                                │ schedule_id     UUID FK  │
                                │ shipper_id      UUID FK  │
                                │ tracking_number TEXT UQ  │
                                │   (TRK-XXXXXXXX形式)     │
                                │ weight_kg       FLOAT8   │
                                │ size_cm         FLOAT8   │
                                │ content_desc    TEXT     │
                                │ recipient_name  TEXT     │
                                │ recipient_phone TEXT     │
                                │ recipient_addr  TEXT     │
                                │ status          TEXT     │
                                │  (accepted/loaded/       │
                                │   in_transit/delivered/  │
                                │   cancelled)             │
                                │ status_updated_at TIMESTAMPTZ│
                                │ created_at      TIMESTAMPTZ│
                                └──────────┬───────────────┘
                                           │ 1:N
                                ┌──────────▼───────────────┐
                                │     booking_status_logs   │
                                ├──────────────────────────┤
                                │ id              UUID PK  │
                                │ booking_id      UUID FK  │
                                │ old_status      TEXT     │
                                │ new_status      TEXT     │
                                │ changed_by      UUID FK→users│
                                │ changed_at      TIMESTAMPTZ│
                                └──────────────────────────┘
```

### DBマイグレーション構成

| ファイル | 内容 |
|---------|------|
| 001_schema.sql | users / schedules / bookings / booking_status_logs テーブル作成 |
| 002_data.sql | テストユーザー・サンプルスケジュール・サンプル予約の初期データ |
| 003_add_arrived_status.sql | schedules.status の CHECK 制約に `arrived` を追加 |
| 004_bus_companies.sql | bus_companies テーブル作成・初期4社登録・users に company_id カラム追加 |
| 005_add_cancelled_status.sql | bookings.status の CHECK 制約に `cancelled` を追加 |
| 006_invite_codes.sql | invite_codes テーブル作成・各バス会社の初期招待コード登録 |
| 007_schedule_cancelled_status.sql | schedules.status の CHECK 制約に `cancelled` を追加 |

### ステータス遷移

```
Schedule（前方向のみ許可）:
  open ──▶ full ──▶ departed ──▶ arrived
  open ──▶ departed（直接遷移も可）
  open/full ──▶ cancelled（Cancel メソッドで別管理、終端状態）
  ※ scheduleStatusOrder マップで順序管理
  ※ cancelled スケジュールの accepted 予約は連動して cancelled になる
  ※ cancelled スケジュールは削除不可（運行中止済みとして保持）

Booking（前方向のみ許可）:
  accepted ──▶ loaded ──▶ in_transit ──▶ delivered
  accepted ──▶ cancelled（Cancel メソッドで別管理、終端状態）
  ※ model.BookingStatus.CanTransitionTo() で前方向遷移を検証
  ※ cancelled は StatusOrder に含めない（終端状態のため）
  ※ QRスキャン時は自動的に次のステータスへ遷移（cancelled はスキャン不可）
```

### システム制限値

| 項目 | 上限 | 適用箇所 |
|------|------|---------|
| 1個あたり重量 | 10kg（0より大きい値） | BookingService.Create（WEIGHT_LIMIT_EXCEEDED） |
| 1個あたりサイズ（3辺合計） | 140cm（0より大きい値） | BookingService.Create（SIZE_LIMIT_EXCEEDED） |
| スケジュールの max_weight_kg | 0より大きい値 | ScheduleService.Create（INVALID_MAX_WEIGHT） |
| スケジュールの max_size_cm | 0より大きい値 | ScheduleService.Create（INVALID_MAX_SIZE） |
| 画像ファイルサイズ | 2MB | フロントエンドバリデーション |

---

## 地図・ジオコーディング設計

### 地名入力→緯度経度自動セット（ScheduleCreate・ScheduleSearch）

```
ユーザー入力（2文字以上）
  → 400ms デバウンス
  → Nominatim API 呼び出し（countrycodes=jp, limit=8, accept-language=ja）
  → 重複排除（lat/lon を小数点4桁で丸めて Set 管理）
  → 表示名を短縮（"東京駅（丸の内, 千代田区）" 形式）
  → ドロップダウン表示（最大5件）
  → 選択時: 緯度経度を origin/dest にセット、地図を flyTo で移動
```

- ScheduleCreate では地図クリックによる地点選択も併用可能
- 地名を編集し直した場合は緯度経度をリセットして再選択を促す
- ScheduleSearch では選択した地点を中心に ±0.3度（約30km）の bounding box を自動生成し、バックエンドの `ScheduleFilter` に渡す

### 地図自動移動（MapView）

`MapView` コンポーネントは以下のプロップで地図ビューを制御する：

| プロップ | 型 | 説明 |
|---------|-----|------|
| `center` | `{ lat, lng }` | 単点への flyTo（ズーム12固定、duration 1秒） |
| `bounds` | `[[lat1,lng1],[lat2,lng2]]` | 2点を含む範囲への flyToBounds（padding 40px） |
| `origin` | `{ lat, lng, name }` | 出発地マーカー（青） |
| `dest` | `{ lat, lng, name }` | 目的地マーカー（赤） |
| `clickable` | boolean | クリック選択モード（ScheduleCreate で使用） |

- ScheduleCreate: 出発地選択時に `center` で地図移動
- ScheduleList（Operator）: 一覧クリック時に出発地・目的地を含む `bounds` で地図移動
- ScheduleSearch（Shipper）: 一覧クリック時に出発地・目的地を含む `bounds` で地図移動

---

## QRコード設計

### QRコード生成（BookingCreate.vue）

- 予約完了後、`tracking_number` を値とするQRコードを `QRCodeDisplay.vue` で表示
- サイズ: 220px
- 印刷: `window.print()` でブラウザ印刷ダイアログを呼び出す

### QRコードスキャン（QRScan.vue）

```
スキャン開始ボタン押下
  → QRScanner コンポーネント（html5-qrcode）でカメラ起動
  → QRコード検出 → tracking_number 取得
  → GET /api/v1/tracking/:tracking_number で現在ステータス確認
  → nextStatus マップで次のステータスを決定
      { accepted→loaded, loaded→in_transit, in_transit→delivered }
  → PATCH /api/v1/bookings/:id/status で更新
  → 更新前・更新後ステータスとアクションメッセージを表示
```

カメラ不使用時のフォールバック: 追跡番号を手動入力して同じフローを実行

---

## バス会社マイページ設計

### 荷物置き場画像の保存方式

- フロントエンドで `FileReader.readAsDataURL()` により画像を Base64 Data URL に変換
- `PATCH /api/v1/companies/me/storage` の `storage_image_url` フィールドに Base64 文字列をそのまま送信
- DBの `bus_companies.storage_image_url` に Base64 または URL を保存
- 表示時は `<img :src="storage_image_url">` でそのまま表示可能

---

## ステータス同期（定期ポーリング）

| 画面 | 間隔 | 開始タイミング | 停止条件 |
|------|------|--------------|---------|
| BookingList（予約一覧） | 60秒 | onMounted | onUnmounted |
| Tracking（荷物追跡） | 30秒 | 照会ボタン押下後 | `delivered` または `cancelled` になった時 / onUnmounted / 追跡番号変更時 |

実装方針：
- `setInterval` で定期実行、`onUnmounted` で `clearInterval`（リソースリーク防止）
- 2回目以降のポーリングはローディング表示なし（ちらつき防止）

---

## エラーハンドリング

### エラーレスポンス形式

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "出発日時は現在時刻より未来である必要があります"
  }
}
```

### エラーコード一覧

| HTTPステータス | コード | 説明 |
|--------------|--------|------|
| 400 | BAD_REQUEST | リクエストボディのパースエラー |
| 400 | VALIDATION_ERROR | 入力値バリデーションエラー（メール形式・パスワード8文字未満・日時・座標等） |
| 400 | WEIGHT_LIMIT_EXCEEDED | 1個あたり重量が0以下または10kg超過 |
| 400 | SIZE_LIMIT_EXCEEDED | 1個あたりサイズが0以下または140cm超過 |
| 400 | INVALID_INVITE_CODE | 招待コードが無効または使用済み |
| 401 | UNAUTHORIZED | 認証失敗・JWT期限切れ |
| 403 | FORBIDDEN | ロール権限不足・他 Shipper の予約へのキャンセル試行 |
| 404 | NOT_FOUND | リソースが存在しない |
| 409 | CAPACITY_EXCEEDED | スケジュールの残余積載量超過 |
| 409 | SIZE_EXCEEDED | スケジュールのサイズ上限超過 |
| 409 | EMAIL_ALREADY_EXISTS | メールアドレス重複 |
| 409 | HAS_BOOKINGS | 有効な予約が存在するスケジュールの削除試行 |
| 409 | INVALID_STATUS | 出発済みスケジュールの削除試行 |
| 409 | ALREADY_CANCELLED | すでにキャンセル済みのスケジュールへのキャンセル試行 |
| 409 | CANNOT_CANCEL | accepted 以外のステータスへのキャンセル試行 |
| 500 | INTERNAL_ERROR | サーバー内部エラー |

### エッジケース対応

| ケース | 対応 |
|--------|------|
| 期限切れJWT | 401 UNAUTHORIZED |
| JWT_SECRET 未設定でサーバー起動 | log.Fatal でサーバー起動を停止 |
| 出発地・目的地未選択でスケジュール登録 | 400 VALIDATION_ERROR |
| max_weight_kg=0 または負値でスケジュール登録 | 400 VALIDATION_ERROR |
| max_size_cm=0 または負値でスケジュール登録 | 400 VALIDATION_ERROR |
| weight_kg=0 または負値で予約登録 | 400 WEIGHT_LIMIT_EXCEEDED |
| size_cm=0 または負値で予約登録 | 400 SIZE_LIMIT_EXCEEDED |
| 不正なメールアドレス形式で登録 | 400 VALIDATION_ERROR |
| パスワード8文字未満で登録 | 400 VALIDATION_ERROR |
| 無効または使用済み招待コードで bus_operator 登録 | 400 INVALID_INVITE_CODE |
| 存在しない tracking_number | 404 NOT_FOUND |
| ルーティングAPI応答なし | 直線距離GeoJSONでフォールバック |
| ステータスの逆方向遷移 | 400 VALIDATION_ERROR |
| メールアドレス重複登録 | 409 EMAIL_ALREADY_EXISTS |
| cancelled 予約のみのスケジュール削除 | 削除可能（有効な予約なしと判定） |
| 有効な予約ありスケジュールの削除 | 409 HAS_BOOKINGS |
| 出発済みスケジュールの削除 | 409 INVALID_STATUS |
| 他オペレーターのスケジュール削除 | 404 NOT_FOUND（存在しないものとして扱う） |
| 配達完了済み荷物のQRスキャン | フロントエンドでエラーメッセージ表示 |
| キャンセル済み荷物のQRスキャン | フロントエンドでエラーメッセージ表示 |
| 画像ファイルが2MB超過 | フロントエンドでエラーメッセージ表示 |
| accepted 以外の予約のキャンセル試行 | 409 CANNOT_CANCEL |
| すでにキャンセル済みのスケジュールへのキャンセル試行 | 409 ALREADY_CANCELLED |
| 他 Shipper の予約のキャンセル試行 | 403 FORBIDDEN |
| 存在しない予約のキャンセル試行 | 404 NOT_FOUND |
| collectSchedules での DB エラー | エラーを上位に伝播（サイレント無視しない） |

### アトミック更新の実装

予約登録時の積載可能重量更新は PostgreSQL のトランザクション + SELECT FOR UPDATE で実装する:

```sql
BEGIN;
SELECT avail_weight_kg, max_size_cm FROM schedules WHERE id = $1 FOR UPDATE;
-- アプリケーション側で残量・サイズチェック
UPDATE schedules SET avail_weight_kg = avail_weight_kg - $2 WHERE id = $1;
INSERT INTO bookings (...) VALUES (...);
COMMIT;
```

予約キャンセル時の積載可能重量回復も同様にトランザクション + SELECT FOR UPDATE で実装する:

```sql
BEGIN;
SELECT id, shipper_id, status, weight_kg, schedule_id FROM bookings WHERE id = $1 FOR UPDATE;
-- アプリケーション側で shipper_id 一致・status = accepted チェック
UPDATE bookings SET status = 'cancelled', status_updated_at = NOW() WHERE id = $1;
UPDATE schedules SET avail_weight_kg = avail_weight_kg + $2 WHERE id = $3;
COMMIT;
```

---

## セキュリティ設計

### 環境変数管理

- `JWT_SECRET` と `POSTGRES_PASSWORD` は `.env` ファイルで管理（`.gitignore` 済み）
- `.env.example` をテンプレートとしてリポジトリに含める
- `config.Load()` は `JWT_SECRET` が空の場合 `log.Fatal` でサーバー起動を停止する

### セキュリティヘッダー（`middleware/security.go`）

バックエンドの全レスポンスに以下のヘッダーを付与する：

| ヘッダー | 値 | 効果 |
|---------|-----|------|
| X-Content-Type-Options | nosniff | MIMEスニッフィング攻撃防止 |
| X-Frame-Options | DENY | クリックジャッキング防止 |
| X-XSS-Protection | 1; mode=block | 古いブラウザ向けXSSフィルター |
| Referrer-Policy | strict-origin-when-cross-origin | リファラー漏洩制限 |
| Permissions-Policy | geolocation=(), microphone=(), camera=(self) | 不要なブラウザ機能無効化 |

### Content-Security-Policy（`frontend/index.html`）

フロントエンドの meta タグで CSP を設定する：

```
default-src 'self'
script-src 'self'
style-src 'self' 'unsafe-inline' https://unpkg.com
img-src 'self' data: blob: https://*.openstreetmap.org
connect-src 'self' https://nominatim.openstreetmap.org
font-src 'self'
object-src 'none'
base-uri 'self'
form-action 'self'
frame-ancestors 'none'
```

### 認証・認可

- JWT は localStorage に保存（XSS 対策として CSP で外部スクリプト読み込みを禁止）
- `bus_operator` 登録は招待コード必須（`invite_codes` テーブルで管理）
- 招待コードは `MarkUsed` の原子的 UPDATE で TOCTOU 競合を防止
- パスワードは bcrypt（DefaultCost）でハッシュ化
- メールアドレスは正規表現でフォーマット検証
- パスワードは最低8文字を強制

---

## テスト戦略

### テスト構成

各レイヤーに `test/` サブパッケージを設け、モックベースのユニットテストを実装する。

```
backend/
├── handler/test/     # ハンドラー層: モックサービスを使ったHTTPレベルテスト
├── service/test/     # サービス層: モックリポジトリを使ったビジネスロジックテスト
├── repository/test/  # リポジトリ層: インターフェース契約・ScheduleFilterロジックテスト
├── model/test/       # モデル層: CanTransitionTo・定数値・構造体テスト
└── integration_test.go  # 統合テスト（-tags=integration）
```

### テスト実行

```bash
# 全ユニットテスト（config テストを含む）
docker compose run --rm backend go test ./handler/test/... ./service/test/... ./repository/test/... ./model/test/... ./config/... -v

# 統合テスト（DB接続が必要）
docker compose run --rm backend go test -tags=integration -v

# フロントエンドテスト
docker compose run --rm frontend npm test

# フロントエンドカバレッジ
docker compose run --rm frontend sh -c "npm test -- --coverage --coverage.provider=v8 --coverage.include='src/views/**'"
```

---

## 正確性プロパティ

### Property 1: 有効な認証情報でJWTが発行される
有効なメールアドレスとパスワードで登録されたユーザーが同じ認証情報でログインしたとき、Auth_Service はJWTトークンを含むレスポンスを返す。
**Validates: Requirements 1.2, 1.3**

### Property 2: 無効な認証情報でエラーが返る
存在しないメールアドレス、または誤ったパスワードでログインしたとき、Auth_Service は認証エラーを返しJWTトークンを発行しない。
**Validates: Requirements 1.4**

### Property 3: パスワードのbcryptハッシュ化
任意のパスワード文字列を登録したとき、DBに保存される値は元のパスワードと異なり、bcrypt.CompareHashAndPassword で検証できる。
**Validates: Requirements 1.6**

### Property 4: スケジュール登録のラウンドトリップ
有効なスケジュールデータを登録したとき、同じIDで取得したスケジュールが登録したデータと一致する。
**Validates: Requirements 2.5**

### Property 5: 過去日時のスケジュール登録はエラー
現在時刻より過去の出発日時を持つスケジュールを登録しようとしたとき、Schedule_Service はバリデーションエラーを返す。
**Validates: Requirements 2.8**

### Property 6: オペレーターは自分のスケジュールのみ取得できる
Bus_Operator が自分のスケジュール一覧を取得したとき、返されるすべてのスケジュールの operator_id が自分のIDと一致する。Shipper が自分の予約一覧を取得したとき、返されるすべての予約の shipper_id が自分のIDと一致する。
**Validates: Requirements 3.1, 5.6**

### Property 7: スケジュール一覧レスポンスに必須フィールドが含まれる
各スケジュールのレスポンスJSONには出発地・目的地・出発日時・ステータス・残り積載可能重量・残り積載可能サイズが含まれる。
**Validates: Requirements 3.2, 4.6**

### Property 8: 検索条件に合致するスケジュールのみ返される
出発地エリア・目的地エリア・日付の検索条件に対して、Schedule_Service が返すすべてのスケジュールはその条件を満たす。
**Validates: Requirements 4.4**

### Property 9: 予約登録のラウンドトリップ（tracking_number発行）
有効な予約データを登録したとき、Booking_Service は一意のtracking_numberを返し、初期ステータスは「受付済み（accepted）」である。
**Validates: Requirements 5.1, 5.2, 8.2**

### Property 10: 積載制限超過でエラーが返る
残り積載可能重量を超える重量、またはサイズ上限を超える荷物を予約しようとしたとき、Booking_Service はエラーを返し予約は登録されない。
**Validates: Requirements 5.3, 5.4, 5.7, 5.8**

### Property 11: 予約登録後の残り積載可能重量の正確な更新
スケジュールに対して複数の予約が順次登録されたとき、残り積載可能重量は各予約の重量の合計分だけ正確に減少する。
**Validates: Requirements 5.5**

### Property 12: 経路データが地図コンポーネントに正しく渡される
スケジュールの経路データ（GeoJSON）が提供されたとき、MapViewコンポーネントはルートラインを描画する。
**Validates: Requirements 2.6, 3.3, 4.5, 6.3**

### Property 13: すべてのAPIレスポンスがJSON形式
任意のAPIエンドポイントへのリクエストに対して、レスポンスのContent-Typeは `application/json` であり、レスポンスボディは有効なJSONである。
**Validates: Requirements 7.2**

### Property 14: 有効なJWTで保護されたエンドポイントにアクセスできる
有効なJWTトークンを持つ認証済みユーザーが保護されたエンドポイントにアクセスしたとき許可される。無効または期限切れのJWTでは401が返る。
**Validates: Requirements 7.5, 1.5**

### Property 15: tracking_numberによるステータス照会のラウンドトリップ
予約のtracking_numberを使ってステータス照会を実行したとき、Tracking_Service は該当予約のステータス・スケジュール情報・最終更新日時を返す。
**Validates: Requirements 8.3**

### Property 16: ステータスの前方向遷移のみ許可
Bookingに対して現在のステータスより前の状態へ戻す更新を実行したとき、Tracking_Service はバリデーションエラーを返しステータスは変更されない。
**Validates: Requirements 8.6**

### Property 17: accepted ステータスの予約のみキャンセルボタンが表示される
*For any* 予約リストにおいて、キャンセルボタンが表示されるのは Booking_Status が「accepted」の予約のみであり、loaded / in_transit / delivered / cancelled の予約にはキャンセルボタンが表示されない。
**Validates: Requirements 11.1, 11.2**

### Property 18: キャンセル後のステータスと追跡照会のラウンドトリップ
*For any* accepted ステータスの Booking に対して Cancel を実行したとき、Booking_Service はステータスを cancelled に更新し、その tracking_number で照会したとき Tracking_Service は cancelled を返す。
**Validates: Requirements 11.3, 11.9**

### Property 19: キャンセル後の avail_weight_kg 回復
*For any* accepted ステータスの Booking（weight_kg = W）に対して Cancel を実行したとき、紐づく Schedule の avail_weight_kg は W だけ増加する。
**Validates: Requirements 11.4**

### Property 20: accepted 以外へのキャンセル試行はエラー
*For any* Booking において、Booking_Status が loaded / in_transit / delivered / cancelled のいずれかの場合に Cancel を実行したとき、Booking_Service はエラーを返しステータスは変更されない。
**Validates: Requirements 11.5, 11.12**

### Property 21: 他 Shipper の予約へのキャンセル試行は 403
*For any* Booking に対して、その shipper_id と異なる shipper_id で Cancel を実行したとき、Booking_Service は 403 FORBIDDEN エラーを返す。
**Validates: Requirements 11.7**
