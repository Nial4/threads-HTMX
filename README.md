# スレッドボード(スレッド)アプリケーション  

## 機能  

0. ユーザー登録とログイン  
1. スレッドを登録、取得、削除できる  
    スレッドは一覧表示できる  
    スレッドのタイトルをリストで表示する  
2. スレッドは詳細表示できる   
    スレッドにたいする投稿を閲覧できる  
3. スレッドにたいして投稿の作成、編集、削除できる  
    編集、削除は自身の投稿のみ可能  
4. スレッドおよび投稿を検索できる  
    全文検索によりスレッドおよび投稿の検索ができる  
5. ページネーションに対応すること  
6. 適切なValidationの処理を実装すること  

## 技術スタック

- バックエンド: Go
- フロントエンド: Pongo2
- データベース: PostgreSQL
- 認証: JWT

## 必要条件

- Go 1.23
- PostgreSQL 16
- Docker

## セットアップ方法

1. リポジトリをクローン：
```bash
git clone このリポジトリ
cd threads-HTMX
```

2. 依存関係のインストール：
```bash
go mod tidy
```

## アプリケーションの起動

```bash
docker-compose down -v
docker-compose up --build
```

アプリケーションは`http://localhost:8080`で起動します。

## テスト方法

```bash
docker-compose -f docker-compose.test.yml down
docker-compose -f docker-compose.test.yml up --build
```

## テストユーザー

- ユーザー名: test
- パスワード: 123456

