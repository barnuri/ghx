package ghuser

type account struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type emailAddress struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}
