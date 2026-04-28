# Implementation Plan

- [ ] 1. Write bug condition exploration test
  - **Property 1: Bug Condition** - JWT_SECRET未設定時の起動失敗
  - **CRITICAL**: このテストは未修正コードで FAIL する — 失敗がバグの存在を証明する
  - **DO NOT attempt to fix the test or the code when it fails**
  - **NOTE**: このテストは期待される挙動をエンコードしている — 実装後にパスすることで修正を検証する
  - **GOAL**: `JWT_SECRET=""` でも `config.Load()` が正常に返ってしまうことを確認し、バグの存在を証明する
  - **Scoped PBT Approach**: 決定論的バグのため、具体的な失敗ケース（`JWT_SECRET=""`）にスコープを絞る
  - `backend/config/` に `config_test.go` を作成する
  - テスト内容: `os.Setenv("JWT_SECRET", "")` した状態で `config.Load()` を呼び出し、`JWTSecret` フィールドが空文字であることを確認（未修正では PASS してしまう = バグ確認）
  - Bug Condition: `isBugCondition(input)` where `input.JWT_SECRET = ""`
  - 未修正コードでテストを実行する
  - **EXPECTED OUTCOME**: テストは FAIL する（空シークレットでも起動できてしまうことを証明）
  - 発見したカウンターエグザンプルを記録する（例: `config.Load()` が空 `JWTSecret` を持つ `Config` を返す）
  - テストを書き、実行し、失敗を記録したらタスク完了とする
  - _Requirements: 1.1, 1.2_

- [ ] 2. Write preservation property tests (BEFORE implementing fix)
  - **Property 2: Preservation** - 既存認証フローの保持
  - **IMPORTANT**: 観察ファーストの方法論に従う
  - 未修正コードで非バグ条件（`JWT_SECRET` が正しく設定されている場合）の挙動を観察する
  - 観察: `JWT_SECRET="dev-jwt-secret-..."` で `config.Load()` を呼ぶと `JWTSecret` に正しい値が返る
  - 観察: 既存の `backend/service/test/auth_service_test.go` が全てパスする
  - 観察: 既存の `backend/handler/test/auth_handler_test.go` が全てパスする
  - プロパティベーステスト: 任意の非空文字列 `JWT_SECRET` に対して `config.Load()` が正しい値を返すことを検証
  - `backend/config/config_test.go` に保持テストを追加する
  - 未修正コードでテストを実行する
  - **EXPECTED OUTCOME**: テストは PASS する（ベースライン挙動を確認）
  - テストを書き、実行し、パスを確認したらタスク完了とする
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 3. Fix for JWT_SECRET未設定時にサーバーが起動してしまうバグ

  - [ ] 3.1 Implement the fix
    - `backend/config/config.go` の `Load()` 関数を修正する
    - `jwtSecret := os.Getenv("JWT_SECRET")` で値を取得する
    - `jwtSecret == ""` の場合に `log.Fatal("JWT_SECRET environment variable is required but not set")` でサーバー起動を停止する
    - `Config` 構造体の `JWTSecret` フィールドに `jwtSecret` 変数をセットする
    - 他のファイル（`middleware/auth.go`、`service/auth_service.go`、`main.go`）への変更は不要
    - _Bug_Condition: isBugCondition(input) where input.JWT_SECRET = "" or input.JWT_SECRET is unset_
    - _Expected_Behavior: config.Load() causes fatal error when JWT_SECRET is empty_
    - _Preservation: JWT_SECRET が正しく設定されている場合の全ての認証フローは変わらない_
    - _Requirements: 2.3, 3.1, 3.2, 3.3, 3.4, 3.5_

  - [ ] 3.2 Verify bug condition exploration test now passes
    - **Property 1: Expected Behavior** - JWT_SECRET未設定時の起動失敗
    - **IMPORTANT**: タスク1で書いた同じテストを再実行する — 新しいテストを書かない
    - タスク1のテストは期待される挙動をエンコードしている
    - このテストがパスすれば、期待される挙動が満たされたことを確認できる
    - タスク1のバグ条件探索テストを実行する
    - **EXPECTED OUTCOME**: テストは PASS する（バグが修正されたことを確認）
    - _Requirements: 2.3_

  - [ ] 3.3 Verify preservation tests still pass
    - **Property 2: Preservation** - 既存認証フローの保持
    - **IMPORTANT**: タスク2で書いた同じテストを再実行する — 新しいテストを書かない
    - タスク2の保持テストを実行する
    - 既存テスト群も実行する: `docker compose run --rm backend go test ./handler/test/... ./service/test/... ./repository/test/... ./model/test/...`
    - **EXPECTED OUTCOME**: 全テストが PASS する（デグレなし）
    - 修正後も全テストがパスすることを確認する

- [ ] 4. Checkpoint - Ensure all tests pass
  - 全テストがパスすることを確認する。疑問点があればユーザーに確認する。
  - `docker compose run --rm backend go test ./...` を実行して全バックエンドテストがパスすることを確認
  - **運用対応の確認**: `.env` ファイルが正しく配置されているか確認（`docker compose up` 実行ディレクトリに `.env` が存在すること）
  - **ユーザー通知**: 既存のログイン済みトークンは無効になるため、再ログインが必要であることをユーザーに伝える
