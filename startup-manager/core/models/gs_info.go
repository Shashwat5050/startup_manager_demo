package models

import "time"

type GameServerInfo struct {
	ID         string     `db:"id"`
	UserID     string     `db:"user_id"`
	ServerName string     `db:"server_name"`
	GameName   string     `db:"game_name"`
	Image      string     `db:"image"`
	Command    string     `db:"command"`
	Status     string     `db:"status"`
	CreatedAt  *time.Time `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}
