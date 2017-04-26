package broker

type WorkSubscriber interface {
	Subscribe(msgBuffer <-chan WorkMsg)
}

type WorkMsg interface {
	Render() string
}
