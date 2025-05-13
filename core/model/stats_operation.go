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

// ExternalOperation represents the type of external operation
type ExternalOperation string

const (
	// ExternalOperationEventUpdate constant
	ExternalOperationEventUpdate ExternalOperation = "event_update"

	// ExternalOperationSocialUpdate constant
	ExternalOperationSocialUpdate ExternalOperation = "social_update"

	// ExternalOperationPollUpdate constant
	ExternalOperationPollUpdate ExternalOperation = "poll_update"
)
