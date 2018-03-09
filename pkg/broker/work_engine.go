//
// Copyright (c) 2018 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package broker

import (
	"errors"

	"github.com/pborman/uuid"
)

// Work - is the interface that wraps the basic run method.
type Work interface {
	Run(token string, msgBuffer chan<- JobMsg)
}

// WorkEngine - a new engine for doing work.
type WorkEngine struct {
	subscribers map[WorkTopic][]WorkSubscriber
	jobs        map[string]chan JobMsg
}

// NewWorkEngine - creates a new work engine
func NewWorkEngine() *WorkEngine {
	return &WorkEngine{jobs: make(map[string]chan JobMsg), subscribers: map[WorkTopic][]WorkSubscriber{}}
}

// StartNewAsyncJob - Starts a job in an new goroutine, reporting to a specific topic.
// returns token, or generated token if an empty token is passed in.
func (engine *WorkEngine) StartNewAsyncJob(
	token string, work Work, topic WorkTopic,
) (string, error) {
	if valid := IsValidWorkTopic(topic); !valid {
		return "", errors.New("invalid work topic")
	}

	if token == "" {
		token = engine.Token()
	}
	go engine.start(token, work, topic)

	return token, nil
}

func (engine *WorkEngine) start(token string, work Work, topic WorkTopic) {
	engine.jobs[token] = make(chan JobMsg)
	// run the job and close the channel in a new routine

	defer func() {
		log.Debug("closing channel for job ", token, engine.jobs)
		close(engine.jobs[token])
		delete(engine.jobs, token)
	}()

	go func() {
		// listen for new messages and hand them off to the subscribers
		for msg := range engine.jobs[token] {
			for _, sub := range engine.subscribers[topic] {
				//TODO edge case consider the fact that a subscriber may never exit and so we would leak go routines
				go sub.Notify(msg)
			}
		}
	}()
	work.Run(token, engine.jobs[token])
}

// StartNewSyncJob - Starts a job and waits for it to finish, reporting to a specific topic.
func (engine *WorkEngine) StartNewSyncJob(
	token string, work Work, topic WorkTopic,
) error {
	if valid := IsValidWorkTopic(topic); !valid {
		return errors.New("invalid work topic")
	}

	if token == "" {
		token = engine.Token()
	}

	engine.start(token, work, topic)
	return nil
}

// Token generates a new work token
func (engine *WorkEngine) Token() string {
	return uuid.New()
}

// AttachSubscriber - Attach a subscriber a specific messaging topic.
// Will send the JobMsg to the subscribers through the message buffer.
func (engine *WorkEngine) AttachSubscriber(
	subscriber WorkSubscriber, topic WorkTopic,
) error {
	if valid := IsValidWorkTopic(topic); !valid {
		return errors.New("invalid work topic")
	}

	engine.subscribers[topic] = append(engine.subscribers[topic], subscriber)

	return nil
}

// GetActiveJobs - Get list of active jobs
func (engine *WorkEngine) GetActiveJobs() map[string]chan JobMsg {
	return engine.jobs
}

// GetSubscribers - Get list of subscribers to a topic
func (engine *WorkEngine) GetSubscribers(topic WorkTopic) []WorkSubscriber {
	return engine.subscribers[topic]
}
