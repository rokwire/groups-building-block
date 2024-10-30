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

package model

import "time"

// MembershipMultiUpdate Defines multi record update for memberships
type MembershipMultiUpdate struct {
	UserIDs      []string   `json:"user_ids"`      // list of user ids
	Status       *string    `json:"status"`        // new status
	Reason       *string    `json:"reason"`        // reason
	DateAttended *time.Time `json:"date_attended"` // date attended
} //@name MembershipMultiUpdate

// IsStatusValid Checks if the status is valid
func (m MembershipMultiUpdate) IsStatusValid() bool {
	return m.Status == nil || (*m.Status == "pending" || *m.Status == "member" || *m.Status == "admin" || *m.Status == "rejected")
}
