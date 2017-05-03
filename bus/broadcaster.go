package bus

// Broadcaster is an interface for broadcasting a message over a fanout channel.
// in other words, over rabbit's fanout exchange.
type Broadcaster interface {
	Broadcast(exName string, msg []byte) error
}
