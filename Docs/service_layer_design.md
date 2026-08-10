# 実装設計書：Service層導入 + タスク所有権対応

対象リポジトリ: practice_go（todo-api）
作成日: 2026-08-09

---

## 0. 前提（決定済み事項）

| # | 論点 | 決定 |
| --- | --- | --- |
| A | スコープ | Service層導入 + memo.mdの本来要件（user_idスコープ・PUT）を実装。ページネーションは次フェーズ |
| B | Repositoryのinterface化 | やる（Task/User両方） |
| C | 既存テストの扱い | 統合テストは維持しService経由に配線し直す。Service単体テストを新規追加 |
| D | エラー方針 | HandlerはGORM/bcrypt/JWTの実装詳細を知らない。Serviceが独自エラー型を返す |
| E | 全削除エンドポイント | 自分のタスクだけ全削除、に変更（旧：全ユーザー分を削除） |
| F | AuthServiceの扱い | 今回のスコープに含める |

---

## 1. 調査で分かった前提のズレ

- `model/task.go`に`UserID`はあるが、`handler.go`のどのメソッドも読んでいない。今のAPIは誰でも全員のタスクを操作できる。→ 今回の本題。
- memo.mdの認証版スペックに`DELETE /tasks`（全削除）は載っていない。今のコードにはある。→ Eの通り、自分のタスクのみ全削除に変更。

---

## 2. 変更対象ファイル

| ファイル | 変更内容 |
| --- | --- |
| `service/task_service.go` | 新規。TaskRepository interface + TaskService |
| `service/auth_service.go` | 新規。UserRepository interface + AuthService |
| `service/errors.go` | 新規。`ErrTaskNotFound` / `ErrForbidden` / `ErrInvalidCredentials` |
| `repository/task.go` | シグネチャ変更（userID引数追加、Update新規） |
| `repository/user.go` | 変更なし（既存メソッドがそのままinterfaceを満たす） |
| `handler/handler.go` | Service経由に変更。UserID取得追加。UpdateTask新規 |
| `handler/auth.go` | bcrypt/JWT処理をAuthServiceに移動 |
| `main.go` | DI配線変更、`PUT /tasks/:id`ルート追加 |
| `handler/handler_test.go` | Service経由に配線し直す（アサーションは維持） |
| `service/task_service_test.go` | 新規（fake repositoryでの単体テスト） |
| `service/auth_service_test.go` | 新規（fake repositoryでの単体テスト） |
| `repository/task_test.go` | シグネチャ変更に合わせて呼び出し部分のみ修正 |

---

## 3. テスト方針

- `handler_test.go`：実DB（in-memory SQLite）を使う統合テストのまま維持。`setupTestRouter`で`service.NewTaskService(repo)`を挟むだけにし、既存アサーションはそのまま使う。
- `service/task_service_test.go` / `auth_service_test.go`：interfaceを満たすfake構造体でDB無しにビジネスロジックのみ検証（所有権チェック、認証失敗など）。interface化（B）の実益がここで出る。
- `repository/task_test.go`：ロジック変更なし、呼び出し部分のみシグネチャに合わせる。

---

## 4. 詳細設計

### 4-1. Repository interface

Serviceパッケージ側で定義する（使う側がinterfaceを持つ、というGoの一般的な慣習[一般論]）。

TaskRepositoryが持つメソッド：
- `Create`：タスク作成
- `GetAllByUser(userID)`：指定ユーザーのタスク一覧取得
- `GetByID(id)`：userIDで絞らない。所有権チェックはService側でやるため
- `Update`：タスク更新（新規追加）
- `DeleteAllByUser(userID)`：指定ユーザーのタスクを全削除
- `DeleteByID(id)`：同上、userIDで絞らない

UserRepositoryが持つメソッド：既存の`Create`・`FindByEmail`のまま（変更不要、そのままinterfaceを満たす）。

なぜGetByID/DeleteByIdにuserIDを渡さないか：SQLの`WHERE id=? AND user_id=?`で絞ると「存在しない」と「他人のタスク」がどちらも0件になり区別できない。memo.mdは他人のタスクへのアクセスを403と明記しているので、「IDだけで取得→Service側でuserIDを比較」の2段階にする。GetAllByUser/DeleteAllByUserは一覧・一括操作なので403の概念がなく、SQLで絞ってよい。

### 4-2. エラー定義

Serviceパッケージに以下の3つのエラーを定義する：
- `ErrTaskNotFound`：指定IDのタスクが存在しない
- `ErrForbidden`：タスクは存在するが、リクエストしたユーザーの所有物ではない
- `ErrInvalidCredentials`：ログイン時のメール／パスワード不一致

Service内部でGORM/bcryptが返すエラーをこれらに変換する。Handlerは`gorm`パッケージをimportしなくなる。

### 4-3. TaskServiceのロジック（Get/Update/Deleteに共通）

1. `repo.GetByID(id)`でタスクを取得する。
2. GORMの「レコードなし」エラーが返ってきたら`ErrTaskNotFound`に変換する。
3. 取得できた場合、`task.UserID`とリクエストの`userID`を比較する。一致しなければ`ErrForbidden`を返す。
4. 一致すれば、そのまま処理（返却／更新／削除）を続ける。

DeleteAllTasksだけは所有権チェック不要で、`repo.DeleteAllByUser(userID)`を呼ぶだけ（E決定）。

### 4-4. AuthServiceのロジック

- Register：受け取ったパスワードをbcryptでハッシュ化してから`repo.Create`。ハッシュ化・保存の失敗は今のところ汎用エラーのまま（要件上、区別する必要がないため）。
- Login：`repo.FindByEmail`でユーザーを検索し、見つからない場合と、bcryptでのパスワード比較が失敗した場合の両方を`ErrInvalidCredentials`にまとめる（メールが存在するかどうかを外部に漏らさないため）。成功したらJWTを発行して返す。

---

## 5. Handler側の変更点

- 各エンドポイント冒頭で `userID := c.MustGet("userID").(uint)`。
- `NewHandler(repo)` → `NewHandler(service)`、`NewAuthHandler(userRepo)` → `NewAuthHandler(authService)`。
- `UpdateTask`新規（`PUT /tasks/:id`、title/statusのみ）。`main.go`に`tasks.PUT("/:id", h.UpdateTask)`追加。
- エラー判定を`gorm.ErrRecordNotFound`から`service.ErrTaskNotFound` / `ErrForbidden` / `ErrInvalidCredentials`に変更。`ErrForbidden`は403。

---

## 6. 次フェーズ

- ページネーション（`?page=&limit=`）
