package corebb

import (
	"encoding/json"
	"fmt"
	"groups/core/model"
	"io/ioutil"
	"log"
	"net/http"
)

// Adapter implements the Storage interface
type Adapter struct {
	coreURL string
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string) *Adapter {
	return &Adapter{coreURL: coreURL}
}

// RetrieveCoreUserAccount retrieves Core user account
func (a *Adapter) RetrieveCoreUserAccount(token string) (*model.CoreAccount, error) {
	if len(token) > 0 {
		url := fmt.Sprintf("%s/services/account", token)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: error creating load user data request - %s", err)
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: error loading user data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveCoreUserAccount: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: error with response code != 200")
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: unable to parse json: %s", err)
		}

		var coreAccount model.CoreAccount
		err = json.Unmarshal(data, &coreAccount)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		return &coreAccount, nil
	}
	return nil, nil
}
