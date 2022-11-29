package local

import (
	"errors"
	"fmt"
)

var (
	ErrAsleep     = errors.New("LocalTransport is dead")
	ErrDisconnect = errors.New("LocalTransport is disconnect")
)

// Using channel to implement a "lock-free" LocalTransport
type LocalTransport struct {
	n             int // total nodes
	id            int // current node id
	peers         map[int]*LocalTransport
	recvCh        chan interface{}
	testRecvCh    chan interface{}
	sendOneCh     chan sendWithDest
	sendBroadcast chan broadcastWithErr
	remoteLogout  chan int
	selfLogout    chan struct{}
	stopCh        chan struct{}
}

// message with destination and err
type sendWithDest struct {
	dest int
	msg  interface{}
	err  chan error
}

// message with err
type broadcastWithErr struct {
	msg interface{}
	err chan error
}

// Create new LocalTransport
func NewLocalTransport(n, id int) *LocalTransport {
	lp := &LocalTransport{n: n, id: id}
	lp.recvCh = make(chan interface{}, lp.n*lp.n)
	lp.testRecvCh = make(chan interface{}, lp.n*lp.n)
	lp.sendOneCh = make(chan sendWithDest, lp.n*lp.n)
	lp.sendBroadcast = make(chan broadcastWithErr, lp.n*lp.n)
	lp.remoteLogout = make(chan int, lp.n)
	lp.selfLogout = make(chan struct{}, 1)
	lp.stopCh = make(chan struct{})

	// start a new goroutine to run
	go lp.run()
	return lp
}

// All operations on map are performed in one goroutine
func (lp *LocalTransport) run() {
	for {
		select {
		case <-lp.stopCh:
			// break for-loop to avoid goroutine leak
			return

		case outToOne := <-lp.sendOneCh:
			receiver, ok := lp.peers[outToOne.dest]
			// if peer is removed, no-op
			if !ok {
				outToOne.err <- ErrDisconnect
				break
			}
			// if peer is deid, no-op
			if receiver.IsAlive() {
				outToOne.err <- ErrAsleep
				break
			}
			receiver.recvFromOtherTransport(outToOne.msg)
			outToOne.err <- nil

		case outToAll := <-lp.sendBroadcast:
			// keep track of which nodes are dead
			deadPeers := make([]int, 0)
			for peerId, receiver := range lp.peers {
				if receiver.IsAlive() {
					deadPeers = append(deadPeers, peerId)
				}
				receiver.recvFromOtherTransport(outToAll.msg)
			}
			var err error
			if len(deadPeers) > 0 {
				err = fmt.Errorf("%v peers are dead", deadPeers)
			} else {
				err = nil
			}
			outToAll.err <- err

		case <-lp.selfLogout:
			for _, peer := range lp.peers {
				peer.unregister(lp.id)
			}
			lp.peers = nil

		case peerId := <-lp.remoteLogout:
			delete(lp.peers, peerId)
		}
	}
}

// Connect to all LocalTransport, for local transport, it's no-op
func (lp *LocalTransport) Connect() error {
	return nil
}

// Send message to other LocalTransport
func (lp *LocalTransport) SendToPeer(peerId int, msg interface{}) error {
	// create channel with buffer to avoid deadlock
	outToOne := sendWithDest{dest: peerId, msg: msg, err: make(chan error, 1)}
	lp.sendOneCh <- outToOne
	err := <-outToOne.err
	if err != nil {
		fmt.Printf("[Sender:%d] send msg to [Receiver:%d] error due to : %s.\n", lp.id, peerId, err)
	}
	return err
}

// Broadcast message
func (lp *LocalTransport) Broadcast(msg interface{}) error {
	// create channel with buffer to avoid deadlock
	outToAll := broadcastWithErr{msg: msg, err: make(chan error, 1)}
	lp.sendBroadcast <- outToAll
	err := <-outToAll.err
	if err != nil {
		fmt.Printf("[Sender:%d] broadcast msg error due to : %s.\n", lp.id, err)
	}
	return err
}

// Assign peers to current LocalTransport
func (lp *LocalTransport) AssignPeers(peers interface{}) {
	lp.peers = make(map[int]*LocalTransport)
	for peerId, peer := range peers.(map[int]*LocalTransport) {
		lp.peers[peerId] = peer
	}
}

// Stop transport
func (lp *LocalTransport) Stop() {
	close(lp.stopCh)
}

// Message entrance
func (lp *LocalTransport) recvFromOtherTransport(msg interface{}) {
	select {
	case <-lp.stopCh:
		return
	default:
		lp.recvCh <- msg
	}
}

// Avoide LocalTransport recvCh is fulled after LocalTransport is stop
func (lp *LocalTransport) IsAlive() bool {
	select {
	case <-lp.stopCh:
		return true
	default:
		return false
	}
}

// Return a "read-only" channel
func (lp *LocalTransport) Consume() <-chan interface{} {
	return lp.recvCh
}

// Bi-directional disconnection
func (lp *LocalTransport) Disconnect() {
	lp.selfLogout <- struct{}{}
}

// remove transport from peers
func (lp *LocalTransport) unregister(peerId int) {
	lp.remoteLogout <- peerId
}
