# Bugfix Requirements Document

## Introduction

JWT_SECRETを環境変数化した変更後、operatorがスケジュール一覧（GET /api/v1/schedules）を取得しようとすると `{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}` が返るバグ。

トークンの発行（ログイン時）と検証（ミドルウェア）で使用される `JWT_SECRET` の値が一致しない場合、署名検証に失敗する。具体的には、`.env` ファイルが正しく読み込まれていない・または環境変数化前後でシークレット値が変わったことにより、既存トークンが無効化される。

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN operatorが有効なJWTトークン（ログイン時に発行済み）をAuthorizationヘッダーに付けてGET /api/v1/schedulesにリクエストを送る THEN システムは `{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}` を返す

1.2 WHEN バックエンドコンテナ起動時に `JWT_SECRET` 環境変数が空文字または未設定の状態でログインしてトークンを取得し、その後 `JWT_SECRET` が正しく設定された状態でそのトークンを検証する THEN システムはトークンの署名検証に失敗する

1.3 WHEN JWT_SECRET環境変数化前に発行されたトークン（ハードコードされた旧シークレットで署名）を、環境変数化後のバックエンド（新シークレットで検証）に送る THEN システムはトークンを無効と判断する

### Expected Behavior (Correct)

2.1 WHEN operatorが正しいJWT_SECRETで署名されたトークンをAuthorizationヘッダーに付けてGET /api/v1/schedulesにリクエストを送る THEN システムはスケジュール一覧を正常に返す（HTTP 200）

2.2 WHEN バックエンドコンテナが `.env` の `JWT_SECRET` を正しく読み込んだ状態でログインしてトークンを取得し、同じシークレットで検証する THEN システムはトークンを有効と判断し認証を通過させる

2.3 WHEN `JWT_SECRET` が空文字または未設定の場合 THEN システムはサーバー起動時にエラーログを出力し、空のシークレットでトークンを発行・検証しない

### Unchanged Behavior (Regression Prevention)

3.1 WHEN 有効期限切れのトークンを送る THEN システムは引き続き `{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}` を返す

3.2 WHEN Authorizationヘッダーが存在しないリクエストを送る THEN システムは引き続き `{"error":{"code":"UNAUTHORIZED","message":"missing or invalid authorization header"}}` を返す

3.3 WHEN shipperロールのトークンでoperator専用エンドポイントにアクセスする THEN システムは引き続き `{"error":{"code":"FORBIDDEN","message":"insufficient permissions"}}` を返す

3.4 WHEN 正しいシークレットで署名された有効なトークンでshipperがGET /api/v1/schedules/searchにアクセスする THEN システムは引き続き正常にスケジュール検索結果を返す

3.5 WHEN POST /api/v1/auth/loginに正しい認証情報を送る THEN システムは引き続きJWTトークンを発行する
