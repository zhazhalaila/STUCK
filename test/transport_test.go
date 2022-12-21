package test

import (
	"testing"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stuck/transport"
	"github.com/stuck/transport/local"
)

var _ transport.Transport = (*local.LocalTransport)(nil)

func initLocalTransports(n int) map[int]transport.Transport {
	// create two maps, one for return and the other for assignment
	peersReturn := make(map[int]transport.Transport, n)
	peers := make(map[int]*local.LocalTransport, n)

	for i := 1; i <= n; i++ {
		peers[i] = local.NewLocalTransport(n, i)
		peersReturn[i] = peers[i]
	}

	// assign peers for each peer
	for _, peer := range peers {
		peer.AssignPeers(peers)
		peer.Connect() // no-op
	}

	return peersReturn
}

// Is it possible to send a message to a single node?
func TestTransportWithSendOne(t *testing.T) {
	// is there a goroutine leak?
	defer leaktest.Check(t)()
	transports := initLocalTransports(2)
	peer1 := transports[1]
	peer2 := transports[2]
	// data flow: peer1 -> peer2
	err := peer1.SendToPeer(2, "Hello")
	assert.Nil(t, err)
	// read once
	peer2Comsumer := peer2.Consume()
	val := <-peer2Comsumer
	assert.Equal(t, "Hello", val.(string))
	peer1.Stop()
	peer2.Stop()
}

// Can I send myself a message?
func TestTransportWithSendMyself(t *testing.T) {
	// is there a goroutine leak?
	defer leaktest.Check(t)()
	transports := initLocalTransports(1)
	peer1 := transports[1]
	// data flow: peer1 -> peer1
	err := peer1.SendToPeer(1, "Hello")
	assert.Nil(t, err)
	// read once
	peer1Consumer := peer1.Consume()
	val := <-peer1Consumer
	assert.Equal(t, "Hello", val.(string))
	peer1.Stop()
}

// Can a node still receive messages after it is disconnected?
func TestTransportAfterDisconnect(t *testing.T) {
	// is there a goroutine leak?
	defer leaktest.Check(t)()
	transports := initLocalTransports(2)
	peer1 := transports[1]
	peer2 := transports[2]
	// data flow: peer1 -> peer2
	err := peer1.SendToPeer(2, "Hello")
	assert.Nil(t, err)
	// read once
	peer2Comsumer := peer2.Consume()
	val := <-peer2Comsumer
	assert.Equal(t, "Hello", val.(string))
	// disconnect between peer1 and peer2
	peer1.Disconnect()
	// case 1: after the connection is closed, the peer1 will report an error when sending again
	err = peer1.SendToPeer(2, 1)
	assert.NotNil(t, err)
	// case 2: after the connection is closed, the peer2 cannot send messages to the peer1
	// data flow: peer2 -> peer1
	err = peer2.SendToPeer(1, 1)
	assert.NotNil(t, err)
	peer1.Stop()
	peer2.Stop()
}

// Can the sender broadcast the message?
func TestTransportBroadcast(t *testing.T) {
	// is there a goroutine leak?
	defer leaktest.Check(t)
	transports := initLocalTransports(5)
	// data flow: sender -> all
	sender := transports[1]
	err := sender.Broadcast("Hello")
	assert.Nil(t, err)
	var consumer <-chan interface{}
	var val interface{}
	// are all nodes able to receive the message?
	for _, peer := range transports {
		consumer = peer.Consume()
		val = <-consumer
		assert.Equal(t, "Hello", val.(string))
		peer.Stop()
	}
}
