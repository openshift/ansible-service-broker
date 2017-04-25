package broker

import (
	"fmt"
)

type WorkSubscriber interface {
	Subscribe(msgBuffer <-chan WorkMsg)
}

type WorkMsg interface {
	Render() string
}

// Example Subscriber
type StdoutWorkSubscriber struct {
	msgBuffer <-chan WorkMsg
}

func (s *StdoutWorkSubscriber) Subscribe(msgBuffer <-chan WorkMsg) {
	// Always drain the buffer if there's a message waiting.
	// Here we're just forwarding to stdout, but of course, the message
	// destination could be anything (ultimate websockets!)
	// NOTE: DON'T FORGET TO GOROUTINE THIS, OR WILL YOU CHOKE THE MAIN PROCESSOR
	s.msgBuffer = msgBuffer
	go func() {
		for {
			msg := <-msgBuffer
			fmt.Printf(msg.Render())
		}
	}()
}
