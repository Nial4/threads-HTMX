package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"message-board/internal/models"

	"github.com/flosch/pongo2/v6"
	"github.com/labstack/echo/v4"
)

const (
	perPage    = 5
	maxMsgSize = 20
)

type MessageHandler struct {
	store *models.MessageStore
}

func NewMessageHandler(store *models.MessageStore) *MessageHandler {
	return &MessageHandler{store: store}
}

func (h *MessageHandler) ListMessages(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	messages, total, err := h.store.List(page, perPage)
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "システムエラー",
			"error_message": "メッセージの取得中にエラーが発生しました。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	totalPages := (total + perPage - 1) / perPage

	tpl := pongo2.Must(pongo2.FromFile("templates/index.html"))
	return tpl.ExecuteWriter(pongo2.Context{
		"messages":    messages,
		"page":        page,
		"total_pages": totalPages,
		"has_prev":    page > 1,
		"has_next":    page < totalPages,
		"user_id":     c.Get("user_id").(int),
		"username":    c.Get("username").(string),
	}, c.Response().Writer)
}

func (h *MessageHandler) GetMessage(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	message, err := h.store.Get(id)
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "メッセージが見つかりません",
			"error_message": "指定されたメッセージは存在しません。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	// 現在のユーザーIDを取得
	userID := c.Get("user_id").(int)

	tpl := pongo2.Must(pongo2.FromFile("templates/detail.html"))
	return tpl.ExecuteWriter(pongo2.Context{
		"message": message,
		"user_id": userID,
	}, c.Response().Writer)
}

func (h *MessageHandler) CreateMessage(c echo.Context) error {
	title := strings.TrimSpace(c.FormValue("title"))
	content := strings.TrimSpace(c.FormValue("content"))

	if len(title) == 0 || len(content) == 0 {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "入力エラー",
			"error_message": "タイトルと内容は必須です。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	if len(title) > maxMsgSize || len(content) > maxMsgSize {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "入力エラー",
			"error_message": "タイトルと内容は20文字以内で入力してください。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	// 現在のユーザーIDを取得
	userID := c.Get("user_id").(int)

	err := h.store.Create(title, content, userID)
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "システムエラー",
			"error_message": "メッセージの作成中にエラーが発生しました。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *MessageHandler) DeleteMessage(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	// 現在のユーザーIDを取得
	userID := c.Get("user_id").(int)

	err := h.store.Delete(id, userID)
	if err != nil {
		if err.Error() == "unauthorized: message belongs to another user" {
			tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
			return tpl.ExecuteWriter(pongo2.Context{
				"error_title":   "権限エラー",
				"error_message": "自分のメッセージのみ削除できます。",
				"back_url":      "/",
			}, c.Response().Writer)
		}
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "システムエラー",
			"error_message": "メッセージの削除中にエラーが発生しました。",
			"back_url":      "/",
		}, c.Response().Writer)
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *MessageHandler) EditMessage(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	message, err := h.store.Get(id)
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "メッセージが見つかりません",
			"error_message": "指定されたメッセージは存在しません。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	// 現在のユーザーIDを取得し、権限をチェック
	userID := c.Get("user_id").(int)
	if message.UserID != userID {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "権限エラー",
			"error_message": "自分のメッセージのみ編集できます。",
			"back_url":      "/messages/" + strconv.Itoa(id),
		}, c.Response().Writer)
	}

	tpl := pongo2.Must(pongo2.FromFile("templates/edit.html"))
	return tpl.ExecuteWriter(pongo2.Context{
		"message": message,
	}, c.Response().Writer)
}

func (h *MessageHandler) UpdateMessage(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	title := strings.TrimSpace(c.FormValue("title"))
	content := strings.TrimSpace(c.FormValue("content"))

	fmt.Println(title, content, id)

	if len(title) == 0 || len(content) == 0 {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "入力エラー",
			"error_message": "タイトルと内容は必須です。",
			"back_url":      "/messages/" + strconv.Itoa(id) + "/edit",
		}, c.Response().Writer)
	}

	if len(title) > maxMsgSize || len(content) > maxMsgSize {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "入力エラー",
			"error_message": "タイトルと内容は20文字以内で入力してください。",
			"back_url":      "/messages/" + strconv.Itoa(id) + "/edit",
		}, c.Response().Writer)
	}

	// 現在のユーザーIDを取得
	userID := c.Get("user_id").(int)

	err := h.store.Update(id, title, content, userID)
	if err != nil {
		if err.Error() == "unauthorized: message belongs to another user" {
			tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
			return tpl.ExecuteWriter(pongo2.Context{
				"error_title":   "権限エラー",
				"error_message": "自分のメッセージのみ編集できます。",
				"back_url":      "/messages/" + strconv.Itoa(id),
			}, c.Response().Writer)
		}
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "システムエラー",
			"error_message": "メッセージの更新中にエラーが発生しました。",
			"back_url":      "/messages/" + strconv.Itoa(id),
		}, c.Response().Writer)
	}

	return c.Redirect(http.StatusSeeOther, "/messages/"+strconv.Itoa(id))
}

func (h *MessageHandler) SearchMessages(c echo.Context) error {
	query := strings.TrimSpace(c.QueryParam("q"))
	if query == "" {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	messages, err := h.store.Search(query)
	if err != nil {
		tpl := pongo2.Must(pongo2.FromFile("templates/error.html"))
		return tpl.ExecuteWriter(pongo2.Context{
			"error_title":   "システムエラー",
			"error_message": "メッセージの検索中にエラーが発生しました。",
			"back_url":      "/",
		}, c.Response().Writer)
	}

	// 現在のユーザーIDを取得
	userID := c.Get("user_id").(int)

	tpl := pongo2.Must(pongo2.FromFile("templates/index.html"))
	return tpl.ExecuteWriter(pongo2.Context{
		"messages": messages,
		"query":    query,
		"user_id":  userID,
	}, c.Response().Writer)
}
