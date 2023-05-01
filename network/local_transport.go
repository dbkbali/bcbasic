package network

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type LocalTransport struct {
	addr       net.Addr
	consumerCh chan RPC
	lock       sync.RWMutex
	peers      map[net.Addr]*LocalTransport
}

func NewLocalTransport(addr net.Addr) *LocalTransport {
	return &LocalTransport{
		addr:       addr,
		consumerCh: make(chan RPC, 1024),
		peers:      make(map[net.Addr]*LocalTransport),
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumerCh
}

func (t *LocalTransport) Connect(tr Transport) error {
	trans := tr.(*LocalTransport)
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[tr.Addr()] = trans

	return nil
}

func (t *LocalTransport) SendMessage(addr net.Addr, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	if t.addr == addr {
		return nil // don't send message to self
	}

	peer, ok := t.peers[addr]
	if !ok {
		return fmt.Errorf("%s: no peer with address %s", t.addr, addr)
	}

	peer.consumerCh <- RPC{
		From:    t.addr,
		Payload: bytes.NewReader(payload),
	}

	return nil
}

func (t *LocalTransport) Broadcast(payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	for _, peer := range t.peers {
		if err := t.SendMessage(peer.Addr(), payload); err != nil {
			return err
		}
	}
	// 	peer.consumerCh <- RPC{
	// 		From:    t.addr,
	// 		Payload: bytes.NewReader(payload),
	// 	}
	// }

	return nil
}

func (t *LocalTransport) Addr() net.Addr {
	return t.addr
}
