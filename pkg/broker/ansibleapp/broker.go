package ansibleapp

type Broker struct {
	name string
}

func NewBroker(name string) (*Broker, error) {
	return &Broker{
		name: name,
	}, nil
}
