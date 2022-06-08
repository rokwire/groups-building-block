package polls

// Adapter implements the Polls interface
type Adapter struct {
	internalAPIKey string
	baseURL        string
}

// NewPollsAdapter creates a new Polls V2 BB adapter instance
func NewPollsAdapter(internalAPIKey string, baseURL string) *Adapter {
	return &Adapter{internalAPIKey: internalAPIKey, baseURL: baseURL}
}

// Keep alive for possible use in the future
