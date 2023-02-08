package users

import "time"

type User struct {
	id        int
	UID       string     `json:"uid"`
	Guild_ID  string     `json:"guild_id"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}
