package domain

func GetDomains() []any {
	return []any{
		&User{},
		&Token{},
	}
}
