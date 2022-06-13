package rewards

import (
	"log"
)

// Adapter implements the Storage interface
type Adapter struct {
	internalAPIKey string
	rewardsHost    string
}

// NewRewardsAdapter creates a new rewards adapter
func NewRewardsAdapter(host string, internalAPIKey string) *Adapter {
	if host != "" {
		return &Adapter{rewardsHost: host, internalAPIKey: internalAPIKey}
	}
	log.Fatal("Error: NewRewardsAdapter - not initialized")
	return nil
}

// SendEmail Sends a transactional email
func (a *Adapter) SendEmail(to string, subject string, body string) error {
	return nil
}
