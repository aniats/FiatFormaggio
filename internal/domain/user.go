package domain

type UserID int64

type User struct {
	UserID    UserID
	Username  string
	FirstName string
	LastName  string
}
