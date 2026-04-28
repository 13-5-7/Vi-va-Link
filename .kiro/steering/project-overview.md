# バス回送物流プラットフォーム — プロジェクト概要

## プロジェクトの目的

バスの回送区間（空車走行）を活用して小荷物を配送するマッチングプラットフォーム。
バス会社（Bus_Operator）が回送スケジュールを登録し、荷主（Shipper）がそのスケジュールに荷物を予約する。

---

## 技術スタック

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.22 / Echo v4 |
| データベース | PostgreSQL 16 |
| フロントエンド | Vue 3 / Vite / Tailwind CSS / Pinia |
| 地図 | Leaflet + OpenStreetMap (Nominatim) |
| ルーティング | OSRM（利用不可時は直線フォールバック） |
| 認証 | JWT (HS256, 有効期限24時間) + bcrypt |
| コンテナ | Docker / Docker Compose |
| テスト (BE) | Go 標準 testing パッケージ（モックベース） |
| テスト (FE) | Vitest + @vue/test-utils + jsdom |

---

## ディレクトリ構成

```
backend/
├── main.go
├── config/        # 環境変数読み込み
├── db/            # pgx 接続プール
├── middleware/    # JWT認証・CORS・エラーハンドラー
├── handler/       # HTTPハンドラー（interfaces.go でインターフェース定義）
├── service/       # ビジネスロジック
├── repository/    # DBアクセス
└── model/         # データモデル・定数

frontend/src/
├── router/        # Vue Router（ロールガード付き）
├── stores/        # Pinia（auth.js）
├── views/         # 画面コンポーネント
├── components/    # 共通コンポーネント
└── views/__tests__/ # Vitest テスト

db/init/           # PostgreSQL 初期化 SQL（番号順に自動適用）
```

---

## ユーザーロール

| ロール | 説明 |
|---|---|
| `bus_operator` | バス会社スタッフ。スケジュール管理・QRスキャン・マイページ |
| `shipper` | 荷主。スケジュール検索・予約・追跡・荷物置き場確認 |

---

## 実装済み機能（現状）

### バックエンド API

| メソッド | パス | 認証 | 説明 |
|---|---|---|---|
| POST | /api/v1/auth/register | なし | ユーザー登録 |
| POST | /api/v1/auth/login | なし | ログイン |
| GET/POST | /api/v1/schedules | operator | スケジュール一覧・作成 |
| GET | /api/v1/schedules/search | shipper | スケジュール検索 |
| GET | /api/v1/schedules/:id | JWT | スケジュール詳細 |
| PATCH | /api/v1/schedules/:id/status | operator | ステータス更新 |
| DELETE | /api/v1/schedules/:id | operator | スケジュール削除 |
| GET/POST | /api/v1/bookings | shipper | 予約一覧・作成 |
| GET | /api/v1/bookings/:id | JWT | 予約詳細 |
| DELETE | /api/v1/bookings/:id | shipper | **予約キャンセル（accepted のみ）** |
| PATCH | /api/v1/bookings/:id/status | operator | 荷物ステータス更新 |
| GET | /api/v1/tracking/:tracking_number | なし | 荷物追跡 |
| GET | /api/v1/routing | なし | ルート取得（OSRM） |
| GET | /api/v1/companies | なし | バス会社一覧 |
| GET | /api/v1/companies/me | operator | 自社情報取得 |
| PATCH | /api/v1/companies/me/storage | operator | 荷物置き場更新 |

### フロントエンド画面

**オペレーター**
- `/operator/login` — ログイン
- `/operator/dashboard` — ダッシュボード
- `/operator/schedules` — スケジュール一覧（ステータス変更・削除・予約管理）
- `/operator/schedules/new` — スケジュール登録（地図クリック or 地名検索）
- `/operator/qrscan` — QRスキャン（ステータス自動更新）
- `/operator/mypage` — マイページ（荷物置き場画像・説明管理）

**荷主**
- `/shipper/login` — ログイン
- `/shipper/dashboard` — ダッシュボード
- `/shipper/schedules` — スケジュール検索
- `/shipper/bookings/new` — 予約登録（QRコード表示・印刷）
- `/shipper/bookings` — 予約一覧（**キャンセルボタン付き**）
- `/shipper/companies` — **荷物置き場一覧（新規追加）**
- `/tracking` — 荷物追跡（認証不要）

---

## ビジネスルール

### 荷物制限
- 1個あたり重量: 最大 **10kg**（`WEIGHT_LIMIT_EXCEEDED`）
- 1個あたりサイズ（3辺合計）: 最大 **140cm**（`SIZE_LIMIT_EXCEEDED`）

### 予約ステータス遷移
```
accepted → loaded → in_transit → delivered  （前方向のみ、QRスキャンで自動遷移）
accepted → cancelled                          （荷主がキャンセル、終端状態）
```
- `cancelled` は `StatusOrder` に含めない（`CanTransitionTo` の対象外）
- キャンセル時は `avail_weight_kg` を weight_kg 分回復（トランザクション）

### スケジュールステータス遷移
```
open → full → departed → arrived  （前方向のみ）
```
- 予約あり・departed/arrived のスケジュールは削除不可

---

## DBマイグレーション構成

| ファイル | 内容 |
|---|---|
| 001_schema.sql | 基本テーブル作成 |
| 002_data.sql | テストデータ |
| 003_add_arrived_status.sql | schedules.status に `arrived` 追加 |
| 004_bus_companies.sql | bus_companies テーブル・users.company_id 追加 |
| 005_add_cancelled_status.sql | bookings.status に `cancelled` 追加 |

---

## テスト状況

### バックエンド（Go）
- `handler/test/` — ハンドラー層（モックサービス使用）
- `service/test/` — サービス層（モックリポジトリ使用）
- `repository/test/` — リポジトリ層（インターフェース契約）
- `model/test/` — モデル層（定数・遷移ロジック）
- `integration_test.go` — 統合テスト（`-tags=integration`）

**実行コマンド:**
```bash
docker compose run --rm backend go test ./handler/test/... ./service/test/... ./repository/test/... ./model/test/...
```

### フロントエンド（Vitest）
- `src/views/__tests__/` — 8ファイル 70テスト（全 PASS）

**実行コマンド:**
```bash
docker compose run --rm frontend npm test
```

---

## 開発コマンド

```bash
# 起動
docker compose up --build

# 完全リセット（DBデータ消去）
docker compose down -v

# バックエンドテスト
docker compose run --rm backend go test ./handler/test/... ./service/test/... ./repository/test/... ./model/test/...

# フロントエンドテスト
docker compose run --rm frontend npm test

# フロントエンドカバレッジ
docker compose run --rm frontend sh -c "npm test -- --coverage --coverage.provider=v8 --coverage.include='src/views/**'"

# DB接続
docker compose exec db psql -U app -d bus_logistics
```

---

## スペックファイル

- 要件定義: `.kiro/specs/bus-logistics-platform/requirements.md`
- 設計書: `.kiro/specs/bus-logistics-platform/design.md`
- タスク: `.kiro/specs/bus-logistics-platform/tasks.md`

---

## 現在の課題・未実装

- ユーザー登録の制限なし（誰でも登録可能）
- パスワードリセット機能なし
- レートリミットなし
- 通知機能なし（メール・プッシュ）
- 決済機能なし
- 画像はBase64でDBに直接保存（大容量時にDBが肥大化するリスク）

## 必ずやること
 - 影響調査、それぞれの実装に影響する箇所を洗い出す
 - デグレ発生禁止、デグレが発生しないよう１ヵ所だけではなく必要箇所修正する

# Role
 あなたは熟練のソフトウェアアーキテクトです。
 提供する機能要件に基づき、開発者が一切迷うことなくコーディングに集中できる「詳細設計書（Markdown）」を作成してください。

# Constraints
 - コード（Go, TypeScript等）の実装そのものは出力しないでください。
 - 代わりに、ロジックのステップ、データ構造、インターフェースの定義を「日本語と言語化された仕様」で記述してください。
 - 出力形式は Markdown とします。

# Target File Path
 docs/design/functions/{機能名}.md

# Document Structure
 ## 1. 概要
 この機能の目的とゴールを簡潔に記述してください。

## 2. シーケンス図
 Mermaid記法を用いて、以下の登場人物間のやり取りを可視化してください。
 (Handler, Service, Model, DB, または外部API等)

## 3. コンポーネント詳細定義
 各レイヤーで「何をすべきか」を箇条書きで定義してください。

### Handler
 - バリデーション項目（型、必須チェック、文字数等）
 - 期待するリクエスト形式とレスポンス形式（ステータスコード含む）

### Service
 - ビジネスロジックの具体的な手順（例：1. パスワードをハッシュ化する 2. 重複チェックを行う...）
 - ここが実装の肝となるため、条件分岐や異常系の発生条件を網羅してください。

### Model / Repository
 - 使用する構造体名とフィールド定義
 - 実行するクエリの論理的な内容（例：usersテーブルへのINSERT、emailによるUNIQ制約チェック）

## 4. 実装対象ファイル一覧（トレーサビリティ）
 実装者がどのファイルを新規作成・修正すべきか、以下の形式でリストアップしてください。
 - {ファイルパス}: {メソッド名/構造体名} の役割と実装内容の要約

## 5. エラーハンドリング
 発生しうるエラー（400 Bad Request, 409 Conflict, 500 Internal Error等）と、その発生条件を定義してください。
