package network

import (
	"fmt"
	"sync"
)

type LocalTransport struct {
	addr       NetAddress
	consumerCh chan RPC
	lock       sync.RWMutex
	peers      map[NetAddress]*LocalTransport
}

func NewLocalTransport(addr NetAddress) Transport {
	return &LocalTransport{
		addr:       addr,
		consumerCh: make(chan RPC, 1024),
		peers:      make(map[NetAddress]*LocalTransport),
	}
}

func (t *LocalTransport) Consume() <-chan RPC {
	return t.consumerCh
}

func (t *LocalTransport) Connect(tr Transport) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.peers[tr.Addr()] = tr.(*LocalTransport)

	return nil
}

func (t *LocalTransport) SendMessage(addr NetAddress, payload []byte) error {
	t.lock.RLock()
	defer t.lock.RUnlock()

	peer, ok := t.peers[addr]
	if !ok {
		return fmt.Errorf("%s: no peer with address %s", t.addr, addr)
	}

	peer.consumerCh <- RPC{
		From:    t.addr,
		Payload: payload,
	}

	return nil
}

func (t *LocalTransport) Addr() NetAddress {
	return t.addr
}
