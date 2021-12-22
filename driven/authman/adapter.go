package authman

// Adapter implements the Storage interface
type Adapter struct {
	authmanURL      string
	authmanUsername string
	authmanPassword string
}

// NewAuthmanAdapter creates a new adapter for Authman API
func NewAuthmanAdapter(authmanURL string, authmanUsername string, authmanPassword string) *Adapter {
	return &Adapter{authmanURL: authmanURL, authmanUsername: authmanUsername, authmanPassword: authmanPassword}
}
