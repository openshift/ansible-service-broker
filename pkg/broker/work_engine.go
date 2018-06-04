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
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pborman/uuid"
)

// Work - is the interface that wraps the basic run method.
type Work interface {
	Run(token string, msgBuffer chan<- JobMsg)
}

// WorkEngine - a new engine for doing work.
type WorkEngine struct {
	subscribers   map[WorkTopic][]WorkSubscriber
	jobChannels   map[string]chan JobMsg
	jobBufferSize int
	// the number of seconds given to each subscriber to complete its task
	subscriberTimeout time.Duration
}

// NewWorkEngine - creates a new work engine
func NewWorkEngine(bufferSize int, subscriberTimeout time.Duration) *WorkEngine {
	return &WorkEngine{
		jobChannels:       make(map[string]chan JobMsg),
		subscribers:       map[WorkTopic][]WorkSubscriber{},
		jobBufferSize:     bufferSize,
		subscriberTimeout: subscriberTimeout}
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
	go engine.startJob(token, work, topic)

	return token, nil
}

func waitForNotify(ctx context.Context, sub WorkSubscriber, msg JobMsg, signal chan<- struct{}) {
	sub.Notify(msg)
	// avoid sending on a closed channel if the context is done
	select {
	case <-ctx.Done():
		return
	default:
		signal <- struct{}{}
	}
}

func (engine *WorkEngine) startJob(token string, work Work, topic WorkTopic) {
	// create a channel specifically for use with this job
	jobChannel := make(chan JobMsg, engine.jobBufferSize)
	engine.jobChannels[token] = jobChannel
	// ensure we always clean up
	defer func() {
		log.Debugf("closing channel for job %v", token)
		close(jobChannel)
		delete(engine.jobChannels, token)
	}()

	go func() {
		// listen for a new message for the job keyed to this token and hand off to the subscribers async.
		// Wait for them all to be done before accepting the next message
		for msg := range jobChannel {
			wg := &sync.WaitGroup{}
			// hand off the msg to all subscribers async
			for _, sub := range engine.subscribers[topic] {
				wg.Add(1)
				go func(msg JobMsg, sub WorkSubscriber) {
					// ensure things don't get locked up.
					// Each subscriber has up to the configured amount of time to complete its action
					ctx, cancel := context.WithTimeout(context.Background(), engine.subscriberTimeout*time.Second)
					// used to tell us when the subscribers notify method is completed
					notifySignal := make(chan struct{})
					// If our subscriber times out or returns normally we will always clean up
					defer func() {
						wg.Done()
						close(notifySignal)
						cancel()
					}()
					// notify the subscriber
					go waitForNotify(ctx, sub, msg, notifySignal)
					// act on whichever happens first the subscriber's notify method completing or the timeout
					select {
					case <-notifySignal:
						return
					case <-ctx.Done():
						log.Errorf("Subscriber %s timeout %v ", sub.ID(), ctx.Err())
						return
					}
				}(msg, sub)
			}
			// ensure we wait until all subs are done before taking on the next message
			wg.Wait()
		}
	}()
	work.Run(token, jobChannel)
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

	engine.startJob(token, work, topic)
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

// GetActiveJobChannels - Get list of active jobs
func (engine *WorkEngine) GetActiveJobChannels() map[string]chan JobMsg {
	return engine.jobChannels
}

// GetSubscribers - Get list of subscribers to a topic
func (engine *WorkEngine) GetSubscribers(topic WorkTopic) []WorkSubscriber {
	return engine.subscribers[topic]
}
