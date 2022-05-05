package polls

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Adapter implements the Storage interface
type Adapter struct {
	internalAPIKey string
	baseURL        string
}

// NewPollsAdapter creates a new Polls V2 BB adapter instance
func NewPollsAdapter(internalAPIKey string, baseURL string) *Adapter {
	return &Adapter{internalAPIKey: internalAPIKey, baseURL: baseURL}
}

// GetPollTroGroupMapping gets poll to group mapping
func (na *Adapter) GetPollTroGroupMapping() (map[string]string, error) {
	url := fmt.Sprintf("%s/api/int/poll-to-group-mapping", na.baseURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error polls.SendNotification() - error creating poll group mapping request - %s", err)
		return nil, err
	}
	req.Header.Set("INTERNAL-API-KEY", na.internalAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error polls.SendNotification() - error loading  poll group mapping data - %s", err)
		return nil, fmt.Errorf("error polls.SendNotification() - error loading  poll group mapping data - %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Error polls.SendNotification() - error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("error polls.SendNotification() - error with response code - %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("RetrieveAuthmanGroupMembersError: unable to read json: %s", err)
		return nil, fmt.Errorf("error polls.SendNotification() - unable to read json - %d", err)
	}

	var mapping map[string]string
	err = json.Unmarshal(data, &mapping)
	if err != nil {
		log.Printf("error polls.SendNotification() - unable to parse json - %d", err)
		return nil, fmt.Errorf("error polls.SendNotification() - unable to parse json - %d", err)
	}

	return mapping, nil
}
