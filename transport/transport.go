package transport

// Transport can be implemented in two ways, one using channel and the other using TCP
type Transport interface {
	AssignPeers(peers interface{})
	Connect() error
	Disconnect()
	SendToPeer(peerId int, msg interface{}) error
	Broadcast(msg interface{}) error
	Stop()
	Consume() <-chan interface{}
}
