package handlers

import (
	"message-board/internal/models"
	"net/http"
	"os"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	userStore *models.UserStore
}

func NewAuthHandler(userStore *models.UserStore) *AuthHandler {
	return &AuthHandler{userStore: userStore}
}

func (h *AuthHandler) ShowLoginPage(c echo.Context) error {
	tpl := pongo2.Must(pongo2.FromFile("templates/login.html"))
	return tpl.ExecuteWriter(pongo2.Context{}, c.Response().Writer)
}

func (h *AuthHandler) ShowRegisterPage(c echo.Context) error {
	tpl := pongo2.Must(pongo2.FromFile("templates/register.html"))
	return tpl.ExecuteWriter(pongo2.Context{}, c.Response().Writer)
}

func (h *AuthHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	user, err := h.userStore.Authenticate(username, password)
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "ログインエラー",
			"error_message": "ユーザー名またはパスワードが正しくありません。",
			"back_url":      "/login",
		}, c.Response().Writer)
	}

	// JWTトークンの作成
	token := jwt.New(jwt.SigningMethodHS256)

	// クレームの設定
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// トークンの生成
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "システムエラー",
			"error_message": "ログイン処理中にエラーが発生しました。",
			"back_url":      "/login",
		}, c.Response().Writer)
	}

	// クッキーの設定
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = t
	cookie.Expires = time.Now().Add(72 * time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) Register(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if err := h.userStore.Create(username, password); err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "登録エラー",
			"error_message": "ユーザー登録に失敗しました。ユーザー名が既に使用されている可能性があります。",
			"back_url":      "/register",
		}, c.Response().Writer)
	}

	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-time.Hour)
	cookie.Path = "/"
	c.SetCookie(cookie)

	return c.Redirect(http.StatusSeeOther, "/login")
}

// JWTミドルウェア
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// クッキーからトークンを取得
		cookie, err := c.Cookie("token")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// トークンの解析
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// ユーザー情報をコンテキストに保存
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Set("username", claims["username"].(string))

		return next(c)
	}
}
