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

// CoreService wrapper record for the corresponding service record response
type CoreService struct {
	Host      string `json:"host"`
	ServiceID string `json:"service_id"`
}

// Creator represents group member entity
type Creator struct {
	UserID string `json:"user_id" bson:"user_id" validate:"required"`
	Name   string `json:"name" bson:"name" validate:"required"`
	Email  string `json:"email" bson:"email" validate:"required"`
} //@name Creator
