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

package core

import (
	"fmt"
	"groups/core/model"
	"groups/driven/storage"
	"log"
	"time"
)

func (app Application) processScheduledPosts() error {

	log.Printf("processScheduledPosts:BEGIN")
	defer log.Printf("processScheduledPosts:END")

	startTime := time.Now()
	syncKey := "scheduled_posts"
	transaction := func(context storage.TransactionContext) error {
		times, err := app.storage.FindSyncTimes(context, "", "scheduled_posts", false)
		if err != nil {
			return err
		}
		if times != nil && times.StartTime != nil {
			if times.EndTime == nil {
				if !startTime.After(times.StartTime.Add(time.Second * time.Duration(60))) {
					log.Println("Another schduled post task process is running for clientID ")
					return fmt.Errorf("another schduled post task  process is running")
				}
				log.Printf("schduled post task past timeout threshold %d\n", 60)
			}
		}

		err = app.storage.SaveSyncTimes(context, model.SyncTimes{StartTime: &startTime, EndTime: nil, Key: syncKey})
		if err != nil {
			return err
		}

		// Finish task
		endTime := time.Now()
		err = app.storage.SaveSyncTimes(context, model.SyncTimes{StartTime: &startTime, EndTime: &endTime, Key: syncKey})
		if err != nil {
			return err
		}

		return nil
	}

	err := app.storage.PerformTransaction(transaction)
	if err != nil {
		log.Printf("processScheduledPosts task running on another instance. error: %s", err)
		return err
	}

	log.Printf("processScheduledPosts started")

	return nil
}

func (app Application) processScheduledPostsWithStartTime(startTime time.Time) error {

}
