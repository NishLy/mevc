package domain

func GetDomains() []any {
	return []any{
		&User{},
		&Token{},
		&Room{},
		&Schedule{},
		&Occurance{},
		&Shortener{},
	}
}
