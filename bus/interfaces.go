package bus

// Broadcaster is an interface for broadcasting a message over a fanout channel.
// in other words, over rabbit's fanout exchange.
type Broadcaster interface {
	Broadcast(exchange, routingKey string, msg []byte) error
}

// Publisher wraps a message bus publish. In our case, rabbitmq's channel.Publish method.
type Publisher interface {
	Publish(qName string, body []byte) error
}

// BroadcastPublisher use when you need to publish & broadcast.
type BroadcastPublisher interface {
	Broadcaster
	Publisher
}
