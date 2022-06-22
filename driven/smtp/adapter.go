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
