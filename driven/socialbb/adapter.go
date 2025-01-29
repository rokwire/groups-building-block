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
	result, err := a.invokePostsOperation("get_posts", &current.ID, &filter.GroupID, map[string]interface{}{
		"filter":               filter,
		"filter_private_posts": filterPrivatePostsValue,
		"filter_by_to_members": filterByToMembers,
	})
	if err != nil {
		a.logger.Errorf("social.GetPosts: error invoking posts operation - %s", err)
		return nil, err
	}

	type postsResponse struct {
		Posts []model.Post `json:"posts"`
		Error error        `json:"error"`
	}

	var posts postsResponse
	err = json.Unmarshal(result, &posts)
	if err != nil {
		a.logger.Errorf("social.GetPosts: error unmarshalling posts - %s", err)
		return nil, err
	}

	return posts.Posts, posts.Error
}

// GetPost retrieves a post from the social BB
func (a *Adapter) GetPost(clientID string, userID *string, groupID string, postID string, skipMembershipCheck bool, filterByToMembers bool) (*model.Post, error) {
	result, err := a.invokePostsOperation("get_post", userID, &groupID, map[string]interface{}{
		"post_id":               postID,
		"skip_membership_check": skipMembershipCheck,
		"filter_by_to_members":  filterByToMembers,
	})
	if err != nil {
		a.logger.Errorf("social.GetPost: error invoking posts operation - %s", err)
		return nil, err
	}

	type postsResponse struct {
		Post  model.Post `json:"post"`
		Error error      `json:"error"`
	}

	var post postsResponse
	err = json.Unmarshal(result, &post)
	if err != nil {
		a.logger.Errorf("social.GetPost: error unmarshalling posts - %s", err)
		return nil, err
	}

	return &post.Post, post.Error
}

// GetUserPostCount retrieves the number of posts for a user
func (a *Adapter) GetUserPostCount(clientID string, userID string) (*int64, error) {
	result, err := a.invokePostsOperation("get_user_post_count", &userID, nil, map[string]interface{}{})
	if err != nil {
		a.logger.Errorf("social.GetUserPostCount: error invoking posts operation - %s", err)
		return nil, err
	}

	type responseData struct {
		Count *int64 `json:"count"`
		Error error  `json:"error"`
	}

	var response responseData
	err = json.Unmarshal(result, &response)
	if err != nil {
		a.logger.Errorf("social.GetUserPostCount: error unmarshalling posts - %s", err)
		return nil, err
	}

	return response.Count, response.Error
}

// CreatePost creates a post
func (a *Adapter) CreatePost(clientID string, current *model.User, post *model.Post, group *model.Group) (*model.Post, error) {
	result, err := a.invokePostsOperation("create_post", &current.ID, &post.GroupID, map[string]interface{}{
		"post": post,
	})
	if err != nil {
		a.logger.Errorf("social.CreatePost: error invoking posts operation - %s", err)
		return nil, err
	}

	type responseData struct {
		Post  *model.Post `json:"post"`
		Error error       `json:"error"`
	}

	var response responseData
	err = json.Unmarshal(result, &response)
	if err != nil {
		a.logger.Errorf("social.CreatePost: error unmarshalling posts - %s", err)
		return nil, err
	}

	return response.Post, response.Error
}

// UpdatePost updates a post
func (a *Adapter) UpdatePost(clientID string, current *model.User, group *model.Group, post *model.Post) (*model.Post, error) {
	result, err := a.invokePostsOperation("update_post", &current.ID, &post.GroupID, map[string]interface{}{
		"post": post,
	})
	if err != nil {
		a.logger.Errorf("social.UpdatePost: error invoking posts operation - %s", err)
		return nil, err
	}

	type responseData struct {
		Post  *model.Post `json:"post"`
		Error error       `json:"error"`
	}

	var response responseData
	err = json.Unmarshal(result, &response)
	if err != nil {
		a.logger.Errorf("social.UpdatePost: error unmarshalling posts - %s", err)
		return nil, err
	}

	return response.Post, response.Error
}

// ReactToPost reacts to a post
func (a *Adapter) ReactToPost(clientID string, current *model.User, groupID string, postID string, reaction string) error {
	result, err := a.invokePostsOperation("react_to_post", &current.ID, &groupID, map[string]interface{}{
		"post_id":  postID,
		"reaction": reaction,
	})
	if err != nil {
		a.logger.Errorf("social.ReactToPost: error invoking posts operation - %s", err)
		return err
	}

	type responseData struct {
		Error error `json:"error"`
	}

	var response responseData
	err = json.Unmarshal(result, &response)
	if err != nil {
		a.logger.Errorf("social.ReactToPost: error unmarshalling posts - %s", err)
		return err
	}

	return response.Error
}

// ReportPostAsAbuse reports a post as abuse
func (a *Adapter) ReportPostAsAbuse(clientID string, current *model.User, group *model.Group, post *model.Post, comment string, sendToDean bool, sendToGroupAdmins bool) error {
	result, err := a.invokePostsOperation("report_abuse_post", &current.ID, &post.GroupID, map[string]interface{}{
		"post_id":              post.ID,
		"comment":              comment,
		"send_to_dean":         sendToDean,
		"send_to_group_admins": sendToGroupAdmins,
	})
	if err != nil {
		a.logger.Errorf("social.ReportPostAsAbuse: error invoking posts operation - %s", err)
		return err
	}

	type responseData struct {
		Error error `json:"error"`
	}

	var response responseData
	err = json.Unmarshal(result, &response)
	if err != nil {
		a.logger.Errorf("social.ReportPostAsAbuse: error unmarshalling posts - %s", err)
		return err
	}

	return response.Error
}

// DeletePost deletes a post
func (a *Adapter) DeletePost(clientID string, userID string, groupID string, postID string, force bool) error {
	result, err := a.invokePostsOperation("delete_post", &userID, &groupID, map[string]interface{}{
		"post_id": postID,
		"force":   force,
	})
	if err != nil {
		a.logger.Errorf("social.DeletePost: error invoking posts operation - %s", err)
		return err
	}

	type responseData struct {
		Error error `json:"error"`
	}

	var response responseData
	err = json.Unmarshal(result, &response)
	if err != nil {
		a.logger.Errorf("social.DeletePost: error unmarshalling posts - %s", err)
		return err
	}

	return response.Error
}

// InvokePostsOperation invokes the posts operation
func (a *Adapter) invokePostsOperation(operation string, userID *string, groupID *string, data map[string]interface{}) ([]byte, error) {

	if a.serviceAccountManager == nil {
		a.logger.Errorf("InvokePostsOperation: service account manager is nil")
		return nil, errors.New("InvokePostsOperation: service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/legacy-proxy", a.socialURL)

	bodyMap := map[string]interface{}{
		"operation":          operation,
		"current_account_id": userID,
		"group_id":           groupID,
		"data":               data,
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

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
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
