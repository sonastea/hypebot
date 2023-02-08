package guilds

type Guild struct {
	id        int
	UID       string `json:"uid"`
	Active    int8   `json:"active"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}
