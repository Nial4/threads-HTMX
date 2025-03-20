package models

import (
	"database/sql"
	"errors"
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageStore struct {
	db *sql.DB
}

func NewMessageStore(db *sql.DB) *MessageStore {
	return &MessageStore{db: db}
}

func (s *MessageStore) List(page, perPage int) ([]Message, int, error) {
	// 総数を取得
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := s.db.Query(`
		SELECT m.id, m.title, m.content, m.user_id, u.username, m.created_at, m.updated_at 
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		ORDER BY m.created_at DESC 
		LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Title, &m.Content, &m.UserID, &m.Username, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		messages = append(messages, m)
	}
	return messages, total, nil
}

func (s *MessageStore) Get(id int) (*Message, error) {
	var m Message
	err := s.db.QueryRow(`
		SELECT m.id, m.title, m.content, m.user_id, u.username, m.created_at, m.updated_at 
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.id = $1`, id).Scan(&m.ID, &m.Title, &m.Content, &m.UserID, &m.Username, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("message not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MessageStore) Create(title, content string, userID int) error {
	_, err := s.db.Exec(`
		INSERT INTO messages (title, content, user_id) 
		VALUES ($1, $2, $3)`, title, content, userID)
	return err
}

func (s *MessageStore) Update(id int, title, content string, userID int) error {
	// まずメッセージがそのユーザーに属しているかチェック
	var messageUserID int
	err := s.db.QueryRow("SELECT user_id FROM messages WHERE id = $1", id).Scan(&messageUserID)
	if err != nil {
		return err
	}
	if messageUserID != userID {
		return errors.New("unauthorized: message belongs to another user")
	}

	result, err := s.db.Exec(`
		UPDATE messages 
		SET title = $1, content = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 AND user_id = $4`, title, content, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (s *MessageStore) Delete(id int, userID int) error {
	// まずメッセージがそのユーザーに属しているかチェック
	var messageUserID int
	err := s.db.QueryRow("SELECT user_id FROM messages WHERE id = $1", id).Scan(&messageUserID)
	if err != nil {
		return err
	}
	if messageUserID != userID {
		return errors.New("unauthorized: message belongs to another user")
	}

	result, err := s.db.Exec(`DELETE FROM messages WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (s *MessageStore) Search(query string) ([]Message, error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.title, m.content, m.user_id, u.username, m.created_at, m.updated_at 
		FROM messages m
		LEFT JOIN users u ON m.user_id = u.id
		WHERE m.title ILIKE $1 OR m.content ILIKE $1 
		ORDER BY m.created_at DESC`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Title, &m.Content, &m.UserID, &m.Username, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
