package model

import "time"

// TelegramVerification represents a pending or completed Telegram verification.
type TelegramVerification struct {
	ID         int       `db:"id"`
	Username   string    `db:"username"`
	Token      string    `db:"token"`
	TelegramID int64     `db:"telegram_id"`
	Phone      string    `db:"phone"`
	Verified   bool      `db:"verified"`
	CreatedAt  time.Time `db:"created_at"`
}

// TelegramDevice represents a registered Telegram device for a user.
type TelegramDevice struct {
	ID         int       `db:"id"`
	Username   string    `db:"username"`
	TelegramID int64     `db:"telegram_id"`
	Phone      string    `db:"phone"`
	ChatID     int64     `db:"chat_id"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	BotUsername string   `db:"bot_username"`
	CreatedAt  time.Time `db:"created_at"`
	LastUsedAt time.Time `db:"last_used_at"`
}
