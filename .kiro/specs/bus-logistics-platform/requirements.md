# 要件定義書

## はじめに

本システムは、バスの回送運転（空車で走る区間）を活用して荷物の配送を同時に行う「バス×物流プラットフォーム」です。
バス会社は回送区間の収益化ができ、荷主は既存の物流網より安価に荷物を送ることができます。

個人開発を前提とし、OSSのみを使用します。技術スタックはDocker + Go言語（Air によるホットデプロイ）を基本とします。

---

## 用語集

- **System**: バス回送物流プラットフォーム全体
- **Bus_Operator**: バス会社のスタッフ。回送スケジュールを登録・管理するユーザー
- **Shipper**: 荷主。荷物を送りたい個人または事業者
- **BusCompany**: バス会社マスター。Bus_Operator が所属する会社エンティティ
- **Route**: 回送バスの出発地から目的地までの経路
- **Schedule**: 特定の日時・経路に紐づく回送バスの運行計画
- **Cargo**: 荷物。3辺合計140cm以下、重量10kg以下の小荷物
- **Booking**: 荷主がスケジュールに対して行う荷物の予約
- **Booking_Status**: Bookingの現在状態。「受付済み（accepted）」「積載済み（loaded）」「輸送中（in_transit）」「配達完了（delivered）」の4種類
- **Tracking_Number**: 予約確認番号。Booking登録時に発行される一意の識別子（TRK-XXXXXXXX形式）
- **Map_View**: 地図上に経路・地点を表示するUIコンポーネント
- **Auth_Service**: 認証・認可を担うサービス
- **Schedule_Service**: 回送スケジュールの登録・管理を担うサービス
- **Booking_Service**: 荷物予約の登録・管理を担うサービス
- **Tracking_Service**: Bookingのステータス照会・更新を担うサービス
- **Company_Service**: バス会社情報（荷物置き場）の管理を担うサービス
- **Geocoding**: 地名テキストから緯度経度座標を取得する処理（Nominatim APIを使用）
- **QR_Scan**: 荷物に貼付されたQRコードをカメラで読み取り、ステータスを自動更新する機能

---

## 要件

### 要件1: ユーザー認証

**ユーザーストーリー:** バス会社スタッフおよび荷主として、専用のログイン画面からシステムにアクセスしたい。そうすることで、自分のロールに応じた機能だけを利用できる。

#### 受け入れ基準

1. THE System SHALL 提供する2種類のログイン画面（Bus_Operator用・Shipper用）を独立したURLで提供する
2. WHEN Bus_Operator がメールアドレスとパスワードを入力してログインを実行したとき、THE Auth_Service SHALL 認証情報を検証し、Bus_Operator ダッシュボードへリダイレクトする
3. WHEN Shipper がメールアドレスとパスワードを入力してログインを実行したとき、THE Auth_Service SHALL 認証情報を検証し、Shipper ダッシュボードへリダイレクトする
4. IF 認証情報が一致しない場合、THEN THE Auth_Service SHALL エラーメッセージを返し、ログイン画面に留まらせる
5. WHEN 認証済みユーザーがセッション有効期限切れの状態でAPIを呼び出したとき、THE Auth_Service SHALL 401ステータスを返す
6. THE Auth_Service SHALL パスワードをbcryptでハッシュ化して保存する
7. THE Auth_Service SHALL ログイン画面のロール（Bus_Operator / Shipper）とDBに登録されたロールが一致しない場合、認証エラーを返す（Bus_Operator 画面から Shipper アカウントでのログインを防ぐ）
8. THE Auth_Service SHALL JWTトークンの有効期限を24時間とする

---

### 要件2: 回送スケジュール登録

**ユーザーストーリー:** Bus_Operator として、出発地・目的地を指定して回送スケジュールを登録したい。そうすることで、荷主が利用可能な回送便を把握できる。

#### 受け入れ基準

1. WHEN Bus_Operator がスケジュール登録画面を開いたとき、THE Map_View SHALL インタラクティブな地図を表示する
2. WHEN Bus_Operator が地図上の任意の地点をクリックしたとき、THE Map_View SHALL その地点を出発地または目的地として選択状態にする（1回目クリック→出発地、2回目クリック→目的地）
3. WHEN Bus_Operator が出発地名または目的地名のテキストフィールドに2文字以上入力したとき、THE System SHALL Nominatim API を使用して候補地名の一覧をドロップダウン表示する（400msデバウンス、最大5件）
4. WHEN Bus_Operator がドロップダウンから候補地名を選択したとき、THE System SHALL その地点の緯度経度を自動的にセットし、地図をその地点に移動する
5. THE Schedule_Service SHALL 出発地・目的地（緯度経度・地名）・出発日時・到着予定日時・積載可能重量・積載可能サイズ（3辺合計上限）を必須項目としてスケジュールを登録する
6. WHEN スケジュールが正常に登録されたとき、THE Map_View SHALL 出発地から目的地までの経路をルートライン（折れ線）として地図上に表示する
7. IF 出発地または目的地が未選択の状態で登録を実行した場合、THEN THE Schedule_Service SHALL バリデーションエラーを返す
8. IF 出発日時が現在時刻より過去の場合、THEN THE Schedule_Service SHALL バリデーションエラーを返す
9. THE System SHALL 積載可能サイズ（max_size_cm）の上限を140cmとし、超過した場合はフォーム上でエラーを表示する

---

### 要件3: 回送スケジュール一覧表示

**ユーザーストーリー:** Bus_Operator として、登録済みの回送スケジュールを一覧で確認したい。そうすることで、スケジュールの管理と修正が容易になる。

#### 受け入れ基準

1. THE Schedule_Service SHALL Bus_Operator がログイン中に登録したスケジュールの一覧を提供する
2. WHEN Bus_Operator がスケジュール一覧画面を開いたとき、THE System SHALL 出発地・目的地・出発日時・ステータス（open/full/departed/arrived）・残重量を一覧表示する
3. WHEN Bus_Operator が一覧からスケジュールを選択したとき、THE Map_View SHALL 出発地・目的地の両方が収まるよう地図ビューを自動調整し、経路を表示する
4. WHEN Bus_Operator が一覧からスケジュールを選択したとき、THE System SHALL そのスケジュールに紐づく Booking の一覧を表示する
5. WHEN Bus_Operator がスケジュールのステータス変更ボタンを押したとき、THE Schedule_Service SHALL ステータスを前方向にのみ更新する（open→full, open→departed, full→departed, departed→arrived）
6. WHEN Bus_Operator がスケジュールの削除ボタンを押したとき、THE Schedule_Service SHALL 予約が0件かつステータスが「departed」または「arrived」でない場合に限りスケジュールを削除する
7. IF 削除対象スケジュールに予約が1件以上存在する場合、THEN THE Schedule_Service SHALL HAS_BOOKINGS エラーを返しスケジュールを削除しない
8. IF 削除対象スケジュールのステータスが「departed」または「arrived」の場合、THEN THE Schedule_Service SHALL INVALID_STATUS エラーを返しスケジュールを削除しない

---

### 要件4: 荷主によるスケジュール検索・閲覧

**ユーザーストーリー:** Shipper として、利用可能な回送スケジュールを検索・閲覧したい。そうすることで、自分の荷物に合った回送便を選べる。

#### 受け入れ基準

1. THE System SHALL Shipper に対して、出発地・目的地の地名入力と日付で絞り込めるスケジュール検索機能を提供する
2. WHEN Shipper が出発地名または目的地名のテキストフィールドに2文字以上入力したとき、THE System SHALL Nominatim API を使用して候補地名の一覧をドロップダウン表示する
3. WHEN Shipper がドロップダウンから候補地名を選択したとき、THE System SHALL その地点を中心に±0.3度の範囲を検索条件として自動設定する
4. WHEN Shipper が検索条件を入力して検索を実行したとき、THE Schedule_Service SHALL 条件に合致するステータスが「受付中（open）」のスケジュールの一覧を返す
5. WHEN Shipper がスケジュール一覧から1件を選択したとき、THE Map_View SHALL 出発地・目的地の両方が収まるよう地図ビューを自動調整し、経路を表示する
6. THE System SHALL 各スケジュールに対して、残り積載可能重量と残り積載可能サイズを表示する
7. WHILE スケジュールのステータスが「満載（full）」または「出発済み（departed）」の場合、THE System SHALL そのスケジュールへの予約ボタンを非活性にする

---

### 要件5: 荷物予約

**ユーザーストーリー:** Shipper として、選択した回送スケジュールに荷物を予約したい。そうすることで、バスに荷物を乗せる手配ができる。

#### 受け入れ基準

1. WHEN Shipper が予約フォームに荷物情報（重量・3辺合計・内容物の概要・受取人情報）を入力して予約を実行したとき、THE Booking_Service SHALL 予約を登録する
2. WHEN 予約が正常に登録されたとき、THE Booking_Service SHALL 予約確認番号（Tracking_Number）を Shipper に返す
3. IF 予約しようとした荷物の重量がスケジュールの残り積載可能重量を超える場合、THEN THE Booking_Service SHALL CAPACITY_EXCEEDED エラーを返す
4. IF 予約しようとした荷物の3辺合計がスケジュールの積載可能サイズ上限を超える場合、THEN THE Booking_Service SHALL SIZE_EXCEEDED エラーを返す
5. THE Booking_Service SHALL 予約登録時にスケジュールの残り積載可能重量をアトミックに更新する（SELECT FOR UPDATE + トランザクション）
6. WHEN Shipper が予約一覧画面を開いたとき、THE System SHALL 自身の予約履歴（予約確認番号・スケジュール情報・ステータス）を表示する
7. IF 予約しようとした荷物の重量が1個あたりのシステム上限（10kg）を超える場合、THEN THE Booking_Service SHALL WEIGHT_LIMIT_EXCEEDED エラーを返す
8. IF 予約しようとした荷物の3辺合計がシステム上限（140cm）を超える場合、THEN THE Booking_Service SHALL SIZE_LIMIT_EXCEEDED エラーを返す
9. WHEN 予約が正常に登録されたとき、THE System SHALL QRコードを画面に表示し、印刷できるようにする

---

### 要件6: 地図表示基盤

**ユーザーストーリー:** システム利用者として、地図上で直感的に出発地・目的地・経路を確認したい。そうすることで、回送ルートを視覚的に把握できる。

#### 受け入れ基準

1. THE Map_View SHALL OSSの地図ライブラリ（Leaflet.js + OpenStreetMap）を使用して地図を表示する
2. THE Map_View SHALL 地図上にマーカーを配置して出発地（青）・目的地（赤）を示す
3. WHEN 経路データが提供されたとき、THE Map_View SHALL 出発地から目的地を結ぶルートラインを地図上に描画する
4. WHERE ルーティング機能が有効な場合、THE Map_View SHALL OSRMを使用して道路に沿った経路を計算する
5. IF ルーティングAPIが応答しない場合、THEN THE Map_View SHALL 出発地と目的地を直線で結ぶフォールバック表示を行う
6. WHEN 地点が選択されたとき、THE Map_View SHALL `center` プロップまたは `bounds` プロップに応じてアニメーション付きで地図ビューを移動する

---

### 要件7: APIの基本品質

**ユーザーストーリー:** 開発者として、安定したAPIを提供したい。そうすることで、フロントエンドとバックエンドの連携が確実に機能する。

#### 受け入れ基準

1. THE System SHALL すべてのAPIエンドポイントをRESTful形式で提供する
2. THE System SHALL すべてのAPIレスポンスをJSON形式で返す
3. IF サーバー内部でエラーが発生した場合、THEN THE System SHALL 500ステータスとエラー詳細を含むJSONを返す
4. THE System SHALL CORSを設定し、フロントエンドオリジンからのリクエストを許可する
5. WHILE Bus_Operator または Shipper が認証済みセッションを持つ場合、THE Auth_Service SHALL JWTトークンによるリクエスト認可を行う

---

### 要件8: 荷物追跡

**ユーザーストーリー:** Shipper として、予約確認番号を使って荷物の現在ステータスを確認したい。そうすることで、荷物がどの段階にあるかをリアルタイムで把握できる。

#### 受け入れ基準

1. THE Tracking_Service SHALL Booking_Status として「受付済み（accepted）」「積載済み（loaded）」「輸送中（in_transit）」「配達完了（delivered）」の4種類の状態を管理する
2. WHEN Booking が正常に登録されたとき、THE Booking_Service SHALL その Booking の Booking_Status を「受付済み（accepted）」に設定する
3. WHEN Shipper が Tracking_Number を入力してステータス照会を実行したとき、THE Tracking_Service SHALL 該当 Booking の Booking_Status・スケジュール情報（出発地・目的地・出発日時）・最終更新日時を返す
4. IF 入力された Tracking_Number に該当する Booking が存在しない場合、THEN THE Tracking_Service SHALL 404ステータスとエラーメッセージを返す
5. WHEN Bus_Operator が Booking の Booking_Status を更新したとき、THE Tracking_Service SHALL 更新後の Booking_Status と更新日時を booking_status_logs テーブルに記録する
6. IF Bus_Operator が現在の Booking_Status より前の状態へ戻す更新を実行した場合、THEN THE Tracking_Service SHALL バリデーションエラーを返す
7. THE Tracking_Service SHALL Shipper の認証なしに Tracking_Number のみでステータス照会を可能にする
8. THE System SHALL Shipper の予約一覧画面において60秒ごとに自動更新し、Bus_Operator が更新した Booking_Status を反映する
9. THE System SHALL Shipper の荷物追跡画面において照会実行後30秒ごとに自動更新し、最新の Booking_Status を反映する。Booking_Status が「配達完了（delivered）」になった場合は自動更新を停止する

---

### 要件9: QRコードによる非対面受け渡し

**ユーザーストーリー:** Bus_Operator として、荷物に貼付されたQRコードをスキャンしてステータスを更新したい。そうすることで、荷主と対面せずに「いつ・誰が」荷物を扱ったかのログを残せる。

#### 受け入れ基準

1. WHEN Shipper が予約を完了したとき、THE System SHALL Tracking_Number を値とするQRコードを画面に表示する
2. THE System SHALL QRコードの印刷機能を提供する（ブラウザの印刷ダイアログを使用）
3. WHEN Bus_Operator がQRスキャン画面でQRコードを読み取ったとき、THE System SHALL Tracking_Number を取得し、現在のステータスを確認する
4. WHEN Bus_Operator がQRコードをスキャンしたとき、THE Tracking_Service SHALL 現在のステータスから次のステータスへ自動的に遷移させる（accepted→loaded→in_transit→delivered）
5. IF スキャンした荷物のステータスがすでに「配達完了（delivered）」の場合、THEN THE System SHALL エラーメッセージを表示しステータスを変更しない
6. THE System SHALL カメラが使用できない場合のフォールバックとして、Tracking_Number を手動入力してステータスを更新できる機能を提供する
7. WHEN ステータスが更新されたとき、THE System SHALL 更新前・更新後のステータスと対応するアクションメッセージを画面に表示する

---

### 要件10: バス会社マイページ（荷物置き場管理）

**ユーザーストーリー:** Bus_Operator として、自社の荷物置き場の写真と説明を登録したい。そうすることで、荷主が迷わず安全な場所へ荷物を届けられる。

#### 受け入れ基準

1. THE System SHALL Bus_Operator に対して、所属バス会社の情報を表示するマイページを提供する
2. WHEN Bus_Operator がマイページを開いたとき、THE System SHALL 所属バス会社名・荷物置き場の画像・説明文を表示する
3. THE Company_Service SHALL Bus_Operator が荷物置き場の画像（JPG/PNG/WebP、最大2MB）をアップロードできる機能を提供する
4. THE Company_Service SHALL Bus_Operator が荷物置き場の説明文を登録・更新できる機能を提供する
5. WHEN Bus_Operator が画像と説明文を保存したとき、THE Company_Service SHALL bus_companies テーブルの storage_image_url と storage_description を更新する
6. THE System SHALL バス会社一覧（GET /api/v1/companies）を認証不要で提供する（荷主が荷物置き場を確認できるようにするため）
7. THE System SHALL 初期データとして沖縄4社（琉球バス交通・那覇バス・沖縄バス・東陽バス）を登録する

---

### 要件11: 予約キャンセル

**ユーザーストーリー:** Shipper として、受付済みの予約をキャンセルしたい。そうすることで、荷物を送る予定がなくなった場合に回送スペースを解放できる。

#### 受け入れ基準

1. WHEN Shipper が自分の予約一覧画面を開いたとき、THE System SHALL Booking_Status が「受付済み（accepted）」の予約に対してキャンセルボタンを表示する
2. WHILE Booking_Status が「積載済み（loaded）」「輸送中（in_transit）」「配達完了（delivered）」の場合、THE System SHALL キャンセルボタンを表示しない
3. WHEN Shipper がキャンセルボタンを押して確認ダイアログで承認したとき、THE Booking_Service SHALL 該当 Booking の Booking_Status を「キャンセル済み（cancelled）」に更新する
4. WHEN Booking_Service が Booking をキャンセルしたとき、THE Booking_Service SHALL 該当 Booking の weight_kg 分だけ紐づく Schedule の avail_weight_kg をアトミックに加算する（SELECT FOR UPDATE + トランザクション）
5. IF キャンセル対象の Booking の Booking_Status が「accepted」以外の場合、THEN THE Booking_Service SHALL INVALID_STATUS_TRANSITION エラーを返しキャンセルを行わない
6. IF キャンセル対象の Booking が存在しない場合、THEN THE Booking_Service SHALL 404ステータスを返す
7. IF キャンセル対象の Booking の shipper_id が操作中の Shipper の ID と一致しない場合、THEN THE Booking_Service SHALL 403ステータスを返す
8. THE System SHALL bookings.status の CHECK 制約に「cancelled」を追加する（DBマイグレーション）
9. WHEN Booking_Status が「キャンセル済み（cancelled）」の予約を荷物追跡画面で照会したとき、THE Tracking_Service SHALL ステータスとして「キャンセル済み（cancelled）」を返す
10. THE System SHALL BookingStatusBadge コンポーネントに「キャンセル済み（cancelled）」ステータスのバッジ表示（グレー系）を追加する
11. WHEN Bus_Operator がスケジュール一覧画面でキャンセル済み予約を確認したとき、THE System SHALL キャンセル済み予約を一覧に表示し、ステータス変更ボタンを表示しない
12. THE Booking_Service SHALL キャンセル済み（cancelled）の Booking に対して、それ以降のステータス遷移（loaded / in_transit / delivered）を拒否する（cancelled は終端状態）
