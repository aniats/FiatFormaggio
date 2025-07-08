package domain

type Message struct {
	ChatID   int64
	UserID   int64
	Username string
	Text     string
}
