package domain

type UserId int64

type User struct {
	UserId       UserId
	Username     string
	FirstName    string
	LastName     string
	LanguageCode string
	Timezone     string
	IsActive     bool
}
