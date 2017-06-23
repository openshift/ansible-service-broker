package broker

import "github.com/pborman/uuid"

// Work - is the interface that wraps the basic run method.
type Work interface {
	Run(token string, msgBuffer chan<- WorkMsg)
}

// WorkEngine - a new engine for doing work.
type WorkEngine struct {
	msgBuffer chan WorkMsg
}

// NewWorkEngine - creates a new work engine
func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{
		msgBuffer: make(chan WorkMsg, bufferSize),
	}
}

// StartNewJob - Starts a job in an new goroutine. returns token, or generated token if an empty token is passed in.
func (engine *WorkEngine) StartNewJob(token string, work Work) string {
	var jobToken string

	if token == "" {
		jobToken = uuid.New()
	} else {
		jobToken = token
	}
	go work.Run(jobToken, engine.msgBuffer)
	return jobToken
}

// AttachSubscriber - Attach a subscriber to the engine. Will send the WorkMsg to the subscribers through the message buffer.
func (engine *WorkEngine) AttachSubscriber(subscriber WorkSubscriber) {
	subscriber.Subscribe(engine.msgBuffer)
}
