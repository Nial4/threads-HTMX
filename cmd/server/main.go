package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"message-board/internal/handlers"
	"message-board/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	// 環境変数を使用してデータベース接続文字列を構築
	dbHost := os.Getenv("POSTGRES_HOST")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	dbURL := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName)

	// データベースに接続
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// データベース接続のテスト
	if err := db.Ping(); err != nil {
		log.Fatal("データベースに接続できません:", err)
	}

	// storeとhandlerの作成
	messageStore := models.NewMessageStore(db)
	userStore := models.NewUserStore(db)
	messageHandler := handlers.NewMessageHandler(messageStore)
	authHandler := handlers.NewAuthHandler(userStore)

	// Echoインスタンスの作成
	e := echo.New()

	// ミドルウェア
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 静的ファイル
	e.Static("/static", "static")

	// パブリックルート
	e.GET("/login", authHandler.ShowLoginPage)
	e.POST("/login", authHandler.Login)
	e.GET("/register", authHandler.ShowRegisterPage)
	e.POST("/register", authHandler.Register)

	// 認証が必要なルートグループ
	auth := e.Group("")
	auth.Use(handlers.JWTMiddleware)

	auth.GET("/", messageHandler.ListMessages)
	auth.POST("/messages", messageHandler.CreateMessage)
	auth.GET("/search", messageHandler.SearchMessages)
	auth.GET("/messages/:id", messageHandler.GetMessage)
	auth.POST("/messages/:id", messageHandler.UpdateMessage)
	auth.GET("/messages/:id/edit", messageHandler.EditMessage)
	auth.POST("/messages/:id/delete", messageHandler.DeleteMessage)
	auth.POST("/logout", authHandler.Logout)

	// サーバーの起動
	e.Logger.Fatal(e.Start(":8080"))
}
