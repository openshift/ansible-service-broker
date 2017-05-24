package broker

import "github.com/pborman/uuid"

type Work interface {
	Run(token string, msgBuffer chan<- WorkMsg)
}

type WorkEngine struct {
	msgBuffer chan WorkMsg
}

func NewWorkEngine(bufferSize int) *WorkEngine {
	return &WorkEngine{
		msgBuffer: make(chan WorkMsg, bufferSize),
	}
}

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

func (engine *WorkEngine) AttachSubscriber(subscriber WorkSubscriber) {
	subscriber.Subscribe(engine.msgBuffer)
}
