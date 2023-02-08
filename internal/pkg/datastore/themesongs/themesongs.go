package themesongs

type Themesong struct {
	id       int
	Guild_ID string `json:"guild_id"`
	User_ID  string `json:"user_id"`
	Filepath string `json:"filepath"`
}
