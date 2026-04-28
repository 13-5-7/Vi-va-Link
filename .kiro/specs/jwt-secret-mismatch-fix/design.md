# JWT Secret Mismatch Bugfix Design

## Overview

JWT_SECRETを環境変数化した変更後、トークンの署名（ログイン時）と検証（ミドルウェア）で使用されるシークレットが一致しなくなり、全ての認証が失敗するバグ。

**影響範囲**: JWT認証が必要な全エンドポイント（operator・shipper両ロール）。特にoperatorのスケジュール一覧取得（GET /api/v1/schedules）で確認されている。

**修正方針**:
1. `JWT_SECRET` が空の場合にサーバー起動を失敗させるガード追加（`config.go`）
2. 既存トークンを再発行するため、ユーザーは再ログインが必要（ドキュメント対応）

---

## Glossary

- **Bug_Condition (C)**: `JWT_SECRET` 環境変数が空文字または未設定の状態でバックエンドが起動し、空シークレットでトークンが発行・検証される条件
- **Property (P)**: 同一の `JWT_SECRET` でトークンが発行・検証されること
- **Preservation**: 既存の認証フロー（ログイン・ロール検証・期限切れ検出）が変わらないこと
- **JWTAuth**: `backend/middleware/auth.go` の `JWTAuth(jwtSecret string)` 関数。Authorizationヘッダーのトークンを検証するミドルウェア
- **NewAuthService**: `backend/service/auth_service.go` の `NewAuthService(...)` 関数。ログイン時にトークンを署名する
- **config.Load()**: `backend/config/config.go` の `Load()` 関数。環境変数を読み込んで `Config` 構造体を返す

---

## Bug Details

### Bug Condition

バグは以下のいずれかの条件で発生する：

**ケース1（最も可能性が高い）**: `.env` ファイルが正しく読み込まれず `JWT_SECRET` が空文字になった状態でバックエンドが起動。ログイン時は空シークレットでトークンを署名し、検証時も空シークレットで検証するため一見動作するが、`.env` が正しく読み込まれた後の再起動で不一致が発生。

**ケース2**: JWT_SECRET環境変数化前（ハードコード値）で発行されたトークンを、環境変数化後（`.env` の値）のバックエンドで検証しようとしている。シークレット値が変わったため署名が一致しない。

**Formal Specification:**
```
FUNCTION isBugCondition(input)
  INPUT: input of type JWTVerificationContext
    - tokenSignedWithSecret: string  // トークン署名時に使ったシークレット
    - verifyWithSecret: string       // 検証時に使うシークレット（os.Getenv("JWT_SECRET")）
  OUTPUT: boolean

  RETURN tokenSignedWithSecret != verifyWithSecret
         OR verifyWithSecret = ""
END FUNCTION
```

### Examples

- **ケース1**: `JWT_SECRET=""` でログイン → トークン発行 → `.env` 読み込み後に再起動 → `JWT_SECRET="dev-jwt-secret-..."` で検証 → 署名不一致 → 401
- **ケース2**: 旧ハードコード値 `"old-secret"` で署名されたトークン → `.env` の `"dev-jwt-secret-..."` で検証 → 署名不一致 → 401
- **正常ケース**: `JWT_SECRET="dev-jwt-secret-..."` でログイン → 同じ値で検証 → 認証成功 → 200

---

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- 有効期限切れトークンは引き続き401を返す
- Authorizationヘッダー欠如は引き続き401を返す
- ロール不一致は引き続き403を返す
- ログイン・登録エンドポイントは認証不要のまま
- shipper・operatorそれぞれのロール制御は変わらない

**Scope:**
`JWT_SECRET` が正しく設定されている場合の全ての認証フローは、この修正によって影響を受けない。修正は `config.go` への起動時バリデーション追加のみ。

---

## Hypothesized Root Cause

調査したファイルから、コードロジック自体に問題はない。根本原因は**環境設定の問題**：

1. **JWT_SECRETが空で起動している**: `config.Load()` は `os.Getenv("JWT_SECRET")` が空でも何もチェックせずそのまま返す。Dockerコンテナが `.env` を読み込む前に起動した、または `.env` ファイルのパスが間違っている場合、空シークレットで動作してしまう。

2. **シークレット値の変更**: JWT_SECRET環境変数化の際に値が変わった（ハードコード値 → `.env` の値）。既存のログイン済みトークンは旧値で署名されているため、新値での検証に失敗する。

3. **`config.go` にバリデーションがない**: `JWT_SECRET` が空でもサーバーが起動してしまい、問題が潜在化する。

---

## Correctness Properties

Property 1: Bug Condition - JWT_SECRET未設定時の起動失敗

_For any_ バックエンド起動コンテキストにおいて `JWT_SECRET` 環境変数が空文字または未設定の場合、固定された `config.Load()` 関数はエラーを返し（またはパニックし）、サーバーは起動しない。これにより空シークレットでのトークン発行・検証が防止される。

**Validates: Requirements 2.3**

Property 2: Preservation - 既存認証フローの保持

_For any_ `JWT_SECRET` が正しく設定されている入力において、修正後のコードは修正前のコードと同一の認証結果を返す。有効トークン→200、期限切れ→401、ヘッダー欠如→401、ロール不一致→403の各挙動が保持される。

**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**

---

## Fix Implementation

### Changes Required

**File**: `backend/config/config.go`

**Function**: `Load()`

**Specific Changes**:
1. **JWT_SECRETの存在チェック追加**: `os.Getenv("JWT_SECRET")` が空文字の場合、`log.Fatal` でサーバー起動を停止する
2. **戻り値の変更（オプション）**: `Load()` をエラーを返す形式に変更するか、`log.Fatal` で即時終了させる

```go
// 修正前
return &Config{
    DatabaseURL: os.Getenv("DATABASE_URL"),
    JWTSecret:   os.Getenv("JWT_SECRET"),
    ...
}

// 修正後（案）
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    log.Fatal("JWT_SECRET environment variable is required but not set")
}
return &Config{
    DatabaseURL: os.Getenv("DATABASE_URL"),
    JWTSecret:   jwtSecret,
    ...
}
```

**File**: `backend/config/config.go` の `Load()` のみ変更。他ファイルへの変更は不要。

### 運用対応（コード変更外）

- 既存のログイン済みトークンは旧シークレットで署名されているため無効。ユーザーは再ログインが必要。
- `.env` ファイルが正しく配置されているか確認（`docker compose up` 実行ディレクトリに `.env` が存在すること）。

---

## Testing Strategy

### Validation Approach

2フェーズアプローチ：まず未修正コードでバグを再現し根本原因を確認、次に修正後に正常動作と既存挙動の保持を検証する。

### Exploratory Bug Condition Checking

**Goal**: `JWT_SECRET` が空の場合にサーバーが起動してしまい、空シークレットでトークンが発行・検証されることを確認する。

**Test Plan**: `config.Load()` に `JWT_SECRET=""` を渡した場合の挙動をテストする。未修正コードでは空シークレットのまま `Config` が返ることを確認する。

**Test Cases**:
1. **空JWT_SECRET起動テスト**: `JWT_SECRET=""` の環境で `config.Load()` を呼び出し、`JWTSecret` フィールドが空文字であることを確認（未修正コードでは PASS してしまう = バグの存在を証明）
2. **空シークレットでのトークン発行・検証テスト**: 空シークレットで署名したトークンを、別のシークレットで検証すると失敗することを確認

**Expected Counterexamples**:
- 未修正の `config.Load()` は `JWT_SECRET=""` でも正常に `Config` を返す（バグ確認）

### Fix Checking

**Goal**: 修正後、`JWT_SECRET` が空の場合にサーバーが起動しないことを確認する。

**Pseudocode:**
```
FOR ALL input WHERE isBugCondition(input) DO
  result := config.Load() with JWT_SECRET=""
  ASSERT result causes fatal error / panic
END FOR
```

### Preservation Checking

**Goal**: `JWT_SECRET` が正しく設定されている場合、修正前後で認証挙動が変わらないことを確認する。

**Pseudocode:**
```
FOR ALL input WHERE NOT isBugCondition(input) DO
  ASSERT config_original(input).JWTSecret = config_fixed(input).JWTSecret
END FOR
```

**Testing Approach**: 既存の `auth_service_test.go` および `auth_handler_test.go` を活用。これらのテストが修正後も全てパスすることを確認する。

**Test Cases**:
1. **正常なJWT_SECRETでのconfig.Load()**: 値が正しく返ることを確認
2. **既存の認証テスト群**: `backend/service/test/auth_service_test.go`、`backend/handler/test/auth_handler_test.go` が全てパスすること

### Unit Tests

- `config.Load()` に `JWT_SECRET=""` を渡した場合のテスト（`backend/config/` に追加）
- `config.Load()` に有効な `JWT_SECRET` を渡した場合のテスト

### Property-Based Tests

- 任意の非空文字列 `JWT_SECRET` に対して `config.Load()` が正常に `Config` を返すことを検証
- 空文字列 `JWT_SECRET` に対して `config.Load()` が失敗することを検証

### Integration Tests

- `docker compose up` 後に `JWT_SECRET` が設定されていることを確認
- ログイン → トークン取得 → GET /api/v1/schedules の一連フローが成功することを確認
