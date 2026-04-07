package jwt

// Claims is a placeholder type for the auth phase.
type Claims struct {
	UserID    string
	CompanyID string
	Role      string
	Kind      string
}
