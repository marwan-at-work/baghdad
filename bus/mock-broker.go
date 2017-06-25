package bus

// MockBroker implements the publisher/consumer interface but does nothing
type MockBroker struct {
}

// Publish satisfies the publisher interface. does nothing. Does nothing.
func (m MockBroker) Publish(qName string, msg []byte) error {
	return nil
}

// Broadcast satisfies the broadcaster interface. Does nothing.
func (m MockBroker) Broadcast(exName string, msg []byte) error {
	return nil
}
