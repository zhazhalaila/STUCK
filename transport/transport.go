package transport

type Transport interface {
	AssignPeers(peers interface{})
	Connect() error
	Disconnect(peerId int)
	SendToPeer(peerId int, msg interface{}) error
	Broadcast(msg interface{}) error
	Stop()
}
