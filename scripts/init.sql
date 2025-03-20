-- 既存のテーブルを削除（存在する場合）
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS users;

-- ユーザーテーブルの作成
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- メッセージテーブルの作成
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    title VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- テストユーザーの作成 (パスワード: 123456)
INSERT INTO users (username, password_hash) VALUES
    ('test', '$2a$10$pbJoSem7uzmXHYJttuL.vuS8IH268ekADVftesFZfcJl6LiFcTe7K');

-- サンプルデータの挿入
INSERT INTO messages (title, content, user_id) VALUES
    ('スレッドボードへようこそ', '最初のメッセージです', 1),
    ('こんにちは世界', 'Hello World!', 1),
    ('テストメッセージ', 'これはテストメッセージです', 1),
    ('ご挨拶', '皆様、今日も良い一日を', 1),
    ('お知らせ', 'これは重要な通知です', 1); 