package broker

// WorkSubscriber - Interface tha wraps the Subscribe method
type WorkSubscriber interface {
	Subscribe(msgBuffer <-chan WorkMsg)
}

// WorkMsg - Interface that wraps the Render method
type WorkMsg interface {
	Render() string
}
