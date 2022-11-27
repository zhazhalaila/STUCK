package local

import (
	"errors"
	"fmt"
)

var (
	ErrAsleep     = errors.New("LocalTransport is died")
	ErrDisconnect = errors.New("LocalTransport is disconnect")
)

// Using channel to implement a "lock-free" LocalTransport
type LocalTransport struct {
	n             int // total nodes
	id            int // current node id
	isTest        bool
	peers         map[int]*LocalTransport
	recvCh        chan interface{}
	testRecvCh    chan interface{}
	sendOneCh     chan sendWithDest
	sendBroadcast chan broadcastWithErr
	stopCh        chan struct{}
}

type sendWithDest struct {
	dest int
	msg  interface{}
	err  chan error
}

type broadcastWithErr struct {
	msg interface{}
	err chan error
}

// Create new LocalTransport
func NewLocalTransport(n, id int, isTest bool) *LocalTransport {
	lp := &LocalTransport{n: n, id: id, isTest: isTest}
	lp.recvCh = make(chan interface{}, n*n)
	lp.testRecvCh = make(chan interface{}, n*n)
	lp.sendOneCh = make(chan sendWithDest, n*n)
	lp.sendBroadcast = make(chan broadcastWithErr, n*n)
	lp.stopCh = make(chan struct{})

	// start a new goroutine to run
	go lp.run()
	return lp
}

// Receive message from revcCh
func (lp *LocalTransport) run() {
	for {
		select {
		case <-lp.stopCh:
			// break for-loop to avoid goroutine leak
			return

		case inMsg := <-lp.recvCh:
			lp.handle(inMsg)

		case outToOne := <-lp.sendOneCh:
			receiver, ok := lp.peers[outToOne.dest]
			// if peer is removed, no-op
			if !ok {
				outToOne.err <- ErrDisconnect
				continue
			}
			// if peer is deid, no-op
			if !receiver.IsAlive() {
				outToOne.err <- ErrAsleep
				continue
			}
			receiver.RecvFromOtherTransport(outToOne.msg)
			outToOne.err <- nil

		case outToAll := <-lp.sendBroadcast:
			diedPeers := make([]int, 0)
			for peerId, receiver := range lp.peers {
				if !receiver.IsAlive() {
					diedPeers = append(diedPeers, peerId)
				}
				receiver.RecvFromOtherTransport(outToAll.msg)
			}
			var err error
			if len(diedPeers) > 0 {
				err = fmt.Errorf("%v peers are died", diedPeers)
			} else {
				err = nil
			}
			outToAll.err <- err
		}
	}
}

// Handle message
func (lp *LocalTransport) handle(msg interface{}) {
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
	fmt.Printf("[Sender:%d] send msg to [Receiver:%d] error due to : %s", lp.id, peerId, err)
	return err
}

// Broadcast message
func (lp *LocalTransport) Broadcast(msg interface{}) error {
	// create channel with buffer to avoid deadlock
	outToAll := broadcastWithErr{msg: msg, err: make(chan error, 1)}
	lp.sendBroadcast <- outToAll
	err := <-outToAll.err
	fmt.Printf("[Sender:%d] broadcast msg error due to : %s", lp.id, err)
	return err
}

// Assign peers to current LocalTransport
func (lp *LocalTransport) AssignPeers(peers interface{}) {
	lp.peers = peers.(map[int]*LocalTransport)
}

// Stop transport
func (lp *LocalTransport) Stop() {
	close(lp.stopCh)
}

// Message entrance
func (lp *LocalTransport) RecvFromOtherTransport(msg interface{}) {
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
