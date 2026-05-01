package users

type Entity struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}
