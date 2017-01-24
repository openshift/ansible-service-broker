package ansibleapp

type Broker struct {
	Name string
}

func NewBroker(name string) (*Broker, error) {
	return &Broker{
		Name: name,
	}, nil
}
