package models

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// 環境変数からデータベース設定を取得
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("POSTGRES_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	if dbPass == "" {
		dbPass = "postgres"
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "messageboard_test"
	}

	dbURL := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("テストデータベースに接続できません: %v", err)
	}

	// 清理数据库
	_, err = db.Exec(`
		DROP TABLE IF EXISTS messages;
		DROP TABLE IF EXISTS users;
		
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE messages (
			id SERIAL PRIMARY KEY,
			title VARCHAR(20) NOT NULL,
			content TEXT NOT NULL,
			user_id INTEGER REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("テストデータベースの初期化に失敗しました: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func TestMessageStore_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewMessageStore(db)

	// 创建测试用户
	_, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES ($1, $2)`, "testuser", "testhash")
	if err != nil {
		t.Fatalf("テストユーザーの作成に失敗しました: %v", err)
	}

	// 测试创建消息
	err = store.Create("テストタイトル", "テスト内容", 1)
	if err != nil {
		t.Errorf("メッセージの作成に失敗しました: %v", err)
	}

	// 验证消息是否创建成功
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&count)
	if err != nil {
		t.Errorf("メッセージ数の取得に失敗しました: %v", err)
	}
	if count != 1 {
		t.Errorf("メッセージ数が1であるべきですが、実際は%dです", count)
	}
}

func TestMessageStore_Get(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewMessageStore(db)

	// 创建测试用户和消息
	_, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (1, 'testuser', 'testhash');
		INSERT INTO messages (id, title, content, user_id) VALUES (1, 'テストタイトル', 'テスト内容', 1);
	`)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// 测试获取消息
	msg, err := store.Get(1)
	if err != nil {
		t.Errorf("メッセージの取得に失敗しました: %v", err)
	}

	// 验证消息内容
	if msg.Title != "テストタイトル" {
		t.Errorf("タイトルが'テストタイトル'であるべきですが、実際は'%s'です", msg.Title)
	}
	if msg.Content != "テスト内容" {
		t.Errorf("内容が'テスト内容'であるべきですが、実際は'%s'です", msg.Content)
	}
	if msg.UserID != 1 {
		t.Errorf("ユーザーIDが1であるべきですが、実際は%dです", msg.UserID)
	}
}

func TestMessageStore_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewMessageStore(db)

	// 创建测试用户和多条消息
	_, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (1, 'testuser', 'testhash');
		INSERT INTO messages (title, content, user_id) VALUES 
			('タイトル1', '内容1', 1),
			('タイトル2', '内容2', 1),
			('タイトル3', '内容3', 1);
	`)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// 测试分页列表
	messages, total, err := store.List(1, 2)
	if err != nil {
		t.Errorf("メッセージリストの取得に失敗しました: %v", err)
	}

	// 验证结果
	if total != 3 {
		t.Errorf("総数が3であるべきですが、実際は%dです", total)
	}
	if len(messages) != 2 {
		t.Errorf("現在のページのメッセージ数が2であるべきですが、実際は%dです", len(messages))
	}
}

func TestMessageStore_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewMessageStore(db)

	// 创建测试用户和消息
	_, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (1, 'testuser', 'testhash');
		INSERT INTO messages (id, title, content, user_id) VALUES (1, '元のタイトル', '元の内容', 1);
	`)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// 测试更新消息
	err = store.Update(1, "新タイトル", "新内容", 1)
	if err != nil {
		t.Errorf("メッセージの更新に失敗しました: %v", err)
	}

	// 验证更新结果
	var title, content string
	err = db.QueryRow("SELECT title, content FROM messages WHERE id = 1").Scan(&title, &content)
	if err != nil {
		t.Errorf("更新後のメッセージの取得に失敗しました: %v", err)
	}
	if title != "新タイトル" {
		t.Errorf("タイトルが'新タイトル'であるべきですが、実際は'%s'です", title)
	}
	if content != "新内容" {
		t.Errorf("内容が'新内容'であるべきですが、実際は'%s'です", content)
	}
}

func TestMessageStore_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewMessageStore(db)

	// 创建测试用户和消息
	_, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (1, 'testuser', 'testhash');
		INSERT INTO messages (id, title, content, user_id) VALUES (1, 'テストタイトル', 'テスト内容', 1);
	`)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// 测试删除消息
	err = store.Delete(1, 1)
	if err != nil {
		t.Errorf("メッセージの削除に失敗しました: %v", err)
	}

	// 验证消息是否已删除
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM messages WHERE id = 1").Scan(&count)
	if err != nil {
		t.Errorf("メッセージの取得に失敗しました: %v", err)
	}
	if count != 0 {
		t.Errorf("メッセージが正常に削除されていません")
	}
}

func TestMessageStore_Search(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewMessageStore(db)

	// 创建测试用户和消息
	_, err := db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (1, 'testuser', 'testhash');
		INSERT INTO messages (title, content, user_id) VALUES 
			('テストタイトル1', 'テスト内容ABC', 1),
			('ABCタイトル2', 'テスト内容2', 1),
			('テストタイトル3', 'テスト内容3', 1);
	`)
	if err != nil {
		t.Fatalf("テストデータの作成に失敗しました: %v", err)
	}

	// 测试搜索消息
	messages, err := store.Search("ABC")
	if err != nil {
		t.Errorf("メッセージの検索に失敗しました: %v", err)
	}

	// 验证搜索结果
	if len(messages) != 2 {
		t.Errorf("検索結果の数が2であるべきですが、実際は%dです", len(messages))
	}
}
