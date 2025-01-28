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

package socialbb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"groups/core/model"
	"io"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v2/authservice"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// Adapter implements the SocialBB interface
type Adapter struct {
	socialURL             string
	serviceAccountManager *authservice.ServiceAccountManager
	logger                logs.Logger
}

// NewSocialAdapter creates a new adapter for Core API
func NewSocialAdapter(socialURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{socialURL: socialURL, serviceAccountManager: serviceAccountManager}
}

// GetPosts retrieves posts from the social BB
func (a *Adapter) GetPosts(clientID string, current *model.User, filter model.PostsFilter, filterPrivatePostsValue *bool, filterByToMembers bool) ([]model.Post, error) {
	result, err := a.invokePostsOperation("get_posts", current, map[string]interface{}{
		"client_id":            clientID,
		"current_user":         current,
		"filter":               filter,
		"filter_private_posts": filterPrivatePostsValue,
		"filter_by_to_members": filterByToMembers,
	})
	if err != nil {
		a.logger.Errorf("social.getPosts: error invoking posts operation - %s", err)
	}

	type postsResponse struct {
		Posts []model.Post `json:"posts"`
		Error error        `json:"error"`
	}

	var posts postsResponse
	err = json.Unmarshal(result, &posts)
	if err != nil {
		a.logger.Errorf("social.getPosts: error unmarshalling posts - %s", err)
		return nil, err
	}

	return posts.Posts, posts.Error
}

func (a *Adapter) getPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	return nil, nil
}

func (a *Adapter) getUserPostCount(clientID string, userID string) (*int64, error) {
	return nil, nil
}

func (a *Adapter) createPost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {
	return nil, nil
}

func (a *Adapter) updatePost(clientID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error) {
	return nil, nil
}

func (a *Adapter) reactToPost(clientID string, current *model.User, groupID string, postID string, reaction string) error {
	return nil
}

func (a *Adapter) reportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {
	return nil
}

func (a *Adapter) deletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	return nil
}

// InvokePostsOperation invokes the posts operation
func (a *Adapter) invokePostsOperation(operation string, current *model.User, data map[string]interface{}) ([]byte, error) {

	if a.serviceAccountManager == nil {
		a.logger.Errorf("InvokePostsOperation: service account manager is nil")
		return nil, errors.New("InvokePostsOperation: service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/posts-proxy", a.socialURL)

	bodyMap := map[string]interface{}{
		"operation": operation,
		"data":      data,
	}

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		a.logger.Errorf("InvokePostsOperation: error marshalling body - %s", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		a.logger.Errorf("InvokePostsOperation: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, current.AppID, current.OrgID)
	if err != nil {
		a.logger.Errorf("InvokePostsOperation: error sending request - %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		a.logger.Errorf("InvokePostsOperation: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("InvokePostsOperation: error with response code != 200")
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		a.logger.Errorf("InvokePostsOperation: unable to read json: %s", err)
		return nil, fmt.Errorf("InvokePostsOperation: unable to parse json: %s", err)
	}

	return responseData, nil
}
