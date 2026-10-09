# CLAUDE.md

このファイルは `todo-api` リポジトリでの Claude Code との協働ルールを定義する。

## プロジェクト概要

- `todo-api`: Go + Gin + GORM + SQLite によるTodo API
- フェーズ1（基本CRUD）完了、フェーズ2（JWT認証 + Service層導入）進行中
- 現在進めている設計は [Docs/service_layer_design.md](Docs/service_layer_design.md) を参照
- レイヤー構成: `handler` → `service` → `repository` → `model`

## 協働スタイル：Claude が実装＋解説

写経はやめ、Claude が直接ファイルを編集して実装する（2026-10-09 に方針変更）。
学習の主眼は「設計を読んで判断する」「動かして運用する」ことに置く。

- コミットできる最小単位ごとに実装し、各単位の後に「何をしたか・なぜそうしたか」を簡潔に解説する。
- コミットはユーザーの依頼があるまで行わない。
- 設計上の判断が生じたら、選択肢と理由を提示し、`dev-log/設計判断ログ.md` への記録を提案する。

## 今後のロードマップ

1. Service層の完成（[Docs/service_layer_design.md](Docs/service_layer_design.md)）
2. SQLite → PostgreSQL 移行（まずDockerでローカル、設定は環境変数化）
3. Dockerfile 作成、Google Cloud (Cloud Run) へデプロイ
4. 動作確認用の簡易Web UI（ログイン・Todo CRUD）

## dev-log の記録ルール

`dev-log/` 全体の運用ルールは [dev-log/README.md](dev-log/README.md) を参照。以下は追加ルール:

- ユーザーが自分の言葉で解釈・説明した内容（技術用語の理解、疑問点など）は、
  要約せず `dev-log/daily/` にそのまま転記する。タイポ修正のみ可、言い換えや圧縮はしない。
- セッション終了時は `/wrapup` スキルを実行し、振り返りを `dev-log/daily/` に記録する。
  （`dev-log/README.md` が参照している「セッション終了時の要約」はこれに対応する）
- 設計上の判断（なぜその設計にしたか）は `dev-log/設計判断ログ.md` に蓄積する。

## 技術スタック

- Web: Gin
- ORM: GORM (SQLite)
- 認証: JWT (`golang-jwt/jwt/v5`) + bcrypt
