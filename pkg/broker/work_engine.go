//
// Copyright (c) 2017 Red Hat, Inc.
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
	topics map[WorkTopic]chan JobMsg
	bufsz  int
}

// NewWorkEngine - creates a new work engine
func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{topics: make(map[WorkTopic]chan JobMsg), bufsz: bufferSize}
}

// StartNewJob - Starts a job in an new goroutine, reporting to a specific topic.
// returns token, or generated token if an empty token is passed in.
func (engine *WorkEngine) StartNewJob(
	token string, work Work, topic WorkTopic,
) (string, error) {
	if valid := IsValidWorkTopic(topic); !valid {
		return "", errors.New("invalid work topic")
	}

	var jobToken string

	if token == "" {
		jobToken = uuid.New()
	} else {
		jobToken = token
	}

	msgBuffer, topicExists := engine.topics[topic]
	if !topicExists {
		msgBuffer = make(chan JobMsg, engine.bufsz)
		engine.topics[topic] = msgBuffer
	}

	go work.Run(jobToken, msgBuffer)
	return jobToken, nil
}

// AttachSubscriber - Attach a subscriber a specific messaging topic.
// Will send the JobMsg to the subscribers through the message buffer.
func (engine *WorkEngine) AttachSubscriber(
	subscriber WorkSubscriber, topic WorkTopic,
) error {
	if valid := IsValidWorkTopic(topic); !valid {
		return errors.New("invalid work topic")
	}

	msgBuffer, topicExists := engine.topics[topic]
	if !topicExists {
		msgBuffer = make(chan JobMsg, engine.bufsz)
		engine.topics[topic] = msgBuffer
	}

	subscriber.Subscribe(msgBuffer)
	return nil
}

// GetActiveTopics - Get list of topics
func (engine *WorkEngine) GetActiveTopics() map[WorkTopic]chan JobMsg {
	return engine.topics
}
