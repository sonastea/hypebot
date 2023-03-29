package guilds

type Guild struct {
	id        int
	UID       string `json:"uid"`
	VCS       map[string][]string
	Playing   bool
	Active    int8   `json:"active"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}
