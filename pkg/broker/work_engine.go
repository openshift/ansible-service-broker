package broker

import "github.com/pborman/uuid"

// Work - is the interface that wraps the basic run method.
type Work interface {
	Run(token string, msgBuffer chan<- WorkMsg)
}

// WorkTopic - Topic jobs can publish messages to, and subscribers can listen to
type WorkTopic string

// WorkEngine - a new engine for doing work.
type WorkEngine struct {
	topics map[WorkTopic]chan WorkMsg
}

// NewWorkEngine - creates a new work engine
func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{topics: make(map[WorkTopic]chan WorkMsg)}
}

//msgBuffer: make(chan WorkMsg, bufferSize),

// StartNewJobOnTopic - Starts a job in an new goroutine, reporting to a specific topic.
// returns token, or generated token if an empty token is passed in.
func (engine *WorkEngine) StartNewJobOnTopic(
	token string, work Work, topic WorkTopic,
) string {
	var jobToken string

	if token == "" {
		jobToken = uuid.New()
	} else {
		jobToken = token
	}

	msgBuffer, topicExists := engine.topics[topic]
	if !topicExists {
		msgBuffer = make(chan WorkMsg)
		engine.topics[topic] = msgBuffer
	}

	go work.Run(jobToken, msgBuffer)
	return jobToken
}

// AttachSubscriberToTopic - Attach a subscriber a specific messaging topic.
// Will send the WorkMsg to the subscribers through the message buffer.
func (engine *WorkEngine) AttachSubscriberToTopic(
	subscriber WorkSubscriber, topic WorkTopic,
) {
	msgBuffer, topicExists := engine.topics[topic]
	if !topicExists {
		msgBuffer = make(chan WorkMsg)
		engine.topics[topic] = msgBuffer
	}

	subscriber.Subscribe(msgBuffer)
}

// GetActiveTopics - Get list of topics
func (engine *WorkEngine) GetActiveTopics() map[WorkTopic]chan WorkMsg {
	return engine.topics
}
