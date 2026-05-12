# 簡易仕様書：TODO管理API

## 1. アプリ概要
シンプルなタスク（TODO）を保存・取得できるバックエンドサーバー。画面は作らず、`curl`コマンドや`Postman`などのツールで動作確認を行います。

## 2. 機能一覧（エンドポイント）
以下の3つのURL（API）を作成します。

| メソッド | パス | 内容 | 期待する動作 |
| :--- | :--- | :--- | :--- |
| **GET** | `/tasks` | タスク一覧取得 | DBから全タスクを取得し、`200 OK`でJSONを返す。 |
| **POST** | `/tasks` | タスク登録 | JSONを受け取り、DBに保存。成功なら`201 Created`。 |
| **GET** | `/tasks/{id}` | 特定タスク取得 | 指定されたIDのタスクを返す。なければ`404 Not Found`。 |

## 3. データ構造（JSON例）
```json
{
  "id": 1,
  "title": "プログラミング学習",
  "status": "pending"
}
```

## 4. 開発ステップ（10hプラン）
1.  **[1-2h] サーバー起動:** 特定のポート（例: 8080）で待ち受けるプログラムを書く。
2.  **[3-4h] Handler作成:** ダミーデータ（メモリ上の変数）を使って、URLにアクセスしたらJSONが返るようにする。
3.  **[5-8h] Repository作成:** SQLiteを導入し、実際にファイルを読み書きする。ここでSQL（`CREATE TABLE`, `INSERT`, `SELECT`）を書く。
4.  **[9-10h] エラーハンドリング:** IDが見つからない時にちゃんと `404` を返す処理を作り込む。

---

# 要件定義：認証付きTodo API（レベル1）

## 1. アプリ概要

現在のTodo APIを拡張し、**ユーザー認証（JWT）・タスク更新**を追加する。
ユーザーはログインして取得したトークンを使い、自分のタスクのみ操作できる。
DBは引き続きSQLiteを使用し、新規ライブラリの追加のみでシンプルに進める。

## 2. 技術スタック

| 分類 | 使用技術 |
| :--- | :--- |
| 言語 | Go |
| フレームワーク | Gin |
| ORM | GORM |
| DB | SQLite |
| 認証 | JWT（`golang-jwt/jwt`） |
| パスワード | bcrypt（`golang.org/x/crypto`） |
| 環境変数 | `.env`（`godotenv`） |

## 3. データモデル

### Userテーブル
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| id | uint | 主キー（自動採番） |
| name | string | ユーザー名 |
| email | string | メールアドレス（UNIQUE） |
| password | string | bcryptハッシュ済みパスワード |
| created_at | time | 作成日時（GORMが自動付与） |
| updated_at | time | 更新日時（GORMが自動付与） |

### Taskテーブル
| カラム | 型 | 説明 |
| :--- | :--- | :--- |
| id | uint | 主キー（自動採番） |
| user_id | uint | 外部キー（Userへの参照） |
| title | string | タスクタイトル |
| status | string | `pending` / `done` |
| created_at | time | 作成日時 |
| updated_at | time | 更新日時 |

**リレーション：** User は Task を複数持つ（1対多）

## 4. APIエンドポイント一覧

### 認証系（認証不要）
| メソッド | パス | 内容 | レスポンス |
| :--- | :--- | :--- | :--- |
| POST | `/auth/register` | ユーザー登録 | `201` + ユーザー情報 |
| POST | `/auth/login` | ログイン | `200` + JWTトークン |

### タスク系（JWT認証必須）
| メソッド | パス | 内容 | レスポンス |
| :--- | :--- | :--- | :--- |
| GET | `/tasks` | 自分のタスク一覧取得（ページネーション対応） | `200` + タスク配列 |
| POST | `/tasks` | タスク作成 | `201` + 作成タスク |
| GET | `/tasks/:id` | 特定タスク取得 | `200` / `404` |
| PUT | `/tasks/:id` | タスク更新（title・status） | `200` + 更新後タスク |
| DELETE | `/tasks/:id` | タスク削除 | `200` |

> 認証済みユーザーは **自分のタスクのみ** 操作可能。他人のタスクへのアクセスは `403 Forbidden`。

## 5. リクエスト／レスポンス例

### POST /auth/register
```json
// Request
{ "name": "Taro", "email": "taro@example.com", "password": "secret123" }

// Response 201
{ "id": 1, "name": "Taro", "email": "taro@example.com" }
```

### POST /auth/login
```json
// Request
{ "email": "taro@example.com", "password": "secret123" }

// Response 200
{ "token": "eyJhbGci..." }
```

### PUT /tasks/:id
```json
// Request Header
Authorization: Bearer eyJhbGci...

// Request Body
{ "title": "Go勉強", "status": "done" }

// Response 200
{ "id": 1, "title": "Go勉強", "status": "done" }
```

### GET /tasks?page=1&limit=10
```json
// Response 200
{
  "tasks": [
    { "id": 1, "title": "Go勉強", "status": "done" },
    { "id": 2, "title": "筋トレ", "status": "pending" }
  ],
  "total": 2,
  "page": 1,
  "limit": 10
}
```

## 6. 認証フロー

```
1. POST /auth/register → パスワードをbcryptでハッシュ化してDBに保存
2. POST /auth/login    → メール＋パスワード検証 → JWTトークン発行（有効期限24h）
3. 認証が必要なAPIへのリクエスト → Authorizationヘッダーを検証するミドルウェアを通過
4. ミドルウェアがトークンからuser_idを取り出し、コンテキストにセット
5. HandlerはコンテキストからリクエストしたユーザーのIDを取得してDB操作
```

## 7. ディレクトリ構成（予定）

```
todo-api/
├── main.go
├── .env
├── handler/
│   ├── auth.go          # register / login
│   ├── task.go          # タスクCRUD
│   └── handler_test.go
├── middleware/
│   └── auth.go          # JWT検証ミドルウェア
├── repository/
│   ├── user.go
│   ├── task.go
│   └── task_test.go
└── model/
    ├── user.go
    └── task.go
```

## 8. 開発ステップ

| ステップ | 内容 | 学べること |
| :--- | :--- | :--- |
| 1 | UserモデルとTaskモデルの定義（1対多リレーション） | GORMのAssociation、外部キー |
| 2 | `/auth/register` 実装 | bcryptハッシュ、ユーザー登録 |
| 3 | `/auth/login` 実装 | パスワード検証、JWTトークン生成 |
| 4 | JWT検証ミドルウェア実装 | Ginミドルウェア、コンテキスト操作 |
| 5 | タスクCRUD実装（既存コードをベースに拡張） | user_idによるフィルタリング |
| 6 | `PUT /tasks/:id` 実装 | GORM Save/Updates |
| 7 | ページネーション追加（`?page=1&limit=10`） | GORMのOffset/Limit |
| 8 | テスト追加 | テーブル駆動テスト |

## 9. 環境変数（.env）

```
JWT_SECRET=your-secret-key
DB_PATH=test.db
```
