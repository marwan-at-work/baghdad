package bus

// Publisher wraps a message bus publish. In our case, rabbitmq's channel.Publish method.
type Publisher interface {
	Publish(qName string, msg []byte) error
}
