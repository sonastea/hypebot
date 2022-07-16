package themesongs

type Themesong struct {
	id       int
	User_ID  string `json:"user_id"`
	Filepath string `json:"filepath"`
}
