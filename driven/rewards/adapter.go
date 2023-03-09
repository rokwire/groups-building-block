// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rewards

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Adapter implements the Rewards interface
type Adapter struct {
	internalAPIKey string
	rewardsHost    string
}

const (
	// GroupsUserCreatedFirstGroup User created its furst group
	GroupsUserCreatedFirstGroup = "groups_user_create_first_group"

	// GroupsUserSubmittedFirstPost User created its furst post
	GroupsUserSubmittedFirstPost = "groups_user_submitted_first_post"

	// GroupsUserSubmittedPost User created a post (after the first one)
	GroupsUserSubmittedPost = "groups_user_submitted_post"
)

// NewRewardsAdapter creates a new rewards adapter
func NewRewardsAdapter(host string, internalAPIKey string) *Adapter {
	if host != "" {
		return &Adapter{rewardsHost: host, internalAPIKey: internalAPIKey}
	}
	log.Fatal("Error: NewRewardsAdapter - not initialized core")
	return nil
}

// createRewardHistoryEntryBody wrapper
type createRewardHistoryEntryBody struct {
	UserID      string `json:"user_id"`
	RewardType  string `json:"reward_type"`
	Description string `json:"description"`
} // @name createRewardHistoryEntryBody

// CreateUserReward retrieves all members for a group
func (a *Adapter) CreateUserReward(userID string, rewardType string, description string) error {
	if userID != "" && rewardType != "" {

		requestBodyStruct := createRewardHistoryEntryBody{
			UserID:      userID,
			RewardType:  rewardType,
			Description: description,
		}
		reqBody, err := json.Marshal(requestBodyStruct)
		if err != nil {
			log.Printf("CreateUserReward: marshal request body - %s", err)
			return err
		}

		url := fmt.Sprintf("%s/api/int/reward_history", a.rewardsHost)
		client := &http.Client{}
		req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBody)))
		req.Header.Add("INTERNAL-API-KEY", a.internalAPIKey)
		if err != nil {
			log.Printf("CreateUserReward: error creating create reward request - %s", err)
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("CreateUserReward: error creating create reward request - %s", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			errorBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("CreateUserReward: unable to read json: %s", err)
				return fmt.Errorf("CreateUserReward: unable to parse json: %s", err)
			}

			log.Printf("CreateUserReward: error with response code - %d body: %s", resp.StatusCode, errorBody)
			return fmt.Errorf("CreateUserReward: error with response code - %d body: %s", resp.StatusCode, errorBody)
		}

		return nil
	}
	return nil
}
