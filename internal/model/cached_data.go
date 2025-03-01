package model

import "time"

type CachedData struct {
	ID        int       `db:"id"`
	Created   time.Time `db:"created_at"`
	Updated   time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Encrypted bool      `db:"encrypted"`
	Value     []byte    `db:"value"`
}
