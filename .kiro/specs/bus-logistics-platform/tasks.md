# 実装計画: バス回送物流プラットフォーム

## 概要

モノリシック構成（Go / Echo / Vue.js 3 / PostgreSQL）で、バスの回送区間を活用した小荷物配送プラットフォームを実装する。
タスクはプロジェクト基盤 → バックエンド → フロントエンド → テスト → 統合確認の順に進める。

---

## タスク

- [x] 1. プロジェクト基盤の構築
  - [x] 1.1 Docker Compose とディレクトリ構成を作成する
    - `docker-compose.yml` を作成し、db / backend / frontend の3サービスを定義する
    - `backend/` と `frontend/` のディレクトリ骨格（Dockerfile, .air.toml）を作成する
    - PostgreSQL の初期化スクリプト配置先 (`db/init/`) を用意する
    - _Requirements: 7.1_

  - [x] 1.2 データベーススキーマとマイグレーションを実装する
    - `db/init/001_schema.sql` に users / schedules / bookings / booking_status_logs テーブルを定義する
    - UUID主キー、外部キー制約、インデックス（email, tracking_number, operator_id, shipper_id）を含める
    - ステータス列は TEXT 型で CHECK 制約を付与する
    - `db/init/003_add_arrived_status.sql` で schedules.status に `arrived` を追加する
    - `db/init/004_bus_companies.sql` で bus_companies テーブルを作成し、users に company_id を追加する
    - _Requirements: 1.6, 2.5, 5.1, 8.1, 10.5, 10.7_

  - [x] 1.3 Go バックエンドの初期セットアップを行う
    - `go mod init` で `backend/go.mod` を作成し、Echo / golang-jwt / bcrypt / pgx などの依存を追加する
    - `backend/config/config.go` で環境変数（DATABASE_URL, JWT_SECRET, OSRM_BASE_URL）を読み込む
    - `backend/db/db.go` で pgx を使った DB 接続プールを実装する
    - `backend/main.go` で Echo インスタンスを起動し、ヘルスチェックエンドポイント `GET /health` を実装する
    - _Requirements: 7.1, 7.2_

  - [x] 1.4 Vue.js フロントエンドの初期セットアップを行う
    - Vite + Vue 3 でプロジェクトを生成する
    - Pinia / Vue Router / Axios / Leaflet / Tailwind CSS をインストールする
    - `frontend/src/router/index.js` にルート定義の骨格を作成する
    - `frontend/src/stores/auth.js` に Pinia ストアの骨格を作成する
    - _Requirements: 1.1_

- [x] 2. バックエンド: 認証機能
  - [x] 2.1 User モデルとリポジトリを実装する
    - `backend/model/user.go` に User 構造体と Role 定数（bus_operator / shipper）を定義する
    - `backend/repository/user_repo.go` に `FindByEmail` / `FindByID` / `Create` メソッドを実装する
    - _Requirements: 1.2, 1.3, 1.6_

  - [x] 2.2 Auth サービスを実装する
    - `backend/service/auth_service.go` に `Register` / `Login` メソッドを実装する
    - `Register`: bcrypt でパスワードをハッシュ化して保存する
    - `Login`: メールアドレスとパスワードを検証し、JWT（有効期限24h）を発行する
    - ログイン画面のロールとDBのロールが一致しない場合は認証エラーを返す
    - _Requirements: 1.2, 1.3, 1.4, 1.6, 1.7, 1.8_

  - [ ]* 2.3 Property 1 のプロパティテストを書く（有効な認証情報でJWTが発行される）
    - **Validates: Requirements 1.2, 1.3**

  - [ ]* 2.4 Property 2 のプロパティテストを書く（無効な認証情報でエラーが返る）
    - **Validates: Requirements 1.4**

  - [ ]* 2.5 Property 3 のプロパティテストを書く（パスワードのbcryptハッシュ化）
    - **Validates: Requirements 1.6**

  - [x] 2.6 JWT ミドルウェアと Auth ハンドラーを実装する
    - `backend/middleware/auth.go` に JWT 検証ミドルウェアと RequireRole ミドルウェアを実装する
    - `backend/middleware/cors.go` に CORS 設定を実装する
    - `backend/handler/auth.go` に `POST /api/v1/auth/register` と `POST /api/v1/auth/login` を実装する
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 7.4, 7.5_

  - [ ]* 2.7 Property 14 のプロパティテストを書く（有効なJWTで保護エンドポイントにアクセスできる）
    - **Validates: Requirements 7.5, 1.5**

- [ ] 3. チェックポイント — 認証APIの動作確認
  - すべてのテストが通ることを確認する。疑問点があればユーザーに確認する。

- [x] 4. バックエンド: スケジュール機能
  - [x] 4.1 Schedule モデルとリポジトリを実装する
    - `backend/model/schedule.go` に Schedule 構造体と ScheduleStatus 型（open/full/departed/arrived）を定義する
    - `backend/repository/schedule_repo.go` に `Create` / `FindByID` / `ListByOperator` / `Search` / `UpdateStatus` / `Delete` を実装する
    - Search は出発地エリア（緯度経度の範囲）・目的地エリア・日付でフィルタリングする
    - _Requirements: 2.5, 3.1, 4.1, 4.4_

  - [x] 4.2 Schedule サービスを実装する
    - `backend/service/schedule_service.go` に `Create` / `ListByOperator` / `Search` / `GetByID` / `UpdateScheduleStatus` / `Delete` を実装する
    - `Create` では出発地・目的地の未選択チェックと出発日時の過去チェックを行う
    - `avail_weight_kg` の初期値は `max_weight_kg` と同値で設定する
    - `UpdateScheduleStatus` は scheduleStatusOrder マップで前方向遷移のみ許可する
    - `Delete` は予約あり・出発済み・到着済みの場合はエラーを返す
    - _Requirements: 2.5, 2.7, 2.8, 3.1, 3.5, 3.6, 3.7, 3.8_

  - [ ]* 4.3 Property 4 のプロパティテストを書く（スケジュール登録のラウンドトリップ）
    - **Validates: Requirements 2.5**

  - [ ]* 4.4 Property 5 のプロパティテストを書く（過去日時のスケジュール登録はエラー）
    - **Validates: Requirements 2.8**

  - [ ]* 4.5 Property 6 のプロパティテストを書く（オペレーターは自分のスケジュールのみ取得できる）
    - **Validates: Requirements 3.1**

  - [ ]* 4.6 Property 7 のプロパティテストを書く（スケジュール一覧レスポンスに必須フィールドが含まれる）
    - **Validates: Requirements 3.2, 4.6**

  - [ ]* 4.7 Property 8 のプロパティテストを書く（検索条件に合致するスケジュールのみ返される）
    - **Validates: Requirements 4.4**

  - [x] 4.8 Schedule ハンドラーを実装する
    - `backend/handler/schedule.go` に以下のエンドポイントを実装する
      - `GET /api/v1/schedules` (JWT: Operator) — 自社スケジュール一覧（予約情報含む）
      - `POST /api/v1/schedules` (JWT: Operator) — スケジュール登録
      - `GET /api/v1/schedules/:id` (JWT) — スケジュール詳細
      - `GET /api/v1/schedules/search` (JWT: Shipper) — スケジュール検索
      - `PATCH /api/v1/schedules/:id/status` (JWT: Operator) — ステータス更新
      - `DELETE /api/v1/schedules/:id` (JWT: Operator) — スケジュール削除
    - _Requirements: 2.5, 3.1, 3.2, 3.5, 3.6, 4.1, 4.4, 4.6_

  - [x] 4.9 OSRM ルーティングプロキシエンドポイントを実装する
    - `GET /api/v1/routing` で OSRM パブリックAPIに転送する（タイムアウト5秒）
    - OSRM が応答しない場合は出発地・目的地の直線距離 GeoJSON を返すフォールバックを実装する
    - _Requirements: 6.4, 6.5_

- [x] 5. バックエンド: 予約機能
  - [x] 5.1 Booking モデルとリポジトリを実装する
    - `backend/model/booking.go` に Booking 構造体・BookingStatus 型・StatusOrder マップ・CanTransitionTo メソッドを定義する
    - `backend/repository/booking_repo.go` に `Create` / `FindByID` / `FindByTrackingNumber` / `ListByShipper` / `UpdateStatus` を実装する
    - _Requirements: 5.1, 5.2, 5.6, 8.1_

  - [x] 5.2 Booking サービスを実装する（アトミック更新）
    - `backend/service/booking_service.go` に `Create` / `ListByShipper` / `GetByID` を実装する
    - `Create` ではシステム制限チェック（10kg / 140cm）をトランザクション前に行う
    - `Create` では PostgreSQL トランザクション + `SELECT FOR UPDATE` で `avail_weight_kg` をアトミックに更新する
    - 重量超過・サイズ超過のチェックをトランザクション内で行う
    - 予約登録時に `tracking_number`（TRK-XXXXXXXX形式）を生成し、初期ステータスを `accepted` に設定する
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.7, 5.8, 8.2_

  - [ ]* 5.3 Property 9 のプロパティテストを書く（予約登録のラウンドトリップ・tracking_number発行）
    - **Validates: Requirements 5.1, 5.2, 8.2**

  - [ ]* 5.4 Property 10 のプロパティテストを書く（積載制限超過でエラーが返る）
    - **Validates: Requirements 5.3, 5.4, 5.7, 5.8**

  - [ ]* 5.5 Property 11 のプロパティテストを書く（予約登録後の残り積載可能重量の正確な更新）
    - **Validates: Requirements 5.5**

  - [ ]* 5.6 Property 6 のプロパティテストを書く（Shipperは自分の予約のみ取得できる）
    - **Validates: Requirements 5.6**

  - [x] 5.7 Booking ハンドラーを実装する
    - `backend/handler/booking.go` に以下のエンドポイントを実装する
      - `GET /api/v1/bookings` (JWT: Shipper) — 自分の予約一覧
      - `POST /api/v1/bookings` (JWT: Shipper) — 荷物予約
      - `GET /api/v1/bookings/:id` (JWT) — 予約詳細
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 6. バックエンド: 追跡機能
  - [x] 6.1 Tracking サービスとリポジトリを実装する
    - `backend/repository/tracking_repo.go` に `InsertStatusLog` を実装する
    - `backend/service/tracking_service.go` に `GetByTrackingNumber` / `UpdateStatus` を実装する
    - `UpdateStatus` ではステータスの逆方向遷移をバリデーションエラーとして拒否する
    - ステータス更新時に `booking_status_logs` にログを挿入する（トランザクション内）
    - _Requirements: 8.1, 8.3, 8.5, 8.6_

  - [ ]* 6.2 Property 15 のプロパティテストを書く（tracking_numberによるステータス照会のラウンドトリップ）
    - **Validates: Requirements 8.3**

  - [ ]* 6.3 Property 16 のプロパティテストを書く（ステータスの前方向遷移のみ許可）
    - **Validates: Requirements 8.6**

  - [x] 6.4 Tracking ハンドラーを実装する
    - `backend/handler/tracking.go` に以下のエンドポイントを実装する
      - `GET /api/v1/tracking/:tracking_number` (認証不要) — 荷物追跡
      - `PATCH /api/v1/bookings/:id/status` (JWT: Operator) — ステータス更新
    - 存在しない tracking_number は 404 を返す
    - _Requirements: 8.3, 8.4, 8.5, 8.6, 8.7_

- [x] 7. バックエンド: バス会社機能
  - [x] 7.1 BusCompany モデルとリポジトリを実装する
    - `backend/model/company.go` に BusCompany 構造体を定義する
    - `backend/repository/company_repo.go` に `List` / `FindByID` / `UpdateStorage` を実装する
    - _Requirements: 10.1, 10.5, 10.6_

  - [x] 7.2 Company ハンドラーを実装する
    - `backend/handler/company.go` に以下のエンドポイントを実装する
      - `GET /api/v1/companies` (認証不要) — バス会社一覧
      - `GET /api/v1/companies/me` (JWT: Operator) — 自社情報取得
      - `PATCH /api/v1/companies/me/storage` (JWT: Operator) — 荷物置き場の画像・説明を更新
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

- [ ] 8. チェックポイント — バックエンド全体の動作確認
  - すべてのテストが通ることを確認する。疑問点があればユーザーに確認する。

- [x] 9. バックエンド: APIの基本品質
  - [x] 9.1 エラーハンドリングミドルウェアを実装する
    - Echo のカスタムエラーハンドラーで統一エラーレスポンス形式（`{"error": {"code": ..., "message": ...}}`）を返す
    - _Requirements: 7.2, 7.3_

  - [ ]* 9.2 Property 13 のプロパティテストを書く（すべてのAPIレスポンスがJSON形式）
    - **Validates: Requirements 7.2**

- [x] 10. フロントエンド: 認証画面
  - [x] 10.1 Vue Router のルート定義を完成させる
    - `/operator/login`, `/operator/dashboard`, `/operator/schedules`, `/operator/schedules/new`, `/operator/qrscan`, `/operator/mypage`
    - `/shipper/login`, `/shipper/dashboard`, `/shipper/schedules`, `/shipper/bookings`, `/shipper/bookings/new`
    - `/tracking`（認証不要）
    - 認証ガード（JWT がない場合はログイン画面へリダイレクト）を実装する
    - _Requirements: 1.1_

  - [x] 10.2 Pinia 認証ストアを完成させる
    - `frontend/src/stores/auth.js` に `login` / `logout` / `token` / `role` / `userId` を実装する
    - JWT をローカルストレージに保存し、Axios のデフォルトヘッダーに設定する
    - _Requirements: 1.2, 1.3, 7.5_

  - [x] 10.3 Bus_Operator ログイン画面を実装する
    - `frontend/src/views/LoginOperator.vue` にメールアドレス・パスワード入力フォームを実装する
    - ログイン成功時は `/operator/dashboard` へリダイレクトする
    - _Requirements: 1.1, 1.2, 1.4_

  - [x] 10.4 Shipper ログイン画面を実装する
    - `frontend/src/views/LoginShipper.vue` にメールアドレス・パスワード入力フォームを実装する
    - ログイン成功時は `/shipper/dashboard` へリダイレクトする
    - _Requirements: 1.1, 1.3, 1.4_

- [x] 11. フロントエンド: 地図基盤コンポーネント
  - [x] 11.1 MapView コンポーネントを実装する
    - `frontend/src/components/MapView.vue` に Leaflet.js + OpenStreetMap タイルを使った地図を実装する
    - props: `origin` / `dest`（マーカー）/ `clickable`（クリック選択モード）/ `center` / `bounds`
    - `clickable` が true のとき、クリックした座標を `emit('point-selected', {lat, lng})` で通知する
    - `center` プロップで flyTo、`bounds` プロップで flyToBounds を実行する
    - _Requirements: 6.1, 6.2, 6.6_

  - [x] 11.2 RouteMap コンポーネントを実装する
    - `frontend/src/components/RouteMap.vue` に経路ライン描画ロジックを実装する
    - OSRM API（`/api/v1/routing`）を呼び出して道路沿いの経路を取得する
    - OSRM が応答しない場合は出発地・目的地を直線で結ぶフォールバックを実装する
    - _Requirements: 6.3, 6.4, 6.5_

  - [ ]* 11.3 Property 12 のフロントエンドテストを書く（経路データが地図コンポーネントに正しく渡される）
    - **Validates: Requirements 2.6, 3.3, 4.5, 6.3**

- [x] 12. フロントエンド: Bus_Operator 画面
  - [x] 12.1 スケジュール登録画面を実装する
    - `frontend/src/views/ScheduleCreate.vue` に RouteMap（clickable モード）と登録フォームを実装する
    - 地名入力（Nominatim、400msデバウンス、最大5件）と地図クリックの両方で地点選択できる
    - 出発日時・到着予定日時・積載可能重量・積載可能サイズ（上限140cm）の入力フォームを実装する
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.9_

  - [x] 12.2 スケジュール一覧画面を実装する
    - `frontend/src/views/ScheduleList.vue` に自社スケジュール一覧を実装する
    - 一覧クリックで RouteMap に経路を表示し、紐づく予約一覧を表示する
    - ステータス変更ボタン・削除ボタンを実装する
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

  - [x] 12.3 Operator ダッシュボードを実装する
    - `frontend/src/views/OperatorDashboard.vue` にスケジュール登録・一覧・QRスキャン・マイページへのナビゲーションを実装する
    - _Requirements: 3.2_

  - [x] 12.4 QRスキャン画面を実装する
    - `frontend/src/views/QRScan.vue` に QRScanner コンポーネントを使ったスキャン機能を実装する
    - スキャン後に自動でステータスを次へ遷移させる（accepted→loaded→in_transit→delivered）
    - 手動入力フォールバックを実装する
    - 更新前・更新後ステータスとアクションメッセージを表示する
    - _Requirements: 9.3, 9.4, 9.5, 9.6, 9.7_

  - [x] 12.5 マイページを実装する
    - `frontend/src/views/OperatorMyPage.vue` に所属バス会社情報と荷物置き場管理フォームを実装する
    - 画像アップロード（JPG/PNG/WebP、最大2MB）と説明文の入力・保存を実装する
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 13. フロントエンド: Shipper 画面
  - [x] 13.1 スケジュール検索画面を実装する
    - `frontend/src/views/ScheduleSearch.vue` に出発地エリア・目的地エリア・日付の検索フォームを実装する
    - 地名入力（Nominatim）で±0.3度の bounding box を自動生成する
    - 検索結果一覧に残り積載可能重量・残り積載可能サイズを表示する
    - ステータスが `full` または `departed` のスケジュールは予約ボタンを非活性にする
    - スケジュール選択時に RouteMap で経路を表示する
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7_

  - [x] 13.2 荷物予約画面を実装する
    - `frontend/src/views/BookingCreate.vue` に荷物情報（重量・3辺合計・内容物・受取人情報）の入力フォームを実装する
    - 重量（最大10kg）・サイズ（最大140cm）のバリデーションをフォーム上で表示する
    - 予約成功時に QRCodeDisplay コンポーネントで QRコードを表示し、印刷ボタンを提供する
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.7, 5.8, 5.9, 9.1, 9.2_

  - [ ]* 13.3 BookingCreate のユニットテストを書く
    - _Requirements: 5.1, 5.3, 5.4_

  - [x] 13.4 予約一覧画面を実装する
    - `frontend/src/views/BookingList.vue` に自分の予約履歴（tracking_number・スケジュール情報・ステータス）を表示する
    - `frontend/src/components/BookingStatusBadge.vue` にステータスバッジコンポーネントを実装する
    - 60秒ごとに自動更新する
    - _Requirements: 5.6, 8.8_

  - [x] 13.5 荷物追跡画面を実装する
    - `frontend/src/views/Tracking.vue` に tracking_number 入力フォームと照会結果表示を実装する
    - 認証不要でアクセスできるようにルートガードを設定する
    - 照会後30秒ごとに自動更新し、`delivered` になったら停止する
    - _Requirements: 8.3, 8.4, 8.7, 8.9_

  - [ ]* 13.6 Tracking 画面のユニットテストを書く
    - _Requirements: 8.3, 8.4_

- [ ] 14. チェックポイント — フロントエンド全体の動作確認
  - すべてのテストが通ることを確認する。疑問点があればユーザーに確認する。

- [x] 15. 統合確認
  - [x] 15.1 Docker Compose で全サービスを結合する
    - `docker-compose up` で db / backend / frontend が正常に起動することを確認する
    - backend の `main.go` に全ルートを登録し、ミドルウェア（CORS, JWT, エラーハンドラー）を適用する
    - _Requirements: 7.1, 7.4_

  - [x] 15.2 統合テストを実装する
    - `backend/integration_test.go` を作成し、`-tags=integration` フラグで実行できるようにする
    - 認証 → スケジュール登録 → 予約 → 追跡の一連のフローをテストする
    - _Requirements: 1.2, 2.5, 5.1, 8.3_

- [x] 16. 最終チェックポイント — すべてのテストが通ることを確認する
  - `go test ./... -tags=integration` を実行してすべてのテストが通ることを確認する。

- [x] 17. 予約キャンセル機能
  - [x] 17.1 DBマイグレーション: cancelled ステータスを追加する
    - `db/init/005_add_cancelled_status.sql` を作成する
    - bookings.status の CHECK 制約に `'cancelled'` を追加する
    - _Requirements: 11.8_

  - [x] 17.2 model/booking.go を更新する
    - `BookingStatusCancelled BookingStatus = "cancelled"` 定数を追加する
    - StatusOrder には追加しない（終端状態のため CanTransitionTo の対象外）
    - _Requirements: 11.12_

  - [x] 17.3 service/booking_service.go を更新する
    - `ErrCannotCancel` エラー変数を追加する
    - `Cancel(ctx, bookingID, shipperID uuid.UUID) error` メソッドを追加する
    - トランザクション内で: booking を FOR UPDATE でロック → shipper_id 確認（不一致は ErrForbidden）→ status = accepted 確認（それ以外は ErrCannotCancel）→ status を cancelled に更新 → schedule の avail_weight_kg を weight_kg 分加算する
    - _Requirements: 11.3, 11.4, 11.5, 11.7_

  - [ ]* 17.4 Property 18 のプロパティテストを書く（キャンセル後のステータスと追跡照会のラウンドトリップ）
    - **Property 18: キャンセル後のステータスと追跡照会のラウンドトリップ**
    - **Validates: Requirements 11.3, 11.9**

  - [ ]* 17.5 Property 19 のプロパティテストを書く（キャンセル後の avail_weight_kg 回復）
    - **Property 19: キャンセル後の avail_weight_kg 回復**
    - **Validates: Requirements 11.4**

  - [ ]* 17.6 Property 20 のプロパティテストを書く（accepted 以外へのキャンセル試行はエラー）
    - **Property 20: accepted 以外へのキャンセル試行はエラー**
    - **Validates: Requirements 11.5, 11.12**

  - [ ]* 17.7 Property 21 のプロパティテストを書く（他 Shipper の予約へのキャンセル試行は 403）
    - **Property 21: 他 Shipper の予約へのキャンセル試行は 403**
    - **Validates: Requirements 11.7**

  - [x] 17.8 handler/interfaces.go を更新する
    - `BookingServiceInterface` に `Cancel(ctx context.Context, bookingID uuid.UUID, shipperID uuid.UUID) error` を追加する
    - _Requirements: 11.3_

  - [x] 17.9 handler/booking.go を更新する
    - `Cancel` ハンドラーを追加する（DELETE /api/v1/bookings/:id）
    - shipper_id を JWT から取得する
    - ErrBookingNotFound → 404、ErrForbidden → 403、ErrCannotCancel → 409 CANNOT_CANCEL
    - _Requirements: 11.3, 11.5, 11.6, 11.7_

  - [x] 17.10 main.go を更新する
    - `bookings.DELETE("/:id", bookingHandler.Cancel, shipperMW)` を追加する
    - _Requirements: 11.3_

  - [x] 17.11 フロントエンド: BookingStatusBadge.vue を更新する
    - `cancelled` ステータスのバッジ（グレー系: `bg-gray-200 text-gray-600`）を追加する
    - _Requirements: 11.10_

  - [x] 17.12 フロントエンド: BookingList.vue を更新する
    - accepted ステータスの予約にキャンセルボタンを追加する
    - クリック時に確認ダイアログを表示する
    - 確認後に `DELETE /api/v1/bookings/:id` を呼び出す
    - 成功後に予約一覧を再取得する
    - _Requirements: 11.1, 11.2, 11.3_

  - [ ]* 17.13 Property 17 のフロントエンドテストを書く（accepted のみキャンセルボタンが表示される）
    - **Property 17: accepted ステータスの予約のみキャンセルボタンが表示される**
    - **Validates: Requirements 11.1, 11.2**

  - [x] 17.14 フロントエンド: ScheduleList.vue を更新する
    - キャンセル済み（cancelled）予約のステータス変更ボタンを非表示にする
    - _Requirements: 11.11_

  - [x] 17.15 フロントエンド: Tracking.vue を更新する
    - ポーリング停止条件に `cancelled` を追加する（delivered と同様に停止）
    - _Requirements: 11.9_

  - [x] 17.16 フロントエンド: QRScan.vue を更新する
    - `nextStatus` マップに `cancelled` が含まれないことを確認し、cancelled ステータスの荷物スキャン時に「この荷物はキャンセル済みです」メッセージを表示する
    - _Requirements: 11.12_

  - [ ]* 17.17 テスト: handler/test/booking_handler_test.go を更新する
    - `TestBookingCancel_*` テストを追加する（正常系・404・403・409）
    - mock_services_test.go の MockBookingService に Cancel メソッドを追加する
    - _Requirements: 11.3, 11.5, 11.6, 11.7_

  - [x]* 17.18 テスト: service/test/booking_service_test.go を更新する
    - `TestBookingService_Cancel_*` テストを追加する（正常系・NotFound・Forbidden・CannotCancel）
    - _Requirements: 11.3, 11.4, 11.5, 11.7_

- [ ] 18. 最終チェックポイント — キャンセル機能のテストが通ることを確認する
  - すべてのテストが通ることを確認する。疑問点があればユーザーに確認する。

---

## 備考

- `*` が付いたサブタスクはオプションであり、MVP を優先する場合はスキップ可能
- 各タスクは対応する要件番号を参照しており、トレーサビリティを確保している
- フロントエンドテストは Vitest + Vue Test Utils を使用する
