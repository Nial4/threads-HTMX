package models

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestUserStore_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)

	// ユーザー作成のテスト
	err := store.Create("testuser", "password123")
	if err != nil {
		t.Errorf("ユーザー作成に失敗しました: %v", err)
	}

	// ユーザーが正しく作成されたか確認
	var username, passwordHash string
	err = db.QueryRow("SELECT username, password_hash FROM users WHERE username = $1", "testuser").Scan(&username, &passwordHash)
	if err != nil {
		t.Errorf("ユーザー検索に失敗しました: %v", err)
	}

	// ユーザー名の確認
	if username != "testuser" {
		t.Errorf("ユーザー名が'testuser'であるべきですが、実際は'%s'です", username)
	}

	// パスワードハッシュの確認
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password123"))
	if err != nil {
		t.Errorf("パスワードハッシュの検証に失敗しました: %v", err)
	}
}

func TestUserStore_GetByUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)

	// テストユーザーの作成
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err := db.Exec(`
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2)
	`, "testuser", string(hashedPassword))
	if err != nil {
		t.Fatalf("テストユーザーの作成に失敗しました: %v", err)
	}

	// ユーザー取得のテスト
	user, err := store.GetByUsername("testuser")
	if err != nil {
		t.Errorf("ユーザー取得に失敗しました: %v", err)
	}

	// ユーザー情報の確認
	if user.Username != "testuser" {
		t.Errorf("ユーザー名が'testuser'であるべきですが、実際は'%s'です", user.Username)
	}
}

func TestUserStore_GetByUsername_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)

	// 存在しないユーザーの取得テスト
	_, err := store.GetByUsername("nonexistent")
	if err == nil {
		t.Error("ユーザーが存在しない場合にエラーを返すことを期待しますが、そうではありません")
	}
	if err.Error() != "user not found" {
		t.Errorf("ユーザーが存在しない場合にエラーを返すことを期待しますが、実際は'%s'です", err.Error())
	}
}

func TestUserStore_Authenticate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)

	// テストユーザーの作成
	err := store.Create("testuser", "password123")
	if err != nil {
		t.Fatalf("テストユーザーの作成に失敗しました: %v", err)
	}

	// ユーザーの正しいパスワード認証のテスト
	user, err := store.Authenticate("testuser", "password123")
	if err != nil {
		t.Errorf("認証に失敗しました: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("ユーザー名が'testuser'であるべきですが、実際は'%s'です", user.Username)
	}

	// ユーザーの誤ったパスワード認証のテスト
	_, err = store.Authenticate("testuser", "wrongpassword")
	if err == nil {
		t.Error("誤ったパスワードを使用した場合にエラーを返すことを期待しますが、そうではありません")
	}
	if err.Error() != "invalid password" {
		t.Errorf("誤ったパスワードを使用した場合にエラーを返すことを期待しますが、実際は'%s'です", err.Error())
	}

	// 存在しないユーザーの認証のテスト
	_, err = store.Authenticate("nonexistent", "password123")
	if err == nil {
		t.Error("存在しないユーザーを認証しようとした場合にエラーを返すことを期待しますが、そうではありません")
	}
	if err.Error() != "user not found" {
		t.Errorf("存在しないユーザーを認証しようとした場合にエラーを返すことを期待しますが、実際は'%s'です", err.Error())
	}
}
