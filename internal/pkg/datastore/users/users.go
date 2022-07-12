package users

import "time"

type User struct {
	id        int
	UID       string     `json:"uid"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}
